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

package goharbor_test

import (
	"fmt"
	"math/rand"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	"github.com/goharbor/harbor-operator/pkg/setup"
	"github.com/ovh/configstore"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	version   string
	log       = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
)

const configDirectory = "../../config/config"

func TestAPIs(t *testing.T) {
	t.Parallel()

	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	rand.Seed(GinkgoRandomSeed())

	version = newName("version")

	logf.SetLogger(log)
	ctx := logger.Context(log)

	application.SetName(&ctx, "test-app")
	application.SetVersion(&ctx, version)
	application.SetGitCommit(&ctx, "test")

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	s, err := scheme.New(ctx)
	Expect(err).ToNot(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: s})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		MetricsBindAddress: "0",
		Scheme:             s,
	})
	Expect(err).NotTo(HaveOccurred(), "failed to create manager")

	configstore.DefaultStore.InMemory(test.NewName("test-")).
		Add(configstore.NewItem(config.TemplateDirectoryKey, path.Join(configDirectory, "assets"), 100)).
		Add(configstore.NewItem(config.CtrlConfigDirectoryKey, path.Join(configDirectory, "controllers"), 100))

	Expect(setup.ControllersWithManager(ctx, mgr)).To(Succeed())

	go func() {
		defer GinkgoRecover()

		err := mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred(), "failed to start manager")
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	ctx.Done()

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// SetupTest will set up a testing environment.
// This includes:
// * creating a Namespace to be used during the test
// * starting the Harbor Reconciler
// * stopping the Harbor Reconciler after the test ends
// Call this function at the start of each of your tests.
func SetupTest() *core.Namespace {
	ctx := logger.Context(log)
	ns := &core.Namespace{}

	BeforeEach(func() {
		*ns = core.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: newName("testns")},
		}

		err := k8sClient.Create(ctx, ns)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
	})

	AfterEach(func() {
		err := k8sClient.Delete(ctx, ns)
		Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
	})

	return ns
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))] //nolint:gosec
	}

	return string(b)
}

const prefixLength = 8

func newName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randStringRunes(prefixLength))
}
