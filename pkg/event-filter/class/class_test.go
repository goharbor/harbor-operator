package class

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"

	// +kubebuilder:scaffold:imports

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
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
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
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
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
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
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: &goharborv1alpha2.Harbor{}})
								Expect(ok).To(BeFalse())
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
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("with other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
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
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
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
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("Harbor with other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Meta: h.GetObjectMeta(), Object: h})
						Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
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
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1alpha2.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1alpha2.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{MetaOld: oldResource.GetObjectMeta(), ObjectOld: oldResource, MetaNew: newResource.GetObjectMeta(), ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})
			})
		})
	})
})
