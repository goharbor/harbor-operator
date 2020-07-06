package registryctl

import (
	"context"
	"path"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
)

const (
	apiPort                = 8080 // https://github.com/goharbor/harbor/blob/2fb1cc89d9ef9313842cc68b4b7c36be73681505/src/common/const.go#L134
	HealthEndpoint         = "/api/health"
	VolumeName             = "registryctl-config"
	ConfigPath             = "/etc/registryctl/"
	CertificatesPath       = ConfigPath + "certificates/"
	certificatesVolumeName = "registryctl-certificates"
)

var (
	varFalse = false
)

// TODO Merge with GetDeployment from registry reconciler

func (r *Reconciler) GetDeployment(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, registryCtl.GetName())
	namespace := registryCtl.GetNamespace()

	envs := []corev1.EnvVar{}

	volumes := []corev1.Volume{
		{
			Name: VolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name,
					},
					Optional: &varFalse,
				},
			},
		},
	}

	volumesMount := []corev1.VolumeMount{
		{
			Name:      VolumeName,
			MountPath: ConfigPath,
		},
	}

	var deploys appsv1.DeploymentList

	err = r.Client.List(ctx, &deploys, client.InNamespace(namespace), client.MatchingLabels{})
	if err != nil {
		return nil, errors.Wrap(err, "cannot list deployment")
	}

	var deploy *appsv1.Deployment

	kinds, _, err := r.Scheme.ObjectKinds(&goharborv1alpha2.Registry{})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get registry kinds")
	}

	for _, d := range deploys.Items {
		d := d

		controller := metav1.GetControllerOf(&d)
		if controller != nil && controller.Name == registryCtl.Spec.RegistryRef {
			for _, kind := range kinds {
				if kind.Kind == controller.Kind {
					deploy = &d
					break
				}
			}
		}
	}

	if deploy == nil {
		key := types.NamespacedName{
			Name:      registryCtl.Spec.RegistryRef,
			Namespace: namespace,
		}

		var registry goharborv1alpha2.Registry

		err := r.Client.Get(ctx, key, &registry)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot find referenced registry %s", key)
		}

		ok, err := statuscheck.BasicCheck(ctx, &registry)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot determine status of registry %s", key)
		}

		if !ok {
			return nil, errors.Errorf("registry %s is not ready yet", key)
		}

		return nil, errors.Errorf("cannot find registry deployment")
	}

	volumes = append(volumes, deploy.Spec.Template.Spec.Volumes...)

	for _, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == "registry" {
			envs = append(envs, container.Env...)
			volumesMount = append(volumesMount, container.VolumeMounts...)

			break
		}
	}

	if registryCtl.Spec.HTTPS.CertificateRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: certificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registryCtl.Spec.HTTPS.CertificateRef,
					Optional:   &varFalse,
				},
			},
		})

		volumesMount = append(volumesMount, corev1.VolumeMount{
			Name:      certificatesVolumeName,
			MountPath: CertificatesPath,
		})
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
			Replicas: registryCtl.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 registryCtl.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					Containers: []corev1.Container{
						{
							Name:  "registryctl",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: apiPort,
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   HealthEndpoint,
										Port:   intstr.FromInt(apiPort),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   HealthEndpoint,
										Port:   intstr.FromInt(apiPort),
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							VolumeMounts: volumesMount,
							Command:      []string{"/home/harbor/harbor_registryctl"},
							Args:         []string{"-c", path.Join(ConfigPath, ConfigName)},
							Env:          envs,
						},
					},
				},
			},
		},
	}, nil
}
