package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var totalReq = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests",
	},
	[]string{"path"},
)

func init() {
	prometheus.Register(totalReq)
}

func Prom(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
		totalReq.WithLabelValues("path").Inc()
	})
}
