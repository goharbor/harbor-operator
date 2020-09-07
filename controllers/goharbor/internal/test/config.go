package test

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"path"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InitSuite() context.Context {
	ginkgo.By("Configuring seed")

	rand.Seed(ginkgo.GinkgoRandomSeed())

	ginkgo.By("Configuring logger")

	ConfigureLoggers(ginkgo.GinkgoWriter)

	ctx := NewContext(path.Join("..", "..", ".."))

	ginkgo.By("bootstrapping test environment")

	var err error
	cfg, err := GetEnvironment(ctx).Start()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(cfg).ToNot(gomega.BeNil())

	ctx = WithRestConfig(ctx, cfg)

	k8sClient, err := client.New(cfg, client.Options{Scheme: GetScheme(ctx)})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(k8sClient).ToNot(gomega.BeNil())

	ctx = WithClient(ctx, k8sClient)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		MetricsBindAddress: "0",
		Scheme:             GetScheme(ctx),
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create manager")

	ctx = WithManager(ctx, mgr)

	return ctx
}

func AfterSuite(ctx context.Context) {
	ginkgo.By("tearing down the test environment")

	gomega.Expect(GetEnvironment(ctx).Stop()).
		To(gomega.Succeed())
}

var keepNamespaceOnFailure bool

func init() {
	flag.BoolVar(&keepNamespaceOnFailure, "keepNamespaceOnFailure", false, "set to true to keep namespaces after tests")
}

func InitNamespace(ctxFactory func() context.Context) *corev1.Namespace {
	ns := &corev1.Namespace{}

	ginkgo.BeforeEach(func() {
		ctx := ctxFactory()
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: NewName("ns")},
		}

		gomega.Expect(GetClient(ctx).Create(ctx, ns)).
			To(SuccessOrExists, "failed to create test namespace")
	})

	ginkgo.AfterEach(func() {
		if ginkgo.CurrentGinkgoTestDescription().Failed && keepNamespaceOnFailure {
			fmt.Fprintf(ginkgo.GinkgoWriter, "keeping namespace %s\n", ns.GetName())
			return
		}

		ctx := ctxFactory()
		gomega.Expect(GetClient(ctx).Delete(ctx, ns)).
			Should(gomega.Succeed(), "failed to delete test namespace")
	})

	return ns
}
