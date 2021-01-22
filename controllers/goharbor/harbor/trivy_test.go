package harbor_test

import (
	"context"
	"os"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovh/configstore"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func trivyFromResource(r graph.Resource) *goharborv1alpha2.Trivy {
	res, ok := r.(*controller.Resource)
	Expect(ok).To(BeTrue())

	u, ok := res.GetResource().(*unstructured.Unstructured)
	Expect(ok).To(BeTrue())

	var trivy goharborv1alpha2.Trivy
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &trivy)
	Expect(err).NotTo(HaveOccurred())

	return &trivy
}

var _ = Describe("Trivy", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = makeContext()

		configstore.DefaultStore = configstore.NewStore()
	})

	Context("AddTrivyConfigurations", func() {
		Context("Global github token empty", func() {
			It("Should pass", func() {
				r := makeReconciler(ctx)

				h := getSpec("./manifests/trivy/default.yaml")

				_, trivyUpdateSecret, err := r.AddTrivyConfigurations(r.NewContext(ctrl.Request{}), h, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(trivyUpdateSecret).To(BeNil())
			})
		})

		Context("Global github token not empty", func() {
			It("Should pass", func() {
				os.Setenv(configstore.ConfigEnvVar, "env")
				os.Setenv("GITHUB_TOKEN", "github-token")
				configstore.InitFromEnvironment()

				r := makeReconciler(ctx)

				h := getSpec("./manifests/trivy/default.yaml")

				_, trivyUpdateSecret, err := r.AddTrivyConfigurations(r.NewContext(ctrl.Request{}), h, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(trivyUpdateSecret).NotTo(BeNil())

				res, ok := trivyUpdateSecret.(*controller.Resource)
				Expect(ok).To(BeTrue())
				Expect(res.GetResource().GetName()).To(BeEquivalentTo("example-harbor-trivy-github"))
			})
		})
	})

	Context("AddTrivy", func() {
		Context("Github token empty", func() {
			Context("Trivy update secret is nil", func() {
				It("Should pass", func() {
					r := makeReconciler(ctx)

					h := getSpec("./manifests/trivy/default.yaml")

					trivy, err := r.AddTrivy(r.NewContext(ctrl.Request{}), h, nil, nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(trivy).NotTo(BeNil())

					t := trivyFromResource(trivy)
					Expect(t.Spec.Update.GithubTokenRef).To(BeEquivalentTo(""))
				})
			})

			Context("Trivy update secret not nil", func() {
				It("Should pass", func() {
					os.Setenv(configstore.ConfigEnvVar, "env")
					os.Setenv("GITHUB_TOKEN", "github-token")
					configstore.InitFromEnvironment()

					r := makeReconciler(ctx)

					h := getSpec("./manifests/trivy/default.yaml")

					c := r.NewContext(ctrl.Request{}) // this ctx has resource manager

					trivyUpdateSecret, err := r.AddTrivyUpdateSecret(c, h)
					Expect(err).NotTo(HaveOccurred())

					trivy, err := r.AddTrivy(c, h, nil, trivyUpdateSecret)
					Expect(err).NotTo(HaveOccurred())
					Expect(trivy).NotTo(BeNil())

					t := trivyFromResource(trivy)
					Expect(t.Spec.Update.GithubTokenRef).To(BeEquivalentTo("example-harbor-trivy-github"))
				})
			})
		})

		Context("Github token not empty", func() {
			It("Should pass", func() {
				os.Setenv(configstore.ConfigEnvVar, "env")
				os.Setenv("GITHUB_TOKEN", "github-token")
				configstore.InitFromEnvironment()

				r := makeReconciler(ctx)

				h := getSpec("./manifests/trivy/github-token.yaml")

				trivy, err := r.AddTrivy(r.NewContext(ctrl.Request{}), h, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(trivy).NotTo(BeNil())

				t := trivyFromResource(trivy)
				Expect(t.Spec.Update.GithubTokenRef).To(BeEquivalentTo("github-token"))
			})
		})
	})
})
