package v1alpha3_test

import (
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("HarborTypes", func() {
	Describe("HarborSpec", func() {
		DescribeTable("ValidateNotary",
			func(spec *goharborv1.HarborSpec, wantErr bool) {
				err := spec.ValidateNotary()
				if wantErr {
					Ω(err).ShouldNot(BeNil())
				} else {
					Ω(err).Should(BeNil())
				}
			},
			Entry("Notary is nil", &goharborv1.HarborSpec{}, false),
			Entry("Expose notary is nil", &goharborv1.HarborSpec{
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					Notary: &goharborv1.NotaryComponentSpec{},
				},
				Expose: goharborv1.HarborExposeSpec{},
			}, true),
			Entry("Expose notary ingress is nil", &goharborv1.HarborSpec{
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					Notary: &goharborv1.NotaryComponentSpec{},
				},
				Expose: goharborv1.HarborExposeSpec{
					Notary: &goharborv1.HarborExposeComponentSpec{},
				},
			}, true),
			Entry("Expose notary ingress tls is nil", &goharborv1.HarborSpec{
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					Notary: &goharborv1.NotaryComponentSpec{},
				},
				Expose: goharborv1.HarborExposeSpec{
					Notary: &goharborv1.HarborExposeComponentSpec{
						Ingress: &goharborv1.HarborExposeIngressSpec{Host: "notary.harbor.domain"},
					},
				},
			}, true),
			Entry("Expose notary ingress tls certificateRef is empty", &goharborv1.HarborSpec{
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					Notary: &goharborv1.NotaryComponentSpec{},
				},
				Expose: goharborv1.HarborExposeSpec{
					Notary: &goharborv1.HarborExposeComponentSpec{
						Ingress: &goharborv1.HarborExposeIngressSpec{Host: "notary.harbor.domain"},
						TLS: &harbormetav1.ComponentsTLSSpec{
							CertificateRef: "",
						},
					},
				},
			}, true),
			Entry("Valid", &goharborv1.HarborSpec{
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					Notary: &goharborv1.NotaryComponentSpec{},
				},
				Expose: goharborv1.HarborExposeSpec{
					Notary: &goharborv1.HarborExposeComponentSpec{
						Ingress: &goharborv1.HarborExposeIngressSpec{Host: "notary.harbor.domain"},
						TLS: &harbormetav1.ComponentsTLSSpec{
							CertificateRef: "cert",
						},
					},
				},
			}, false),
		)

		DescribeTable("ValidateRegistryController",
			func(spec *goharborv1.HarborSpec, wantErr bool) {
				err := spec.ValidateRegistryController()
				if wantErr {
					Ω(err).ShouldNot(BeNil())
				} else {
					Ω(err).Should(BeNil())
				}
			},
			Entry("RegistryController is nil", &goharborv1.HarborSpec{}, false),
			Entry("RegistryController is not nil", &goharborv1.HarborSpec{
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					RegistryController: &harbormetav1.ComponentSpec{
						ServiceAccountName: "account",
					},
				},
			}, false),
			Entry("Storage is file system, nodeSelector and tolerations matched", &goharborv1.HarborSpec{
				ImageChartStorage: &goharborv1.HarborStorageImageChartStorageSpec{
					FileSystem: &goharborv1.HarborStorageImageChartStorageFileSystemSpec{},
				},
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					RegistryController: &harbormetav1.ComponentSpec{
						ServiceAccountName: "account",
					},
				},
			}, false),
			Entry("Storage is file system, nodeSelector not matched", &goharborv1.HarborSpec{
				ImageChartStorage: &goharborv1.HarborStorageImageChartStorageSpec{
					FileSystem: &goharborv1.HarborStorageImageChartStorageFileSystemSpec{},
				},
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					RegistryController: &harbormetav1.ComponentSpec{
						ServiceAccountName: "account",
						NodeSelector:       map[string]string{"hostname": "host"},
					},
				},
			}, true),
			Entry("Storage is file system, tolerations not matched", &goharborv1.HarborSpec{
				ImageChartStorage: &goharborv1.HarborStorageImageChartStorageSpec{
					FileSystem: &goharborv1.HarborStorageImageChartStorageFileSystemSpec{},
				},
				HarborComponentsSpec: goharborv1.HarborComponentsSpec{
					RegistryController: &harbormetav1.ComponentSpec{
						ServiceAccountName: "account",
						Tolerations:        []corev1.Toleration{{Key: "key", Value: "value"}},
					},
				},
			}, true),
		)
	})
})
