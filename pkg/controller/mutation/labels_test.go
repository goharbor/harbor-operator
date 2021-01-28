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

var _ = Describe("Mutate 1 label", func() {
	var labelMutation resources.Mutable
	var labelName string
	var labelValue string

	BeforeEach(func() {
		labelName, labelValue = "the-label", "the-value"
		labelMutation = GetLabelsMutation(labelName, labelValue)
	})

	Context("With a metav1 object", func() {
		var resource *corev1.Secret

		BeforeEach(func() {
			resource = &corev1.Secret{}
		})

		Context("Without labels", func() {
			BeforeEach(func() {
				resource.SetLabels(nil)
			})

			It("Should add the right label", func() {
				err := labelMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				labels := resource.GetLabels()
				Expect(labels).To(HaveKeyWithValue(labelName, labelValue))
			})
		})

		Context("With different labels", func() {
			var initialLabels map[string]string

			BeforeEach(func() {
				resource.SetLabels(map[string]string{
					"rap-label": "Booster",
					"rap-song":  "MAP",
				})
				initialLabels = resource.GetLabels()
			})

			It("Should merge all labels", func() {
				expectedLabels := initialLabels
				expectedLabels[labelName] = labelValue

				err := labelMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				labels := resource.GetLabels()
				Expect(labels).To(BeEquivalentTo(expectedLabels))
			})
		})

		Context("With the same label", func() {
			var initialLabels map[string]string

			BeforeEach(func() {
				resource.SetLabels(map[string]string{
					labelName: labelValue,
				})
				initialLabels = resource.GetLabels()
			})

			It("Should merge all labels", func() {
				expectedLabels := initialLabels
				expectedLabels[labelName] = labelValue

				err := labelMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				labels := resource.GetLabels()
				Expect(labels).To(BeEquivalentTo(expectedLabels))
			})
		})
	})
})

var _ = Describe("Mutate multiples label", func() {
	var getLabelMutation resources.Mutable
	var labelName1 string
	var labelValue1 string
	var labelName2 string
	var labelValue2 string

	BeforeEach(func() {
		labelName1, labelValue1 = "the-first-label", "the-first-value"
		labelName2, labelValue2 = "the-second-label", "the-second-value"
		getLabelMutation = GetLabelsMutation(labelName1, labelValue1, labelName2, labelValue2)
	})

	Context("With a metav1 object", func() {
		var resource *corev1.Secret

		BeforeEach(func() {
			resource = &corev1.Secret{}
		})

		Context("Without labels", func() {
			BeforeEach(func() {
				resource.SetLabels(nil)
			})

			It("Should add the right labels", func() {
				err := getLabelMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				labels := resource.GetLabels()
				Expect(labels).To(HaveKeyWithValue(labelName1, labelValue1))
				Expect(labels).To(HaveKeyWithValue(labelName2, labelValue2))
			})
		})

		Context("With different labels", func() {
			var initialLabels map[string]string

			BeforeEach(func() {
				resource.SetLabels(map[string]string{
					"rap-label": "Booster",
					"rap-song":  "MAP",
				})
				initialLabels = resource.GetLabels()
			})

			It("Should merge all labels", func() {
				expectedLabels := initialLabels
				expectedLabels[labelName1] = labelValue1
				expectedLabels[labelName2] = labelValue2

				err := getLabelMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				labels := resource.GetLabels()
				Expect(labels).To(BeEquivalentTo(expectedLabels))
			})
		})

		Context("With the same label", func() {
			var initialLabels map[string]string

			BeforeEach(func() {
				resource.SetLabels(map[string]string{
					labelName1: labelValue1,
					labelName2: labelValue2,
				})
				initialLabels = resource.GetLabels()
			})

			It("Should merge all labels", func() {
				expectedLabels := initialLabels
				expectedLabels[labelName1] = labelValue1
				expectedLabels[labelName2] = labelValue2

				err := getLabelMutation(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())

				labels := resource.GetLabels()
				Expect(labels).To(BeEquivalentTo(expectedLabels))
			})
		})
	})
})
