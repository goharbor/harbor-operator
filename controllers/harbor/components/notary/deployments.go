package notary

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	containerregistryv1alpha1 "github.com/goharbor/harbor-core-operator/api/v1alpha1"
	"github.com/goharbor/harbor-core-operator/pkg/factories/application"
)

const (
	migrationDatabaseURL = "postgresql://$(username):$(password)@$(host):$(port)/$(database)?sslmode=$(ssl)"
	initImage            = "hairyhenderson/gomplate"
	notaryServerPort     = 4443
	notarySignerPort     = 7899
)

var (
	revisionHistoryLimit     int32 = 0 // nolint:golint
	varFalse                       = false
	notarySignerKeyAlgorithm       = "ecdsa"
)

// https://github.com/goharbor/harbor-helm/blob/master/templates/notary/notary-server.yaml
// https://github.com/goharbor/harbor-helm/blob/master/templates/notary/notary-signer.yaml

func (n *Notary) GetDeployments(ctx context.Context) []*appsv1.Deployment { // nolint:funlen
	operatorName := application.GetName(ctx)
	harborName := n.harbor.GetName()

	return []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(NotaryServerName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":      NotaryServerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      NotaryServerName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: n.harbor.Spec.Components.Notary.Server.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"configuration/checksum": n.GetConfigMapsCheckSum(),
							"secret/checksum":        n.GetSecretsCheckSum(),
							"operator/version":       application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      NotaryServerName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector:                 n.harbor.Spec.Components.Notary.Server.NodeSelector,
						AutomountServiceAccountToken: &varFalse,
						Volumes: []corev1.Volume{
							{
								Name: "config-template",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: n.harbor.NormalizeComponentName(NotaryServerName),
										},
										Items: []corev1.KeyToPath{
											{
												Key:  serverConfigKey,
												Path: serverConfigKey,
											},
										},
									},
								},
							}, {
								Name:         "config",
								VolumeSource: corev1.VolumeSource{},
							}, {
								Name: "notary-certificate",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: n.harbor.NormalizeComponentName(notaryCertificateName),
									},
								},
							}, {
								Name: "token-certificate",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: n.harbor.NormalizeComponentName(containerregistryv1alpha1.CertificateName),
									},
								},
							},
						},
						InitContainers: []corev1.Container{
							{
								Name:  "init-db",
								Image: n.harbor.Spec.Components.Notary.DBMigrator.GetImage(),
								Args: []string{
									"-c",
									"server",
									"-d",
									migrationDatabaseURL,
								},
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: n.harbor.Spec.Components.Notary.Server.DatabaseSecret,
											},
										},
									},
								},
							}, {
								Name:       "configuration",
								Image:      initImage,
								WorkingDir: "/workdir",
								Args:       []string{"--input-dir", "/workdir", "--output-dir", "/processed"},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "config-template",
										MountPath: "/workdir",
										ReadOnly:  true,
									}, {
										Name:      "config",
										MountPath: "/processed",
										ReadOnly:  false,
									},
								},
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: n.harbor.Spec.Components.Notary.Server.DatabaseSecret,
											},
										},
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "core_public_url",
										Value: n.harbor.Spec.PublicURL,
									}, {
										Name:  "notary_server_port",
										Value: strconv.Itoa(notaryServerPort),
									}, {
										Name:  "notary_signer_url",
										Value: n.harbor.NormalizeComponentName(NotarySignerName),
									}, {
										Name:  "notary_signer_key_algorithm",
										Value: notarySignerKeyAlgorithm,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "notary-server",
								Image: n.harbor.Spec.Components.Notary.Server.GetImage(),
								Args: []string{
									"notary-server",
									"-config",
									"/etc/notary/server-config.json",
								},
								ImagePullPolicy: corev1.PullAlways,
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "token-certificate",
										MountPath: "/etc/ssl/notary/auth-token.crt",
										SubPath:   "tls.crt",
									}, {
										Name:      "notary-certificate",
										MountPath: "/etc/ssl/notary/ca.crt",
										SubPath:   "ca.crt",
									}, {
										Name:      "config",
										MountPath: "/etc/notary/server-config.json",
										SubPath:   serverConfigKey,
									},
								},
							},
						},
						Priority: n.Option.GetPriority(),
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               n.harbor.Spec.Paused,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      n.harbor.NormalizeComponentName(NotarySignerName),
				Namespace: n.harbor.Namespace,
				Labels: map[string]string{
					"app":      NotarySignerName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":      NotarySignerName,
						"harbor":   harborName,
						"operator": operatorName,
					},
				},
				Replicas: n.harbor.Spec.Components.Notary.Signer.Replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"configuration/checksum": n.GetConfigMapsCheckSum(),
							"secret/checksum":        n.GetSecretsCheckSum(),
							"operator/version":       application.GetVersion(ctx),
						},
						Labels: map[string]string{
							"app":      NotarySignerName,
							"harbor":   harborName,
							"operator": operatorName,
						},
					},
					Spec: corev1.PodSpec{
						NodeSelector:                 n.harbor.Spec.Components.Notary.Signer.NodeSelector,
						AutomountServiceAccountToken: &varFalse,
						Volumes: []corev1.Volume{
							{
								Name: "config-template",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: n.harbor.NormalizeComponentName(NotarySignerName),
										},
										Items: []corev1.KeyToPath{
											{
												Key:  signerConfigKey,
												Path: signerConfigKey,
											},
										},
									},
								},
							}, {
								Name:         "config",
								VolumeSource: corev1.VolumeSource{},
							}, {
								Name: "notary-certificate",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: n.harbor.NormalizeComponentName(notaryCertificateName),
									},
								},
							},
						},
						InitContainers: []corev1.Container{
							{
								Name:  "init-db",
								Image: n.harbor.Spec.Components.Notary.DBMigrator.GetImage(),
								Args: []string{
									"-c",
									"signer",
									"-d",
									migrationDatabaseURL,
								},
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: n.harbor.Spec.Components.Notary.Signer.DatabaseSecret,
											},
										},
									},
								},
							}, {
								Name:       "configuration",
								Image:      initImage,
								WorkingDir: "/workdir",
								Args:       []string{"--input-dir", "/workdir", "--output-dir", "/processed"},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "config-template",
										MountPath: "/workdir",
										ReadOnly:  true,
									}, {
										Name:      "config",
										MountPath: "/processed",
										ReadOnly:  false,
									},
								},
								EnvFrom: []corev1.EnvFromSource{
									{
										SecretRef: &corev1.SecretEnvSource{
											Optional: &varFalse,
											LocalObjectReference: corev1.LocalObjectReference{
												Name: n.harbor.Spec.Components.Notary.Signer.DatabaseSecret,
											},
										},
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "notary_signer_port",
										Value: strconv.Itoa(notarySignerPort),
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "notary-signer",
								Image: n.harbor.Spec.Components.Notary.Signer.GetImage(),
								Args: []string{
									"notary-signer",
									"-config",
									"/etc/notary/signer-config.json",
								},
								ImagePullPolicy: corev1.PullAlways,
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "notary-certificate",
										MountPath: "/etc/ssl/notary/tls.crt",
										SubPath:   "tls.crt",
									}, {
										Name:      "notary-certificate",
										MountPath: "/etc/ssl/notary/tls.key",
										SubPath:   "tls.key",
									}, {
										Name:      "config",
										MountPath: "/etc/notary/signer-config.json",
										SubPath:   signerConfigKey,
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "NOTARY_SIGNER_DEFAULTALIAS",
										Value: "defaultalias",
									},
								},
							},
						},
						Priority: n.Option.GetPriority(),
					},
				},
				RevisionHistoryLimit: &revisionHistoryLimit,
				Paused:               n.harbor.Spec.Paused,
			},
		},
	}
}
