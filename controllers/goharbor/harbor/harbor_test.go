package harbor_test

import (
	"context"
	"io/ioutil"
	"strings"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/controllers"
	"github.com/goharbor/harbor-operator/controllers/goharbor/harbor"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"
)

func fileString(filePath string) string {
	content, err := ioutil.ReadFile(filePath)
	Expect(err).NotTo(HaveOccurred())

	return strings.TrimSpace(string(content))
}

var _ = Describe("Harbor", func() {
	var (
		ctx     context.Context
		decoder runtime.Decoder
		r       *harbor.Reconciler

		input        string
		getComponent func(*goharborv1alpha2.Harbor) runtime.Object
		output       string
	)

	BeforeEach(func() {
		ctx = logger.Context(zap.LoggerTo(GinkgoWriter, true))
		application.SetName(&ctx, "operator")
		application.SetVersion(&ctx, "dev")

		name := controllers.Harbor.String()

		configStore := config.NewConfigWithDefaults()
		configStore.Env(name)

		i, err := harbor.New(ctx, name, configStore)
		Expect(err).NotTo(HaveOccurred())
		r = i.(*harbor.Reconciler)

		sch := runtime.NewScheme()
		_ = goharborv1alpha2.AddToScheme(sch)

		decoder = serializer.NewCodecFactory(sch).UniversalDeserializer()

		input = ""
		getComponent = nil
		output = ""
	})

	JustBeforeEach(func() {
		obj, _, err := decoder.Decode([]byte(input), nil, nil)
		Expect(err).NotTo(HaveOccurred())

		h, ok := obj.(*goharborv1alpha2.Harbor)
		Expect(ok).To(BeTrue())

		bytes, err := yaml.Marshal(getComponent(h))
		Expect(err).NotTo(HaveOccurred())

		output = strings.TrimSpace(string(bytes))
	})

	Context("GetJobService", func() {
		BeforeEach(func() {
			getComponent = func(h *goharborv1alpha2.Harbor) runtime.Object {
				j, err := r.GetJobService(ctx, h)
				Expect(err).NotTo(HaveOccurred())

				j.SetGroupVersionKind(schema.FromAPIVersionAndKind("goharbor.io/v1alpha2", "JobService"))

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
})
