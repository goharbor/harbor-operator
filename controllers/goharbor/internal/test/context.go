package test

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var k8sClientKey = true

func GetClient(ctx context.Context) client.Client {
	return ctx.Value(&k8sClientKey).(client.Client)
}

func WithClient(ctx context.Context, c client.Client) context.Context {
	return context.WithValue(ctx, &k8sClientKey, c)
}

var k8sSchemeKey = true

func GetScheme(ctx context.Context) *runtime.Scheme {
	return ctx.Value(&k8sSchemeKey).(*runtime.Scheme)
}

func WithScheme(ctx context.Context, s *runtime.Scheme) context.Context {
	return context.WithValue(ctx, &k8sSchemeKey, s)
}

var environmentKey = true

func GetEnvironment(ctx context.Context) *envtest.Environment {
	return ctx.Value(&environmentKey).(*envtest.Environment)
}

func WithEnvironment(ctx context.Context, env *envtest.Environment) context.Context {
	return context.WithValue(ctx, &environmentKey, env)
}

var managerKey = true

func GetManager(ctx context.Context) manager.Manager {
	return ctx.Value(&managerKey).(manager.Manager)
}

func WithManager(ctx context.Context, mgr manager.Manager) context.Context {
	return context.WithValue(ctx, &managerKey, mgr)
}

var restConfigKey = true

func GetRestConfig(ctx context.Context) *rest.Config {
	return ctx.Value(&restConfigKey).(*rest.Config)
}

func WithRestConfig(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, &restConfigKey, cfg)
}

func NewContext() context.Context {
	log := zap.New(zap.WriteTo(ginkgo.GinkgoWriter), zap.UseDevMode(true))
	logf.SetLogger(log)

	ctx := logger.Context(log)

	application.SetName(&ctx, NewName("app"))
	application.SetVersion(&ctx, NewName("version"))
	application.SetGitCommit(&ctx, NewName("commit"))

	s, err := scheme.New(ctx)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	ctx = WithScheme(ctx, s)

	testEnv := &envtest.Environment{}

	ctx = WithEnvironment(ctx, testEnv)

	return ctx
}
