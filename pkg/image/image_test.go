package image_test

import (
	"context"
	"fmt"
	"os"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	. "github.com/goharbor/harbor-operator/pkg/image"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovh/configstore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Get image", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = logger.Context(zap.LoggerTo(GinkgoWriter, true))
	})

	Describe("Get image for unknow component", func() {
		It("Should fail", func() {
			_, err := GetImage(ctx, "unknow")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get default image", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core")
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:%s", config.DefaultHarborVersion)))
		})
	})

	Describe("Get image from spec", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithImageFromSpec("docker.io/goharbor/harbor-core:latest"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("docker.io/goharbor/harbor-core:latest"))
		})
	})

	Describe("Get image with repository", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithRepository("quay.io/goharbor/"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("quay.io/goharbor/harbor-core:%s", config.DefaultHarborVersion)))
		})
	})

	Describe("Get image with tag", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithTag(config.DefaultHarborVersion+"-suffix"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:%s-suffix", config.DefaultHarborVersion)))
		})
	})

	Describe("Get image with repository and tag", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithRepository("quay.io/goharbor/"), WithTag(config.DefaultHarborVersion+"-suffix"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("quay.io/goharbor/harbor-core:%s-suffix", config.DefaultHarborVersion)))
		})
	})

	Describe("Get image from config store", func() {
		It("Should pass", func() {
			os.Setenv(configstore.ConfigEnvVar, "env")
			os.Setenv(ConfigImageKey, "goharbor/harbor-core:latest")
			configStore := configstore.NewStore()
			configStore.InitFromEnvironment()

			image, err := GetImage(ctx, "core", WithConfigstore(configStore))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("goharbor/harbor-core:latest"))
		})
	})

	Describe("Get image from config store but not found", func() {
		It("Should pass", func() {
			configStore := configstore.NewStore()

			image, err := GetImage(ctx, "core", WithConfigstore(configStore))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:%s", config.DefaultHarborVersion)))
		})
	})

	Describe("Get image with harbor version", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithHarborVersion("2.0.0"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("goharbor/harbor-core:v2.0.0"))
		})
	})
})
