package class_test

import (
	"context"
	"fmt"

	. "github.com/goharbor/harbor-operator/pkg/event-filter/class"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// These tests use Ginkgo (BDD-style Go testing framework). Rcfer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("class-filter", func() {
	Context("With no harbor-class", func() {
		var cf *Filter

		BeforeEach(func() {
			cf, _ = setupTest(context.TODO())
		})

		Describe("Creation event", func() {
			Context("For an Harbor resource", func() {
				type Object interface {
					runtime.Object
					metav1.ObjectMetaAccessor
					metav1.Object
				}
				var h Object

				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						annotationKey := "not-an-ingress-class-annotation"
						Ω(annotationKey).
							ShouldNot(BeEquivalentTo(goharborv1alpha2.HarborClassAnnotation))

						h.SetAnnotations(map[string]string{
							annotationKey: "a class?",
						})
					})

					It("Should match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With other-class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})
			})
		})

		Describe("Deletion event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1alpha2.Harbor

				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						annotationKey := "not-an-ingress-class-annotation"
						Ω(annotationKey).
							ShouldNot(BeEquivalentTo(goharborv1alpha2.HarborClassAnnotation))

						h.SetAnnotations(map[string]string{
							annotationKey: "a class?",
						})
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With other-class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("With the right class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})
			})
		})

		Describe("Generic event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1alpha2.Harbor

				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						annotationKey := "not-an-ingress-class-annotation"
						Ω(annotationKey).
							ShouldNot(BeEquivalentTo(goharborv1alpha2.HarborClassAnnotation))

						h.SetAnnotations(map[string]string{
							annotationKey: "a class?",
						})
					})

					It("Should match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})

				Context("With other-class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("With the right class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})
			})
		})

		Describe("Update event", func() {
			Context("From an Harbor resource", func() {
				var oldResource *goharborv1alpha2.Harbor

				BeforeEach(func() {
					oldResource = &goharborv1alpha2.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(nil)
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})

						Context("To resource with other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("To resource with other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("To resource with empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("To resource with other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})
					})
				})

				Context("With other-class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})).
									Should(BeFalse())
							})
						})
					})
				})
			})
		})
	})

	Context("With a specified harbor-class", func() {
		var cf *Filter

		BeforeEach(func() {
			cf, _ = setupTest(context.TODO())
			cf.ClassName = "harbor-class-name"
		})

		Describe("Creation event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1alpha2.Harbor

				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("with other annotations", func() {
					BeforeEach(func() {
						annotationKey := "not-an-ingress-class-annotation"
						Ω(annotationKey).
							ShouldNot(BeEquivalentTo(goharborv1alpha2.HarborClassAnnotation))

						h.SetAnnotations(map[string]string{
							annotationKey: "a class?",
						})
					})

					It("Should match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with other-class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with the right class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						Ω(cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})
			})
		})

		Describe("Deletion event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1alpha2.Harbor

				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						annotationKey := "not-an-ingress-class-annotation"
						Ω(annotationKey).
							ShouldNot(BeEquivalentTo(goharborv1alpha2.HarborClassAnnotation))

						h.SetAnnotations(map[string]string{
							annotationKey: "a class?",
						})
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with other-class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with the right class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						Ω(cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})
			})
		})

		Describe("Generic event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1alpha2.Harbor

				BeforeEach(func() {
					h = &goharborv1alpha2.Harbor{}
				})

				Context("Harbor with no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should not match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("Harbor with other annotations", func() {
					BeforeEach(func() {
						annotationKey := "not-an-ingress-class-annotation"
						Ω(annotationKey).
							ShouldNot(BeEquivalentTo(goharborv1alpha2.HarborClassAnnotation))

						h.SetAnnotations(map[string]string{
							annotationKey: "a class?",
						})
					})

					It("Should not match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should not match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with other-class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeFalse())
					})
				})

				Context("resource with the right class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						Ω(cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})).
							Should(BeTrue())
					})
				})
			})
		})

		Describe("Update event", func() {
			Context("From an Harbor resource", func() {
				var oldResource *goharborv1alpha2.Harbor

				BeforeEach(func() {
					oldResource = &goharborv1alpha2.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(nil)
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})
					})
				})

				Context("With other-class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should not match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})
					})
				})

				Context("With the right class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1alpha2.Harbor

						BeforeEach(func() {
							newResource = &goharborv1alpha2.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								className := "other-class"
								Ω(className).ShouldNot(BeEquivalentTo(cf.ClassName))

								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: className,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								Ω(cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})).
									Should(BeTrue())
							})
						})
					})
				})
			})
		})
	})
})
