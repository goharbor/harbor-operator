package portforward

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test/pods"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Closer struct {
	stopCh chan struct{}
}

func (c *Closer) Close() error {
	close(c.stopCh)

	return nil
}

type InfoLogger struct {
	Logger logr.Logger
}

func (p *InfoLogger) Write(d []byte) (n int, err error) {
	d = bytes.TrimRight(d, "\r\n")

	p.Logger.Info(string(d))

	return len(d), nil
}

type errorForLogger struct{}

func (err *errorForLogger) Error() string {
	return "error occurred"
}

type ErrorLogger struct {
	Logger logr.Logger
}

func (p *ErrorLogger) Write(d []byte) (n int, err error) {
	d = bytes.TrimRight(d, "\r\n")

	p.Logger.Error(&errorForLogger{}, string(d))

	return len(d), nil
}

func New(ctx context.Context, namespacedName types.NamespacedName, remotePort int) (int, io.Closer) {
	config := test.NewRestConfig(ctx)

	client, err := rest.UnversionedRESTClientFor(config)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	transport, upgrader, err := spdy.RoundTripperFor(config)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})
	logger := ctrlzap.New(ctrlzap.WriteTo(ginkgo.GinkgoWriter), ctrlzap.UseDevMode(true)).
		WithName("port-forward").
		WithValues("deployment", namespacedName)

	var portforwarder *portforward.PortForwarder

	go func() {
		defer ginkgo.GinkgoRecover()

		gomega.Eventually(func() error {
			pod := pods.List(ctx, namespacedName).Ready(ctx).Latest(ctx)
			gomega.Expect(pod).ToNot(gomega.BeNil())

			portForwardURL := client.Post().
				Resource("pods").
				Namespace(pod.GetNamespace()).
				Name(pod.GetName()).
				SubResource("portforward").
				URL()

			dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, portForwardURL)

			var err error
			portforwarder, err = portforward.New(dialer, []string{fmt.Sprintf(":%d", remotePort)}, stopCh, readyCh, &InfoLogger{
				Logger: logger,
			}, &ErrorLogger{
				Logger: logger,
			})
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

			return portforwarder.ForwardPorts()
		}, 30*time.Second, 400*time.Millisecond).ShouldNot(gomega.HaveOccurred())
	}()

	select {
	case <-readyCh:
		// nothing to do
	case <-time.After(time.Minute):
		ginkgo.Fail("port-forward: timedout")
	}

	ports, err := portforwarder.GetPorts()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(ports).To(gomega.HaveLen(1))

	for _, port := range ports {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port.Local))
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		conn, err := net.DialTCP("tcp", nil, addr)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		gomega.Expect(conn.Close()).To(gomega.Succeed())

		return int(port.Local), &Closer{stopCh: stopCh}
	}

	close(stopCh)
	ginkgo.Fail("port-forward: no local port found")

	return 0, nil
}
