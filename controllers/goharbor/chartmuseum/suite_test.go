package chartmuseum_test

import (
	"context"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	internalconfig "github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/config"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

var (
	stopCh     chan struct{}
	ctx        context.Context
	reconciler *chartmuseum.Reconciler
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	ctx = test.InitSuite()

	name := controllers.ChartMuseum.String()

	configStore, provider := internalconfig.New(ctx, chartmuseum.ConfigTemplatePathKey, path.Base(chartmuseum.DefaultConfigTemplatePath))
	provider.Add(configstore.NewItem(config.HarborClassKey, test.NewName("class"), 100))
	configStore.Env(name)

	r, err := chartmuseum.New(ctx, name, configStore)
	Expect(err).ToNot(HaveOccurred())

	reconciler = r.(*chartmuseum.Reconciler)

	ctx, stopCh = test.StartController(ctx, reconciler)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	defer test.AfterSuite(ctx)

	if stopCh != nil {
		close(stopCh)
	}
})
