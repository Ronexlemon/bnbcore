package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)


type MetricsHandler struct{
	Server         *http.ServeMux

}

func NewMetrics(server *http.ServeMux)*MetricsHandler{
	m:= &MetricsHandler{
		Server: server,
	}
	m.registerRoutes()

	return m
}

func (h *MetricsHandler) registerRoutes() {

	h.Server.Handle("/metrics", promhttp.Handler())
	
}
