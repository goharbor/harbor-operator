package portforward

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/controllers/goharbor/internal/test"
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

	pods := test.GetPods(ctx, namespacedName)
	gomega.Expect(pods).ToNot(gomega.BeEmpty())

	pod := pods[0]

	portForwardURL := client.Post().
		Resource("pods").
		Namespace(pod.GetNamespace()).
		Name(pod.GetName()).
		SubResource("portforward").
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(config)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, portForwardURL)
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})
	logger := ctrlzap.New(ctrlzap.WriteTo(ginkgo.GinkgoWriter), ctrlzap.UseDevMode(true)).
		WithName("port-forward").
		WithValues("deployment", namespacedName)

	portforwarder, err := portforward.New(dialer, []string{fmt.Sprintf(":%d", remotePort)}, stopCh, readyCh, &InfoLogger{
		Logger: logger,
	}, &ErrorLogger{
		Logger: logger,
	})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	go func() {
		defer ginkgo.GinkgoRecover()

		gomega.Expect(portforwarder.ForwardPorts()).To(gomega.Succeed())
	}()

	gomega.Eventually(readyCh).Should(gomega.BeClosed())

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
