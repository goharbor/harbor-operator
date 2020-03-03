package statuscheck

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// +kubebuilder:scaffold:imports

	"github.com/goharbor/harbor-operator/pkg/scheme"
)

// These tests use Ginkgo (BDD-style Go testing framework). Rcfer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Check the status", func() {
	Context("Of a certificate resource", func() {
		var resource *certv1.Certificate
		var data *certv1.Certificate

		BeforeEach(func() {
			s, err := scheme.New(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			data = &certv1.Certificate{}
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
				data.Status = certv1.CertificateStatus{}
			})

			It("Should not be ready", func() {
				ok, err := CertificateCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With NotAfter status < Now", func() {
			BeforeEach(func() {
				expiredAt := metav1.NewTime(time.Now().Add(-24 * time.Hour))
				data.Status.NotAfter = &expiredAt
			})

			It("Should not be ready", func() {
				ok, err := CertificateCheck(context.TODO(), resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		Context("With Ready condition", func() {
			Context("To True", func() {
				BeforeEach(func() {
					data.Status.Conditions = append(data.Status.Conditions, certv1.CertificateCondition{
						Type:   certv1.CertificateConditionReady,
						Status: cmmeta.ConditionTrue,
					})
				})

				It("Should not be ready", func() {
					ok, err := CertificateCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeTrue())
				})
			})
			Context("To False", func() {
				BeforeEach(func() {
					data.Status.Conditions = append(data.Status.Conditions, certv1.CertificateCondition{
						Type:   certv1.CertificateConditionReady,
						Status: cmmeta.ConditionFalse,
					})
				})

				It("Should not be ready", func() {
					ok, err := CertificateCheck(context.TODO(), resource)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(BeFalse())
				})
			})
		})
	})
})
