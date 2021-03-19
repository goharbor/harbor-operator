package version_test

import (
	"github.com/goharbor/harbor-operator/pkg/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	BeforeEach(func() {
		version.RegisterKnowVersions("2.1.0", "2.1.1", "2.1.2")
	})

	Describe("Default version", func() {
		It("Should pass", func() {
			v := version.Default()
			Expect(v).To(Equal("2.1.2"))
		})
	})

	Describe("Validate version 2.1.0", func() {
		It("Should fail", func() {
			err := version.Validate("2.1.0")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Validate default version", func() {
		It("Should pass", func() {
			err := version.Validate(version.Default())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UpgradeAllowed from unknow version to default version", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed("unknow", version.Default())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpgradeAllowed from default version to unknow version", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed(version.Default(), "unknow")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpgradeAllowed from unknow version to unknow version", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed("unknow", "unknow")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpgradeAllowed from default version to 2.1.0", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed(version.Default(), "2.1.0")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpgradeAllowed from 2.1.0 to default version", func() {
		It("Should pass", func() {
			err := version.UpgradeAllowed("2.1.0", version.Default())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetVersion from nil annotations", func() {
		It("Should pass", func() {
			v := version.GetVersion(nil)
			Expect(v).To(Equal(""))
		})
	})

	Describe("GetVersion from annotations without version", func() {
		It("Should pass", func() {
			v := version.GetVersion(map[string]string{"key": "value"})
			Expect(v).To(Equal(""))
		})
	})

	Describe("GetVersion from annotations with version", func() {
		It("Should pass", func() {
			v := version.GetVersion(version.SetVersion(nil, version.Default()))
			Expect(v).To(Equal(version.Default()))
		})
	})

	Describe("NewVersionAnnotations from nil annotations", func() {
		It("Should pass", func() {
			annotations := version.NewVersionAnnotations(nil)
			Expect(annotations).To(Equal(map[string]string(nil)))
		})
	})

	Describe("NewVersionAnnotations from annotations without version", func() {
		It("Should pass", func() {
			annotations := version.NewVersionAnnotations(map[string]string{"key": "value"})
			Expect(annotations).To(Equal(map[string]string(nil)))
		})
	})

	Describe("NewVersionAnnotations from annotations with version", func() {
		It("Should pass", func() {
			annotations := version.NewVersionAnnotations(version.SetVersion(nil, version.Default()))
			Expect(annotations).NotTo(Equal(map[string]string(nil)))
			Expect(len(annotations)).To(Equal(1))
		})
	})
})
