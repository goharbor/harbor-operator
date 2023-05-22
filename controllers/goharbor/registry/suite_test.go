package registry_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/registry"
)

var (
	ctx        context.Context
	reconciler *registry.Reconciler
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	ctx = test.InitSuite()
	className := test.NewName("class")

	reconciler = controllers.NewRegistry(ctx, className)
	Expect(reconciler).ToNot(BeNil())

	test.StartManager(ctx)
})

var _ = AfterSuite(func() {
	defer test.AfterSuite(ctx)

	ctx.Done()
})
