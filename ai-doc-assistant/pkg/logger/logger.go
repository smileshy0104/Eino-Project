package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

// Init 初始化日志
func Init(level, filepath string) {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 配置编码器
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	
	if filepath != "" {
		config.OutputPaths = append(config.OutputPaths, filepath)
	}

	// 创建logger
	zapLogger, err := config.Build()
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	logger = zapLogger.Sugar()
}

// S 返回SugaredLogger
func S() *zap.SugaredLogger {
	if logger == nil {
		// 使用默认配置初始化
		Init("info", "")
	}
	return logger
}

// Sync 同步日志
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}