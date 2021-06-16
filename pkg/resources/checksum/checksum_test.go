package checksum_test

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	"github.com/goharbor/harbor-operator/pkg/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Checksum", func() {
	var (
		ctx        context.Context
		depManager *checksum.Dependencies

		d1       *appsv1.Deployment
		owner    *appsv1.Deployment
		resource *appsv1.Deployment

		addDependencies func()
	)

	BeforeEach(func() {
		ctx = logger.Context(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

		scheme, _ := scheme.New(ctx)

		depManager = checksum.New(scheme)

		d1 = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "d1",
				Namespace: "namespace",
			},
		}

		owner = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "owner",
				Namespace: "namespace",
			},
		}

		resource = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "resource",
				Namespace: "namespace",
			},
		}

		addDependencies = func() {
			depManager.Add(ctx, d1, false)
			depManager.Add(ctx, owner, true)
		}
	})

	Describe("ChangedFor without dependencies", func() {
		It("Should pass", func() {
			changed := depManager.ChangedFor(ctx, resource)
			Expect(changed).To(Equal(false))
		})
	})

	Describe("ChangedFor with dependencies but without checksum", func() {
		It("Should pass", func() {
			addDependencies()

			changed := depManager.ChangedFor(ctx, resource)
			Expect(changed).To(Equal(true))
		})
	})

	Describe("ChangedFor with dependencies and checksum", func() {
		It("Should pass", func() {
			addDependencies()

			depManager.AddAnnotations(resource)

			changed := depManager.ChangedFor(ctx, resource)
			Expect(changed).To(Equal(false))
		})
	})

	Describe("ChangedFor with dependencies and checksum but not equal", func() {
		It("Should pass", func() {
			addDependencies()

			depManager.AddAnnotations(resource)

			owner.SetGeneration(100)

			changed := depManager.ChangedFor(ctx, resource)
			Expect(changed).To(Equal(true))
		})
	})

	Describe("ChangedFor with version annotation but without version checksum annotation", func() {
		It("Should pass", func() {
			resource.Annotations = version.SetVersion(resource.Annotations, "v1")
			Expect(len(resource.Annotations)).To(Equal(1))

			changed := depManager.ChangedFor(ctx, resource)
			Expect(changed).To(Equal(true))
		})
	})

	Describe("ChangedFor version annotation and version checksum equal", func() {
		It("Should pass", func() {
			resource.Annotations = version.SetVersion(resource.Annotations, "v1")
			Expect(len(resource.Annotations)).To(Equal(1))

			depManager.AddAnnotations(resource)
			Expect(len(resource.Annotations)).To(Equal(2))

			changed := depManager.ChangedFor(ctx, resource)
			Expect(changed).To(Equal(false))
		})
	})

	Describe("ChangedFor version annotation and version checksum not equal", func() {
		It("Should pass", func() {
			resource.Annotations = version.SetVersion(resource.Annotations, "v1")
			Expect(len(resource.Annotations)).To(Equal(1))

			depManager.AddAnnotations(resource)
			Expect(len(resource.Annotations)).To(Equal(2))

			resource.Annotations = version.SetVersion(resource.Annotations, "v2")

			changed := depManager.ChangedFor(ctx, resource)
			Expect(changed).To(Equal(true))
		})
	})
})
