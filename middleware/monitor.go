package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
		[]string{"method", "endpoint"},
	)
	RequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func MonitorMiddleware(c *gin.Context) {
	start := time.Now()
	method := c.Request.Method
	path := c.Request.URL.Path

	c.Next()

	duration := time.Since(start).Seconds()
	RequestCount.WithLabelValues(method, path).Inc()
	RequestLatency.WithLabelValues(method, path).Observe(duration)
}
