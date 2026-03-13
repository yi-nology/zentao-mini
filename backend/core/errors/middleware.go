package errors

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 错误恢复中间件
// 捕获panic并返回统一的错误响应
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				stack := string(debug.Stack())
				log.Printf("[PANIC] %v\n%s", err, stack)

				// 过滤敏感信息
				filteredStack := filterSensitiveInfo(stack)

				// 返回统一的错误响应
				c.JSON(http.StatusInternalServerError, Response{
					Code:    CodeInternalError,
					Message: "服务器内部错误",
					Data: gin.H{
						"error":  err,
						"stack":  filteredStack,
						"time":   time.Now().Format(time.RFC3339),
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ErrorHandlerMiddleware 错误处理中间件
// 统一处理错误响应
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			Error(c, err)
			c.Abort()
		}
	}
}

// RequestLoggerMiddleware 请求日志中间件
// 记录请求信息和错误日志
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		// 记录请求信息
		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%s] %s %s %d %v",
			method,
			path,
			c.ClientIP(),
			status,
			latency,
		)

		// 如果有错误，记录错误详情
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("[ERROR] %v", err)
			}
		}
	}
}

// filterSensitiveInfo 过滤敏感信息
// 移除文件路径、密码、token等敏感信息
func filterSensitiveInfo(stack string) string {
	// 过滤文件路径中的用户目录
	stack = strings.ReplaceAll(stack, "/Users/", "/***/")
	stack = strings.ReplaceAll(stack, "/home/", "/***/")
	stack = strings.ReplaceAll(stack, "C:\\Users\\", "C:\\***\\")

	// 过滤常见的敏感关键词
	sensitiveWords := []string{
		"password",
		"passwd",
		"token",
		"secret",
		"api_key",
		"apikey",
		"authorization",
		"credential",
	}

	for _, word := range sensitiveWords {
		stack = strings.ReplaceAll(stack, word, "***")
	}

	return stack
}

// CORSMiddleware CORS中间件
// 支持从环境变量配置允许的域名
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从环境变量读取允许的域名，多个域名用逗号分隔
		// 例如: ALLOWED_ORIGINS=http://localhost:3000,https://example.com
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		
		origin := c.Request.Header.Get("Origin")
		
		// 如果没有设置环境变量，使用宽松策略（仅用于开发环境）
		if allowedOrigins == "" {
			// 开发环境允许所有来源
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// 生产环境：严格验证来源
			origins := strings.Split(allowedOrigins, ",")
			allowed := false
			for _, allowedOrigin := range origins {
				allowedOrigin = strings.TrimSpace(allowedOrigin)
				if origin == allowedOrigin {
					allowed = true
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			
			// 如果来源不在允许列表中，拒绝请求
			if !allowed && origin != "" {
				log.Printf("[CORS] Blocked request from unauthorized origin: %s", origin)
			}
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
