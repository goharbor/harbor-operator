package goharbor_test

import (
	"context"
	"net/url"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const storageRequest = "10Mi"

var _ = Context("Harbor reconciler", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = logger.Context(log)
	})

	Describe("Creating resources with invalid public url", func() {
		It("Should raise an error", func() {
			harbor := &goharborv1.Harbor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "harbor-invalid-url",
					Namespace: ns.Name,
				},
				Spec: goharborv1.HarborSpec{
					ExternalURL: "123::bad::dns",
				},
			}

			err := k8sClient.Create(ctx, harbor)
			Expect(err).To(HaveOccurred())
			Expect(err).To(WithTransform(apierrs.IsInvalid, BeTrue()))
		})
	})
})

func newHarborController() controllerTest {
	return controllerTest{
		Setup:         setupValidHarbor,
		Update:        updateHarbor,
		GetStatusFunc: getHarborStatusFunc,
	}
}

func setupHarborResourceDependencies(ctx context.Context, ns string) (string, string, string, string) {
	adminSecretName := newName("admin-secret")

	err := k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      adminSecretName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.SharedSecretKey: "th3Adm!nPa$$w0rd",
		},
		Type: harbormetav1.SecretTypeSingle,
	})
	Expect(err).ToNot(HaveOccurred())

	tokenIssuerName := newName("token-issuer")

	err = k8sClient.Create(ctx, &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenIssuerName,
			Namespace: ns,
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	})
	Expect(err).ToNot(HaveOccurred())

	registryPvcName := newName("registry-pvc")

	err = k8sClient.Create(ctx, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryPvcName,
			Namespace: ns,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storageRequest),
				},
			},
		},
	})
	Expect(err).ToNot(HaveOccurred())

	chartPvcName := newName("chart-pvc")

	err = k8sClient.Create(ctx, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chartPvcName,
			Namespace: ns,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(storageRequest),
				},
			},
		},
	})
	Expect(err).ToNot(HaveOccurred())

	return registryPvcName, chartPvcName, adminSecretName, tokenIssuerName
}

func setupValidHarbor(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	registryPvcName, chartPvcName, adminSecretName, tokenIssuerName := setupHarborResourceDependencies(ctx, ns)

	database := postgresql.New(ctx, ns, "core")
	redis := redis.New(ctx, ns)

	name := newName("harbor")
	publicURL := url.URL{
		Scheme: "http",
		Host:   "the.dns",
	}

	harbor := &goharborv1.Harbor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: goharborv1.HarborSpec{
			ExternalURL:            publicURL.String(),
			HarborAdminPasswordRef: adminSecretName,
			Version:                test.GetVersion(),
			ImageChartStorage: &goharborv1.HarborStorageImageChartStorageSpec{
				FileSystem: &goharborv1.HarborStorageImageChartStorageFileSystemSpec{
					RegistryPersistentVolume: goharborv1.HarborStorageRegistryPersistentVolumeSpec{
						HarborStoragePersistentVolumeSpec: goharborv1.HarborStoragePersistentVolumeSpec{
							PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: registryPvcName,
							},
						},
					},
					ChartPersistentVolume: &goharborv1.HarborStoragePersistentVolumeSpec{
						PersistentVolumeClaimVolumeSource: corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: chartPvcName,
						},
					},
				},
			},
			HarborComponentsSpec: goharborv1.HarborComponentsSpec{
				Core: goharborv1.CoreComponentSpec{
					TokenIssuer: cmmeta.ObjectReference{
						Name: tokenIssuerName,
					},
				},
				Database: &goharborv1.HarborDatabaseSpec{
					PostgresCredentials: database.PostgresCredentials,
					Hosts:               database.Hosts,
					SSLMode:             harbormetav1.PostgresSSLMode(database.Parameters[harbormetav1.PostgresSSLModeKey]),
				},
				Redis: &goharborv1.ExternalRedisSpec{
					RedisHostSpec:    redis.RedisHostSpec,
					RedisCredentials: redis.RedisCredentials,
				},
			},
		},
	}

	Expect(k8sClient.Create(ctx, harbor)).To(Succeed())

	return harbor, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateHarbor(ctx context.Context, object Resource) {
	harbor, ok := object.(*goharborv1.Harbor)
	Expect(ok).To(BeTrue())

	u, err := url.Parse(harbor.Spec.ExternalURL)
	Expect(err).ToNot(HaveOccurred())

	u.Host = "new." + u.Host
	harbor.Spec.ExternalURL = u.String()
}

func getHarborStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var harbor goharborv1.Harbor

		err := k8sClient.Get(ctx, key, &harbor)

		Expect(err).ToNot(HaveOccurred())

		return harbor.Status
	}
}
