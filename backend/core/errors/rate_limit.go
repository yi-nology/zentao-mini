package errors

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 请求限流器
type RateLimiter struct {
	requests map[string]*ClientInfo
	mu       sync.RWMutex
	config   RateLimitConfig
}

// ClientInfo 客户端请求信息
type ClientInfo struct {
	Count     int
	ResetTime time.Time
	Blocked   bool
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerMinute int           // 每分钟允许的请求数
	BlockDuration     time.Duration // 超限后的封禁时长
	CleanupInterval   time.Duration // 清理过期记录的间隔
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() RateLimitConfig {
	// 从环境变量读取配置
	requestsPerMinute := 60 // 默认每分钟60次
	if val := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			requestsPerMinute = num
		}
	}

	blockDuration := 5 * time.Minute // 默认封禁5分钟
	if val := os.Getenv("RATE_LIMIT_BLOCK_DURATION_MINUTES"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			blockDuration = time.Duration(num) * time.Minute
		}
	}

	return RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		BlockDuration:     blockDuration,
		CleanupInterval:   1 * time.Minute,
	}
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	limiter := &RateLimiter{
		requests: make(map[string]*ClientInfo),
		config:   config,
	}

	// 启动后台清理任务
	go limiter.cleanupTask()

	return limiter
}

// cleanupTask 定期清理过期记录
func (rl *RateLimiter) cleanupTask() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, info := range rl.requests {
			// 清理已过期的记录
			if now.After(info.ResetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[ip]

	if !exists {
		// 首次请求
		rl.requests[ip] = &ClientInfo{
			Count:     1,
			ResetTime: now.Add(1 * time.Minute),
			Blocked:   false,
		}
		return true, rl.config.RequestsPerMinute - 1, rl.requests[ip].ResetTime
	}

	// 检查是否被封禁
	if info.Blocked {
		if now.Before(info.ResetTime) {
			// 仍在封禁期内
			return false, 0, info.ResetTime
		}
		// 封禁期已过，重置
		info.Blocked = false
		info.Count = 0
		info.ResetTime = now.Add(1 * time.Minute)
	}

	// 检查是否需要重置计数器（新的一分钟）
	if now.After(info.ResetTime) {
		info.Count = 0
		info.ResetTime = now.Add(1 * time.Minute)
	}

	// 检查是否超过限制
	if info.Count >= rl.config.RequestsPerMinute {
		// 超过限制，封禁IP
		info.Blocked = true
		info.ResetTime = now.Add(rl.config.BlockDuration)
		log.Printf("[RATE_LIMIT] IP %s blocked for %v due to exceeding rate limit", ip, rl.config.BlockDuration)
		return false, 0, info.ResetTime
	}

	// 允许请求，增加计数
	info.Count++
	remaining := rl.config.RequestsPerMinute - info.Count
	return true, remaining, info.ResetTime
}

// RateLimitMiddleware 请求限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	limiter := NewRateLimiter(config)

	log.Printf("[RATE_LIMIT] Rate limiter initialized: %d requests/minute, block duration: %v",
		config.RequestsPerMinute, config.BlockDuration)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 检查是否允许请求
		allowed, remaining, resetTime := limiter.Allow(ip)

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			log.Printf("[RATE_LIMIT] Request blocked from IP: %s, Path: %s", ip, c.Request.URL.Path)
			c.JSON(http.StatusTooManyRequests, Response{
				Code:    http.StatusTooManyRequests,
				Message: "请求过于频繁，请稍后再试",
				Data: gin.H{
					"retryAfter": resetTime.Sub(time.Now()).Seconds(),
					"resetAt":    resetTime.Format(time.RFC3339),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddlewareWithConfig 带自定义配置的请求限流中间件
func RateLimitMiddlewareWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)

	log.Printf("[RATE_LIMIT] Rate limiter initialized with custom config: %d requests/minute, block duration: %v",
		config.RequestsPerMinute, config.BlockDuration)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 检查是否允许请求
		allowed, remaining, resetTime := limiter.Allow(ip)

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			log.Printf("[RATE_LIMIT] Request blocked from IP: %s, Path: %s", ip, c.Request.URL.Path)
			c.JSON(http.StatusTooManyRequests, Response{
				Code:    http.StatusTooManyRequests,
				Message: "请求过于频繁，请稍后再试",
				Data: gin.H{
					"retryAfter": resetTime.Sub(time.Now()).Seconds(),
					"resetAt":    resetTime.Format(time.RFC3339),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
