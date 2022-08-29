package storage

import (
	"context"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const initBucketScript = `
for i in {1..10}; do
  if [[ $(mc alias set minio $MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY 2>&1 1>/dev/null) ]]; then
    if [ $i -eq 10 ]; then
      echo "Initialize bucket $MINIO_BUCKET failed for set alias failed with many retries"
      exit 1
    fi
    echo "Set minio alias for $MINIO_ENDPOINT failed, try again after $(($i+$i)) seconds"
    sleep $(($i+$i))
  else
    echo "Set minio alias for $MINIO_ENDPOINT success"
    break
  fi
done

for i in {1..10}; do
  if [[ $(mc ls minio/$MINIO_BUCKET 2>&1 1>/dev/null) ]]; then
    mc mb minio/$MINIO_BUCKET >/dev/null 2>&1
    if [ $? -eq 0 ]; then
      echo "Initialize bucket $MINIO_BUCKET success"
      exit 0
    else
      echo "Create bucket $MINIO_BUCKET failed, try again after $(($i+$i)) seconds"
      sleep $(($i+$i))
    fi
  else
    echo "Bucket $MINIO_BUCKET found"
    exit 0
  fi
done

echo "Initialize bucket $MINIO_BUCKET failed"
exit 1
`

func (m *MinIOController) generateMinIOInitJob(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*batchv1.Job, error) { //nolint:funlen
	image, err := m.getMinIOClientImage(ctx, harborcluster)
	if err != nil {
		return nil, err
	}

	minioEndpoint := fmt.Sprintf("http://%s.%s.svc:%d", m.getTenantsServiceName(harborcluster), harborcluster.Namespace, m.getServicePort())

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.getMinIONamespacedName(harborcluster).Name,
			Namespace: m.getMinIONamespacedName(harborcluster).Namespace,
			Labels: map[string]string{
				"job-type": "minio-init",
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(harborcluster, goharborv1.HarborClusterGVK),
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"job-type": "minio-init",
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: m.getMinIOClientImagePullSecrets(ctx, harborcluster),
					RestartPolicy:    corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{{
						Name:            "init-minio",
						Image:           image,
						ImagePullPolicy: m.getMinIOClientImagePullPolicy(ctx, harborcluster),
						Command:         []string{"bash", "-c", initBucketScript},
						Env: []corev1.EnvVar{{
							Name:  "MINIO_BUCKET",
							Value: DefaultBucket,
						}, {
							Name:  "MINIO_ENDPOINT",
							Value: minioEndpoint,
						}, {
							Name: "MINIO_ACCESS_KEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: m.getMinIOSecretNamespacedName(harborcluster).Name,
									},
									Key: "accesskey",
								},
							},
						}, {
							Name: "MINIO_SECRET_KEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: m.getMinIOSecretNamespacedName(harborcluster).Name,
									},
									Key: "secretkey",
								},
							},
						}},
					}},
				},
			},
		},
	}

	job.Spec.Template.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(job, batchv1.SchemeGroupVersion.WithKind("Job")),
	}

	dependencies := checksum.New(m.Scheme)
	dependencies.Add(ctx, harborcluster, true)
	dependencies.AddAnnotations(job)

	return job, nil
}

func (m *MinIOController) createMinIOInitJob(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	job, err := m.generateMinIOInitJob(ctx, harborcluster)
	if err != nil {
		return minioNotReadyStatus(CreateDefaultBucketError, err.Error()), err
	}

	if err := m.KubeClient.Create(ctx, job); err != nil {
		return minioNotReadyStatus(CreateInitJobError, err.Error()), err
	}

	m.Log.Info("MinIO init job is created")

	return minioUnknownStatus(), nil
}

func (m *MinIOController) applyMinIOInitJob(ctx context.Context, harborcluster *goharborv1.HarborCluster) (*lcm.CRStatus, error) {
	initJob := &batchv1.Job{}

	err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), initJob)
	if err != nil {
		if errors.IsNotFound(err) {
			m.Log.Info("Start creating minio init job.")

			return m.createMinIOInitJob(ctx, harborcluster)
		}

		return minioNotReadyStatus(GetInitJobError, err.Error()), err
	}

	// Recreate if necessary
	if !common.Equals(ctx, m.Scheme, harborcluster, initJob) {
		// can't change the template after the job has been created, so delete it first and the recreate it
		// https://github.com/kubernetes/kubernetes/issues/89657
		if err := m.KubeClient.Delete(ctx, initJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
			return minioNotReadyStatus(DeleteInitJobError, err.Error()), err
		}

		return m.createMinIOInitJob(ctx, harborcluster)
	}

	return minioUnknownStatus(), nil
}

func (m *MinIOController) checkMinIOInitJobReady(ctx context.Context, harborcluster *goharborv1.HarborCluster) (bool, error) {
	job := &batchv1.Job{}
	if err := m.KubeClient.Get(ctx, m.getMinIONamespacedName(harborcluster), job); err != nil {
		return false, err
	}

	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete {
			return condition.Status == corev1.ConditionTrue, nil
		}
	}

	return false, nil
}
