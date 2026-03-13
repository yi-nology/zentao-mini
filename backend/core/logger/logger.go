package logger

import (
	"context"
	"os"
	"strings"

	"chandao-mini/backend/core/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局日志实例
var globalLogger *zap.Logger
var sugarLogger *zap.SugaredLogger

// ContextKey 用于在context中存储trace ID
type ContextKey string

const (
	TraceIDKey ContextKey = "trace_id"
)

// Init 初始化日志
func Init(cfg *config.LogConfig) error {
	// 创建日志编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据日志格式选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		// 生产环境使用JSON格式
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// 开发环境使用Console格式（彩色输出）
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志级别
	level := getZapLevel(cfg.Level)

	// 创建核心组件
	core := zapcore.NewCore(
		encoder,
		getOutputWriter(cfg.OutputPath),
		level,
	)

	// 创建logger选项
	options := []zap.Option{}
	if cfg.EnableCaller {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(1))
	}
	if cfg.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 创建logger实例
	globalLogger = zap.New(core, options...)
	sugarLogger = globalLogger.Sugar()

	return nil
}

// getZapLevel 将字符串日志级别转换为zap级别
func getZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// getOutputWriter 获取日志输出writer
func getOutputWriter(outputPath string) zapcore.WriteSyncer {
	if outputPath == "" || outputPath == "stdout" {
		return zapcore.AddSync(os.Stdout)
	}

	// 打开日志文件
	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// 如果打开文件失败，回退到标准输出
		return zapcore.AddSync(os.Stdout)
	}

	return zapcore.AddSync(file)
}

// GetLogger 获取全局logger实例
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置
		panic("logger not initialized, please call Init() first")
	}
	return globalLogger
}

// GetSugarLogger 获取SugarLogger实例
func GetSugarLogger() *zap.SugaredLogger {
	if sugarLogger == nil {
		panic("logger not initialized, please call Init() first")
	}
	return sugarLogger
}

// Sync 刷新日志缓冲
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// WithTraceID 添加追踪ID到日志字段
func WithTraceID(traceID string) *zap.Logger {
	return GetLogger().With(zap.String("trace_id", traceID))
}

// WithContext 从context中获取trace ID并添加到日志
func WithContext(ctx context.Context) *zap.Logger {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if !ok || traceID == "" {
		return GetLogger()
	}
	return WithTraceID(traceID)
}

// 便捷方法 - 使用全局logger

// Debug 记录Debug级别日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 记录Info级别日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 记录Warn级别日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 记录Error级别日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 记录Fatal级别日志并退出
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Debugf 使用格式化字符串记录Debug日志
func Debugf(template string, args ...interface{}) {
	GetSugarLogger().Debugf(template, args...)
}

// Infof 使用格式化字符串记录Info日志
func Infof(template string, args ...interface{}) {
	GetSugarLogger().Infof(template, args...)
}

// Warnf 使用格式化字符串记录Warn日志
func Warnf(template string, args ...interface{}) {
	GetSugarLogger().Warnf(template, args...)
}

// Errorf 使用格式化字符串记录Error日志
func Errorf(template string, args ...interface{}) {
	GetSugarLogger().Errorf(template, args...)
}

// Fatalf 使用格式化字符串记录Fatal日志并退出
func Fatalf(template string, args ...interface{}) {
	GetSugarLogger().Fatalf(template, args...)
}

// WithContextFromGin 从Gin Context中获取trace ID并添加到日志
func WithContextFromGin(c *gin.Context) *zap.Logger {
	traceID := c.GetString(string(TraceIDKey))
	if traceID == "" {
		return GetLogger()
	}
	return WithTraceID(traceID)
}

// LogRequest 记录HTTP请求日志
func LogRequest(c *gin.Context, statusCode int, latency int64, method, path string) {
	traceID := c.GetString(string(TraceIDKey))
	
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", statusCode),
		zap.Int64("latency_ms", latency),
		zap.String("client_ip", c.ClientIP()),
	}
	
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if statusCode >= 400 {
		Warn("HTTP Request", fields...)
	} else {
		Info("HTTP Request", fields...)
	}
}

// LogError 记录错误日志（带堆栈信息）
func LogError(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	Error(msg, allFields...)
}

// LogPanic 记录panic日志
func LogPanic(msg string, fields ...zap.Field) {
	fields = append(fields, zap.Stack("stack"))
	Error(msg, fields...)
}
