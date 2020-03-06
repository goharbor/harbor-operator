package components

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	containerregistryv1alpha1 "github.com/goharbor/harbor-core-operator/api/v1alpha1"
	"github.com/goharbor/harbor-core-operator/pkg/factories/logger"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestComponents(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Components Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = Context("With minimal harbor", func() {
	log := zap.LoggerTo(GinkgoWriter, true)

	harbor := &containerregistryv1alpha1.Harbor{
		Spec: containerregistryv1alpha1.HarborSpec{
			HarborVersion: "1.9.1",
			PublicURL:     "http://localhost",
		},
	}
	harbor.Default()

	Measure("get components", func(b Benchmarker) {
		runtime := b.Time("runtime", func() {
			components, err := GetComponents(logger.Context(log), harbor)
			Expect(err).ToNot(HaveOccurred())
			Expect(components).ToNot(BeNil())
		})

		Expect(runtime.Seconds()).Should(BeNumerically("<", 0.05), "GetComponents() should not take too long")
	}, 1000)

	var components *Components
	It("get components should succeed", func() {
		var err error
		components, err = GetComponents(logger.Context(log), harbor)
		Expect(err).ToNot(HaveOccurred())
	})

	Measure("parallel run", func(b Benchmarker) {
		runtime := b.Time("runtime", func() {
			err := components.ParallelRun(logger.Context(log), harbor, func(context.Context, *containerregistryv1alpha1.Harbor, *ComponentRunner) error {
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})

		Expect(runtime.Seconds()).Should(BeNumerically("<", 0.1), "ParallelRun() should not take too long")
	}, 1000)
})

var _ = Context("With full harbor", func() {
	log := zap.LoggerTo(GinkgoWriter, true)

	harbor := &containerregistryv1alpha1.Harbor{
		Spec: containerregistryv1alpha1.HarborSpec{
			HarborVersion: "1.9.1",
			PublicURL:     "http://localhost",
			Components: containerregistryv1alpha1.HarborComponents{
				ChartMuseum: &containerregistryv1alpha1.ChartMuseumComponent{},
				Clair:       &containerregistryv1alpha1.ClairComponent{},
				Notary:      &containerregistryv1alpha1.NotaryComponent{},
			},
		},
	}
	harbor.Default()

	Measure("get components", func(b Benchmarker) {
		runtime := b.Time("runtime", func() {
			components, err := GetComponents(logger.Context(log), harbor)
			Expect(err).ToNot(HaveOccurred())
			Expect(components).ToNot(BeNil())
		})

		Expect(runtime.Seconds()).Should(BeNumerically("<", 0.05), "GetComponents() should not take too long")
	}, 1000)

	var components *Components
	It("get components should succeed", func() {
		var err error
		components, err = GetComponents(logger.Context(log), harbor)
		Expect(err).ToNot(HaveOccurred())
	})

	Measure("parallel run", func(b Benchmarker) {
		runtime := b.Time("runtime", func() {
			err := components.ParallelRun(logger.Context(log), harbor, func(context.Context, *containerregistryv1alpha1.Harbor, *ComponentRunner) error {
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})

		Expect(runtime.Seconds()).Should(BeNumerically("<", 0.1), "ParallelRun() should not take too long")
	}, 1000)
})
