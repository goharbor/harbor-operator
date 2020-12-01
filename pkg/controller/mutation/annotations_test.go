package mutation_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/controller/mutation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/goharbor/harbor-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Rcfer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Mutate 1 annotation", func() {
	var getAnnotationMutation resources.Mutable
	var annotationName string
	var annotationValue string

	BeforeEach(func() {
		annotationName, annotationValue = "the-annotation", "the-value"
		getAnnotationMutation = GetAnnotationsMutation(annotationName, annotationValue)
	})

	Context("With a metav1 object", func() {
		var resource *corev1.Secret

		BeforeEach(func() {
			resource = &corev1.Secret{}
		})

		Context("Without annotations", func() {
			BeforeEach(func() {
				resource.SetAnnotations(nil)
			})

			It("Should add the right annotation", func() {
				err := getAnnotationMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				annotations := resource.GetAnnotations()
				Expect(annotations).To(HaveKeyWithValue(annotationName, annotationValue))
			})
		})

		Context("With different annotations", func() {
			var initialAnnotations map[string]string

			BeforeEach(func() {
				resource.SetAnnotations(map[string]string{
					"rap-annotation": "Booster",
					"rap-song":       "MAP",
				})
				initialAnnotations = resource.GetAnnotations()
			})

			It("Should merge all annotations", func() {
				expectedAnnotations := initialAnnotations
				expectedAnnotations[annotationName] = annotationValue

				err := getAnnotationMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				annotations := resource.GetAnnotations()
				Expect(annotations).To(BeEquivalentTo(expectedAnnotations))
			})
		})

		Context("With the same annotation", func() {
			var initialAnnotations map[string]string

			BeforeEach(func() {
				resource.SetAnnotations(map[string]string{
					annotationName: annotationValue,
				})
				initialAnnotations = resource.GetAnnotations()
			})

			It("Should merge all annotations", func() {
				expectedAnnotations := initialAnnotations
				expectedAnnotations[annotationName] = annotationValue

				err := getAnnotationMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				annotations := resource.GetAnnotations()
				Expect(annotations).To(BeEquivalentTo(expectedAnnotations))
			})
		})
	})
})

var _ = Describe("Mutate multiples annotation", func() {
	var getAnnotationMutation resources.Mutable
	var annotationName1 string
	var annotationValue1 string
	var annotationName2 string
	var annotationValue2 string

	BeforeEach(func() {
		annotationName1, annotationValue1 = "the-first-annotation", "the-first-value"
		annotationName2, annotationValue2 = "the-second-annotation", "the-second-value"
		getAnnotationMutation = GetAnnotationsMutation(annotationName1, annotationValue1, annotationName2, annotationValue2)
	})

	Context("With a metav1 object", func() {
		var resource *corev1.Secret

		BeforeEach(func() {
			resource = &corev1.Secret{}
		})

		Context("Without annotations", func() {
			BeforeEach(func() {
				resource.SetAnnotations(nil)
			})

			It("Should add the right annotations", func() {
				err := getAnnotationMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				annotations := resource.GetAnnotations()
				Expect(annotations).To(HaveKeyWithValue(annotationName1, annotationValue1))
				Expect(annotations).To(HaveKeyWithValue(annotationName2, annotationValue2))
			})
		})

		Context("With different annotations", func() {
			var initialAnnotations map[string]string

			BeforeEach(func() {
				resource.SetAnnotations(map[string]string{
					"anno-1800-editor": "Ubisoft",
					"anno-1800-player": "Zerator",
				})
				initialAnnotations = resource.GetAnnotations()
			})

			It("Should merge all annotations", func() {
				expectedAnnotations := initialAnnotations
				expectedAnnotations[annotationName1] = annotationValue1
				expectedAnnotations[annotationName2] = annotationValue2

				err := getAnnotationMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				annotations := resource.GetAnnotations()
				Expect(annotations).To(BeEquivalentTo(expectedAnnotations))
			})
		})

		Context("With the same annotation", func() {
			var initialAnnotations map[string]string

			BeforeEach(func() {
				resource.SetAnnotations(map[string]string{
					annotationName1: annotationValue1,
					annotationName2: annotationValue2,
				})
				initialAnnotations = resource.GetAnnotations()
			})

			It("Should merge all annotations", func() {
				expectedAnnotations := initialAnnotations
				expectedAnnotations[annotationName1] = annotationValue1
				expectedAnnotations[annotationName2] = annotationValue2

				err := getAnnotationMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				annotations := resource.GetAnnotations()
				Expect(annotations).To(BeEquivalentTo(expectedAnnotations))
			})
		})
	})
})
