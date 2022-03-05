package redis

import (
	"context"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// New deploy a service, a deployment and a secret to run a redis instance.
// Based on https://hub.docker.com/_/redis
func New(ctx context.Context, ns string) harbormetav1.RedisConnection {
	k8sClient := test.GetClient(ctx)

	redisName := test.NewName("redis")
	redisPasswordName := test.NewName("redis-password")

	gomega.Expect(k8sClient.Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisName,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "http",
				Port: 6379,
			}},
			Selector: map[string]string{
				"pod-selector": redisName,
			},
		},
	})).To(gomega.Succeed())

	gomega.Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisPasswordName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.RedisPasswordKey: "th3Adm!nPa$$w0rd",
		},
		Type: harbormetav1.SecretTypeRedis,
	})).To(gomega.Succeed())

	gomega.Expect(k8sClient.Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisName,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"pod-selector": redisName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"pod-selector": redisName,
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
						Name:  "redis",
						Image: "bitnami/redis:6.2.6",
						Env: []corev1.EnvVar{{
							Name: "REDIS_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: redisPasswordName,
									},
									Key: harbormetav1.RedisPasswordKey,
								},
							},
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 6379,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/data",
							Name:      "data",
						}},
					}},
				},
			},
		},
	})).To(gomega.Succeed())

	return harbormetav1.RedisConnection{
		RedisHostSpec: harbormetav1.RedisHostSpec{
			Host: redisName,
			Port: 6379,
		},
		Database: 0,
		RedisCredentials: harbormetav1.RedisCredentials{
			PasswordRef: redisPasswordName,
		},
	}
}
