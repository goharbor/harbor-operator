package main

import (
	"os"

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

const (
	OperatorName    = "harbor-operator"
	OperatorVersion = "devel"
)

const (
	exitCodeFailure = 1
)

var exitCode = 0

func SetExitCode(value int) {
	exitCode = value
}

func GetExitCode() int {
	return exitCode
}

func main() {
	defer func() { os.Exit(GetExitCode()) }()

	setupLog := ctrl.Log.WithName("setup")
	ctx := logger.Context(setupLog)

	err := setup.Logger(ctx, OperatorName, OperatorVersion)
	if err != nil {
		setupLog.Error(err, "unable to create logger")
		SetExitCode(exitCodeFailure)

		return
	}

	configstore.InitFromEnvironment()

	application.SetName(&ctx, OperatorName)
	application.SetVersion(&ctx, OperatorVersion)

	scheme, err := scheme.New(ctx)
	if err != nil {
		setupLog.Error(err, "unable to create scheme")
		SetExitCode(exitCodeFailure)

		return
	}

	mgr, err := manager.New(ctx, scheme)
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		SetExitCode(exitCodeFailure)

		return
	}

	traCon, err := tracing.New(ctx)
	if err != nil {
		setupLog.Error(err, "unable to create tracer")
		SetExitCode(exitCodeFailure)

		return
	}
	defer traCon.Close()

	if err := (setup.WithManager(ctx, mgr)); err != nil {
		setupLog.Error(err, "unable to setup controllers")
		SetExitCode(exitCodeFailure)

		return
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager", "version", OperatorVersion)

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "cannot start manager")
		SetExitCode(exitCodeFailure)

		return
	}
}
