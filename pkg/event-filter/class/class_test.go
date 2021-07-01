package class_test

import (
	"context"
	"fmt"

	. "github.com/goharbor/harbor-operator/pkg/event-filter/class"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
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
					h = &goharborv1.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})
			})
		})

		Describe("Deletion event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1.Harbor

				BeforeEach(func() {
					h = &goharborv1.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})
			})
		})

		Describe("Generic event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1.Harbor

				BeforeEach(func() {
					h = &goharborv1.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})

				Context("With other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})
			})
		})

		Describe("Update event", func() {
			Context("From an Harbor resource", func() {
				var oldResource *goharborv1.Harbor

				BeforeEach(func() {
					oldResource = &goharborv1.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(nil)
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("To resource with other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})

				Context("With other-class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
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
				var h *goharborv1.Harbor

				BeforeEach(func() {
					h = &goharborv1.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("with other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Create(event.CreateEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})
			})
		})

		Describe("Deletion event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1.Harbor

				BeforeEach(func() {
					h = &goharborv1.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Delete(event.DeleteEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})
			})
		})

		Describe("Generic event", func() {
			Context("For an Harbor resource", func() {
				var h *goharborv1.Harbor

				BeforeEach(func() {
					h = &goharborv1.Harbor{}
				})

				Context("Harbor with no annotation", func() {
					BeforeEach(func() {
						h.SetAnnotations(nil)
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("Harbor with other annotations", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with empty class", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with other-class should not match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					It("Should not match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeFalse())
					})
				})

				Context("resource with the right class should match", func() {
					BeforeEach(func() {
						h.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: cf.ClassName,
						})
					})

					It("Should match", func() {
						ok := cf.Generic(event.GenericEvent{Object: h})
						Expect(ok).To(BeTrue())
					})
				})
			})
		})

		Describe("Update event", func() {
			Context("From an Harbor resource", func() {
				var oldResource *goharborv1.Harbor

				BeforeEach(func() {
					oldResource = &goharborv1.Harbor{}
				})

				Context("With no annotation", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(nil)
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})

				Context("With other annotations", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})

				Context("With empty class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})

				Context("With other-class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: "other-class",
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should not match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeFalse())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})

				Context("With the right class", func() {
					BeforeEach(func() {
						oldResource.SetAnnotations(map[string]string{
							goharborv1.HarborClassAnnotation: cf.ClassName,
						})
					})

					Context("To resource", func() {
						var newResource *goharborv1.Harbor

						BeforeEach(func() {
							newResource = &goharborv1.Harbor{}
						})

						Context("With no annotation", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(nil)
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other annotations", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									fmt.Sprintf("%s-false", goharborv1.HarborClassAnnotation): "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With empty class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With other-class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: "other-class",
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})

						Context("With the right class", func() {
							BeforeEach(func() {
								newResource.SetAnnotations(map[string]string{
									goharborv1.HarborClassAnnotation: cf.ClassName,
								})
							})

							It("Should match", func() {
								ok := cf.Update(event.UpdateEvent{ObjectOld: oldResource, ObjectNew: newResource})
								Expect(ok).To(BeTrue())
							})
						})
					})
				})
			})
		})
	})
})
