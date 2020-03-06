package tracing

import (
	"context"
	"io"

	kit_log "github.com/go-kit/kit/log"
	jaeger "github.com/jaegertracing/jaeger-lib/client/log/go-kit"
	"github.com/opentracing/opentracing-go"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	jaeger_client "github.com/uber/jaeger-client-go"
	jaeger_cnf "github.com/uber/jaeger-client-go/config"
	jaeger_metrics "github.com/uber/jaeger-lib/metrics"
	jaeger_prom "github.com/uber/jaeger-lib/metrics/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
)

func Init(jaegerConfig *jaeger_cnf.Configuration, logger jaeger_client.Logger, metrics jaeger_metrics.Factory) (io.Closer, error) {
	tracer, con, err := jaegerConfig.NewTracer(jaeger_cnf.Metrics(metrics), jaeger_cnf.Logger(logger))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create tracing")
	}

	opentracing.SetGlobalTracer(tracer)

	return con, err
}

func New(ctx context.Context, name, version string) (io.Closer, error) {
	var jaegerConfig *jaeger_cnf.Configuration

	log := logger.Get(ctx)
	traceLogger := ctrl.Log.WithName("tracing").WithName("jaeger")

	item, err := configstore.Filter().
		Slice("jaeger").
		Unmarshal(func() interface{} { return &jaeger_cnf.Configuration{} }).
		GetFirstItem()
	if err == nil {
		config, err := item.Unmarshaled()
		if err != nil {
			return nil, errors.Wrap(err, "invalid configuration")
		}

		jaegerConfig = config.(*jaeger_cnf.Configuration)
	} else {
		jaegerConfig, err = jaeger_cnf.FromEnv()
		if err != nil {
			return nil, errors.Wrap(err, "unable to configure from env")
		}
	}

	jaegerConfig.Tags = append(jaegerConfig.Tags, opentracing.Tag{
		Key:   "version",
		Value: version,
	})

	log.Info("Tracing initialized", "host", jaegerConfig.Reporter.LocalAgentHostPort)

	if jaegerConfig.ServiceName == "" {
		jaegerConfig.ServiceName = name
	}

	traCon, err := Init(
		jaegerConfig,
		jaeger.NewLogger(kit_log.LoggerFunc(func(i ...interface{}) error {
			traceLogger.Info("message", i...)
			return nil
		})),
		jaeger_prom.New(jaeger_prom.WithRegisterer(prometheus.DefaultRegisterer)),
	)

	return traCon, errors.Wrap(err, "unable to init tracer")
}
