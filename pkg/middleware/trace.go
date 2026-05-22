package middleware

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/gin-gonic/gin"
)

const TraceIDKey = "trace_id"

// generateTraceID 生成简单的 TraceID 避免引入第三方库
func generateTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// TraceMiddleware generates a TraceID for every request and sets it in context
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()

			// Also create a derived context.Context with the trace info so it can be passed downward
			ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
			c.Request = c.Request.WithContext(ctx)

			c.Header("X-Trace-ID", traceID)
			c.Next()
		}
	}
}

// GetTraceID extracts TraceID from context
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(TraceIDKey).(string); ok {
		return val
	}
	return ""
}
