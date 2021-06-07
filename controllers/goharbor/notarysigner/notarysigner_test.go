package notarysigner_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/certificate"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/pods"
	portforward "github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/port-forward"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/postgresql"
	"github.com/theupdateframework/notary"
	notary_client "github.com/theupdateframework/notary/signer/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const defaultGenerationNumber int64 = 1

var _ = Describe("NotarySigner", func() {
	var (
		ns           = test.InitNamespace(func() context.Context { return ctx })
		notarysigner goharborv1.NotarySigner
		ca           *certificate.CA
	)

	BeforeEach(func() {
		className, err := reconciler.GetClassName(ctx)
		Expect(err).ToNot(HaveOccurred())

		ca = certificate.NewCA()

		notarysigner.ObjectMeta = metav1.ObjectMeta{
			Name:      test.NewName("notarysigner"),
			Namespace: ns.GetName(),
			Annotations: test.AddVersionAnnotations(map[string]string{
				goharborv1.HarborClassAnnotation: className,
			}),
		}
	})

	JustAfterEach(pods.LogsAll(&ctx, func() types.NamespacedName {
		return types.NamespacedName{
			Name:      reconciler.NormalizeName(ctx, notarysigner.GetName()),
			Namespace: notarysigner.GetNamespace(),
		}
	}))

	Context("Without TLS", func() {
		BeforeEach(func() {
			namespace := notarysigner.GetNamespace()

			certificateName := test.NewName("certificate")
			aliasesName := test.NewName("aliases")

			notarysigner.Spec = goharborv1.NotarySignerSpec{
				Authentication: goharborv1.NotarySignerAuthenticationSpec{
					CertificateRef: certificateName,
				},
				Storage: goharborv1.NotarySignerStorageSpec{
					AliasesRef: aliasesName,
					NotaryStorageSpec: goharborv1.NotaryStorageSpec{
						Postgres: postgresql.New(ctx, namespace),
					},
				},
			}

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      certificateName,
					Namespace: namespace,
				},
				Data: ca.NewCert(reconciler.NormalizeName(ctx, notarysigner.GetName()), "localhost").ToMap(),
				Type: corev1.SecretTypeTLS,
			})).To(Succeed())

			Expect(test.GetClient(ctx).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      aliasesName,
					Namespace: namespace,
				},
				StringData: map[string]string{
					harbormetav1.DefaultAliasSecretKey: "the-key",
				},
				Type: harbormetav1.SecretTypeNotarySignerAliases,
			})).To(Succeed())
		})

		It("Should works", func() {
			By("Creating new resource", func() {
				Ω(test.GetClient(ctx).Create(ctx, &notarysigner)).
					Should(test.SuccessOrExists)

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&notarysigner), &notarysigner)
				}, time.Minute, 5*time.Second).
					Should(Succeed(), "resource should exists")

				Ω(notarysigner.GetGeneration()).
					Should(Equal(defaultGenerationNumber), "Generation should not be updated")

				test.EnsureReady(ctx, &notarysigner, time.Minute, 5*time.Second)

				IntegTest(ctx, &notarysigner, ca)
			})

			By("Updating resource spec", func() {
				oldGeneration := notarysigner.GetGeneration()

				test.ScaleUp(ctx, &notarysigner)

				Ω(notarysigner.GetGeneration()).
					Should(BeNumerically(">", oldGeneration), "ObservedGeneration should be updated")

				Ω(test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&notarysigner), &notarysigner)).
					Should(Succeed(), "resource should still be accessible")

				test.EnsureReady(ctx, &notarysigner, time.Minute, 5*time.Second)

				IntegTest(ctx, &notarysigner, ca)
			})

			By("Deleting resource", func() {
				Ω(test.GetClient(ctx).Delete(ctx, &notarysigner)).
					Should(Succeed())

				Eventually(func() error {
					return test.GetClient(ctx).Get(ctx, test.GetNamespacedName(&notarysigner), &notarysigner)
				}, time.Minute, 5*time.Second).
					ShouldNot(Succeed(), "Resource should no more exist")
			})
		})
	})
})

func IntegTest(ctx context.Context, notarysigner *goharborv1.NotarySigner, ca *certificate.CA) {
	namespacedName := types.NamespacedName{
		Name:      reconciler.NormalizeName(ctx, notarysigner.GetName()),
		Namespace: notarysigner.GetNamespace(),
	}

	localPort, pf := portforward.New(ctx, namespacedName, goharborv1.NotarySignerAPIPort)
	defer pf.Close()

	rootPool := x509.NewCertPool()
	Ω(rootPool.AppendCertsFromPEM(ca.PEM)).Should(BeTrue())

	cert := ca.NewCert()
	tlsCert, err := tls.X509KeyPair(cert.PEM, cert.PrivKey)
	Ω(err).ShouldNot(HaveOccurred())

	clientConn, err := notary_client.NewGRPCConnection("localhost", fmt.Sprintf("%d", localPort), &tls.Config{
		RootCAs:      rootPool,
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS13,
	})
	Ω(err).ShouldNot(HaveOccurred())

	notaClient := notary_client.NewNotarySigner(clientConn)

	Ω(notaClient.CheckHealth(10*time.Second, notary.HealthCheckOverall)).Should(Succeed())
	Ω(notaClient.ListAllKeys()).Should(BeEmpty())
}
