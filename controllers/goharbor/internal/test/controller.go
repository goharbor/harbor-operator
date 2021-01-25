package test

import (
	"context"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Controller interface {
	SetupWithManager(context.Context, manager.Manager) error
}

func StartController(ctx context.Context, controller Controller) (context.Context, string, chan struct{}) {
	ginkgo.By("Starting controller")

	mgr := GetManager(ctx)
	harborClass := NewName("class")

	gomega.Expect(controller.SetupWithManager(ctx, mgr)).
		To(gomega.Succeed())

	stopCh := make(chan struct{})

	go func() {
		defer ginkgo.GinkgoRecover()

		ginkgo.By("Starting manager")

		gomega.Expect(mgr.Start(stopCh)).
			To(gomega.Succeed(), "failed to start manager")
	}()

	return ctx, harborClass, stopCh
}
