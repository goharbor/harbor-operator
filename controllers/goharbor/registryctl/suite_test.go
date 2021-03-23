package registryctl_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registryctl"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

var (
	stopCh             chan struct{}
	ctx                context.Context
	reconciler         *registryctl.Reconciler
	registryReconciler *registry.Reconciler
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	ctx = test.InitSuite()

	className := test.NewName("class")

	reconciler = controllers.NewRegistryCtl(ctx, className)
	registryReconciler = controllers.NewRegistry(ctx, className)

	ctx, stopCh = test.StartManager(ctx)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	defer test.AfterSuite(ctx)

	if stopCh != nil {
		close(stopCh)
	}
})
