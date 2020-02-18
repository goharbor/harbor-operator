package harbor

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/event"

	// +kubebuilder:scaffold:imports

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/scheme"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestEventFilter(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"EventFilter Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

// SetupTest will set up a testing environment.
// This includes:
// * creating a Namespace to be used during the test
// * starting the Harbor Reconciler
// * stopping the Harbor Reconciler after the test ends
// Call this function at the start of each of your tests.
func SetupTest(ctx context.Context) *EventFilter {
	ef := &EventFilter{}

	BeforeEach(func() {
		s, err := scheme.New(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to initialize scheme")

		ef.Scheme = s
	})

	return ef
}

func getEventTester(ef *EventFilter, resource runtime.Object, matchResult types.GomegaMatcher, factory func(*metav1.ObjectMeta, runtime.Object) bool) func() {
	return func() {
		It("resource with no annotation", func() {
			ok := factory(&metav1.ObjectMeta{}, resource.DeepCopyObject())
			Expect(ok).To(matchResult)
		})

		It("resource with other annotations", func() {
			ok := factory(&metav1.ObjectMeta{
				Annotations: map[string]string{
					fmt.Sprintf("%s-false", containerregistryv1alpha1.HarborClassAnnotation): "",
				},
			}, resource.DeepCopyObject())
			Expect(ok).To(matchResult)
		})

		It("resource with empty class", func() {
			ok := factory(&metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: "",
				},
			}, resource.DeepCopyObject())
			Expect(ok).To(matchResult)
		})

		It("resource with other-class should not match", func() {
			ok := factory(&metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: "other-class",
				},
			}, resource.DeepCopyObject())
			Expect(ok).To(BeFalse())
		})

		It("resource with the right class should match", func() {
			ok := factory(&metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: ef.ClassName,
				},
			}, resource.DeepCopyObject())
			Expect(ok).To(BeTrue())
		})
	}
}

func getUpdateEventTester(ef *EventFilter, resource runtime.Object, matchResult types.GomegaMatcher, oldMeta metav1.Object, oldResource runtime.Object, subMatchResult types.GomegaMatcher) func() {
	return func() {
		It("resource with no annotation", func() {
			ok := ef.Update(event.UpdateEvent{MetaOld: oldMeta, ObjectOld: oldResource, MetaNew: &metav1.ObjectMeta{}, ObjectNew: resource.DeepCopyObject()})
			Expect(ok).To(SatisfyAny(matchResult, subMatchResult))
		})

		It("resource with other annotations", func() {
			ok := ef.Update(event.UpdateEvent{MetaOld: oldMeta, ObjectOld: oldResource, MetaNew: &metav1.ObjectMeta{
				Annotations: map[string]string{
					fmt.Sprintf("%s-false", containerregistryv1alpha1.HarborClassAnnotation): "",
				},
			}, ObjectNew: resource.DeepCopyObject()})
			Expect(ok).To(SatisfyAny(matchResult, subMatchResult))
		})

		It("resource with empty class", func() {
			ok := ef.Update(event.UpdateEvent{MetaOld: oldMeta, ObjectOld: oldResource, MetaNew: &metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: "",
				},
			}, ObjectNew: resource.DeepCopyObject()})
			Expect(ok).To(SatisfyAny(matchResult, subMatchResult))
		})

		It("resource with other-class", func() {
			ok := ef.Update(event.UpdateEvent{MetaOld: oldMeta, ObjectOld: oldResource, MetaNew: &metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: "other-class",
				},
			}, ObjectNew: resource.DeepCopyObject()})
			Expect(ok).To(subMatchResult)
		})

		It("resource with the right class should match", func() {
			ok := ef.Update(event.UpdateEvent{MetaOld: oldMeta, ObjectOld: oldResource, MetaNew: &metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: ef.ClassName,
				},
			}, ObjectNew: resource.DeepCopyObject()})
			Expect(ok).To(BeTrue())
		})
	}
}

func matcher(ef *EventFilter, resource runtime.Object, matchResult types.GomegaMatcher) func() {
	return func() {
		Context("For a creation event", getEventTester(ef, resource, matchResult, func(meta *metav1.ObjectMeta, resource runtime.Object) bool {
			return ef.Create(event.CreateEvent{Meta: meta, Object: resource})
		}))

		Context("For a deletion event", getEventTester(ef, resource, matchResult, func(meta *metav1.ObjectMeta, resource runtime.Object) bool {
			return ef.Delete(event.DeleteEvent{Meta: meta, Object: resource})
		}))

		Context("For a generic event", getEventTester(ef, resource, matchResult, func(meta *metav1.ObjectMeta, resource runtime.Object) bool {
			return ef.Generic(event.GenericEvent{Meta: meta, Object: resource})
		}))

		Context("For a update event", func() {
			Context("From a resource with no annotation", getUpdateEventTester(ef, resource, matchResult, &metav1.ObjectMeta{}, resource.DeepCopyObject(), matchResult))

			Context("From a resource with other annotations", getUpdateEventTester(ef, resource, matchResult, &metav1.ObjectMeta{
				Annotations: map[string]string{
					fmt.Sprintf("%s-false", containerregistryv1alpha1.HarborClassAnnotation): "",
				},
			}, resource.DeepCopyObject(), matchResult))

			Context("From a resource with empty class", getUpdateEventTester(ef, resource, matchResult, &metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: "",
				},
			}, resource.DeepCopyObject(), matchResult))

			Context("From a resource with other-class", getUpdateEventTester(ef, resource, matchResult, &metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: "other-class",
				},
			}, resource.DeepCopyObject(), BeFalse()))

			Context("From a resource with the right class", getUpdateEventTester(ef, resource, matchResult, &metav1.ObjectMeta{
				Annotations: map[string]string{
					containerregistryv1alpha1.HarborClassAnnotation: ef.ClassName,
				},
			}, resource.DeepCopyObject(), BeTrue()))
		})
	}
}

var _ = Context("With no harbor-class", func() {
	ef := SetupTest(context.TODO())

	Context("For an Harbor resource", matcher(ef, &containerregistryv1alpha1.Harbor{}, BeTrue()))
	Context("For an Deployment resource", matcher(ef, &appsv1.Deployment{}, BeTrue()))
})

var _ = Context("With a specified harbor-class", func() {
	ef := SetupTest(context.TODO())
	ef.ClassName = "the-class-name"

	Context("For an Harbor resource", matcher(ef, &containerregistryv1alpha1.Harbor{}, BeFalse()))
	Context("For an Deployment resource", matcher(ef, &appsv1.Deployment{}, BeFalse()))
})
