package harbor_test

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/harbor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Harbor Client Configuration", func() {
	var (
		client harbor.Client
		server *ghttp.Server
		err    error
	)

	BeforeSuite(func() {
		server = ghttp.NewTLSServer()
		client = harbor.NewClient(server.URL())
	})

	AfterSuite(func() {
		server.Close()
	})

	Context("Apply configuration", func() {
		BeforeEach(func() {
			server.RouteToHandler("PUT", "/api/v2.0/configurations", ghttp.RespondWith(200, nil))
		})

		It("should apply success", func() {
			payload := `{"email_ssl": true}`
			err = client.ApplyConfiguration(context.TODO(), []byte(payload))
			Expect(err).To(BeNil())
		})
	})
})
