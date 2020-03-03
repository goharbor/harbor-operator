package mutation

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	// +kubebuilder:scaffold:imports

	"github.com/goharbor/harbor-operator/pkg/resources"
	"github.com/goharbor/harbor-operator/pkg/scheme"
)

// These tests use Ginkgo (BDD-style Go testing framework). Rcfer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Mutate the owner reference", func() {
	var getOwnerMutation resources.Mutable
	var owner *corev1.Namespace
	var matcher interface{}

	BeforeEach(func() {
		s, err := scheme.New(context.TODO())
		Expect(err).ToNot(HaveOccurred())

		owner = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "unesco",
				UID:  "775665789",
			},
		}
		gvks, _, err := s.ObjectKinds(owner)
		Expect(err).ToNot(HaveOccurred())

		gvk := gvks[0]
		owner.SetGroupVersionKind(gvk)

		getOwnerMutation = GetOwnerMutation(s, owner)
		varTrue := true
		matcher = gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"APIVersion": BeEquivalentTo(gvk.Version),
			"Kind":       BeEquivalentTo(gvk.Kind),
			"Controller": BeEquivalentTo(&varTrue),
			"Name":       BeEquivalentTo(owner.GetName()),
		})
	})

	Context("With a metav1 object", func() {
		var resource *corev1.Secret
		var result *corev1.Secret
		var mutate controllerutil.MutateFn

		BeforeEach(func() {
			resource = &corev1.Secret{}
			result = resource.DeepCopy()
			mutate = getOwnerMutation(context.TODO(), resource, result)
		})

		JustBeforeEach(func() {
			resource.DeepCopyInto(result)
		})

		Context("Without owner", func() {
			BeforeEach(func() {
				resource.SetOwnerReferences(nil)
			})

			It("Should add the right owner", func() {
				err := mutate()
				Expect(err).ToNot(HaveOccurred())

				Expect(result.GetOwnerReferences()).To(ContainElement(matcher))
			})
		})

		Context("With no-controller owners", func() {
			var initialOwners []metav1.OwnerReference

			BeforeEach(func() {
				varFalse := false
				version, kind := owner.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
				initialOwners = []metav1.OwnerReference{
					{
						APIVersion: version,
						Kind:       kind,
						Controller: &varFalse,
						Name:       "owner",
						UID:        types.UID("the-uid"),
					},
				}
				resource.SetOwnerReferences(initialOwners)
			})

			It("Should add the owner", func() {
				err := mutate()
				Expect(err).ToNot(HaveOccurred())

				ownerReferences := result.GetOwnerReferences()
				Expect(ownerReferences).To(ContainElement(matcher))
				for _, owner := range initialOwners {
					Expect(ownerReferences).To(ContainElement(owner))
				}
			})
		})

		Context("With controller owner", func() {
			BeforeEach(func() {
				varTrue := true
				version, kind := owner.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
				resource.SetOwnerReferences([]metav1.OwnerReference{
					{
						APIVersion: version,
						Kind:       kind,
						Controller: &varTrue,
						Name:       "owner",
						UID:        types.UID("the-uid"),
					},
				})
			})

			It("Should failed", func() {
				err := mutate()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("With the same owner", func() {
			var initialOwners []metav1.OwnerReference

			BeforeEach(func() {
				varTrue := true
				version, kind := owner.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
				initialOwners = []metav1.OwnerReference{
					{
						APIVersion: version,
						Kind:       kind,
						Controller: &varTrue,
						Name:       owner.Name,
						UID:        owner.GetUID(),
					},
				}
				resource.SetOwnerReferences(initialOwners)
			})

			It("Should pass", func() {
				err := mutate()
				Expect(err).ToNot(HaveOccurred())

				ownerReferences := result.GetOwnerReferences()
				Expect(ownerReferences).To(ConsistOf(matcher))
			})
		})
	})
})
