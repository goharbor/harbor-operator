package registry

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

const (
	defaultRegistryConfigName = "config.yml"
	registryConfigName        = "config.yaml"
)

var (
	configContent []byte
)

func (r *Reconciler) InitConfigMaps() error {
	{
		file, err := pkger.Open("/assets/templates/registry/config.yaml")
		if err != nil {
			return errors.Wrapf(err, "cannot open Registry configuration template %s", "/assets/templates/registry/config.yaml")
		}
		defer file.Close()

		configContent, err = ioutil.ReadAll(file)
		if err != nil {
			return errors.Wrapf(err, "cannot read Registry configuration template %s", "/assets/templates/registry/config.yaml")
		}
	}

	return nil
}

func (r *Reconciler) GetConfigMap(ctx context.Context, registry *goharborv1alpha2.Registry) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-registry", registry.GetName()),
			Namespace: registry.GetNamespace(),
		},

		BinaryData: map[string][]byte{
			registryConfigName: configContent,
		},
	}, nil
}
