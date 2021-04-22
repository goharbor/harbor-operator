package tracing

import (
	"context"
	"io"
	"sync"

	kit_log "github.com/go-kit/kit/log"
	"github.com/goharbor/harbor-operator/pkg/factories/application"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	jaeger_kit "github.com/jaegertracing/jaeger-lib/client/log/go-kit"
	"github.com/opentracing/opentracing-go"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	jaeger_cnf "github.com/uber/jaeger-client-go/config"
	jaeger_metrics "github.com/uber/jaeger-lib/metrics"
	jaeger_prom "github.com/uber/jaeger-lib/metrics/prometheus"
)

var jaegerProm jaeger_metrics.Factory

var once sync.Once

func Init(ctx context.Context, jaegerConfig *jaeger_cnf.Configuration) (io.Closer, error) {
	once.Do(func() {
		jaegerProm = jaeger_prom.New(jaeger_prom.WithRegisterer(prometheus.DefaultRegisterer))
	})

	traceLogger := logger.Get(ctx).WithName("tracing")
	l := jaeger_kit.NewLogger(kit_log.LoggerFunc(func(i ...interface{}) error {
		traceLogger.Info("message", i...)

		return nil
	}))

	tracer, con, err := jaegerConfig.NewTracer(
		jaeger_cnf.Metrics(jaegerProm),
		jaeger_cnf.Logger(l),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create tracing")
	}

	opentracing.SetGlobalTracer(tracer)

	traceLogger.Info("Initialized", "host", jaegerConfig.Reporter.LocalAgentHostPort)

	return con, err
}

func New(ctx context.Context) (io.Closer, error) {
	jaegerConfig := &jaeger_cnf.Configuration{}

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
	}

	jaegerConfig, err = jaegerConfig.FromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "unable to configure from env")
	}

	jaegerConfig.Tags = append(jaegerConfig.Tags, opentracing.Tag{
		Key:   "version",
		Value: application.GetVersion(ctx),
	})

	if jaegerConfig.ServiceName == "" {
		jaegerConfig.ServiceName = application.GetName(ctx)
	}

	traCon, err := Init(ctx, jaegerConfig)

	return traCon, errors.Wrap(err, "unable to init tracer")
}
