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
	name    = "harbor-operator"
	version = "dev"
	commit  = "none"
)

const (
	exitCodeFailure = 1
)

func main() {
	setupLog := ctrl.Log.WithName("setup")
	ctx := logger.Context(setupLog)

	// Check non-nil error then log and exit with non-zero code
	fail := func(err error, msg string) {
		// If err is non-nil, then exit with non-zero code,
		// otherwise move on
		if err != nil {
			setupLog.Error(err, msg)
			os.Exit(exitCodeFailure)
		}
	}

	err := setup.Logger(ctx, name, version)
	fail(err, "unable to create logger")

	configstore.InitFromEnvironment()

	application.SetName(&ctx, name)
	application.SetVersion(&ctx, version)

	scheme, err := scheme.New(ctx)
	fail(err, "unable to create scheme")

	mgr, err := manager.New(ctx, scheme)
	fail(err, "unable to create manager")

	traCon, err := tracing.New(ctx)
	fail(err, "unable to create tracer")
	defer traCon.Close()

	err = setup.WithManager(ctx, mgr)
	fail(err, "unable to setup controllers")

	// +kubebuilder:scaffold:builder

	// Log
	setupLog.Info("starting manager", "version", version, "commit", commit)
	err = mgr.Start(ctrl.SetupSignalHandler())
	fail(err, "cannot start manager")
}
