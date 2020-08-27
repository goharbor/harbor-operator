package test

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	k8sClient         = true
	k8sScheme         = true
	controllerVersion = true
	environment       = true
)

func GetClient(ctx context.Context) client.Client {
	return ctx.Value(&k8sClient).(client.Client)
}

func SetClient(ctx context.Context, c client.Client) context.Context {
	return context.WithValue(ctx, &k8sClient, c)
}

func GetScheme(ctx context.Context) *runtime.Scheme {
	return ctx.Value(&k8sScheme).(*runtime.Scheme)
}

func SetScheme(ctx context.Context, s *runtime.Scheme) context.Context {
	return context.WithValue(ctx, &k8sScheme, s)
}

func GetEnvironment(ctx context.Context) *envtest.Environment {
	return ctx.Value(&environment).(*envtest.Environment)
}

func SetEnvironment(ctx context.Context, env *envtest.Environment) context.Context {
	return context.WithValue(ctx, &environment, env)
}

func NewContext(pathToModule string) context.Context {
	log := zap.LoggerTo(GinkgoWriter, true)
	logf.SetLogger(log)

	ctx := logger.Context(log)

	application.SetName(&ctx, NewName("app"))
	application.SetVersion(&ctx, NewName("version"))

	s, err := scheme.New(ctx)
	Expect(err).ToNot(HaveOccurred())

	ctx = SetScheme(ctx, s)

	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join(pathToModule, "config", "crd", "bases")},
	}

	ctx = SetEnvironment(ctx, testEnv)

	return ctx
}
