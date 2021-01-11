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

package goharbor_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	//	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func bytesToPEM(certBytes []byte, certPrivKey *rsa.PrivateKey) (*bytes.Buffer, *bytes.Buffer) {
	certPEM := new(bytes.Buffer)
	err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	Expect(err).ToNot(HaveOccurred())

	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	Expect(err).ToNot(HaveOccurred())

	return certPEM, certPrivKeyPEM
}

func verifyCertificate(caPEM *bytes.Buffer, certPEM *bytes.Buffer, dnsNames ...string) {
	roots := x509.NewCertPool()
	Expect(roots.AppendCertsFromPEM(caPEM.Bytes())).To(BeTrue())

	block, _ := pem.Decode(certPEM.Bytes())
	Expect(block).ToNot(BeNil())

	cert, err := x509.ParseCertificate(block.Bytes)
	Expect(err).ToNot(HaveOccurred())

	_, err = cert.Verify(x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
		Roots:                     roots,
		CurrentTime:               time.Now(),
		MaxConstraintComparisions: 0,
	})
	Expect(err).ToNot(HaveOccurred())

	for _, dnsName := range dnsNames {
		_, err = cert.Verify(x509.VerifyOptions{
			Roots:   roots,
			DNSName: dnsName,
		})
		Expect(err).ToNot(HaveOccurred())
	}
}

func generateCertificate(dnsNames ...string) map[string][]byte {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"goharbor.io"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Minute * 30),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).ToNot(HaveOccurred())

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	Expect(err).ToNot(HaveOccurred())

	caPEM, _ := bytesToPEM(caBytes, caPrivKey)

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"goharbor.io"},
		},
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Minute * 30),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).ToNot(HaveOccurred())

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	Expect(err).ToNot(HaveOccurred())

	certPEM, certPrivKeyPEM := bytesToPEM(certBytes, certPrivKey)
	verifyCertificate(caPEM, certPEM, dnsNames...)

	return map[string][]byte{
		corev1.TLSPrivateKeyKey:        certPrivKeyPEM.Bytes(),
		corev1.TLSCertKey:              certPEM.Bytes(),
		corev1.ServiceAccountRootCAKey: caPEM.Bytes(),
	}
}

// setupPostgresql deploy a servicea deployment and a secret to run a postgresql instance
// Based on https://hub.docker.com/_/postgres
func setupPostgresql(ctx context.Context, ns string, databases ...string) harbormetav1.PostgresConnectionWithParameters {
	pgName := newName("pg")
	pgPasswordName := newName("pg-password")
	pgConfigMapName := newName("harbor-init-db")

	Expect(k8sClient.Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgName,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "http",
				Port: 5432,
			}},
			Selector: map[string]string{
				"pod-selector": pgName,
			},
		},
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgPasswordName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.PostgresqlPasswordKey: "th3Adm!nPa$$w0rd",
		},
		Type: harbormetav1.SecretTypePostgresql,
	})).To(Succeed())

	sql := ""
	for _, database := range databases {
		sql += fmt.Sprintf("CREATE DATABASE %s WITH OWNER postgres;", database)
	}

	Expect(k8sClient.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgConfigMapName,
			Namespace: ns,
		},
		Data: map[string]string{
			"init-db.sql": sql,
		},
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &appsv1.Deployment{
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
						Image: "bitnami/postgresql",
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
	})).To(Succeed())

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

// setupRedis deploy a service, a deployment and a secret to run a redis instance
// Based on https://hub.docker.com/_/redis
func setupRedis(ctx context.Context, ns string) harbormetav1.RedisConnection {
	redisName := newName("redis")
	redisPasswordName := newName("redis-password")

	Expect(k8sClient.Create(ctx, &corev1.Service{
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
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redisPasswordName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.RedisPasswordKey: "th3Adm!nPa$$w0rd",
		},
		Type: harbormetav1.SecretTypeRedis,
	})).To(Succeed())

	Expect(k8sClient.Create(ctx, &appsv1.Deployment{
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
						Image: "bitnami/redis",
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
	})).To(Succeed())

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
