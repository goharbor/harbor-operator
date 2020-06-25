package jobservice

import (
	"context"
	"io"
	"io/ioutil"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Masterminds/sprig"
	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

const (
	ConfigName = "config.yaml"
)

const (
	logsDirectory = "/var/log/jobs"
)

func (r *Reconciler) GetConfigMap(ctx context.Context, jobservice *goharborv1alpha2.JobService) (*corev1.ConfigMap, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetConfigMap", opentracing.Tags{})
	defer span.Finish()

	templateConfig, err := r.ConfigStore.GetItemValue(ConfigTemplateKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get template")
	}

	template, err := template.New(ConfigName).
		Funcs(sprig.TxtFuncMap()).
		Funcs(r.Funcs(ctx, jobservice)).
		Parse(templateConfig)
	if err != nil {
		return nil, errors.Wrap(err, "invalid template")
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	var errTemplate error

	go func() {
		defer writer.Close()

		errTemplate = template.Execute(writer, jobservice)
	}()

	configContent, err := ioutil.ReadAll(reader)

	if errTemplate != nil {
		return nil, errors.Wrap(errTemplate, "cannot process config template")
	}

	if err != nil {
		return nil, errors.Wrap(err, "cannot read processed config")
	}

	name := r.NormalizeName(ctx, jobservice.GetName())
	namespace := jobservice.GetNamespace()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},

		BinaryData: map[string][]byte{
			ConfigName: configContent,
		},
	}, nil
}
