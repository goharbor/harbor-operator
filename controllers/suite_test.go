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

package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	harborCtrl "github.com/goharbor/harbor-operator/controllers/harbor"
	"github.com/goharbor/harbor-operator/pkg/controllers/common"
	"github.com/goharbor/harbor-operator/pkg/controllers/health"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/scheme"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	stopCh    chan struct{}
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	rand.Seed(time.Now().UnixNano())
	log := zap.LoggerTo(GinkgoWriter, true)
	logf.SetLogger(log)
	ctx := logger.Context(log)

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

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

	controller := &harborCtrl.Reconciler{
		HealthClient: health.Client{
			Scheme: s,
		},
		Log:        logf.Log,
		Controller: common.NewController("test-operator", "test"),
		Scheme:     s,
		RestConfig: mgr.GetConfig(),
	}
	err = controller.SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred(), "failed to setup controller")

	go func() {
		defer GinkgoRecover()

		err := mgr.Start(stopCh)
		Expect(err).NotTo(HaveOccurred(), "failed to start manager")
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	close(stopCh)

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
func SetupTest(ctx context.Context) *core.Namespace {
	ns := &core.Namespace{}

	BeforeEach(func() {
		stopCh = make(chan struct{})
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
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

const prefixLength = 8

func newName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randStringRunes(prefixLength))
}
