package jobservice

import (
	"context"
	"crypto/sha256"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	logsDirectory = "/var/log/jobs"

	// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/jobservice/config.yml.jinja
	config = `
protocol: "http"
port: 8080

worker_pool:
  backend: "redis"

  redis_pool:
    namespace: jobservice

job_loggers:
  - name: STD_OUTPUT
    level: INFO # INFO/DEBUG/WARNING/ERROR/FATAL

    # JobService read files to expose logs
  - name: FILE
    level: INFO
    settings: # Customized settings of logger
      base_dir: "` + logsDirectory + `"
    sweeper:
      duration: 7 #days
      settings: # Customized settings of sweeper
        work_dir: "` + logsDirectory + `"

loggers:
  - name: STD_OUTPUT
    level: INFO`
)

func (j *JobService) GetConfigMaps(ctx context.Context) []*corev1.ConfigMap {
	operatorName := application.GetName(ctx)
	harborName := j.harbor.Name

	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      j.harbor.NormalizeComponentName(containerregistryv1alpha1.JobServiceName),
				Namespace: j.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.JobServiceName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Data: map[string]string{
				"config.yml": config,
			},
		},
	}
}

func (j *JobService) GetConfigCheckSum() string {
	h := sha256.New()
	return fmt.Sprintf("%x", h.Sum([]byte(j.harbor.Spec.PublicURL)))
}
