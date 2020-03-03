package notaryserver

import (
	"context"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/pkg/errors"
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

func (r *Reconciler) GetDeployment(ctx context.Context, notary *goharborv1alpha2.NotaryServer) (*appsv1.Deployment, error) { // nolint:funlen
	image, err := notary.Spec.GetImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get image")
	}

	dbMigratorImage, err := notary.Spec.GetDBMigratorImage()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get dbMigrator image")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-notaryserver", notary.GetName()),
			Namespace: notary.GetNamespace(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"notaryserver-name":      notary.GetName(),
					"notaryserver-namespace": notary.GetNamespace(),
				},
			},
			Replicas: notary.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"notaryserver-name":      notary.GetName(),
						"notaryserver-namespace": notary.GetNamespace(),
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 notary.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes: []corev1.Volume{
						{
							Name: "config-template",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: notary.Name,
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
									SecretName: notary.Spec.CertificateSecret,
								},
							},
						}, {
							Name: "token-certificate",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: notary.Spec.TokenSecret,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "init-db",
							Image: dbMigratorImage,
							Args:  []string{"-c", "server", "-d", migrationDatabaseURL},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										Optional: &varFalse,
										LocalObjectReference: corev1.LocalObjectReference{
											Name: notary.Spec.DatabaseSecret,
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
											Name: notary.Spec.DatabaseSecret,
										},
									},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "core_public_url",
									Value: notary.Spec.PublicURL,
								}, {
									Name:  "notary_server_port",
									Value: strconv.Itoa(notaryServerPort),
								}, {
									Name:  "notary_signer_url",
									Value: notary.Spec.SignerURL,
								}, {
									Name:  "notary_signer_key_algorithm",
									Value: notarySignerKeyAlgorithm,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "notary-server",
							Image:           image,
							Args:            []string{"notary-server", "-config", "/etc/notary/server-config.json"},
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
					Priority: notary.Spec.Priority,
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit,
		},
	}, nil
}
