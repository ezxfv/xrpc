package trace

import (
	"io"
	"log"

	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

func registerJaeger() (c io.Closer) {
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		log.Printf("Could not parse Jaeger env vars: %s", err.Error())
		return
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	tracer, c, err := cfg.NewTracer(jaegercfg.Logger(jLogger), jaegercfg.Metrics(jMetricsFactory))
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}

	opentracing.SetGlobalTracer(tracer)
	return
}
