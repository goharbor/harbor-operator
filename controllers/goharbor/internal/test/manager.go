package test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func StartManager(ctx context.Context) {
	ginkgo.By("Starting controller")

	go func() {
		defer ginkgo.GinkgoRecover()

		ginkgo.By("Starting manager")

		gomega.Expect(GetManager(ctx).Start(ctx)).
			To(gomega.Succeed(), "failed to start manager")
	}()
}

func NewManager(ctx context.Context) manager.Manager {
	mgr, err := ctrl.NewManager(GetRestConfig(ctx), ctrl.Options{
		Metrics: server.Options{
			BindAddress: "0",
		},
		Scheme: GetScheme(ctx),
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create manager")

	return mgr
}
