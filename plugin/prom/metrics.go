package prom

import "github.com/prometheus/client_golang/prometheus"

type EndPoint string

const (
	Server EndPoint = "server"
	Client EndPoint = "client"
)

func newDefaultMetrics(point EndPoint, labels map[string]string) *DefaultMetrics {
	if point == Client {
		return &DefaultMetrics{
			point:       point,
			constLabels: labels,
			startedCounter: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name:        "xrpc_client_started_total",
					Help:        "Total number of RPCs started on the client.",
					ConstLabels: labels,
				}, []string{"xrpc_type", "xrpc_service", "xrpc_method"}),
			handledCounter: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name:        "xrpc_client_handled_total",
					Help:        "Total number of RPCs completed on the client, regardless of success or failure.",
					ConstLabels: labels,
				}, []string{"xrpc_type", "xrpc_service", "xrpc_method", "xrpc_code"}),
			sampleCounter: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name:        "xrpc_server_sample_times",
					Help:        "Gauge of response latency (seconds) of xrpc that had been application-level handled by the client.",
					ConstLabels: labels,
				}),
		}
	}
	return &DefaultMetrics{
		point:       point,
		constLabels: labels,
		startedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "xrpc_server_started_total",
				Help:        "Total number of RPCs started on the server.",
				ConstLabels: labels,
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method"}),
		handledCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "xrpc_server_handled_total",
				Help:        "Total number of RPCs completed on the server, regardless of success or failure.",
				ConstLabels: labels,
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method", "xrpc_code"}),
		sampleCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:        "xrpc_server_sample_times",
				Help:        "Gauge of response latency (seconds) of xrpc that had been application-level handled by the server.",
				ConstLabels: labels,
			}),
	}
}

type DefaultMetrics struct {
	point EndPoint

	startedCounter   *prometheus.CounterVec
	handledCounter   *prometheus.CounterVec
	handledGauge     *prometheus.GaugeVec
	sampleCounter    prometheus.Counter
	handledHistogram *prometheus.HistogramVec

	enableDelay bool
	constLabels map[string]string
}

func (dm *DefaultMetrics) EnableDelay(buckets []float64) {
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	dm.enableDelay = true
	if dm.point == Client {
		dm.handledHistogram = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "xrpc_client_handling_seconds",
				Help:        "Histogram of response latency (seconds) of xrpc that had been application-level handled by the client.",
				ConstLabels: dm.constLabels,
				Buckets:     buckets,
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method"})
		dm.handledGauge = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "xrpc_client_handling_delay_sample",
				Help:        "Gauge of response latency (seconds) of xrpc that had been application-level handled by the client.",
				ConstLabels: dm.constLabels,
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method"})
	} else {
		dm.handledHistogram = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "xrpc_server_handling_seconds",
				Help:        "Histogram of response latency (seconds) of xrpc that had been application-level handled by the server.",
				ConstLabels: dm.constLabels,
				Buckets:     buckets,
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method"})
		dm.handledGauge = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "xrpc_server_handling_delay_sample",
				Help:        "Gauge of response latency (seconds) of xrpc that had been application-level handled by the server.",
				ConstLabels: dm.constLabels,
			}, []string{"xrpc_type", "xrpc_service", "xrpc_method"})
	}
}

func (dm *DefaultMetrics) Describe(ch chan<- *prometheus.Desc) {
	dm.startedCounter.Describe(ch)
	dm.handledCounter.Describe(ch)
	dm.sampleCounter.Describe(ch)
	if dm.enableDelay {
		dm.handledHistogram.Describe(ch)
		dm.handledGauge.Describe(ch)
	}
}

func (dm *DefaultMetrics) Collect(ch chan<- prometheus.Metric) {
	dm.startedCounter.Collect(ch)
	dm.handledCounter.Collect(ch)
	dm.sampleCounter.Inc()
	dm.sampleCounter.Collect(ch)
	if dm.enableDelay {
		dm.handledHistogram.Collect(ch)
		dm.handledGauge.Collect(ch)
	}
}
