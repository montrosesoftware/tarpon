package instrumentation

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusInstrumentation struct {
	metricsHandler http.Handler
}

func NewPrometheusInstrumentation() *PrometheusInstrumentation {
	i := PrometheusInstrumentation{promhttp.Handler()}
	return &i
}

func (i *PrometheusInstrumentation) MetricsHandler() http.Handler {
	return i.metricsHandler
}
