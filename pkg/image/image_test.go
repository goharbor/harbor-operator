package image_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	. "github.com/goharbor/harbor-operator/pkg/image"
	"github.com/goharbor/harbor-operator/pkg/version"
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
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:v%s", version.Default())))
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
			image, err := GetImage(ctx, "core", WithRepository("ghcr.io/goharbor/"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("ghcr.io/goharbor/harbor-core:v%s", version.Default())))
		})
	})

	Describe("Get image with tag suffix", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithTagSuffix("-suffix"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:v%s-suffix", version.Default())))
		})
	})

	Describe("Get image with repository and tag suffix", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithRepository("ghcr.io/goharbor/"), WithTagSuffix("-suffix"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("ghcr.io/goharbor/harbor-core:v%s-suffix", version.Default())))
		})
	})

	Describe("Get image from config store", func() {
		It("Should pass", func() {
			os.Setenv(configstore.ConfigEnvVar, "env")
			os.Setenv(ConfigImageKey+"_"+strings.ReplaceAll(version.Default(), ".", "_"), "goharbor/harbor-core:latest")
			configStore := configstore.NewStore()
			configStore.InitFromEnvironment()

			image, err := GetImage(ctx, "core", WithConfigstore(configStore))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("goharbor/harbor-core:latest"))
		})
	})

	Describe("Get image from config store with image key", func() {
		It("Should pass", func() {
			os.Setenv(configstore.ConfigEnvVar, "env")
			os.Setenv("key"+"_"+strings.ReplaceAll(version.Default(), ".", "_"), "goharbor/harbor-core:latest")
			configStore := configstore.NewStore()
			configStore.InitFromEnvironment()

			image, err := GetImage(ctx, "core", WithConfigstore(configStore), WithConfigImageKey("key"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("goharbor/harbor-core:latest"))
		})
	})

	Describe("Get image from config store but not found", func() {
		It("Should pass", func() {
			configStore := configstore.NewStore()

			image, err := GetImage(ctx, "core", WithConfigstore(configStore))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:v%s", version.Default())))
		})
	})

	Describe("Get image from config store but failed", func() {
		It("Should pass", func() {
			configStore := configstore.NewStore()
			configStore.RegisterProvider("foo", func() (configstore.ItemList, error) {
				return configstore.ItemList{}, fmt.Errorf("failed")
			})

			_, err := GetImage(ctx, "core", WithConfigstore(configStore))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get image with harbor version", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "core", WithHarborVersion("2.0.0"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("goharbor/harbor-core:v2.0.0"))
		})
	})

	Describe("Get image for in cluster redis", func() {
		It("Should pass", func() {
			image, err := GetImage(ctx, "cluster-redis")
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("redis:5.0-alpine"))
		})
	})
})
