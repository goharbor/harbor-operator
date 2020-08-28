package notaryserver

import (
	"context"
	"path"
	"time"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	HealthPath           = "/_notary_server/health"
	VolumeName           = "config"
	ConfigPath           = "/etc/notary-server"
	HTTPSVolumeName      = "certificates"
	HTTPSCertificatePath = ConfigPath + "/certificates"
	TrustVolumeName      = "trust-certificates"
	TrustCertificatePath = ConfigPath + "/trust-certificates"
	AuthVolumeName       = "auth-certificates"
	AuthCertificatePath  = ConfigPath + "/auth-certificates"

	initialDelayReadiness = 10 * time.Second
)

var varFalse = false

const apiPort = 4443

func (r *Reconciler) GetDeployment(ctx context.Context, notary *goharborv1alpha2.NotaryServer) (*appsv1.Deployment, error) { // nolint:funlen
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

	if notary.Spec.TrustService.Remote != nil {
		volumes = append(volumes, corev1.Volume{
			Name: TrustVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: notary.Spec.TrustService.Remote.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      TrustVolumeName,
			MountPath: TrustCertificatePath,
		})
	}

	if notary.Spec.TLS.Enabled() {
		volumes = append(volumes, corev1.Volume{
			Name: HTTPSVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: notary.Spec.TLS.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      HTTPSVolumeName,
			MountPath: HTTPSCertificatePath,
		})
	}

	if notary.Spec.Authentication != nil {
		volumes = append(volumes, corev1.Volume{
			Name: AuthVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: notary.Spec.Authentication.Token.CertificateRef,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      AuthVolumeName,
			MountPath: AuthCertificatePath,
		})
	}

	initContainers := []corev1.Container{}

	if notary.Spec.Migration.Enabled() {
		migrationContainer, migrationVolumes, err := notary.Spec.Migration.GetMigrationContainer(ctx, &notary.Spec.Storage)
		if err != nil {
			return nil, errors.Wrap(err, "migrationContainer")
		}

		if migrationContainer != nil {
			initContainers = append(initContainers, *migrationContainer)
		}

		volumes = append(volumes, migrationVolumes...)
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(harbormetav1.NotaryServerAPIPortName),
		Scheme: notary.Spec.TLS.GetScheme(),
	}

	deploy := &appsv1.Deployment{
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
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					InitContainers:               initContainers,
					Containers: []corev1.Container{{
						Name:  controllers.NotaryServer.String(),
						Image: image,
						Args:  []string{"notary-server", "-config", path.Join(ConfigPath, ConfigName)},
						Ports: []corev1.ContainerPort{{
							ContainerPort: apiPort,
							Name:          harbormetav1.NotaryServerAPIPortName,
						}},
						VolumeMounts: volumeMounts,

						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: httpGET,
							},
						},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: httpGET,
							},
							// App health endpoint is ready before checking database access and grpc connection
							InitialDelaySeconds: int32(initialDelayReadiness.Seconds()),
						},
					}},
				},
			},
		},
	}

	notary.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}
