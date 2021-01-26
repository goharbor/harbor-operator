package scheme

import (
	"context"
	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/api/v1"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	certv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
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

	err = certv1alpha2.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure certificate-manager scheme certv1alpha2")
	}

	err = certv1beta1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure certificate-manager scheme certv1beta1")
	}

	err = certv1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure certificate-manager scheme certv1")
	}

	err = goharborv1alpha2.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure Harbor scheme")
	}

	err = redisfailoverv1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure minio scheme")
	}

	err = postgresv1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure minio scheme")
	}

	err = minio.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure minio scheme")
	}

	// +kubebuilder:scaffold:scheme

	return scheme, nil
}
