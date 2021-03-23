package harborcore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	nginxRootPath   = "/usr/share/nginx/html"
	internalAPIPath = "/api/internal"
	v2APIPath       = "/api/v2.0"
	port            = 80
)

func DeployDatabase(ctx context.Context, ns string, coreConfig *config.CfgManager) harbormetav1.PostgresConnectionWithParameters {
	pg := postgresql.New(ctx, ns)

	coreConfig.Set(common.PostGreSQLHOST, pg.Hosts[0].Host)
	coreConfig.Set(common.PostGreSQLPassword, pg.Hosts[0].Port)
	coreConfig.Set(common.PostGreSQLDatabase, pg.Database)
	coreConfig.Set(common.DatabaseType, "postgresql")

	if sslMode, ok := pg.Parameters[harbormetav1.PostgresSSLModeKey]; ok {
		coreConfig.Set(common.PostGreSQLSSLMode, sslMode)
	}

	pgPassword := &corev1.Secret{}
	gomega.Expect(test.GetClient(ctx).Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      pg.PasswordRef,
	}, pgPassword)).To(gomega.Succeed())
	coreConfig.Set(common.PostGreSQLPassword, string(pgPassword.Data[harbormetav1.PostgresqlPasswordKey]))
	coreConfig.Set(common.PostGreSQLUsername, pg.Username)

	return pg
}

func New(ctx context.Context, ns string, coreConfig *config.CfgManager) *url.URL {
	k8sClient := test.GetClient(ctx)

	coreName := test.NewName("core")
	internalAPIMock := test.NewName("core-internal-api")
	v2APIMock := test.NewName("core-api-v2")

	localURL := coreConfig.Get(common.CoreLocalURL).GetString()
	if localURL == "" {
		localURL = fmt.Sprintf("http://%s:%d", coreName, port)
	}

	u, err := url.Parse(localURL)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	publicPort := int32(port)

	if port := u.Port(); port != "" {
		result, err := strconv.ParseInt(u.Port(), 10, 32)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		publicPort = int32(result)
	}

	gomega.Expect(k8sClient.Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      u.Hostname(),
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       publicPort,
				TargetPort: intstr.FromInt(port),
			}},
			Selector: map[string]string{
				"pod-selector": coreName,
			},
		},
	})).To(gomega.Succeed())

	gomega.Expect(coreConfig.Save()).To(gomega.Succeed())

	internalConfigurations, err := json.Marshal(coreConfig.GetAll())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	gomega.Expect(k8sClient.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      internalAPIMock,
			Namespace: ns,
		},
		BinaryData: map[string][]byte{
			"configurations": internalConfigurations,
		},
	})).To(gomega.Succeed())

	gomega.Expect(k8sClient.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v2APIMock,
			Namespace: ns,
		},
		Data: map[string]string{
			"health": `{"status":"healthy","components":[]}`,
		},
	})).To(gomega.Succeed())

	gomega.Expect(k8sClient.Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      coreName,
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"pod-selector": coreName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"pod-selector": coreName,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "internal-api",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: internalAPIMock,
								},
							},
						},
					}, {
						Name: "api-v2",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: v2APIMock,
								},
							},
						},
					}},
					Containers: []corev1.Container{{
						Name:  "mock",
						Image: "nginx",
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: nginxRootPath + internalAPIPath,
							Name:      "internal-api",
						}, {
							MountPath: nginxRootPath + v2APIPath,
							Name:      "api-v2",
						}},
					}},
				},
			},
		},
	})).To(gomega.Succeed())

	return u
}
