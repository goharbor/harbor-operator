package common_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	// +kubebuilder:scaffold:imports

	. "github.com/goharbor/harbor-operator/pkg/controllers/common"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t, "Controller", []Reporter{printer.NewlineReporter{}})
}

func setupTest(ctx context.Context) (*Controller, context.Context) {
	return NewController("test", "version", nil, nil), ctx
}
