package notarysigner_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/plotly/harbor-operator/controllers/goharbor/internal/test"
	"github.com/plotly/harbor-operator/controllers/goharbor/internal/test/controllers"
	"github.com/plotly/harbor-operator/controllers/goharbor/notarysigner"
)

var (
	ctx        context.Context
	reconciler *notarysigner.Reconciler
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	ctx = test.InitSuite()

	className := test.NewName("class")

	reconciler = controllers.NewNotarySigner(ctx, className)

	test.StartManager(ctx)
})

var _ = AfterSuite(func() {
	defer test.AfterSuite(ctx)

	ctx.Done()
})
