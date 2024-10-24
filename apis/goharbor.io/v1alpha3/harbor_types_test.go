package v1alpha3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	goharborv1 "github.com/plotly/harbor-operator/apis/goharbor.io/v1alpha3"
	harbormetav1 "github.com/plotly/harbor-operator/apis/meta/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("HarborTypes", func() {
	Describe("HarborSpec", func() {
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
