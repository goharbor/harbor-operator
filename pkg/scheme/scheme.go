package scheme

import (
	"context"

	goharborv1alpha3 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	goharborv1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	minio "github.com/goharbor/harbor-operator/pkg/cluster/controllers/storage/minio/apis/minio.min.io/v2"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/pkg/errors"
	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgresv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
		return nil, errors.Wrap(err, "unable to configure certificate-manager scheme certv1")
	}

	err = goharborv1alpha3.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure Harbor v1alpha3 scheme")
	}

	err = goharborv1beta1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure Harbor v1beta1 scheme")
	}

	err = redisfailoverv1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure redis failover scheme")
	}

	err = postgresv1.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure postgres scheme")
	}

	err = minio.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure minio scheme")
	}

	err = apiextensions.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure apiextensions scheme")
	}

	// +kubebuilder:scaffold:scheme

	return scheme, nil
}
