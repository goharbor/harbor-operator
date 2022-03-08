package postgresql

import (
	"context"
	"fmt"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const postgresPort = 5432

// New deploy a servicea deployment and a secret to run a postgresql instance.
// Based on https://hub.docker.com/_/postgres
func New(ctx context.Context, ns string, databases ...string) harbormetav1.PostgresConnectionWithParameters {
	k8sClient := test.GetClient(ctx)

	pgName := test.NewName("pg")
	pgPasswordName := test.NewName("pg-password")
	pgConfigMapName := test.NewName("init-db")

	gomega.Expect(k8sClient.Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgName,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "http",
				Port: postgresPort,
			}},
			Selector: map[string]string{
				"pod-selector": pgName,
			},
		},
	})).To(gomega.Succeed())

	gomega.Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgPasswordName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.PostgresqlPasswordKey: "th3Adm1nPa55w0rd",
		},
		Type: harbormetav1.SecretTypePostgresql,
	})).To(gomega.Succeed())

	sql := ""
	for _, database := range databases {
		sql += fmt.Sprintf("CREATE DATABASE %s WITH OWNER postgres;", database)
	}

	gomega.Expect(k8sClient.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgConfigMapName,
			Namespace: ns,
		},
		Data: map[string]string{
			"init-db.sql": sql,
		},
	})).To(gomega.Succeed())

	gomega.Expect(k8sClient.Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgName,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"pod-selector": pgName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"pod-selector": pgName,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "custom-init-scripts",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: pgConfigMapName,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{{
						Name:  "database",
						Image: "bitnami/postgresql:13.6.0",
						Env: []corev1.EnvVar{
							{
								Name: "POSTGRESQL_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: pgPasswordName,
										},
										Key: harbormetav1.PostgresqlPasswordKey,
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 5432,
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/var/lib/postgresql/data",
								Name:      "data",
							},
							{
								MountPath: "/docker-entrypoint-initdb.d",
								Name:      "custom-init-scripts",
							},
						},
					}},
				},
			},
		},
	})).To(gomega.Succeed())

	return harbormetav1.PostgresConnectionWithParameters{
		PostgresConnection: harbormetav1.PostgresConnection{
			PostgresCredentials: harbormetav1.PostgresCredentials{
				PasswordRef: pgPasswordName,
				Username:    "postgres",
			},
			Database: "postgres",
			Hosts: []harbormetav1.PostgresHostSpec{{
				Host: pgName,
				Port: 5432,
			}},
		},
		Parameters: map[string]string{
			harbormetav1.PostgresSSLModeKey: string(harbormetav1.PostgresSSLModeDisable),
		},
	}
}
