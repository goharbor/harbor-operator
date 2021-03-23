package test

import (
	"context"
	"flag"
	"fmt"
	"math/rand"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitSuite() context.Context {
	initFlag()

	ginkgo.By("Configuring seed", func() {
		rand.Seed(ginkgo.GinkgoRandomSeed())
	})

	ginkgo.By("Configuring logger", func() {
		ConfigureLoggers(ginkgo.GinkgoWriter)
	})

	var ctx context.Context

	ginkgo.By("bootstrapping test environment", func() {
		ctx = NewContext()

		cfg, err := GetEnvironment(ctx).Start()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(cfg).ToNot(gomega.BeNil())

		ctx = WithRestConfig(ctx, cfg)

		ctx = WithClient(ctx, NewClient(ctx))
		ctx = WithManager(ctx, NewManager(ctx))
	})

	gomega.Expect(ctx).ToNot(gomega.BeNil())

	return ctx
}

func AfterSuite(ctx context.Context) {
	ginkgo.By("tearing down the test environment", func() {
		if ctx != nil {
			gomega.Expect(GetEnvironment(ctx).Stop()).
				To(gomega.Succeed())
		}
	})
}

var keepNamespaceOnFailure bool

func initFlag() {
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
