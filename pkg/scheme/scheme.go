package scheme

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func New(ctx context.Context) (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure native scheme")
	}

	err = certv1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure certificate-manager scheme")
	}

	err = goharborv1alpha2.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure Harbor scheme")
	}

	if err := admissionv1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to configure admissionv1 scheme")
	}

	if err := admissionv1beta1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to configure admissionv1beta1 scheme")
	}

	// +kubebuilder:scaffold:scheme

	return scheme, nil
}
