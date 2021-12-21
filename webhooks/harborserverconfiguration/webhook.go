package harborserverconfiguration

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/umisama/go-regexpcache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-hsc,mutating=false,failurePolicy=fail,groups="goharbor.io",resources=harborserverconfigurations,verbs=create;update,sideEffects=None,admissionReviewVersions=v1beta1,versions=v1beta1,name=hsc.goharbor.io

type Validator struct {
	Client  client.Client
	Log     logr.Logger
	decoder *admission.Decoder
}

var (
	_ admission.Handler         = (*Validator)(nil)
	_ admission.DecoderInjector = (*Validator)(nil)
)

func (h *Validator) Handle(ctx context.Context, req admission.Request) admission.Response {
	hsc := &goharborv1.HarborServerConfiguration{}

	err := h.decoder.Decode(req, hsc)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	for _, rule := range hsc.Spec.Rules {
		registryRegex := rule[:strings.LastIndex(rule, ",")+1]

		if _, err := regexpcache.Compile(registryRegex); err != nil {
			return admission.ValidationResponse(false, fmt.Sprintf("%s can not be validated, %q is not a valid regular expression: %s", hsc.Name, registryRegex, err.Error()))
		}
	}
	// Check for duplicate default configurations
	if hsc.Spec.Default {
		hscList := &goharborv1.HarborServerConfigurationList{}
		if err := h.Client.List(ctx, hscList); err != nil {
			return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to list harbor server configurations: %w", err))
		}

		for _, harborConf := range hscList.Items {
			if harborConf.Name != hsc.Name && harborConf.Spec.Default {
				return admission.ValidationResponse(false, fmt.Sprintf("%q can not be set as default, %q is the default harbor server configuration", hsc.Name, harborConf.Name))
			}
		}
	}

	return admission.Allowed("")
}

func (h *Validator) InjectDecoder(decoder *admission.Decoder) error {
	h.decoder = decoder

	return nil
}

func (h *Validator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&goharborv1.HarborServerConfiguration{}).Complete()
}
