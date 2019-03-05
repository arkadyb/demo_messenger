package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// middleware handler for Prometheus
func PrometheusMiddleware(handler http.Handler) http.Handler {
	return promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handler)
}
