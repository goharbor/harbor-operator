package image_test

import (
	"context"
	"fmt"
	"os"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	. "github.com/goharbor/harbor-operator/pkg/image"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Get image", func() {
	var (
		ctx           context.Context
		harborVersion string
		getImage      func(ctx context.Context, component string, options ...Option) (string, error)
	)

	BeforeEach(func() {
		ctx = logger.Context(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
		harborVersion = "2.3.0"
		getImage = func(ctx context.Context, component string, options ...Option) (string, error) {
			options = append([]Option{WithHarborVersion(harborVersion)}, options...)

			return GetImage(ctx, component, options...)
		}

		os.Unsetenv(ImageSourceRepositoryEnvKey)
		os.Unsetenv(ImageSourceTagSuffixEnvKey)
	})

	Describe("Get image for unknow component", func() {
		It("Should fail", func() {
			_, err := getImage(ctx, "unknow")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get default image", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "core")
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:v%s", harborVersion)))
		})
	})

	Describe("Get image from spec", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "core", WithImageFromSpec("docker.io/goharbor/harbor-core:latest"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("docker.io/goharbor/harbor-core:latest"))
		})
	})

	Describe("Get image with repository", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "core", WithRepository("ghcr.io/goharbor/"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("ghcr.io/goharbor/harbor-core:v%s", harborVersion)))
		})
	})

	Describe("Get image with tag suffix", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "core", WithTagSuffix("-suffix"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:v%s-suffix", harborVersion)))
		})
	})

	Describe("Get image with repository and tag suffix", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "core", WithRepository("ghcr.io/goharbor/"), WithTagSuffix("-suffix"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("ghcr.io/goharbor/harbor-core:v%s-suffix", harborVersion)))
		})
	})

	Describe("Get image with repository from env", func() {
		It("Should pass", func() {
			os.Setenv(ImageSourceRepositoryEnvKey, "ghcr.io/goharbor")

			image, err := getImage(ctx, "core")
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("ghcr.io/goharbor/harbor-core:v%s", harborVersion)))
		})
	})

	Describe("Get image with tag suffix from env", func() {
		It("Should pass", func() {
			os.Setenv(ImageSourceTagSuffixEnvKey, "-suffix")

			image, err := getImage(ctx, "core")
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal(fmt.Sprintf("goharbor/harbor-core:v%s-suffix", harborVersion)))
		})
	})

	Describe("Get image with harbor version", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "core", WithHarborVersion("2.0.0"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("goharbor/harbor-core:v2.0.0"))
		})
	})

	Describe("Get image without harbor version", func() {
		It("Should fail", func() {
			_, err := getImage(ctx, "core", WithHarborVersion(""))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get image for in cluster redis", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "cluster-redis")
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("redis:5.0-alpine"))
		})
	})

	Describe("Get image for in cluster redis with repository", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "cluster-redis", WithRepository("ghcr.io/goharbor"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("ghcr.io/goharbor/redis:5.0-alpine"))
		})
	})

	Describe("Get image for in cluster postgresql", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "cluster-postgresql")
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("registry.opensource.zalan.do/acid/spilo-13:2.1-p1"))
		})
	})

	Describe("Get image for in cluster postgresql with repository", func() {
		It("Should pass", func() {
			image, err := getImage(ctx, "cluster-postgresql", WithRepository("ghcr.io/goharbor"))
			Expect(err).ToNot(HaveOccurred())
			Expect(image).To(Equal("ghcr.io/goharbor/spilo-13:2.1-p1"))
		})
	})
})
