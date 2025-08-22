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

	metricsInitialized = false
)

func PrometheusInit() {
	if metricsInitialized {
		return
	}

	// Try to register the collectors, but don't panic if they're already registered
	if err := prometheus.Register(RequestCount); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			panic(err)
		}
	}

	if err := prometheus.Register(ErrorCount); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			panic(err)
		}
	}

	metricsInitialized = true
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
