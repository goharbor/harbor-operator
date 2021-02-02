package test

import (
	"context"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func StartManager(ctx context.Context) (context.Context, chan struct{}) {
	ginkgo.By("Starting controller")

	stopCh := make(chan struct{})

	go func() {
		defer ginkgo.GinkgoRecover()

		ginkgo.By("Starting manager")

		gomega.Expect(GetManager(ctx).Start(stopCh)).
			To(gomega.Succeed(), "failed to start manager")
	}()

	return ctx, stopCh
}

func NewManager(ctx context.Context) manager.Manager {
	mgr, err := ctrl.NewManager(GetRestConfig(ctx), ctrl.Options{
		MetricsBindAddress: "0",
		Scheme:             GetScheme(ctx),
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create manager")

	return mgr
}
