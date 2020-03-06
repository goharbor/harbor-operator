package manager

import (
	"context"
	"fmt"
	"net/http"

	nettracing "github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/transport"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

const (
	WebHookPort = 9443
	MetricsPort = 8080
)

func New(ctx context.Context, scheme *runtime.Scheme) (manager.Manager, error) {
	var mgrConfig *manager.Options = &ctrl.Options{
		MetricsBindAddress: fmt.Sprintf(":%d", MetricsPort),
		LeaderElection:     false,
		Port:               WebHookPort,
	}

	log := logger.Get(ctx)

	item, err := configstore.Filter().
		Slice("operator").
		Unmarshal(func() interface{} { return &manager.Options{} }).
		GetFirstItem()
	if err == nil {
		// todo
		config, err := item.Unmarshaled()
		if err != nil {
			return nil, errors.Wrap(err, "unable to get configuration")
		}

		mgrConfig = config.(*manager.Options)
	}

	mgrConfig.Scheme = scheme

	log.Info("Manager initialized", "Metrics.Address", mgrConfig.MetricsBindAddress, "LeaderElection.Enabled", mgrConfig.LeaderElection, "LeaderElection.Namespace", mgrConfig.LeaderElectionNamespace, "LeaderElection.ID", mgrConfig.LeaderElectionID)

	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get rest configuration")
	}

	config.WrapTransport = transport.Wrappers(func(rt http.RoundTripper) http.RoundTripper {
		return &nettracing.Transport{RoundTripper: rt}
	})

	mgr, err := ctrl.NewManager(config, *mgrConfig)

	return mgr, errors.Wrap(err, "unable to get the manager")
}
