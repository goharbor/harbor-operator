package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/harbor-cluster-trait-mutate-v1-pod,mutating=true,failurePolicy=fail,groups=core,resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,admissionReviewVersions={"v1"},sideEffects=None

// PodAnnotator annotates Pods.
type PodAnnotator struct {
	Client  client.Client
	decoder *admission.Decoder
}

var TraitMap sync.Map

// Log used this webhook.
var clog = logf.Log.WithName("harborclustertrait-webhook")

// Handle podAnnotator adds an annotation to every incoming pods.
func (a *PodAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	clog.Info("catch harbor cluster pod", "name", pod.Name, "namespace", pod.Namespace)

	if pod.Labels == nil {
		return admission.Allowed("not found label,skip to inject affinities")
	}

	for k, v := range pod.Labels {
		key := strings.Join([]string{k, v}, "=")

		policy, found := TraitMap.Load(key)
		if !found {
			continue
		}

		affinity, ok := policy.(*corev1.Affinity)
		if !ok {
			continue
		}

		pod.Spec.Affinity = affinity
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	pod.Annotations["harbor-cluster-trait-affinities"] = "injected"

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// podAnnotator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *PodAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d

	return nil
}
