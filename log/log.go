package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

type Config struct {
	OutputPaths []string
	ErrorOutputPaths []string
	Development bool
}

func Init(c *Config) {
	var err error
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = c.OutputPaths
	cfg.ErrorOutputPaths=c.OutputPaths
	cfg.Development = c.Development
	cfg.DisableStacktrace=true
	cfg.DisableCaller=true
	cfg.EncoderConfig.EncodeTime = TimeEncoder
	logger, err = cfg.Build()
	logger.WithOptions()
	if err != nil {
		return
	}
	defer logger.Sync()
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339Nano))
}

// Info send log info to logstash
func Info(msg string, fields ...zap.Field) {
		logger.Info(msg, fields...)
}

// Error send log error to logstash
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Warn send log warn to logstash
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Debug send log debug to logstash
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Fatal send log fatal to logstash
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
