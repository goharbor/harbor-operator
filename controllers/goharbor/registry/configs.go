package registry

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	ConfigName = "config.yml"
)

func (r *Reconciler) GetConfigMap(ctx context.Context, registry *goharborv1alpha2.Registry) (*corev1.ConfigMap, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetConfigMap", opentracing.Tags{})
	defer span.Finish()

	templateConfig, err := r.ConfigStore.GetItemValue(ConfigTemplateKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get template")
	}

	template, err := template.New(ConfigName).
		Funcs(sprig.TxtFuncMap()).
		Funcs(r.Funcs(ctx, registry)).
		Parse(templateConfig)
	if err != nil {
		return nil, errors.Wrap(err, "invalid template")
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	var errTemplate error

	go func() {
		defer writer.Close()

		errTemplate = template.Execute(writer, registry)
	}()

	configContent, err := ioutil.ReadAll(reader)

	if errTemplate != nil {
		return nil, errors.Wrap(errTemplate, "cannot process config template")
	}

	if err != nil {
		return nil, errors.Wrap(err, "cannot read processed config")
	}

	name := r.NormalizeName(ctx, registry.GetName())

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: registry.GetNamespace(),
			Annotations: map[string]string{
				"template.registry.goharbor.io/checksum": fmt.Sprintf("%x", sha256.Sum256([]byte(templateConfig))),
				"regsitry.goharbor.io/uid":               fmt.Sprintf("%v", registry.GetUID()),
				"registry.goharbor.io/generation":        fmt.Sprintf("%v", registry.GetGeneration()),
			},
		},

		BinaryData: map[string][]byte{
			ConfigName: configContent,
		},
	}, nil
}
