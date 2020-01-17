package prom

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func newDefaultReporter(c prometheus.Collector, params ...string) *defaultReporter {
	metrics, ok := c.(*DefaultMetrics)
	if !ok {
		return nil
	}
	if len(params) < 2 {
		return nil
	}
	rpcType := params[0]
	fullMethod := params[1]
	serviceName, methodName := splitMethodName(fullMethod)
	metrics.startedCounter.WithLabelValues(rpcType, serviceName, methodName).Inc()
	r := &defaultReporter{
		rpcType:   rpcType,
		service:   serviceName,
		method:    methodName,
		startTime: time.Now(),
		metrics:   metrics,
	}
	return r
}

type defaultReporter struct {
	rpcType   string
	service   string
	method    string
	startTime time.Time
	metrics   *DefaultMetrics
}

// Handled 更新metrics信息
func (r *defaultReporter) Handled(code string) {
	if r.metrics.promEndPoint == Server {
		r.metrics.serverHandledCounter.WithLabelValues(r.rpcType, r.service, r.method, code).Inc()
		r.metrics.serverHandledHistogram.WithLabelValues(r.rpcType, r.service, r.method).Observe(float64(time.Since(r.startTime).Milliseconds()))
		return
	}
	r.metrics.clientHandledCounter.WithLabelValues(r.rpcType, r.service, r.method, code).Inc()
	r.metrics.clientHandledHistogram.WithLabelValues(r.rpcType, r.service, r.method).Observe(float64(time.Since(r.startTime).Milliseconds()))
}
