package test

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewRestConfig(ctx context.Context) *rest.Config {
	config := rest.CopyConfig(GetRestConfig(ctx))
	config = rest.AddUserAgent(config, fmt.Sprintf("%s(%s)", application.GetName(ctx), application.GetVersion(ctx)))
	config.APIPath = "api"
	config.NegotiatedSerializer = serializer.NewCodecFactory(GetScheme(ctx))
	config.GroupVersion = &corev1.SchemeGroupVersion

	return config
}

func NewClient(ctx context.Context) client.Client {
	k8sClient, err := client.New(GetRestConfig(ctx), client.Options{Scheme: GetScheme(ctx)})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(k8sClient).ToNot(gomega.BeNil())

	return k8sClient
}
