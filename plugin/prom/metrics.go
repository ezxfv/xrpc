package prom

import "github.com/prometheus/client_golang/prometheus"

type EndPoint string

const (
	Server EndPoint = "server"
	Client EndPoint = "client"
)

type DefaultMetrics struct {
	promEndPoint EndPoint

	// 标准的数值上报，上报进程级别的计数类型测量数据
	startedCounter *prometheus.CounterVec
	// 服务端数值上报，上报作为Server时的计数类型测量数据
	serverHandledCounter *prometheus.CounterVec
	// 客户端数值上报，上报作为Client时的计数类型测量数据
	clientHandledCounter *prometheus.CounterVec

	serverHandledHistogram *prometheus.HistogramVec

	clientHandledHistogram *prometheus.HistogramVec
}

func (dm *DefaultMetrics) Describe(ch chan<- *prometheus.Desc) {
	dm.startedCounter.Describe(ch)
	if dm.promEndPoint == Server {
		dm.serverHandledCounter.Describe(ch)
		dm.serverHandledHistogram.Describe(ch)
		return
	}
	dm.clientHandledCounter.Describe(ch)
	dm.clientHandledHistogram.Describe(ch)
}

func (dm *DefaultMetrics) Collect(ch chan<- prometheus.Metric) {
	dm.startedCounter.Collect(ch)
	if dm.promEndPoint == Server {
		dm.serverHandledCounter.Collect(ch)
		dm.serverHandledHistogram.Collect(ch)
		return
	}
	dm.clientHandledCounter.Collect(ch)
	dm.clientHandledHistogram.Collect(ch)
}
