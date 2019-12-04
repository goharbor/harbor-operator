package clair

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

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	configKey = "config.yaml"
)

var (
	once   sync.Once
	config []byte
)

func InitConfigMaps() {
	file, err := pkger.Open("/assets/templates/clair/config.yaml")
	if err != nil {
		panic(errors.Wrapf(err, "cannot open Clair configuration template %s", "/assets/templates/clair/config.yaml"))
	}
	defer file.Close()

	config, err = ioutil.ReadAll(file)
	if err != nil {
		panic(errors.Wrapf(err, "cannot read Clair configuration template %s", "/assets/templates/clair/config.yaml"))
	}
}

func (c *Clair) GetConfigMaps(ctx context.Context) []*corev1.ConfigMap {
	once.Do(InitConfigMaps)

	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.ClairName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.ClairName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			BinaryData: map[string][]byte{
				configKey: config,
			},
		},
	}
}

func (c *Clair) GetConfigCheckSum() string {
	h := sha256.New()
	// todo get generation of the secret
	return fmt.Sprintf("%x", h.Sum([]byte(c.harbor.Spec.Components.Clair.DatabaseSecret)))
}
