package middleware

import (
	"context"
	"fmt"
	"net/http"
	"order-payment-system/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// redis+lua实现令牌桶限流
const rateLimitLua = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local last_tokens = tonumber(redis.call("hget", key, "tokens"))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("hget", key, "last_refreshed"))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now - last_refreshed)
local filled_tokens = math.min(capacity, last_tokens + (delta * rate))

local allowed = filled_tokens >= requested
local new_tokens = filled_tokens

if allowed then
    new_tokens = filled_tokens - requested
end

redis.call("hset", key, "tokens", new_tokens)
redis.call("hset", key, "last_refreshed", now)
-- 设置过期时间，防止死key堆积 (稍微大一点以容错)
redis.call("expire", key, math.ceil(capacity/rate) + 10)

return {allowed and 1 or 0, new_tokens}
`

// RateLimit 令牌桶限流中间件
// capacity: 桶的容量 (最大并发或存量)
// rate: 每秒生成令牌的数量
func RateLimit(capacity int, rate float64, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		now := time.Now().Unix()

		res, err := rdb.Eval(context.Background(), rateLimitLua, []string{key}, capacity, rate, now, 1).Result()

		if err != nil {
			c.Next()
			return
		}

		result, ok := res.([]interface{})
		if !ok || len(result) == 0 {
			c.Next()
			return
		}

		allowed, ok := result[0].(int64)
		if ok && allowed != 1 {
			response.Error(c, http.StatusTooManyRequests, 429, "当前访问人数过多，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
