package notarysigner_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/notarysigner"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

var (
	ctx        context.Context
	reconciler *notarysigner.Reconciler
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

	reconciler = controllers.NewNotarySigner(ctx, className)

	test.StartManager(ctx)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	defer test.AfterSuite(ctx)

	ctx.Done()
})
