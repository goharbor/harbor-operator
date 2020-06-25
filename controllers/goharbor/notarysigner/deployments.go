package notarysigner

import (
	"context"
	"path"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

const (
	port                 = 7899
	VolumeName           = "config"
	ConfigPath           = "/etc/notary-signer"
	HTTPSVolumeName      = "certificates"
	HTTPSCertificatePath = ConfigPath + "/certificates"
)

var (
	varFalse = false
)

func (r *Reconciler) GetDeployment(ctx context.Context, notary *goharborv1alpha2.NotarySigner) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, notary.GetName())
	namespace := notary.GetNamespace()

	volumes := []corev1.Volume{{
		Name: VolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
			},
		},
	}}

	volumeMounts := []corev1.VolumeMount{{
		Name:      VolumeName,
		MountPath: ConfigPath,
	}}

	if notary.Spec.HTTPS.CertificateRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: HTTPSVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: notary.Spec.HTTPS.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      HTTPSVolumeName,
			MountPath: HTTPSCertificatePath,
		})
	}

	initContainers := []corev1.Container{}

	if !notary.Spec.Migration.Disabled {
		migrationContainer, err := notary.Spec.Migration.GetMigrationContainer(ctx, &notary.Spec.Storage.NotaryStorageSpec)
		if err != nil {
			return nil, errors.Wrap(err, "migrationContainer")
		}

		initContainers = append(initContainers, migrationContainer)
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"):      name,
					r.Label("namespace"): namespace,
				},
			},
			Replicas: notary.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 notary.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					InitContainers:               initContainers,
					Containers: []corev1.Container{
						{
							Name:            "notary-signer",
							Image:           image,
							Args:            []string{"notary-signer", "-config", path.Join(ConfigPath, ConfigName)},
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts:    volumeMounts,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt(port),
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt(port),
									},
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									Prefix: "NOTARY_SIGNER_",
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: notary.Spec.Storage.AliasesRef,
										},
										Optional: &varFalse,
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
