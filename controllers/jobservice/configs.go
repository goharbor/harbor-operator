package jobservice

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
	configName = "config.yaml"
)

const (
	logsDirectory = "/var/log/jobs"
)

var (
	once          sync.Once
	configContent []byte
	hookMaxRetry  = 5
)

func (r *Reconciler) InitConfigMaps() error {
	file, err := pkger.Open("/assets/templates/jobservice/config.yaml")
	if err != nil {
		return errors.Wrapf(err, "cannot open JobService configuration template %s", "/assets/templates/jobservice/config.yaml")
	}
	defer file.Close()

	configContent, err = ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "cannot read JobService configuration template %s", "/assets/templates/jobservice/config.yaml")
	}

	return nil
}

func (r *Reconciler) GetConfigMap(ctx context.Context, jobservice *goharborv1alpha2.JobService) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-jobservice", jobservice.GetName()),
			Namespace: jobservice.GetNamespace(),
		},
		BinaryData: map[string][]byte{
			configName: configContent,
		},
	}, nil
}
