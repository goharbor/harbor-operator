package scheme

import (
	"context"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/api/v1"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/pkg/errors"
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

	err = minio.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure minio scheme")
	}

	// +kubebuilder:scaffold:scheme

	return scheme, nil
}
