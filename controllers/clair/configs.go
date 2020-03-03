package clair

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
	configKey = "config.yaml"
)

var (
	once          sync.Once
	configContent []byte
)

func (r *Reconciler) InitConfigMaps() error {
	file, err := pkger.Open("/assets/templates/clair/config.yaml")
	if err != nil {
		return errors.Wrapf(err, "cannot open Clair configuration template %s", "/assets/templates/clair/config.yaml")
	}
	defer file.Close()

	configContent, err = ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "cannot read Clair configuration template %s", "/assets/templates/clair/config.yaml")
	}

	return nil
}

func (r *Reconciler) GetConfigMap(ctx context.Context, clair *goharborv1alpha2.Clair) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-clair", clair.GetName()),
			Namespace: clair.GetNamespace(),
		},
		BinaryData: map[string][]byte{
			configKey: configContent,
		},
		// https://github.com/goharbor/harbor-scanner-clair#configuration
		// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/clair/clair_env.jinja
		Data: map[string]string{
			"SCANNER_CLAIR_URL":                   fmt.Sprintf("http://%s", clair.GetName()),
			"SCANNER_LOG_LEVEL":                   "debug",
			"SCANNER_STORE_REDIS_POOL_MAX_ACTIVE": "5",
			"SCANNER_STORE_REDIS_POOL_MAX_IDLE":   "5",
			"SCANNER_STORE_REDIS_SCAN_JOB_TTL":    "1h",
			"SCANNER_API_SERVER_ADDR":             fmt.Sprintf(":%d", adapterPort),
		},
	}, nil
}
