package middleware

import (
	"time"

	"chandao-mini/backend/core/logger"
	"chandao-mini/backend/core/metrics"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceIDMiddleware 请求追踪ID中间件
// 为每个请求生成唯一的追踪ID，便于日志追踪和问题排查
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取trace ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			// 如果请求头没有，生成新的trace ID
			traceID = uuid.New().String()
		}

		// 将trace ID存入context
		c.Set(string(logger.TraceIDKey), traceID)

		// 将trace ID添加到响应头
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
// 记录请求日志，包括请求方法、路径、状态码、耗时等信息
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		traceID := c.GetString(string(logger.TraceIDKey))

		// 记录请求日志
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("trace_id", traceID),
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Error("HTTP Request", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP Request", fields...)
		} else {
			logger.Info("HTTP Request", fields...)
		}
	}
}

// MetricsMiddleware 性能监控中间件
// 收集HTTP请求的性能指标
func MetricsMiddleware() gin.HandlerFunc {
	return metrics.Middleware()
}

// RecoveryMiddleware 恢复中间件
// 捕获panic并记录日志，防止服务崩溃
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := c.GetString(string(logger.TraceIDKey))

				// 记录panic日志
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("trace_id", traceID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Stack("stack"),
				)

				// 返回500错误
				c.AbortWithStatusJSON(500, gin.H{
					"error":    "Internal Server Error",
					"trace_id": traceID,
				})
			}
		}()

		c.Next()
	}
}
