package version_test

import (
	"github.com/goharbor/harbor-operator/pkg/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var v string

	BeforeEach(func() {
		v = "2.2.1"

		version.RegisterKnownConstraints("~2.1.x", "~2.2.x")
	})

	Describe("Validate version xyz", func() {
		It("Should fail", func() {
			err := version.Validate("xyz")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Validate version 2.1.0", func() {
		It("Should fail", func() {
			err := version.Validate("2.1.0")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Validate version 2.2.0", func() {
		It("Should pass", func() {
			err := version.Validate("2.2.0")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UpgradeAllowed from 2.0.0 to 2.2.0", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed("2.0.0", "2.2.0")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unknown version 2.0.0"))
		})
	})

	Describe("UpgradeAllowed from 2.1.0 to 2.3.0", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed("2.1.0", "2.3.0")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unknown version 2.3.0"))
		})
	})

	Describe("UpgradeAllowed from 2.0.0 version to 2.3.0", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed("2.0.0", "2.3.0")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unknown version 2.0.0"))
		})
	})

	Describe("UpgradeAllowed from 2.1.0 to 2.1.2", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed("2.1.0", "2.1.2")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("upgrade from 2.1.0 to 2.1.2 is not allowed, error: 2.1.2 does not have same major and minor version as 2.2.x"))
		})
	})

	Describe("UpgradeAllowed from 2.2.1 to 2.2.0", func() {
		It("Should fail", func() {
			err := version.UpgradeAllowed("2.2.1", "2.2.0")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("downgrade from 2.2.1 to 2.2.0 is not allowed"))
		})
	})

	Describe("UpgradeAllowed from 2.1.0 to 2.2.0", func() {
		It("Should pass", func() {
			err := version.UpgradeAllowed("2.1.0", "2.2.0")
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
			v := version.GetVersion(version.SetVersion(nil, v))
			Expect(v).To(Equal(v))
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
			annotations := version.NewVersionAnnotations(version.SetVersion(nil, v))
			Expect(annotations).NotTo(Equal(map[string]string(nil)))
			Expect(len(annotations)).To(Equal(1))
		})
	})
})
