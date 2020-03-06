package scheme

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	containerregistryv1alpha1 "github.com/goharbor/harbor-core-operator/api/v1alpha1"
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

	err = containerregistryv1alpha1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure OVH scheme")
	}

	// +kubebuilder:scaffold:scheme

	return scheme, nil
}
