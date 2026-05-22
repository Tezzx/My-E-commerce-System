package middleware

import (
	"time"

	"order-payment-system/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		cost := time.Since(start)

		logger.Log.Info("http request",
			zap.String("trace_id", GetTraceID(c.Request.Context())),
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("cost", cost),
		)
	}
}
