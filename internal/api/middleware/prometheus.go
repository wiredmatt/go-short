package middleware

import (
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "go_short_requests_total",
			Help: "Total number of requests processed by the go_short web server.",
		},
		[]string{"path", "status"},
	)

	ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "go_short_requests_errors_total",
			Help: "Total number of error requests processed by the go_short web server.",
		},
		[]string{"path", "status"},
	)
)

func PrometheusInit() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(ErrorCount)
}

func TrackMetrics(ctx huma.Context, next func(huma.Context)) {
	next(ctx)

	path := ctx.Operation().Path
	status := ctx.Status()

	RequestCount.WithLabelValues(path, fmt.Sprintf("%d", status)).Inc()
	if status >= 400 {
		ErrorCount.WithLabelValues(path, fmt.Sprintf("%d", status)).Inc()
	}
}
