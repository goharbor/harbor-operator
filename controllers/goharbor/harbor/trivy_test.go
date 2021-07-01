package harbor_test

import (
	"context"
	"os"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/factories/owner"
	"github.com/goharbor/harbor-operator/pkg/graph"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ovh/configstore"
	ctrl "sigs.k8s.io/controller-runtime"
)

func trivyFromResource(r graph.Resource) *goharborv1.Trivy {
	res, ok := r.(*controller.Resource)
	Expect(ok).To(BeTrue())

	trivy, ok := res.GetResource().(*goharborv1.Trivy)
	Expect(ok).To(BeTrue(), "resource is not trivy")

	return trivy
}

var _ = Describe("Trivy", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = test.NewContext()

		configstore.DefaultStore = configstore.NewStore()
	})

	Context("AddTrivyConfigurations", func() {
		Context("Global github token empty", func() {
			It("Should pass", func() {
				r := makeReconciler(ctx)

				h := getSpec("./manifests/trivy/default.yaml")

				c := r.PopulateContext(context.TODO(), ctrl.Request{}) // this ctx has resource manager
				owner.Set(&c, h)

				_, trivyUpdateSecret, err := r.AddTrivyConfigurations(c, h, nil)
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

				c := r.PopulateContext(context.TODO(), ctrl.Request{}) // this ctx has resource manager
				owner.Set(&c, h)

				_, trivyUpdateSecret, err := r.AddTrivyConfigurations(c, h, nil)
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

					c := r.PopulateContext(context.TODO(), ctrl.Request{}) // this ctx has resource manager
					owner.Set(&c, h)

					trivy, err := r.AddTrivy(c, h, nil, nil)
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

					c := r.PopulateContext(context.TODO(), ctrl.Request{}) // this ctx has resource manager
					owner.Set(&c, h)

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

				c := r.PopulateContext(context.TODO(), ctrl.Request{}) // this ctx has resource manager
				owner.Set(&c, h)

				trivy, err := r.AddTrivy(c, h, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(trivy).NotTo(BeNil())

				t := trivyFromResource(trivy)
				Expect(t.Spec.Update.GithubTokenRef).To(BeEquivalentTo("github-token"))
			})
		})
	})
})
