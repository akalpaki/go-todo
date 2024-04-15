package web

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingResponseWritter struct {
	http.ResponseWriter
	Status int
}

func (rw *loggingResponseWritter) WriteHeader(s int) {
	rw.Status = s
	rw.ResponseWriter.WriteHeader(s)
}

func Access(next http.HandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := loggingResponseWritter{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}
		next(&lrw, r)
		logger.Info("access", "endpoint", r.URL.Path, "method", r.Method, "status", lrw.Status, "latency", time.Since(start))
	}
}
