package statuscheck_test

import (
	"context"
	"fmt"

	. "github.com/goharbor/harbor-operator/pkg/resources/statuscheck"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kstatus/status"
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

		AfterEach(func() {
			if !CurrentGinkgoTestDescription().Failed {
				return
			}

			if data == nil {
				return
			}

			fmt.Fprintf(GinkgoWriter, "%+v", data.Status)
		})

		JustBeforeEach(func() {
			data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(data)
			Expect(err).ToNot(HaveOccurred())

			resource.SetUnstructuredContent(data)
		})

		JustAfterEach(func() {
			if resource != nil {
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.UnstructuredContent(), data)
				Expect(err).ToNot(HaveOccurred())

				resource = nil
			}
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
					data.Status.Conditions = []corev1.PodCondition{{
						Type:   corev1.PodReady,
						Status: corev1.ConditionFalse,
					}}
				})

				It("Should not be ready", func() {
					ok, err := UnstructuredCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeFalse())
				})
			})

			Context("To True", func() {
				BeforeEach(func() {
					data.Status.Conditions = []corev1.PodCondition{{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					}}
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
		var data *goharborv1.Portal

		BeforeEach(func() {
			s, err := scheme.New(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			data = &goharborv1.Portal{}
			gvks, _, err := s.ObjectKinds(data)
			Expect(err).ToNot(HaveOccurred())

			gvk := gvks[0]
			data.SetGroupVersionKind(gvk)

			resource = &unstructured.Unstructured{}
		})

		AfterEach(func() {
			if !CurrentGinkgoTestDescription().Failed {
				return
			}

			if data == nil {
				return
			}

			fmt.Fprintf(GinkgoWriter, "%+v", data.Status)
		})

		JustBeforeEach(func() {
			data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(data)
			Expect(err).ToNot(HaveOccurred())

			resource.SetUnstructuredContent(data)
		})

		JustAfterEach(func() {
			if resource != nil {
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.UnstructuredContent(), data)
				Expect(err).ToNot(HaveOccurred())

				resource = nil
			}
		})

		Context("With empty status", func() {
			BeforeEach(func() {
				data.Status = harbormetav1.ComponentStatus{}
			})

			It("Should not be ready", func() {
				ok, err := UnstructuredCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
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

		Context("With Observed Generation up to date", func() {
			JustBeforeEach(func() {
				data.Status.ObservedGeneration = data.GetGeneration()
			})

			Context("With basic inprogress status", func() {
				Context("To True", func() {
					BeforeEach(func() {
						data.Status.Conditions = []harbormetav1.Condition{{
							Type:   status.ConditionInProgress,
							Status: corev1.ConditionTrue,
						}}
					})

					It("Should not be ready", func() {
						ok, err := UnstructuredCheck(context.TODO(), resource)
						Expect(err).ToNot(HaveOccurred())
						Expect(ok).To(BeFalse())
					})
				})

				Context("To False", func() {
					BeforeEach(func() {
						data.Status.Conditions = []harbormetav1.Condition{{
							Type:   status.ConditionInProgress,
							Status: corev1.ConditionFalse,
						}}
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
})
