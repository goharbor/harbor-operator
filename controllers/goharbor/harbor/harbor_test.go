package harbor_test

import (
	"context"
	"os"
	"strings"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harbor"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/image"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Harbor", func() {
	var (
		ctx     context.Context
		decoder runtime.Decoder
		r       *harbor.Reconciler

		input        string
		getComponent func(*goharborv1.Harbor) runtime.Object
		output       string
	)

	BeforeEach(func() {
		ctx = test.NewContext()

		r = makeReconciler(ctx)

		sch := runtime.NewScheme()
		_ = goharborv1.AddToScheme(sch)

		decoder = serializer.NewCodecFactory(sch).UniversalDeserializer()

		input = ""
		getComponent = nil
		output = ""

		os.Unsetenv(image.ImageSourceRepositoryEnvKey)
		os.Unsetenv(image.ImageSourceTagSuffixEnvKey)
	})

	JustBeforeEach(func() {
		obj, _, err := decoder.Decode([]byte(input), nil, nil)
		Expect(err).NotTo(HaveOccurred())

		h, ok := obj.(*goharborv1.Harbor)
		Expect(ok).To(BeTrue())

		bytes, err := yaml.Marshal(getComponent(h))
		Expect(err).NotTo(HaveOccurred())

		output = strings.TrimSpace(string(bytes))
	})

	Context("GetJobService", func() {
		BeforeEach(func() {
			getComponent = func(h *goharborv1.Harbor) runtime.Object {
				j, err := r.GetJobService(ctx, h)
				Expect(err).NotTo(HaveOccurred())

				j.SetGroupVersionKind(schema.FromAPIVersionAndKind("goharbor.io/v1beta1", "JobService"))

				return j
			}
		})

		for _, text := range []string{
			"Default",
			"Repository and tag suffix not empty",
			"Repository not empty",
			"Tag suffix not empty",
			"Version and repository not empty",
			"Version not empty",
		} {
			text := text

			filename := strings.Join(strings.Split(strings.ToLower(text), " "), "-")

			Context(text, func() {
				BeforeEach(func() {
					input = fileString("./manifests/jobservice/" + filename + ".yaml")
				})

				It("Should pass", func() {
					expected := fileString("./manifests/jobservice/" + filename + "-expected.yaml")
					Expect(output).To(BeEquivalentTo(expected))
				})
			})
		}
	})

	Context("GetTrivy", func() {
		BeforeEach(func() {
			getComponent = func(h *goharborv1.Harbor) runtime.Object {
				t, err := r.GetTrivy(ctx, h, false)
				Expect(err).NotTo(HaveOccurred())

				t.SetGroupVersionKind(schema.FromAPIVersionAndKind("goharbor.io/v1beta1", "Trivy"))

				return t
			}
		})

		for _, text := range []string{
			"Default",
			"Expose core with tls",
			"Github token",
		} {
			text := text

			filename := strings.Join(strings.Split(strings.ToLower(text), " "), "-")

			Context(text, func() {
				BeforeEach(func() {
					input = fileString("./manifests/trivy/" + filename + ".yaml")
				})

				It("Should pass", func() {
					expected := fileString("./manifests/trivy/" + filename + "-expected.yaml")
					Expect(output).To(BeEquivalentTo(expected))
				})
			})
		}
	})
})
