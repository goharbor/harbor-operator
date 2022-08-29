package pod

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/rule"
	"github.com/goharbor/harbor-operator/pkg/utils/consts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-image-path,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,sideEffects=NoneOnDryRun,admissionReviewVersions=v1beta1,versions=v1,name=mimg.kb.io

// ImagePathRewriter implements webhook logic to mutate the image path of deploying pods.
type ImagePathRewriter struct {
	Client  client.Client
	Log     logr.Logger
	decoder *admission.Decoder
}

// Handle the admission webhook for mutating the image path of deploying pods.
func (ipr *ImagePathRewriter) Handle(ctx context.Context, req admission.Request) admission.Response { //nolint:funlen,gocognit
	pod := &corev1.Pod{}

	err := ipr.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Get namespace of pod
	podNS, err := ipr.getPodNamespace(ctx, req.Namespace)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("get pod namespace object error: %w", err))
	}

	ipr.Log.Info("receive pod request", "pod", pod.Name, "namespace", podNS.Name)

	// whether to rewrite image path is dependent on rules
	// the rules could be in assigned hsc or default hsc
	// assigned hsc has higher priority
	ipr.Log.Info("try find image rewrite rules that will be applied to this pod")

	var allRules []rule.Rule

	// check if the configmap exist
	if cmName, ok := podNS.Annotations[consts.AnnotationImageRewriteRuleConfigMapRef]; ok { //nolint:nestif
		cm, err := ipr.getConfigMap(ctx, cmName, podNS.Name)
		if err != nil {
			// The resource may have been deleted after reconcile request coming in
			if apierr.IsNotFound(err) {
				return admission.Errored(http.StatusBadRequest, fmt.Errorf("the ConfigMap %s/%s is not found: %w", podNS.Name, cmName, err))
			}

			return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to get ConfigMap %s/%s:,%w", podNS.Name, cmName, err))
		}

		// skip if rewriting is off
		if enable, yes := cm.Data[consts.ConfigMapKeyRewriting]; yes {
			if enable == consts.ConfigMapValueRewritingOff {
				return admission.Allowed("no change")
			} else if enable != consts.ConfigMapValueRewritingOn {
				return admission.Errored(http.StatusBadRequest, errors.Errorf("the rewriting value in configmap %s/%s '%s' is unacceptable", podNS.Name, cmName, enable))
			}
		}

		// check the hsc that the configmap is referring to
		if hscKey, yes := cm.Data[consts.ConfigMapKeyHarborServer]; yes {
			hsc, err := ipr.getHarborServerConfig(ctx, hscKey)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}

			// check selector, error out if assigned HSC doesn't select current namespace
			if hsc.Spec.NamespaceSelector != nil {
				if match := checkNamespaceSelector(podNS.Labels, hsc.Spec.NamespaceSelector.MatchLabels); !match {
					return admission.Errored(http.StatusBadRequest, errors.New("the selector specified in HSC doesn't match the current namespace"))
				}
			}

			// append rules of configMap to rules of hsc.
			rulesFromHSC, err := rule.StringToRules(hsc.Spec.Rules, hsc.Spec.ServerURL)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to parse rule, error: %w", err))
			}

			rulesFromConfigMap, err := rule.StringToRules(strings.Split(strings.TrimSpace(cm.Data[consts.ConfigMapKeyRules]), "\n"), hsc.Spec.ServerURL)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to parse rule, error: %w", err))
			}

			allRules = rule.MergeRules(rulesFromConfigMap, rulesFromHSC)
		} else if _, yes := cm.Data[consts.ConfigMapKeyRules]; yes && strings.TrimSpace(cm.Data[consts.ConfigMapKeyRules]) != "" {
			return admission.Errored(http.StatusBadRequest, errors.New("rule are defined in configMap but there is no hsc associated with it"))
		}
	}

	defaultHSC, err := ipr.lookupDefaultHarborServerConfig(ctx)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("get default hsc object error: %w", err))
	}

	if defaultHSC != nil && defaultHSC.Spec.NamespaceSelector != nil {
		// check selector, if there is match, add the default rule to it. it has lowerest priority
		if match := checkNamespaceSelector(podNS.Labels, defaultHSC.Spec.NamespaceSelector.MatchLabels); match {
			ruleFromDefaultHSC, err := rule.StringToRules(defaultHSC.Spec.Rules, defaultHSC.Spec.ServerURL)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to parse rule, error: %w", err))
			}

			allRules = rule.MergeRules(allRules, ruleFromDefaultHSC)
		} else {
			// it's ok to not match the default hsc
			ipr.Log.Info("default hsc ", defaultHSC.Namespace, "/", defaultHSC.Name, " doesn't match current namespace")
		}
	}

	// there is no rule that will be applied to the current namespace, skip
	if len(allRules) == 0 {
		return admission.Allowed("no change")
	}

	ipr.Log.Info("try rewrite the image path")

	return ipr.rewriteContainers(req, allRules, pod)
}

func checkNamespaceSelector(nsLabels, hscLabelSelector map[string]string) bool {
	if nsLabels != nil && hscLabelSelector != nil {
		for k, v := range nsLabels {
			if _, ok := hscLabelSelector[k]; ok && v == hscLabelSelector[k] {
				return true
			}
		}
	}

	return false
}

func (ipr *ImagePathRewriter) getConfigMap(ctx context.Context, name, namespace string) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	cmNamespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	if err := ipr.Client.Get(ctx, cmNamespacedName, cm); err != nil {
		return nil, err
	}

	return cm, nil
}

func (ipr *ImagePathRewriter) rewriteContainers(req admission.Request, rules []rule.Rule, pod *corev1.Pod) admission.Response {
	for i, c := range pod.Spec.Containers {
		rewrittenImage, err := rewriteContainer(c.Image, rules)
		if err != nil {
			ipr.Log.Error(err, "invalid container image format", "image", c.Image)

			continue
		}

		if rewrittenImage != "" {
			rewrittenContainer := c.DeepCopy()
			rewrittenContainer.Image = rewrittenImage
			pod.Spec.Containers[i] = *rewrittenContainer

			ipr.Log.Info("rewrite container image", "original", c.Image, "rewrite", rewrittenImage)
		}
	}

	for i, c := range pod.Spec.InitContainers {
		rewrittenImage, err := rewriteContainer(c.Image, rules)
		if err != nil {
			ipr.Log.Error(err, "invalid container image format", "image", c.Image)

			continue
		}

		if rewrittenImage != "" {
			rewrittenContainer := c.DeepCopy()
			rewrittenContainer.Image = rewrittenImage
			pod.Spec.InitContainers[i] = *rewrittenContainer

			ipr.Log.Info("rewrite init image", "original", c.Image, "rewrite", rewrittenImage)
		}
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (ipr *ImagePathRewriter) lookupDefaultHarborServerConfig(ctx context.Context) (*goharborv1.HarborServerConfiguration, error) {
	hscList := &goharborv1.HarborServerConfigurationList{}
	if err := ipr.Client.List(ctx, hscList); err != nil {
		return nil, err
	}

	for _, hsc := range hscList.Items {
		if hsc.Spec.Default {
			return &hsc, nil
		}
	}

	return nil, nil
}

// A decoder will be automatically injected.
// InjectDecoder injects the decoder.
func (ipr *ImagePathRewriter) InjectDecoder(d *admission.Decoder) error {
	ipr.decoder = d

	return nil
}

func (ipr *ImagePathRewriter) getPodNamespace(ctx context.Context, ns string) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{}

	nsName := types.NamespacedName{
		Namespace: "",
		Name:      ns,
	}
	if err := ipr.Client.Get(ctx, nsName, namespace); err != nil {
		return nil, fmt.Errorf("get namesapce error: %w", err)
	}

	return namespace, nil
}

func (ipr *ImagePathRewriter) getHarborServerConfig(ctx context.Context, issuer string) (*goharborv1.HarborServerConfiguration, error) {
	hsc := &goharborv1.HarborServerConfiguration{}
	nsName := types.NamespacedName{
		Name: issuer,
	}

	if err := ipr.Client.Get(ctx, nsName, hsc); err != nil {
		return nil, err
	}

	return hsc, nil
}
