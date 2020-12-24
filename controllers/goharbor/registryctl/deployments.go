package registryctl

import (
	"context"
	"fmt"
	"path"
	"strings"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/pkg/config"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	VolumeName                            = "registryctl-config"
	ConfigPath                            = "/etc/registryctl"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
	InternalCertificatesPath              = ConfigPath + "/ssl"
	HealthPath                            = "/api/health"
)

var varFalse = false

const (
	httpsPort = 8443
	httpPort  = 8080
)

const (
	AffinityWeightConfigKey     = "affinity-weight"
	AffinityWeightConfigDefault = 50
	AffinityTopology            = "kubernetes.io/hostname"
)

func (r *Reconciler) GetDeployment(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := r.GetImage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	name := r.NormalizeName(ctx, registryCtl.GetName())
	namespace := registryCtl.GetNamespace()

	reg, err := r.GetRegistry(ctx, registryCtl)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get registry")
	}

	deploy, err := r.Reconciler.GetDeployment(ctx, reg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get registry deployment")
	}

	deploy.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
	deploy.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			r.Label("name"):      name,
			r.Label("namespace"): namespace,
		},
	}
	deploy.Spec.Replicas = registryCtl.Spec.Replicas
	deploy.Spec.Template.ObjectMeta = metav1.ObjectMeta{
		Labels: map[string]string{
			r.Label("name"):      name,
			r.Label("namespace"): namespace,
		},
	}
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: VolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Optional: &varFalse,
			},
		},
	})

	affinityWeight, err := r.ConfigStore.GetItemValueInt(AffinityWeightConfigKey)
	if err != nil {
		if !config.IsNotFound(err, AffinityWeightConfigKey) {
			return nil, errors.Wrap(err, "cannot get affinity weight")
		}

		affinityWeight = AffinityWeightConfigDefault
	}

	deploy.Spec.Template.Spec.Affinity = &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
				Weight: int32(affinityWeight),
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							r.Reconciler.Label("name"): r.Reconciler.NormalizeName(ctx, reg.GetName()),
						},
					},
					TopologyKey: AffinityTopology,
				},
			}},
		},
	}

	registryContainer, err := r.getRegistryContainer(deploy)
	if err != nil {
		return nil, serrors.UnrecoverrableError(errors.Wrap(err, "registry container not found"), serrors.OperatorReason, "cannot find registry container spec")
	}

	registryContainer.Name = controllers.RegistryController.String()
	registryContainer.Image = image
	registryContainer.Args = nil
	registryContainer.Command = nil
	registryContainer.VolumeMounts = append(registryContainer.VolumeMounts, corev1.VolumeMount{
		Name:      VolumeName,
		MountPath: ConfigPath,
	})

	if registryCtl.Spec.Authentication.JobServiceSecretRef != "" {
		registryContainer.Env = append(registryContainer.Env, corev1.EnvVar{
			Name: "JOBSERVICE_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registryCtl.Spec.Authentication.JobServiceSecretRef,
					},
					Key: harbormetav1.SharedSecretKey,
				},
			},
		})
	}

	if registryCtl.Spec.Authentication.CoreSecretRef != "" {
		registryContainer.Env = append(registryContainer.Env, corev1.EnvVar{
			Name: "CORE_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: registryCtl.Spec.Authentication.CoreSecretRef,
					},
					Key: harbormetav1.SharedSecretKey,
				},
			},
		})
	}

	if registryCtl.Spec.TLS.Enabled() {
		r.applyTLSVolumeConfig(ctx, registryCtl, deploy)
		r.applyTLSVolumeMountConfig(ctx, registryContainer)
	}

	registryContainer.Ports = []corev1.ContainerPort{{
		Name:          harbormetav1.RegistryControllerHTTPPortName,
		ContainerPort: httpPort,
	}, {
		Name:          harbormetav1.RegistryControllerHTTPSPortName,
		ContainerPort: httpsPort,
	}}

	port := harbormetav1.RegistryControllerHTTPPortName
	if registryCtl.Spec.TLS.Enabled() {
		port = harbormetav1.RegistryControllerHTTPSPortName
	}

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   HealthPath,
				Port:   intstr.FromString(port),
				Scheme: registryCtl.Spec.TLS.GetScheme(),
			},
		},
	}

	registryContainer.LivenessProbe = probe
	registryContainer.ReadinessProbe = probe

	registryCtl.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}

var errContainerNotFound = fmt.Errorf("container %s not found", controllers.Registry.String())

func (r *Reconciler) getRegistryContainer(deploy *appsv1.Deployment) (*corev1.Container, error) {
	expectedName := controllers.Registry.String()

	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == expectedName {
			return &(deploy.Spec.Template.Spec.Containers[i]), nil
		}
	}

	return nil, errContainerNotFound
}

func (r *Reconciler) applyTLSVolumeMountConfig(ctx context.Context, container *corev1.Container) {
	found := false

	for i, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == registry.InternalCertificatesVolumeName || volumeMount.Name == InternalCertificatesVolumeName {
			container.VolumeMounts[i].Name = InternalCertificatesVolumeName

			if volumeMount.MountPath == registry.InternalCertificatesPath {
				logger.Get(ctx).V(0).Info("tls volumemount found, updating it")

				container.VolumeMounts[i].MountPath = InternalCertificatesPath
				found = true
			}
		}
	}

	if !found {
		logger.Get(ctx).Info("tls volume mount not found, adding new one")

		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificateAuthorityDirectory, corev1.ServiceAccountRootCAKey),
			SubPath:   strings.TrimLeft(corev1.ServiceAccountRootCAKey, "/"),
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificatesPath,
			ReadOnly:  true,
		})
	}
}

func (r *Reconciler) applyTLSVolumeConfig(ctx context.Context, registryCtl *goharborv1alpha2.RegistryController, deploy *appsv1.Deployment) {
	for i, volume := range deploy.Spec.Template.Spec.Volumes {
		if volume.Name == registry.InternalCertificatesVolumeName {
			logger.Get(ctx).V(0).Info("tls volume found, updating it")

			deploy.Spec.Template.Spec.Volumes[i].Name = InternalCertificatesVolumeName
			deploy.Spec.Template.Spec.Volumes[i].VolumeSource = corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registryCtl.Spec.TLS.CertificateRef,
				},
			}

			return
		}
	}

	logger.Get(ctx).Info("tls volume not found, adding new one")

	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: InternalCertificatesVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: registryCtl.Spec.TLS.CertificateRef,
			},
		},
	})
}
