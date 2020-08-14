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
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/kustomize/kstatus/status"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	defaultGenerationNumber int64 = 1
)

type controllerTest struct {
	Setup         func(context.Context, string) (Resource, client.ObjectKey)
	Update        func(context.Context, Resource)
	GetStatusFunc func(ctx context.Context, key client.ObjectKey) func() goharborv1alpha2.ComponentStatus
}

type Resource interface {
	runtime.Object
	metav1.Object

	DeepCopyObject() runtime.Object
}

var ns = SetupTest()
var ctx context.Context

var _ = BeforeEach(func() {
	ctx = logger.Context(log)
})

var _ = DescribeTable(
	"Controller",
	func(resourceController controllerTest, timeouts ...interface{}) {
		By("Creating new resource")

		resource, key := resourceController.Setup(ctx, ns.Name)

		if resource, ok := resource.(webhook.Validator); ok {
			Expect(resource.ValidateCreate()).To(Succeed())
		}

		Eventually(func() error { return k8sClient.Get(ctx, key, resource) }, timeouts...).
			Should(Succeed(), "resource should exists")

		Expect(resource.GetGeneration()).
			Should(Equal(defaultGenerationNumber), "ObservedGeneration should not be updated")

		By("Updating resource spec")

		old := resource.DeepCopyObject()

		resourceController.Update(ctx, resource)

		if resource, ok := resource.(webhook.Validator); ok {
			Expect(resource.ValidateUpdate(old)).To(Succeed())
		}

		Expect(k8sClient.Get(ctx, key, resource)).To(Succeed(), "resource should still be accessible")

		Eventually(resourceController.GetStatusFunc(ctx, key), timeouts...).
			Should(MatchFields(IgnoreExtras, Fields{
				"ObservedGeneration": BeEquivalentTo(resource.GetGeneration()),
				"Conditions": ContainElements(MatchFields(IgnoreExtras, Fields{
					"Type":   BeEquivalentTo(status.ConditionInProgress),
					"Status": BeEquivalentTo(corev1.ConditionFalse),
				}), MatchFields(IgnoreExtras, Fields{
					"Type":   BeEquivalentTo(status.ConditionFailed),
					"Status": BeEquivalentTo(corev1.ConditionFalse),
				})),
				"Operator": MatchFields(IgnoreExtras, Fields{
					"ControllerVersion": BeEquivalentTo(version),
				}),
			}), "resource should be applied")

		By("Deleting resource")

		if resource, ok := resource.(webhook.Validator); ok {
			Expect(resource.ValidateDelete()).To(Succeed())
		}

		Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

		Eventually(func() error { return k8sClient.Get(ctx, key, resource) }, timeouts...).
			ShouldNot(Succeed(), "Resource should no more exist")
	},
	Entry("Portal", newPortalController(), 30*time.Second, 2*time.Second),
	Entry("Registry", newRegistryController(), time.Minute, 5*time.Second),
	Entry("RegistryCtl", newRegistryCtlController(), 2*time.Minute, 5*time.Second),
	Entry("ChartMuseum", newChartMuseumController(), time.Minute, 5*time.Second),
	Entry("NotaryServer", newNotaryServerController(), time.Minute, 5*time.Second),
	Entry("NotarySigner", newNotarySignerController(), time.Minute, 5*time.Second),
	Entry("Core", newCoreController(), time.Minute, 5*time.Second),
	Entry("JobService", newJobServiceController(), time.Minute, 5*time.Second),
	// Following tests require redis
	PEntry("Harbor", newHarborController(), 5*time.Minute, 10*time.Second),
)
