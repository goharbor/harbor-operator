package main

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/ovh/configstore"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	goharborv1alpha2 "github.com/goharbor/harbor-operator/api/v1alpha2"
	"github.com/goharbor/harbor-operator/pkg/controllers/setup"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/manager"
	"github.com/goharbor/harbor-operator/pkg/scheme"
	"github.com/goharbor/harbor-operator/pkg/tracing"
)

const (
	OperatorName    = "harbor-operator"
	OperatorVersion = "devel"
)

const (
	exitCodeFailure = 1
)

func getLogger() logr.Logger {
	development, err := configstore.Filter().GetItemValueBool("dev-mode")
	if err != nil {
		development = true
	}

	return zap.Logger(development)
}

func main() {
	// uses env var CONFIGURATION_FROM=... to initialize config
	// examples of possible values:
	// CONFIGURATION_FROM=file:/etc/cfg1.conf,file:/etc/cfg2.conf
	// CONFIGURATION_FROM=env
	// ...
	configstore.InitFromEnvironment()

	setupLog := ctrl.Log.WithName("setup")

	ctx := logger.Context(setupLog)

	logger := getLogger()
	ctrl.SetLogger(logger)

	scheme, err := scheme.New(ctx)
	if err != nil {
		setupLog.Error(err, "unable to create scheme")
		os.Exit(exitCodeFailure)
	}

	mgr, err := manager.New(ctx, scheme)
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(exitCodeFailure)
	}

	traCon, err := tracing.New(ctx, OperatorName, OperatorVersion)
	if err != nil {
		setupLog.Error(err, "unable to create tracer")
		os.Exit(exitCodeFailure)
	}
	defer traCon.Close()

	if err := (&goharborv1alpha2.Harbor{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Harbor")
		os.Exit(exitCodeFailure)
	}

	if err := (setup.SetupWithManager(ctx, mgr, OperatorVersion)); err != nil {
		setupLog.Error(err, "unable to setup controllers")
		os.Exit(exitCodeFailure)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager", "version", OperatorVersion)

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "cannot start manager")
		os.Exit(exitCodeFailure)
	}
}
