package manager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	nettracing "github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/transport"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	WebHookPort = 9443
	MetricsPort = 8080
	ProbePort   = 5000

	ManagerConfigKey = "operator"
)

func New(ctx context.Context, scheme *runtime.Scheme) (manager.Manager, error) {
	var mgrConfig manager.Options = ctrl.Options{
		MetricsBindAddress:     fmt.Sprintf(":%d", MetricsPort),
		LeaderElection:         false,
		Port:                   WebHookPort,
		HealthProbeBindAddress: fmt.Sprintf(":%d", ProbePort),
		Scheme:                 scheme,
	}

	item, err := configstore.Filter().
		Slice(ManagerConfigKey).
		Unmarshal(func() interface{} {
			// Duplicate mgrConfig
			config := mgrConfig

			return &config
		}).
		GetFirstItem()
	if err != nil {
		if _, ok := err.(configstore.ErrItemNotFound); !ok {
			return nil, errors.Wrap(err, "cannot get configuration")
		}
	} else {
		config, err := item.Unmarshaled()
		if err != nil {
			return nil, errors.Wrap(err, "unable to get configuration")
		}

		mgrConfig = *config.(*manager.Options)
	}

	logger.Get(ctx).Info(
		"Manager initialized",
		"Webhook.Port", mgrConfig.Port,
		"Metrics.Address", mgrConfig.MetricsBindAddress,
		"Probe.Address", mgrConfig.HealthProbeBindAddress,
		"LeaderElection.Enabled", mgrConfig.LeaderElection,
		"LeaderElection.Namespace", mgrConfig.LeaderElectionNamespace,
		"LeaderElection.ID", mgrConfig.LeaderElectionID,
	)

	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get rest configuration")
	}

	config.WrapTransport = transport.Wrappers(func(rt http.RoundTripper) http.RoundTripper {
		return &nettracing.Transport{RoundTripper: rt}
	})

	mgr, err := ctrl.NewManager(config, mgrConfig)

	return mgr, errors.Wrap(err, "unable to get the manager")
}
