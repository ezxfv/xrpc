package prom

import "github.com/prometheus/client_golang/prometheus"

type EndPoint string

const (
	Server EndPoint = "server"
	Client EndPoint = "client"
)

func newDefaultMetrics(point EndPoint) *DefaultMetrics {
	if point == Client {
		return &DefaultMetrics{
			startedCounter: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "xrpc_client_started_total",
					Help: "Total number of RPCs started on the client.",
				}, []string{"xrpc_type", "xrpc_service", "xrpc_method"}),
			handledCounter: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "xrpc_client_handled_total",
					Help: "Total number of RPCs completed on the client, regardless of success or failure.",
				}, []string{"xrpc_type", "xrpc_service", "xrpc_method", "xrpc_code"}),
			handledHistogram: prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "xrpc_client_handling_seconds",
					Help:    "Histogram of response latency (seconds) of gRPC that had been application-level handled by the client.",
					Buckets: prometheus.DefBuckets,
				}, []string{"xrpc_type", "xrpc_service", "xrpc_method"}),
		}
	}
	return &DefaultMetrics{
		startedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "xrpc_server_started_total",
				Help: "Total number of RPCs started on the server.",
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method"}),
		handledCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "xrpc_server_handled_total",
				Help: "Total number of RPCs completed on the server, regardless of success or failure.",
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method", "xrpc_code"}),
		handledHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "xrpc_server_handling_seconds",
				Help:    "Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.",
				Buckets: prometheus.DefBuckets,
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method"}),
	}
}

type DefaultMetrics struct {
	startedCounter   *prometheus.CounterVec
	handledCounter   *prometheus.CounterVec
	handledHistogram *prometheus.HistogramVec
}

func (dm *DefaultMetrics) Describe(ch chan<- *prometheus.Desc) {
	dm.startedCounter.Describe(ch)
	dm.handledCounter.Describe(ch)
	dm.handledHistogram.Describe(ch)
}

func (dm *DefaultMetrics) Collect(ch chan<- prometheus.Metric) {
	dm.startedCounter.Collect(ch)
	dm.handledCounter.Collect(ch)
	dm.handledHistogram.Collect(ch)
}
