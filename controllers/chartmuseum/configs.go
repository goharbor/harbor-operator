package chartmuseum

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
	configName = "config.yaml"
)

var (
	configContent []byte
)

func (r *Reconciler) InitConfigMaps() error {
	file, err := pkger.Open("/assets/templates/chartmuseum/config.yaml")
	if err != nil {
		return errors.Wrapf(err, "cannot open ChartMuseum configuration template %s", "/assets/templates/chartmuseum/config.yaml")
	}
	defer file.Close()

	configContent, err = ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "cannot read ChartMuseum configuration template %s", "/assets/templates/chartmuseum/config.yaml")
	}

	return nil
}

// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/chartserver/env.jinja

func (r *Reconciler) GetConfigMap(ctx context.Context, chartMuseum *goharborv1alpha2.ChartMuseum) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-chartmuseum", chartMuseum.GetName()),
			Namespace: chartMuseum.GetNamespace(),
		},
		BinaryData: map[string][]byte{
			configName: configContent,
		},
	}, nil
}
