package common

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"

	// +kubebuilder:scaffold:imports

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("status", func() {
	const (
		reason         = "the-reason"
		message        = "A human readable message"
		extraParameter = "extra-parameter"
	)

	var r *Controller
	var ctx context.Context

	BeforeEach(func() {
		r, ctx = setupTest(context.TODO())
	})

	Describe("An Harbor resource", func() {
		var h *goharborv1alpha2.Harbor

		BeforeEach(func() {
			h = &goharborv1alpha2.Harbor{}
		})

		Context("With no conditions", func() {
			JustBeforeEach(func() {
				h.Status.Conditions = nil
			})

			Describe("Update Ready condition to False", func() {
				var conditionType status.ConditionType
				var conditionValue corev1.ConditionStatus

				JustBeforeEach(func() {
					conditionType = status.ConditionInProgress
					conditionValue = corev1.ConditionFalse
				})

				It("Should be added", func() {
					err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue)
					Expect(err).ToNot(HaveOccurred(), "failed to update condition")

					Expect(h.Status.Conditions).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Type":   BeEquivalentTo(conditionType),
						"Status": BeEquivalentTo(conditionValue),
					})))
				})

				Context("With a reason", func() {
					It("Should be added", func() {
						err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue, reason)
						Expect(err).ToNot(HaveOccurred(), "failed to update condition")

						Expect(h.Status.Conditions).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
							"Type":   BeEquivalentTo(conditionType),
							"Status": BeEquivalentTo(conditionValue),
							"Reason": BeEquivalentTo(reason),
						})))
					})

					Context("With a message", func() {
						It("Should be added", func() {
							err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue, reason, message)
							Expect(err).ToNot(HaveOccurred(), "failed to update condition")

							Expect(h.Status.Conditions).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
								"Type":    BeEquivalentTo(conditionType),
								"Status":  BeEquivalentTo(conditionValue),
								"Reason":  BeEquivalentTo(reason),
								"Message": BeEquivalentTo(message),
							})))
						})

						Context("With an extra parameter", func() {
							It("Should return an error", func() {
								err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue, reason, message, extraParameter)
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
					result := r.GetCondition(ctx, &h.Status, conditionType)
					Expect(result).To(BeEquivalentTo(status.Condition{
						Type:   conditionType,
						Status: corev1.ConditionUnknown,
					}))
				})
			})

			Describe("Get Ready status", func() {
				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				It("Should return unknown", func() {
					result := r.GetConditionStatus(ctx, &h.Status, status.ConditionInProgress)
					Expect(result).To(BeEquivalentTo(corev1.ConditionUnknown))
				})
			})
		})

		Context("With Applied condition to True", func() {
			var condition goharborv1alpha2.Condition

			JustBeforeEach(func() {
				condition = goharborv1alpha2.Condition{
					Type:   status.ConditionFailed,
					Reason: "",
					Status: corev1.ConditionTrue,
				}
				h.Status.Conditions = []goharborv1alpha2.Condition{condition}
			})

			Describe("Update Ready condition to False", func() {
				var conditionType status.ConditionType
				var conditionValue corev1.ConditionStatus

				JustBeforeEach(func() {
					conditionType = status.ConditionInProgress
					conditionValue = corev1.ConditionFalse
				})

				It("Should be added", func() {
					err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue)
					Expect(err).ToNot(HaveOccurred())

					Expect(h.Status.Conditions).To(ConsistOf(condition, gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Type":   BeEquivalentTo(conditionType),
						"Status": BeEquivalentTo(conditionValue),
					})))
				})

				Context("With a reason", func() {
					It("Should be added", func() {
						err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue, reason)
						Expect(err).ToNot(HaveOccurred())

						Expect(h.Status.Conditions).To(ConsistOf(condition, gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
							"Type":   BeEquivalentTo(conditionType),
							"Status": BeEquivalentTo(conditionValue),
							"Reason": BeEquivalentTo(reason),
						})))
					})

					Context("With a message", func() {
						It("Should be added", func() {
							err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue, reason, message)
							Expect(err).ToNot(HaveOccurred())

							Expect(h.Status.Conditions).To(ConsistOf(condition, gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
								"Type":    BeEquivalentTo(status.ConditionInProgress),
								"Status":  BeEquivalentTo(corev1.ConditionFalse),
								"Reason":  BeEquivalentTo(reason),
								"Message": BeEquivalentTo(message),
							})))
						})

						Context("With an extra parameter", func() {
							var extraParameter string

							JustBeforeEach(func() {
								extraParameter = "extra-parameter"
							})

							It("Should return an error", func() {
								err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue, reason, message, extraParameter)
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
					result := r.GetCondition(ctx, &h.Status, conditionType)
					Expect(result).To(BeEquivalentTo(status.Condition{
						Type:   conditionType,
						Status: corev1.ConditionUnknown,
					}))
				})
			})

			Describe("Get Ready status", func() {
				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				It("Should return unknown", func() {
					result := r.GetConditionStatus(ctx, &h.Status, status.ConditionInProgress)
					Expect(result).To(BeEquivalentTo(corev1.ConditionUnknown))
				})
			})
		})

		Context("With Ready condition to True", func() {
			var condition goharborv1alpha2.Condition

			JustBeforeEach(func() {
				condition = goharborv1alpha2.Condition{
					Type:   status.ConditionInProgress,
					Reason: "",
					Status: corev1.ConditionTrue,
				}
				h.Status.Conditions = []goharborv1alpha2.Condition{condition}
			})

			Describe("Update Ready condition to False", func() {
				var conditionType status.ConditionType
				var conditionValue corev1.ConditionStatus

				JustBeforeEach(func() {
					conditionType = status.ConditionInProgress
					conditionValue = corev1.ConditionFalse
				})

				It("Should update the value", func() {
					err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue)
					Expect(err).ToNot(HaveOccurred())

					Expect(h.Status.Conditions).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Type":   BeEquivalentTo(status.ConditionInProgress),
						"Status": BeEquivalentTo(corev1.ConditionFalse),
					})))
				})

				Context("With a reason", func() {
					It("Should update the status", func() {
						err := r.UpdateCondition(ctx, &h.Status, status.ConditionInProgress, corev1.ConditionFalse, "the-reason")
						Expect(err).ToNot(HaveOccurred())

						Expect(h.Status.Conditions).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
							"Type":   BeEquivalentTo(status.ConditionInProgress),
							"Status": BeEquivalentTo(corev1.ConditionFalse),
							"Reason": BeEquivalentTo(reason),
						})))
					})

					Context("With a message", func() {
						It("Should update the status", func() {
							err := r.UpdateCondition(ctx, &h.Status, status.ConditionInProgress, corev1.ConditionFalse, "the-reason", "A human readable message")
							Expect(err).ToNot(HaveOccurred())

							Expect(h.Status.Conditions).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
								"Type":    BeEquivalentTo(status.ConditionInProgress),
								"Status":  BeEquivalentTo(corev1.ConditionFalse),
								"Reason":  BeEquivalentTo(reason),
								"Message": BeEquivalentTo(message),
							})))
						})

						Context("With an extra parameter", func() {
							It("Should return an error", func() {
								err := r.UpdateCondition(ctx, &h.Status, conditionType, conditionValue, reason, message, extraParameter)
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

				It("Should return the condition", func() {
					result := r.GetCondition(ctx, &h.Status, conditionType)
					Expect(result).To(BeEquivalentTo(condition))
				})
			})

			Describe("Get Ready condition", func() {
				var conditionType status.ConditionType

				JustBeforeEach(func() {
					conditionType = status.ConditionInProgress
				})

				It("Should return the status", func() {
					result := r.GetConditionStatus(ctx, &h.Status, conditionType)
					Expect(result).To(BeEquivalentTo(condition.Status))
				})
			})
		})

		Context("With multiple conditions", func() {
			var readyCondition goharborv1alpha2.Condition

			JustBeforeEach(func() {
				readyCondition = goharborv1alpha2.Condition{
					Type:   status.ConditionInProgress,
					Reason: "",
					Status: corev1.ConditionTrue,
				}
				appliedCondition := goharborv1alpha2.Condition{
					Type:   status.ConditionFailed,
					Reason: "",
					Status: corev1.ConditionTrue,
				}
				h.Status.Conditions = []goharborv1alpha2.Condition{readyCondition, appliedCondition}
			})

			Describe("Get Ready condition", func() {
				var conditionType status.ConditionType

				JustBeforeEach(func() {
					conditionType = status.ConditionInProgress
				})

				It("Should return the right status", func() {
					result := r.GetCondition(ctx, &h.Status, conditionType)
					Expect(result).To(BeEquivalentTo(readyCondition))
				})
			})
		})
	})
})
