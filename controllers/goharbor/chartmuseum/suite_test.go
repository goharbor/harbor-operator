/*
Copyright 2019 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package chartmuseum_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	stopCh     chan struct{}
	ctx        context.Context
	reconciler *chartmuseum.Reconciler
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	ctx = test.InitSuite()

	By("Configuring controller")

	mgr := test.GetManager(ctx)
	name := controllers.ChartMuseum.String()

	configStore := config.NewConfigWithDefaults()
	configStore.Env(name)

	commonReconciler, err := chartmuseum.New(ctx, name, configStore)
	Expect(err).ToNot(HaveOccurred())

	var ok bool
	reconciler, ok = commonReconciler.(*chartmuseum.Reconciler)
	Expect(ok).To(BeTrue())

	Expect(reconciler.SetupWithManager(ctx, mgr)).
		To(Succeed())

	stopCh = make(chan struct{})

	go func() {
		defer GinkgoRecover()

		By("Starting manager")

		Expect(mgr.Start(stopCh)).
			To(Succeed(), "failed to start manager")
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	close(stopCh)

	test.AfterSuite(ctx)
})
