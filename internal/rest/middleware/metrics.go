package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	namespace = "realworld"
	subsystem = "http_incoming"
)

var (
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_histogram_seconds",
		Help:      "Request time duration.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"code", "method"})

	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "Total number of requests received.",
	}, []string{"code", "method"})

	requestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_size_histogram_bytes",
			Help:      "Request size in bytes.",
			Buckets:   []float64{100, 1000, 2000, 5000, 10000},
		}, []string{},
	)

	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "response_size_histogram_bytes",
			Help:      "Response size in bytes.",
			Buckets:   []float64{100, 1000, 2000, 5000, 10000},
		}, []string{},
	)

	inflight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "in_flight_requests",
			Help:      "Number of http requests which are currently running.",
		},
	)
)

type (
	MetricMiddleware struct {
		fn gin.HandlerFunc
	}
)

func (m MetricMiddleware) GinHandlerFunc() gin.HandlerFunc {
	return m.fn
}

func NewMetricMiddleware() MetricMiddleware {
	var next http.HandlerFunc
	metricsH := promhttp.InstrumentHandlerInFlight(
		inflight,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}),
	)
	metricsH = promhttp.InstrumentHandlerResponseSize(responseSize, metricsH)
	metricsH = promhttp.InstrumentHandlerRequestSize(requestSize, metricsH)
	metricsH = promhttp.InstrumentHandlerCounter(requestsTotal, metricsH)
	metricsH = promhttp.InstrumentHandlerDuration(requestDuration, metricsH)
	metricsChain := gin.WrapH(metricsH)

	return MetricMiddleware{
		fn: func(ctx *gin.Context) {
			next = func(w http.ResponseWriter, r *http.Request) {
				ctx.Next()
			}
			metricsChain(ctx)
		},
	}
}
