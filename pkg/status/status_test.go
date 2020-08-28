package common_test

import (
	"context"

	. "github.com/goharbor/harbor-operator/pkg/status"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// +kubebuilder:scaffold:imports
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("A resource", func() {
	const (
		reason         = "the-reason"
		message        = "A human readable message"
		extraParameter = "extra-parameter"
	)

	var ctx context.Context
	var conditions []interface{}

	BeforeEach(func() {
		ctx = context.TODO()
	})

	Context("With no conditions", func() {
		JustBeforeEach(func() {
			conditions = nil
		})

		Describe("Update Ready condition to False", func() {
			var conditionType status.ConditionType
			var conditionValue corev1.ConditionStatus

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
				conditionValue = corev1.ConditionFalse
			})

			It("Should be added", func() {
				conditions, err := UpdateCondition(ctx, conditions, conditionType, conditionValue)
				Expect(err).ToNot(HaveOccurred(), "failed to update condition")

				Expect(conditions).To(ContainElement(SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
					HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
				)))
			})

			Context("With a reason", func() {
				It("Should be added", func() {
					conditions, err := UpdateCondition(ctx, conditions, conditionType, conditionValue, reason)
					Expect(err).ToNot(HaveOccurred(), "failed to update condition")

					Expect(conditions).To(ContainElement(SatisfyAll(
						HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
						HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
						HaveKeyWithValue("reason", BeEquivalentTo(reason)),
					)))
				})

				Context("With a message", func() {
					It("Should be added", func() {
						conditions, err := UpdateCondition(ctx, conditions, conditionType, conditionValue, reason, message)
						Expect(err).ToNot(HaveOccurred(), "failed to update condition")

						Expect(conditions).To(ContainElement(SatisfyAll(
							HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
							HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
							HaveKeyWithValue("reason", BeEquivalentTo(reason)),
							HaveKeyWithValue("message", BeEquivalentTo(message)),
						)))
					})

					Context("With an extra parameter", func() {
						It("Should return an error", func() {
							_, err := UpdateCondition(ctx, conditions, conditionType, conditionValue, reason, message, extraParameter)
							Expect(err).To(HaveOccurred())
						})
					})
				})
			})
		})

		Describe("Get Ready condition", func() {
			var conditionType status.ConditionType

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
			})

			It("Should return unknown status", func() {
				result, err := GetCondition(ctx, conditions, conditionType)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
					HaveKeyWithValue("status", BeEquivalentTo(corev1.ConditionUnknown)),
				))
			})
		})

		Describe("Get Ready status", func() {
			It("Should return unknown", func() {
				result, err := GetConditionStatus(ctx, conditions, status.ConditionInProgress)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(BeEquivalentTo(corev1.ConditionUnknown))
			})
		})
	})

	Context("With Failed condition to True", func() {
		var condition types.GomegaMatcher

		JustBeforeEach(func() {
			conditions = []interface{}{}

			var err error
			conditions, err = UpdateCondition(ctx, conditions, status.ConditionFailed, corev1.ConditionTrue, "")
			Expect(err).ToNot(HaveOccurred())

			condition = SatisfyAll(
				HaveKeyWithValue("type", BeEquivalentTo(status.ConditionFailed)),
				HaveKeyWithValue("status", BeEquivalentTo(corev1.ConditionTrue)),
			)
		})

		Describe("Update Ready condition to False", func() {
			var conditionType status.ConditionType
			var conditionValue corev1.ConditionStatus

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
				conditionValue = corev1.ConditionFalse
			})

			It("Should be added", func() {
				conditions, err := UpdateCondition(ctx, conditions, conditionType, conditionValue)
				Expect(err).ToNot(HaveOccurred())

				Expect(conditions).To(ConsistOf(condition, SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
					HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
				)))
			})

			Context("With a reason", func() {
				It("Should be added", func() {
					conditions, err := UpdateCondition(ctx, conditions, conditionType, conditionValue, reason)
					Expect(err).ToNot(HaveOccurred())

					Expect(conditions).To(ConsistOf(condition, SatisfyAll(
						HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
						HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
						HaveKeyWithValue("reason", BeEquivalentTo(reason)),
					)))
				})

				Context("With a message", func() {
					It("Should be added", func() {
						conditions, err := UpdateCondition(ctx, conditions, conditionType, conditionValue, reason, message)
						Expect(err).ToNot(HaveOccurred())

						Expect(conditions).To(ConsistOf(condition, SatisfyAll(
							HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
							HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
							HaveKeyWithValue("reason", BeEquivalentTo(reason)),
							HaveKeyWithValue("message", BeEquivalentTo(message)),
						)))
					})

					Context("With an extra parameter", func() {
						var extraParameter string

						JustBeforeEach(func() {
							extraParameter = "extra-parameter"
						})

						It("Should return an error", func() {
							_, err := UpdateCondition(ctx, conditions, conditionType, conditionValue, reason, message, extraParameter)
							Expect(err).To(HaveOccurred())
						})
					})
				})
			})
		})

		Describe("Get Ready condition", func() {
			var conditionType status.ConditionType

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
			})

			It("Should return unknown", func() {
				result, err := GetCondition(ctx, conditions, conditionType)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
					HaveKeyWithValue("status", BeEquivalentTo(corev1.ConditionUnknown)),
				))
			})
		})

		Describe("Get Ready status", func() {
			It("Should return unknown", func() {
				result, err := GetConditionStatus(ctx, conditions, status.ConditionInProgress)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(BeEquivalentTo(corev1.ConditionUnknown))
			})
		})
	})

	Context("With InProgress condition to True", func() {
		JustBeforeEach(func() {
			conditions = []interface{}{}

			var err error
			conditions, err = UpdateCondition(ctx, conditions, status.ConditionInProgress, corev1.ConditionTrue, "")
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("Update Ready condition to False", func() {
			var conditionType status.ConditionType
			var conditionValue corev1.ConditionStatus

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
				conditionValue = corev1.ConditionFalse
			})

			It("Should update the value", func() {
				conditions, err := UpdateCondition(ctx, conditions, conditionType, conditionValue)
				Expect(err).ToNot(HaveOccurred())

				Expect(conditions).To(ContainElement(SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
					HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
				)))
			})

			Context("With a reason", func() {
				It("Should update the status", func() {
					conditions, err := UpdateCondition(ctx, conditions, status.ConditionInProgress, corev1.ConditionFalse, "the-reason")
					Expect(err).ToNot(HaveOccurred())

					Expect(conditions).To(ContainElement(SatisfyAll(
						HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
						HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
						HaveKeyWithValue("reason", BeEquivalentTo(reason)),
					)))
				})

				Context("With a message", func() {
					It("Should update the status", func() {
						conditions, err := UpdateCondition(ctx, conditions, status.ConditionInProgress, corev1.ConditionFalse, "the-reason", "A human readable message")
						Expect(err).ToNot(HaveOccurred())

						Expect(conditions).To(ContainElement(SatisfyAll(
							HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
							HaveKeyWithValue("status", BeEquivalentTo(conditionValue)),
							HaveKeyWithValue("reason", BeEquivalentTo(reason)),
							HaveKeyWithValue("message", BeEquivalentTo(message)),
						)))
					})

					Context("With an extra parameter", func() {
						It("Should return an error", func() {
							_, err := UpdateCondition(ctx, conditions, conditionType, conditionValue, reason, message, extraParameter)
							Expect(err).To(HaveOccurred())
						})
					})
				})
			})
		})

		Describe("Get InProgress condition", func() {
			var conditionType status.ConditionType

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
			})

			It("Should return the condition", func() {
				result, err := GetCondition(ctx, conditions, conditionType)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(status.ConditionInProgress)),
					HaveKeyWithValue("status", BeEquivalentTo(corev1.ConditionTrue)),
				))
			})
		})

		Describe("Get InProgress condition", func() {
			var conditionType status.ConditionType

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
			})

			It("Should return the status", func() {
				result, err := GetConditionStatus(ctx, conditions, conditionType)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(BeEquivalentTo(corev1.ConditionTrue))
			})
		})
	})

	Context("With multiple conditions", func() {
		JustBeforeEach(func() {
			conditions = []interface{}{}

			var err error

			conditions, err = UpdateCondition(ctx, conditions, status.ConditionInProgress, corev1.ConditionTrue, "")
			Expect(err).ToNot(HaveOccurred())

			conditions, err = UpdateCondition(ctx, conditions, status.ConditionFailed, corev1.ConditionFalse, "")
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("Get InProgress condition", func() {
			var conditionType status.ConditionType

			JustBeforeEach(func() {
				conditionType = status.ConditionInProgress
			})

			It("Should return the right status", func() {
				result, err := GetCondition(ctx, conditions, conditionType)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
					HaveKeyWithValue("status", BeEquivalentTo(corev1.ConditionTrue)),
				))
			})
		})

		Describe("Get Failed condition", func() {
			var conditionType status.ConditionType

			JustBeforeEach(func() {
				conditionType = status.ConditionFailed
			})

			It("Should return the right status", func() {
				result, err := GetCondition(ctx, conditions, conditionType)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(SatisfyAll(
					HaveKeyWithValue("type", BeEquivalentTo(conditionType)),
					HaveKeyWithValue("status", BeEquivalentTo(corev1.ConditionFalse)),
				))
			})
		})
	})
})
