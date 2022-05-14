package main

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/goharbor/harbor-operator/pkg/exit"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/manager"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	"github.com/goharbor/harbor-operator/pkg/setup"
	"github.com/goharbor/harbor-operator/pkg/tracing"
	"github.com/ovh/configstore"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	name      = "harbor-operator"
	version   = "dev"
	gitCommit = "none"
)

const (
	LoggerExitCode int = iota + 1
	ManagerExitCode
	SchemeExitCode
	TracingExitCode
	ControllersExitCode
	RunExitCode
)

func setupContextAndLogger() (context.Context, logr.Logger, error) {
	setupLog := ctrl.Log.WithName("setup")
	ctx := logger.Context(setupLog)

	err := setup.Logger(ctx, name, version)
	if err != nil {
		return nil, ctrl.Log, err
	}

	application.SetName(&ctx, name)
	application.SetVersion(&ctx, version)
	application.SetGitCommit(&ctx, gitCommit)

	return ctx, setupLog, nil
}

func main() {
	defer exit.Exit()

	ctx, setupLog, err := setupContextAndLogger()
	if err != nil {
		setupLog.Error(err, "unable to create logger")
		exit.SetCode(LoggerExitCode)

		return
	}

	configstore.InitFromEnvironment()

	scheme, err := scheme.New(ctx)
	if err != nil {
		setupLog.Error(err, "unable to create scheme")
		exit.SetCode(SchemeExitCode)

		return
	}

	mgr, err := manager.New(ctx, scheme)
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		exit.SetCode(ManagerExitCode)

		return
	}

	traCon, err := tracing.New(ctx)
	if err != nil {
		setupLog.Error(err, "unable to create tracer")
		exit.SetCode(TracingExitCode)

		return
	}
	defer traCon.Close()

	if err := (setup.WithManager(ctx, mgr)); err != nil {
		setupLog.Error(err, "unable to setup controllers")
		exit.SetCode(ControllersExitCode)

		return
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager", "version", version, "commit", gitCommit)

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "cannot start manager")

		if exit.GetCode() == exit.SuccessExitCode {
			exit.SetCode(RunExitCode)
		}

		return
	}
}
