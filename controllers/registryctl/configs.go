package registryctl

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

const (
	defaultRegistryConfigName = "config.yml"
	registryConfigName        = "config.yaml"
	registryCtlConfigName     = "ctl-config.yaml"
)

var (
	once              sync.Once
	registryConfig    []byte
	registryCtlConfig []byte
)

func (r *Reconciler) InitConfigMaps() error {
	{
		file, err := pkger.Open("/assets/templates/registry/config.yaml")
		if err != nil {
			return errors.Wrapf(err, "cannot open Registry configuration template %s", "/assets/templates/registry/config.yaml")
		}
		defer file.Close()

		registryConfig, err = ioutil.ReadAll(file)
		if err != nil {
			return errors.Wrapf(err, "cannot read Registry configuration template %s", "/assets/templates/registry/config.yaml")
		}
	}
	{
		file, err := pkger.Open("/assets/templates/registry/ctl-config.yaml")
		if err != nil {
			return errors.Wrapf(err, "cannot open Registry configuration template %s", "/assets/templates/registry/ctl-config.yaml")
		}
		defer file.Close()

		registryCtlConfig, err = ioutil.ReadAll(file)
		if err != nil {
			return errors.Wrapf(err, "cannot read Registry configuration template %s", "/assets/templates/registry/ctl-config.yaml")
		}
	}

	return nil
}

func (r *Reconciler) GetConfigMap(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-registryctl", registryCtl.GetName()),
			Namespace: registryCtl.GetNamespace(),
		},

		BinaryData: map[string][]byte{
			registryConfigName:    registryConfig,
			registryCtlConfigName: registryCtlConfig,
		},
	}, nil
}
