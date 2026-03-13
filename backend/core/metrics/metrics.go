package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics 性能指标收集器
type Metrics struct {
	// HTTP请求相关指标
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight *prometheus.GaugeVec

	// 缓存相关指标
	CacheHits    *prometheus.CounterVec
	CacheMisses  *prometheus.CounterVec
	CacheLatency *prometheus.HistogramVec

	// 禅道API相关指标
	ZentaoAPIRequests    *prometheus.CounterVec
	ZentaoAPIDuration    *prometheus.HistogramVec
	ZentaoAPIErrors      *prometheus.CounterVec
	ZentaoTokenRefreshes prometheus.Counter

	// 业务指标
	BugsTotal       *prometheus.GaugeVec
	StoriesTotal    *prometheus.GaugeVec
	TasksTotal      *prometheus.GaugeVec
	TimelogTotal    *prometheus.CounterVec
}

// 全局metrics实例
var globalMetrics *Metrics

// Init 初始化性能指标收集器
func Init() error {
	m := &Metrics{
		// HTTP请求指标
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

		// 缓存指标
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

		// 禅道API指标
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

		// 业务指标
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

	// 注册所有指标
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

// Get 获取全局metrics实例
func Get() *Metrics {
	if globalMetrics == nil {
		panic("metrics not initialized, please call Init() first")
	}
	return globalMetrics
}

// Middleware Gin中间件，用于收集HTTP请求指标
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// 增加正在处理的请求数
		Get().RequestsInFlight.WithLabelValues(c.Request.Method).Inc()
		defer Get().RequestsInFlight.WithLabelValues(c.Request.Method).Dec()

		// 处理请求
		c.Next()

		// 记录请求指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		Get().RequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		Get().RequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// Handler 返回Prometheus metrics handler
func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit(cacheType string) {
	Get().CacheHits.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss(cacheType string) {
	Get().CacheMisses.WithLabelValues(cacheType).Inc()
}

// RecordCacheOperation 记录缓存操作耗时
func RecordCacheOperation(cacheType, operation string, duration time.Duration) {
	Get().CacheLatency.WithLabelValues(cacheType, operation).Observe(duration.Seconds())
}

// RecordZentaoAPIRequest 记录禅道API请求
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

// RecordTokenRefresh 记录Token刷新
func RecordTokenRefresh() {
	Get().ZentaoTokenRefreshes.Inc()
}

// UpdateBugsTotal 更新Bug总数
func UpdateBugsTotal(product, project, status string, count float64) {
	Get().BugsTotal.WithLabelValues(product, project, status).Set(count)
}

// UpdateStoriesTotal 更新需求总数
func UpdateStoriesTotal(product, project, status string, count float64) {
	Get().StoriesTotal.WithLabelValues(product, project, status).Set(count)
}

// UpdateTasksTotal 更新任务总数
func UpdateTasksTotal(project, execution, status string, count float64) {
	Get().TasksTotal.WithLabelValues(project, execution, status).Set(count)
}

// RecordTimelog 记录工时
func RecordTimelog(user, project string, hours float64) {
	Get().TimelogTotal.WithLabelValues(user, project).Add(hours)
}

// GetCacheHitRate 计算缓存命中率
func GetCacheHitRate(cacheType string) float64 {
	// 注意：这需要从Counter中获取值，Prometheus客户端库不直接支持读取Counter值
	// 这个函数主要用于演示，实际使用时应该通过Prometheus查询
	return 0.0
}
