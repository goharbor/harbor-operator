package cache

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRedisHealthChecker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HealthChecker")
}

var _ = Describe("Test HealthChecker", func() {
	var (
		checker lcm.HealthChecker
		server  *miniredis.Miniredis
		err     error
	)

	BeforeSuite(func() {
		checker = &RedisHealthChecker{}
		server, err = miniredis.Run()
		if err != nil {
			Fail(fmt.Errorf("mock redis server error: %s", err.Error()).Error())
		}

	})

	AfterSuite(func() {
		if server != nil {
			server.Close()
		}
	})

	Context("Test CheckHealth", func() {
		It("bad address should return error", func() {
			config := &lcm.ServiceConfig{
				Endpoint: &lcm.Endpoint{
					Host: "127.0.0.1",
					Port: 6379,
				},
			}
			resp, err := checker.CheckHealth(context.TODO(), config)
			Expect(err).ShouldNot(BeNil())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.Status).Should(Equal(lcm.UnHealthy))
		})

		It("right address should check health successfully", func() {
			host, port, err := net.SplitHostPort(server.Addr())
			Expect(err).To(BeNil())
			p, err := strconv.Atoi(port)
			Expect(err).To(BeNil())
			config := &lcm.ServiceConfig{
				Endpoint: &lcm.Endpoint{
					Host: host,
					Port: uint(p),
				},
			}
			resp, err := checker.CheckHealth(context.TODO(), config)
			Expect(err).Should(BeNil())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.Status).Should(Equal(lcm.Healthy))
		})
	})
})
