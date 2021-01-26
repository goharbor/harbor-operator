package chartmuseum_test

import (
	"context"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovh/configstore"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	stopCh      chan struct{}
	ctx         context.Context
	reconciler  *chartmuseum.Reconciler
	harborClass string
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	ctx = test.InitSuite()

	By("Configuring controller")

	mgr := test.GetManager(ctx)
	name := controllers.ChartMuseum.String()
	harborClass = test.NewName(name)

	configStore, provider := test.NewConfig(ctx, chartmuseum.ConfigTemplatePathKey, path.Base(chartmuseum.DefaultConfigTemplatePath))
	provider.Add(configstore.NewItem(config.HarborClassKey, harborClass, 100))
	configStore.Env(name)

	commonReconciler, err := chartmuseum.New(ctx, name, configStore)
	Expect(err).ToNot(HaveOccurred())

	var ok bool
	reconciler, ok = commonReconciler.(*chartmuseum.Reconciler)
	Expect(ok).To(BeTrue())

	Expect(reconciler.SetupWithManager(ctx, mgr)).
		To(Succeed())

	stopCh = make(chan struct{})

	go func() {
		defer GinkgoRecover()

		By("Starting manager")

		Expect(mgr.Start(stopCh)).
			To(Succeed(), "failed to start manager")
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	defer test.AfterSuite(ctx)

	close(stopCh)
})
