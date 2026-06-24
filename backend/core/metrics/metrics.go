package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight *prometheus.GaugeVec

	CacheHits    *prometheus.CounterVec
	CacheMisses  *prometheus.CounterVec
	CacheLatency *prometheus.HistogramVec

	ZentaoAPIRequests    *prometheus.CounterVec
	ZentaoAPIDuration    *prometheus.HistogramVec
	ZentaoAPIErrors      *prometheus.CounterVec
	ZentaoTokenRefreshes prometheus.Counter

	BugsTotal    *prometheus.GaugeVec
	StoriesTotal *prometheus.GaugeVec
	TasksTotal   *prometheus.GaugeVec
	TimelogTotal *prometheus.CounterVec
}

var globalMetrics *Metrics

func Init() error {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		RequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method"},
		),

		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type"},
		),
		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type"},
		),
		CacheLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "cache_operation_duration_seconds",
				Help:    "Cache operation duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
			},
			[]string{"cache_type", "operation"},
		),

		ZentaoAPIRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zentao_api_requests_total",
				Help: "Total number of Zentao API requests",
			},
			[]string{"endpoint", "method"},
		),
		ZentaoAPIDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zentao_api_duration_seconds",
				Help:    "Zentao API request duration in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
			},
			[]string{"endpoint", "method"},
		),
		ZentaoAPIErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zentao_api_errors_total",
				Help: "Total number of Zentao API errors",
			},
			[]string{"endpoint", "method", "error_type"},
		),
		ZentaoTokenRefreshes: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "zentao_token_refreshes_total",
				Help: "Total number of Zentao token refreshes",
			},
		),

		BugsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "bugs_total",
				Help: "Total number of bugs",
			},
			[]string{"product", "project", "status"},
		),
		StoriesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "stories_total",
				Help: "Total number of stories",
			},
			[]string{"product", "project", "status"},
		),
		TasksTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tasks_total",
				Help: "Total number of tasks",
			},
			[]string{"project", "execution", "status"},
		),
		TimelogTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "timelog_hours_total",
				Help: "Total hours logged",
			},
			[]string{"user", "project"},
		),
	}

	prometheus.MustRegister(
		m.RequestsTotal,
		m.RequestDuration,
		m.RequestsInFlight,
		m.CacheHits,
		m.CacheMisses,
		m.CacheLatency,
		m.ZentaoAPIRequests,
		m.ZentaoAPIDuration,
		m.ZentaoAPIErrors,
		m.ZentaoTokenRefreshes,
		m.BugsTotal,
		m.StoriesTotal,
		m.TasksTotal,
		m.TimelogTotal,
	)

	globalMetrics = m
	return nil
}

func Get() *Metrics {
	if globalMetrics == nil {
		panic("metrics not initialized, please call Init() first")
	}
	return globalMetrics
}

func Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if string(c.Request.URI().Path()) == "/metrics" {
			c.Next(ctx)
			return
		}

		start := time.Now()
		path := string(c.Request.URI().Path())

		Get().RequestsInFlight.WithLabelValues(string(c.Request.Method())).Inc()
		defer Get().RequestsInFlight.WithLabelValues(string(c.Request.Method())).Dec()

		c.Next(ctx)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response.StatusCode())

		Get().RequestsTotal.WithLabelValues(string(c.Request.Method()), path, status).Inc()
		Get().RequestDuration.WithLabelValues(string(c.Request.Method()), path).Observe(duration)
	}
}

func Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		stdReq, _ := adaptor.GetCompatRequest(&c.Request)
		stdResp := adaptor.GetCompatResponseWriter(&c.Response)
		promhttp.Handler().ServeHTTP(stdResp, stdReq)
	}
}

func RecordCacheHit(cacheType string) {
	Get().CacheHits.WithLabelValues(cacheType).Inc()
}

func RecordCacheMiss(cacheType string) {
	Get().CacheMisses.WithLabelValues(cacheType).Inc()
}

func RecordCacheOperation(cacheType, operation string, duration time.Duration) {
	Get().CacheLatency.WithLabelValues(cacheType, operation).Observe(duration.Seconds())
}

func RecordZentaoAPIRequest(endpoint, method string, duration time.Duration, err error) {
	m := Get()
	m.ZentaoAPIRequests.WithLabelValues(endpoint, method).Inc()
	m.ZentaoAPIDuration.WithLabelValues(endpoint, method).Observe(duration.Seconds())

	if err != nil {
		errorType := "unknown"
		if err != nil {
			errorType = "error"
		}
		m.ZentaoAPIErrors.WithLabelValues(endpoint, method, errorType).Inc()
	}
}

func RecordTokenRefresh() {
	Get().ZentaoTokenRefreshes.Inc()
}

func UpdateBugsTotal(product, project, status string, count float64) {
	Get().BugsTotal.WithLabelValues(product, project, status).Set(count)
}

func UpdateStoriesTotal(product, project, status string, count float64) {
	Get().StoriesTotal.WithLabelValues(product, project, status).Set(count)
}

func UpdateTasksTotal(project, execution, status string, count float64) {
	Get().TasksTotal.WithLabelValues(project, execution, status).Set(count)
}

func RecordTimelog(user, project string, hours float64) {
	Get().TimelogTotal.WithLabelValues(user, project).Add(hours)
}

func GetCacheHitRate(cacheType string) float64 {
	return 0.0
}
