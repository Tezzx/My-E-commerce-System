package logger

import (
	"context"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Ctx 提取上下文中的 TraceID 并返回绑定了该字段的子 Logger
func Ctx(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Log
	}

	val := ctx.Value("trace_id")
	if traceID, ok := val.(string); ok && traceID != "" {
		return Log.With(zap.String("trace_id", traceID))
	}
	return Log
}

func Init() error {
	env := strings.ToLower(os.Getenv("ENV"))

	var cfg zap.Config
	if env == "dev" || env == "development" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	l, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = l
	return nil
}
