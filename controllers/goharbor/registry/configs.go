package registry

import (
	"context"
	"crypto/sha256"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	conftemplate "github.com/goharbor/harbor-operator/pkg/config/template"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/goharbor/harbor-operator/pkg/utils/strings"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConfigName      = "config.yml"
	RegistryCtlName = "registryctl"
)

func (r *Reconciler) GetConfigMap(ctx context.Context, registry *goharborv1.Registry) (*corev1.ConfigMap, error) {
	templateConfig, err := r.ConfigStore.GetItemValue(conftemplate.ConfigTemplateKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get template")
	}

	content, err := r.GetTemplatedConfig(ctx, templateConfig, registry)
	if err != nil {
		return nil, err
	}

	name := r.NormalizeName(ctx, registry.GetName())
	namespace := registry.GetNamespace()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				checksum.GetStaticID("template"): fmt.Sprintf("%x", sha256.Sum256([]byte(templateConfig))),
			},
		},

		BinaryData: map[string][]byte{
			ConfigName: content,
		},
	}, nil
}

func (r *Reconciler) GetCtlConfigMap(ctx context.Context, registryCtl *goharborv1.RegistryController) (*corev1.ConfigMap, error) {
	templateConfig, err := r.ConfigStore.GetItemValue(CtlConfigTemplateKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get template")
	}

	content, err := r.GetTemplatedConfig(ctx, templateConfig, registryCtl)
	if err != nil {
		return nil, err
	}

	name := strings.NormalizeName(registryCtl.GetName(), RegistryCtlName)
	namespace := registryCtl.GetNamespace()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				checksum.GetStaticID("template"): fmt.Sprintf("%x", sha256.Sum256([]byte(templateConfig))),
			},
		},
		BinaryData: map[string][]byte{
			ConfigName: content,
		},
	}, nil
}
