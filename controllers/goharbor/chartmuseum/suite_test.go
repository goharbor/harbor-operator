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
	"math/rand"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	// +kubebuilder:scaffold:imports

	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/chartmuseum"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/config"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg        *rest.Config
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
	rand.Seed(GinkgoRandomSeed())

	ctx = test.NewContext(path.Join("..", "..", ".."))

	By("bootstrapping test environment")
	var err error
	cfg, err = test.GetEnvironment(ctx).Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err := client.New(cfg, client.Options{Scheme: test.GetScheme(ctx)})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	ctx = test.SetClient(ctx, k8sClient)

	// +kubebuilder:scaffold:scheme

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		MetricsBindAddress: "0",
		Scheme:             test.GetScheme(ctx),
	})
	Expect(err).NotTo(HaveOccurred(), "failed to create manager")

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

	go func() {
		defer GinkgoRecover()

		Expect(mgr.Start(stopCh)).
			To(Succeed(), "failed to start manager")
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	close(stopCh)

	By("tearing down the test environment")
	Expect(test.GetEnvironment(ctx).Stop()).
		To(Succeed())
})

// SetupTest will set up a testing environment.
// This includes:
// * creating a Namespace to be used during the test
// * starting the Harbor Reconciler
// * stopping the Harbor Reconciler after the test ends
// Call this function at the start of each of your tests.
func SetupTest() *core.Namespace {
	ns := &core.Namespace{}

	BeforeEach(func() {
		stopCh = make(chan struct{})
		*ns = core.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: test.NewName("ns")},
		}

		err := test.GetClient(ctx).Create(ctx, ns)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
	})

	AfterEach(func() {
		err := test.GetClient(ctx).Delete(ctx, ns)
		Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
	})

	return ns
}
