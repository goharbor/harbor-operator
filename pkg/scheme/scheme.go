package scheme

import (
	"context"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
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

	// +kubebuilder:scaffold:scheme

	return scheme, nil
}
