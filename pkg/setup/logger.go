package setup

import (
	"context"

	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	DevModeConfigKey     = "dev-mode"
	DevModeConfigDefault = true
	LogLevelConfigKey    = "log-level"
)

func FromLogrusToZapLevel(level logrus.Level) zapcore.Level {
	return zapcore.Level(int8(logrus.InfoLevel) - int8(level) + int8(zapcore.InfoLevel))
}

func Logger(ctx context.Context, name, version string) error {
	level := logrus.InfoLevel

	store := configstore.NewStore()
	store.Env("")

	development, err := store.GetItemValueBool(DevModeConfigKey)
	if err != nil {
		if !config.IsNotFound(err, DevModeConfigKey) {
			return errors.Wrap(err, "development mode")
		}

		development = DevModeConfigDefault
		level = logrus.DebugLevel
	}

	if !development {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	levelValue, err := store.GetItemValueInt(LogLevelConfigKey)
	if err != nil {
		if !config.IsNotFound(err, LogLevelConfigKey) {
			return errors.Wrap(err, "level")
		}
	} else {
		level = logrus.Level(levelValue)
	}

	logger := ctrlzap.New(
		ctrlzap.UseDevMode(development),
		ctrlzap.Level(FromLogrusToZapLevel(level)),
	)

	ctrl.SetLogger(logger)
	logrus.SetLevel(level)

	return nil
}
