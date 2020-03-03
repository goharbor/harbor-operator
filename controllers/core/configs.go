package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
)

const (
	configName = "app.conf"
)

var (
	once          sync.Once
	configContent []byte
)

func (r *Reconciler) InitConfigMaps() error {
	file, err := pkger.Open("/assets/templates/core/app.conf")
	if err != nil {
		return errors.Wrapf(err, "cannot open Core configuration template %s", "/assets/templates/core/app.conf")
	}
	defer file.Close()

	configContent, err = ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "cannot read Core configuration template %s", "/assets/templates/core/app.conf")
	}

	return nil
}

func (r *Reconciler) GetConfigMap(ctx context.Context, core *goharborv1alpha2.Core) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-core", core.GetName()),
			Namespace: core.GetNamespace(),
		},

		BinaryData: map[string][]byte{
			configName: configContent,
		},
	}, nil
}
