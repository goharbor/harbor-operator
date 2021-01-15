/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"context"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetupPostgresql deploy a servicea deployment and a secret to run a postgresql instance.
// Based on https://hub.docker.com/_/postgres
func SetupPostgresql(ctx context.Context, ns string) harbormetav1.PostgresConnectionWithParameters {
	k8sClient := GetClient(ctx)

	pgName := NewName("pg")
	pgPasswordName := NewName("pg-password")

	gomega.Expect(k8sClient.Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgName,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "http",
				Port: 5432,
			}},
		},
	})).To(gomega.Succeed())

	gomega.Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgPasswordName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.PostgresqlPasswordKey: "th3Adm!nPa$$w0rd",
		},
		Type: harbormetav1.SecretTypePostgresql,
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
					Volumes: []corev1.Volume{{
						Name: "data",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
					Containers: []corev1.Container{{
						Name:  "database",
						Image: "postgres",
						Env: []corev1.EnvVar{{
							Name: "POSTGRES_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: pgPasswordName,
									},
									Key: harbormetav1.PostgresqlPasswordKey,
								},
							},
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 5432,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/var/lib/postgresql/data",
							Name:      "data",
						}},
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
			Database: "postgresql",
			Hosts: []harbormetav1.PostgresHostSpec{{
				Host: pgName,
				Port: 5432,
			}},
		},
		Parameters: map[string]string{
			harbormetav1.PostgresSSLModeKey: string(harbormetav1.PostgresSSLModeRequire),
		},
	}
}
