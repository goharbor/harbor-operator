package statuscheck

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	// +kubebuilder:scaffold:imports

	"github.com/goharbor/harbor-operator/pkg/scheme"
)

// These tests use Ginkgo (BDD-style Go testing framework). Rcfer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Check the status", func() {
	Context("Of a deployment resource", func() {
		var resource *appsv1.Deployment
		var data *appsv1.Deployment

		BeforeEach(func() {
			s, err := scheme.New(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			data = &appsv1.Deployment{}
			gvks, _, err := s.ObjectKinds(data)
			Expect(err).ToNot(HaveOccurred())

			gvk := gvks[0]
			data.SetGroupVersionKind(gvk)

			resource = data.DeepCopy()
		})

		JustBeforeEach(func() {
			data.DeepCopyInto(resource)
		})

		Context("With empty status", func() {
			BeforeEach(func() {
				data.Status = appsv1.DeploymentStatus{}
			})

			It("Should not be ready", func() {
				ok, err := BasicCheck(context.TODO(), resource)
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
				ok, err := BasicCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With missing replicas", func() {
			BeforeEach(func() {
				var replicasCount int32 = 3
				data.Spec.Replicas = &replicasCount
				data.Status.Replicas = 0
				data.Status.UpdatedReplicas = 0
				data.Status.AvailableReplicas = 0
				data.Status.ReadyReplicas = 0
			})

			It("Should not be ready", func() {
				ok, err := BasicCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With missing updated replicas", func() {
			BeforeEach(func() {
				var replicasCount int32 = 3
				data.Spec.Replicas = &replicasCount
				data.Status.Replicas = replicasCount
				data.Status.UpdatedReplicas = replicasCount - 1
				data.Status.AvailableReplicas = replicasCount
				data.Status.ReadyReplicas = replicasCount
			})

			It("Should not be ready", func() {
				ok, err := BasicCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With missing available replicas", func() {
			BeforeEach(func() {
				var replicasCount int32 = 3
				data.Spec.Replicas = &replicasCount
				data.Status.Replicas = replicasCount
				data.Status.UpdatedReplicas = replicasCount
				data.Status.AvailableReplicas = replicasCount - 1
				data.Status.ReadyReplicas = replicasCount - 1
			})

			It("Should not be ready", func() {
				ok, err := BasicCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With missing ready replicas", func() {
			BeforeEach(func() {
				var replicasCount int32 = 3
				data.Spec.Replicas = &replicasCount
				data.Status.Replicas = replicasCount
				data.Status.UpdatedReplicas = replicasCount
				data.Status.AvailableReplicas = replicasCount
				data.Status.ReadyReplicas = replicasCount - 1
			})

			It("Should not be ready", func() {
				ok, err := BasicCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With 2/3 replicas", func() {
			BeforeEach(func() {
				var replicasCount int32 = 3
				data.Spec.Replicas = &replicasCount
				data.Status.Replicas = replicasCount - 1
				data.Status.UpdatedReplicas = replicasCount - 1
				data.Status.AvailableReplicas = replicasCount - 1
				data.Status.ReadyReplicas = replicasCount - 1
			})

			It("Should not be ready", func() {
				ok, err := BasicCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With matching replicas count", func() {
			BeforeEach(func() {
				var replicasCount int32 = 3
				data.Spec.Replicas = &replicasCount
				data.Status.Replicas = replicasCount
				data.Status.UpdatedReplicas = replicasCount
				data.Status.AvailableReplicas = replicasCount
				data.Status.ReadyReplicas = replicasCount
			})

			Context("With available condition", func() {
				Context("To True", func() {
					BeforeEach(func() {
						data.Status.Conditions = append(data.Status.Conditions, appsv1.DeploymentCondition{
							Type:   appsv1.DeploymentAvailable,
							Status: corev1.ConditionTrue,
						})
					})

					Context("With processing condition", func() {
						Context("To False", func() {
							BeforeEach(func() {
								data.Status.Conditions = append(data.Status.Conditions, appsv1.DeploymentCondition{
									Type:   appsv1.DeploymentProgressing,
									Status: corev1.ConditionFalse,
								})
							})

							It("Should be ready", func() {
								ok, err := BasicCheck(context.TODO(), resource)
								Expect(err).ToNot(HaveOccurred())
								Expect(ok).To(BeTrue())
							})
						})

						Context("To True", func() {
							BeforeEach(func() {
								data.Status.Conditions = []appsv1.DeploymentCondition{
									{
										Type:   appsv1.DeploymentProgressing,
										Status: corev1.ConditionTrue,
									},
								}
							})

							It("Should not be ready", func() {
								ok, err := BasicCheck(context.TODO(), resource)
								Expect(err).ToNot(HaveOccurred())
								Expect(ok).To(BeFalse())
							})
						})
					})
				})

				Context("To False", func() {
					BeforeEach(func() {
						data.Status.Conditions = append(data.Status.Conditions, appsv1.DeploymentCondition{
							Type:   appsv1.DeploymentAvailable,
							Status: corev1.ConditionFalse,
						})
					})

					Context("With processing condition", func() {
						Context("To False", func() {
							BeforeEach(func() {
								data.Status.Conditions = append(data.Status.Conditions, appsv1.DeploymentCondition{
									Type:   appsv1.DeploymentProgressing,
									Status: corev1.ConditionFalse,
								})
							})

							It("Should not be ready", func() {
								ok, err := BasicCheck(context.TODO(), resource)
								Expect(err).ToNot(HaveOccurred())
								Expect(ok).To(BeFalse())
							})
						})

						Context("To True", func() {
							BeforeEach(func() {
								data.Status.Conditions = []appsv1.DeploymentCondition{
									{
										Type:   appsv1.DeploymentProgressing,
										Status: corev1.ConditionTrue,
									},
								}
							})

							It("Should not be ready", func() {
								ok, err := BasicCheck(context.TODO(), resource)
								Expect(err).ToNot(HaveOccurred())
								Expect(ok).To(BeFalse())
							})
						})
					})
				})
			})
		})
	})

	Context("Of a pod resource", func() {
		var resource *corev1.Pod
		var data *corev1.Pod

		BeforeEach(func() {
			s, err := scheme.New(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			data = &corev1.Pod{}
			gvks, _, err := s.ObjectKinds(data)
			Expect(err).ToNot(HaveOccurred())

			gvk := gvks[0]
			data.SetGroupVersionKind(gvk)

			resource = data.DeepCopy()
		})

		JustBeforeEach(func() {
			data.DeepCopyInto(resource)
		})

		Context("With empty status", func() {
			BeforeEach(func() {
				data.Status = corev1.PodStatus{}
			})

			It("Should not be ready", func() {
				ok, err := BasicCheck(context.TODO(), resource)
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
					ok, err := BasicCheck(context.TODO(), resource)
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
					ok, err := BasicCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeTrue())
				})
			})
		})
	})
})
