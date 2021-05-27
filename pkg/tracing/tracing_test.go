package tracing_test

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	. "github.com/goharbor/harbor-operator/pkg/tracing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/opentracing/opentracing-go"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = BeforeSuite(func() {
	// TODO check that before each test once
	// it is possible to *unregister* or reset the GlobalTracer
	ok := opentracing.IsGlobalTracerRegistered()
	Expect(ok).To(BeFalse(), "tracing should not be registered")
})

var _ = Describe("Intializing tracing", func() {
	var ctx context.Context

	var i int32

	BeforeEach(func() {
		ctx = logger.Context(zap.New(zap.UseDevMode(true)))

		application.SetName(&ctx, fmt.Sprintf("test-default-service-%d", atomic.AddInt32(&i, 1)))
		application.SetVersion(&ctx, "test-version")
		application.SetGitCommit(&ctx, "test-commit")
	})

	Context("With no value", func() {
		It("Should be registered", func() {
			tracer, err := New(ctx)
			Expect(err).ToNot(HaveOccurred())

			ok := opentracing.IsGlobalTracerRegistered()
			Expect(ok).To(BeTrue(), "tracing should be registered")

			err = tracer.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	PContext("Two times", func() {
		It("Should override the first tracer", func() {
			tracer1, err := New(ctx)
			Expect(err).ToNot(HaveOccurred())

			ok := opentracing.IsGlobalTracerRegistered()
			Expect(ok).To(BeTrue(), "tracing should be registered")

			tracer2, err := New(ctx)
			Expect(err).ToNot(HaveOccurred())

			ok = opentracing.IsGlobalTracerRegistered()
			Expect(ok).To(BeTrue(), "tracing should be registered")

			// Check that the global tracer is tracer2

			err = tracer1.Close()
			Expect(err).ToNot(HaveOccurred())

			err = tracer2.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("With environment", func() {
		It("Should be registered", func() {
			// https://github.com/jaegertracing/jaeger-client-go#environment-variables
			os.Setenv("JAEGER_SERVICE_NAME", "test-service")
			os.Setenv("JAEGER_SAMPLER_TYPE", "const")
			os.Setenv("JAEGER_SAMPLER_PARAM", "1")

			tracer, err := New(ctx)
			Expect(err).ToNot(HaveOccurred())

			ok := opentracing.IsGlobalTracerRegistered()
			Expect(ok).To(BeTrue(), "tracing should be registered")

			err = tracer.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
