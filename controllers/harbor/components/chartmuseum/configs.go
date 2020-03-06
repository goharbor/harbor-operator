package chartmuseum

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha1 "github.com/goharbor/harbor-operator/api/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
)

const (
	configName = "config.yaml"
)

var (
	once   sync.Once
	config []byte
)

func InitConfigMaps() {
	file, err := pkger.Open("/assets/templates/chartmuseum/config.yaml")
	if err != nil {
		panic(errors.Wrapf(err, "cannot open ChartMuseum configuration template %s", "/assets/templates/chartmuseum/config.yaml"))
	}
	defer file.Close()

	config, err = ioutil.ReadAll(file)
	if err != nil {
		panic(errors.Wrapf(err, "cannot read ChartMuseum configuration template %s", "/assets/templates/chartmuseum/config.yaml"))
	}
}

// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/chartserver/env.jinja

func (c *ChartMuseum) GetConfigMaps(ctx context.Context) []*corev1.ConfigMap {
	once.Do(InitConfigMaps)

	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(goharborv1alpha1.ChartMuseumName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      goharborv1alpha1.ChartMuseumName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			BinaryData: map[string][]byte{
				configName: config,
			},
			Data: map[string]string{
				"PORT":      fmt.Sprintf("%d", port),
				"CHART_URL": fmt.Sprintf("%s/chartrepo", c.harbor.Spec.PublicURL),
			},
		},
	}
}

func (c *ChartMuseum) GetConfigMapsCheckSum() string {
	value := fmt.Sprintf("%s\n%d\n%x", c.harbor.Spec.PublicURL, port, config)
	sum := sha256.New().Sum([]byte(value))

	// todo get generation of the secret
	return fmt.Sprintf("%x", sum)
}
