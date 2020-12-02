package graph_test

import (
	"context"
	"testing"

	. "github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// +kubebuilder:scaffold:imports
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestSuite(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t, "resourceManager", []Reporter{printer.NewlineReporter{}})
}

func setupTest(ctx context.Context) (Manager, context.Context) {
	return NewResourceManager(), ctx
}
