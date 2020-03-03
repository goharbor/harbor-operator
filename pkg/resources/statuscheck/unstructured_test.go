package statuscheck

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kstatus/status"

	// +kubebuilder:scaffold:imports

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/scheme"
)

// These tests use Ginkgo (BDD-style Go testing framework). Rcfer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Check the status", func() {
	Context("Of a pod resource", func() {
		var resource *unstructured.Unstructured
		var data *corev1.Pod

		BeforeEach(func() {
			s, err := scheme.New(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			data = &corev1.Pod{}
			gvks, _, err := s.ObjectKinds(data)
			Expect(err).ToNot(HaveOccurred())

			gvk := gvks[0]
			data.SetGroupVersionKind(gvk)

			resource = &unstructured.Unstructured{}
		})

		JustBeforeEach(func() {
			data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(data)
			Expect(err).ToNot(HaveOccurred())

			resource.SetUnstructuredContent(data)
		})

		Context("With empty status", func() {
			BeforeEach(func() {
				data.Status = corev1.PodStatus{}
			})

			It("Should not be ready", func() {
				ok, err := UnstructuredCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With podReady condition", func() {
			Context("To False", func() {
				BeforeEach(func() {
					data.Status.Conditions = []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					}
				})

				It("Should not be ready", func() {
					ok, err := UnstructuredCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeFalse())
				})
			})
			Context("To True", func() {
				BeforeEach(func() {
					data.Status.Conditions = []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					}
				})

				It("Should be ready", func() {
					ok, err := UnstructuredCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeTrue())
				})
			})
		})
	})

	Context("Of a portal resource", func() {
		var resource *unstructured.Unstructured
		var data *goharborv1alpha2.Portal

		BeforeEach(func() {
			s, err := scheme.New(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			data = &goharborv1alpha2.Portal{}
			gvks, _, err := s.ObjectKinds(data)
			Expect(err).ToNot(HaveOccurred())

			gvk := gvks[0]
			data.SetGroupVersionKind(gvk)

			resource = &unstructured.Unstructured{}
		})

		JustBeforeEach(func() {
			data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(data)
			Expect(err).ToNot(HaveOccurred())

			resource.SetUnstructuredContent(data)
		})

		Context("With empty status", func() {
			BeforeEach(func() {
				data.Status = goharborv1alpha2.ComponentStatus{}
			})

			It("Should be ready", func() {
				ok, err := UnstructuredCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeTrue())
			})
		})

		Context("With observedGeneration mismatching generation", func() {
			BeforeEach(func() {
				data.SetGeneration(882)
				data.Status.ObservedGeneration = 1013
			})

			It("Should not be ready", func() {
				ok, err := UnstructuredCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With basic inprogress status", func() {
			Context("To True", func() {
				BeforeEach(func() {
					data.Status.Conditions = []goharborv1alpha2.Condition{
						{
							Type:   status.ConditionInProgress,
							Status: corev1.ConditionTrue,
						},
					}
				})

				It("Should not be ready", func() {
					ok, err := UnstructuredCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeFalse())
				})
			})

			Context("To False", func() {
				BeforeEach(func() {
					data.Status.Conditions = []goharborv1alpha2.Condition{
						{
							Type:   status.ConditionInProgress,
							Status: corev1.ConditionFalse,
						},
					}
				})

				It("Should be ready", func() {
					ok, err := UnstructuredCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeTrue())
				})
			})
		})
	})
})
