package goharbor_test

import (
	"context"
	"fmt"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/certificate"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/redis"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newTrivyController() controllerTest {
	return controllerTest{
		Setup:         setupValidTrivy,
		Update:        updateTrivy,
		GetStatusFunc: getTrivyStatusFunc,
	}
}

func setupTrivyResourceDependencies(ctx context.Context, ns string, name string) (string, string) {
	trivyCertName := newName("trivy-certificate")
	trivyCertCommonName := fmt.Sprintf("%s-%s", name, controllers.Trivy.String())
	trivyGithubTokenName := newName("trivy-github-token")

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      trivyCertName,
			Namespace: ns,
		},
		Data: certificate.NewCA().NewCert(trivyCertCommonName).ToMap(),
		Type: corev1.SecretTypeTLS,
	})).ToNot(HaveOccurred())

	Expect(k8sClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      trivyGithubTokenName,
			Namespace: ns,
		},
		StringData: map[string]string{
			harbormetav1.GithubTokenKey: "github-token",
		},
		Type: harbormetav1.SecretTypeGithubToken,
	})).To(Succeed())

	return trivyCertName, trivyGithubTokenName
}

func setupValidTrivy(ctx context.Context, ns string) (Resource, client.ObjectKey) {
	var replicas int32 = 1

	name := newName("trivy")
	redis := redis.New(ctx, ns)
	trivyCertName, trivyGithubTokenName := setupTrivyResourceDependencies(ctx, ns, name)

	trivy := &goharborv1.Trivy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: test.AddVersionAnnotations(nil),
		},

		Spec: goharborv1.TrivySpec{
			ComponentSpec: harbormetav1.ComponentSpec{
				Replicas: &replicas,
			},
			Redis: goharborv1.TrivyRedisSpec{
				RedisConnection: redis,
			},
			Server: goharborv1.TrivyServerSpec{
				TLS: &harbormetav1.ComponentsTLSSpec{
					CertificateRef: trivyCertName,
				},
			},
			Update: goharborv1.TrivyUpdateSpec{
				GithubTokenRef: trivyGithubTokenName,
				Skip:           false,
			},
		},
	}

	Expect(k8sClient.Create(ctx, trivy)).To(Succeed())

	return trivy, client.ObjectKey{
		Name:      name,
		Namespace: ns,
	}
}

func updateTrivy(ctx context.Context, object Resource) {
	trivy, ok := object.(*goharborv1.Trivy)
	Expect(ok).To(BeTrue())

	var replicas int32 = 1

	if trivy.Spec.Replicas != nil {
		replicas = *trivy.Spec.Replicas + 1
	}

	trivy.Spec.Replicas = &replicas
}

func getTrivyStatusFunc(ctx context.Context, key client.ObjectKey) func() harbormetav1.ComponentStatus {
	return func() harbormetav1.ComponentStatus {
		var trivy goharborv1.Trivy

		err := k8sClient.Get(ctx, key, &trivy)

		Expect(err).ToNot(HaveOccurred())

		return trivy.Status
	}
}
