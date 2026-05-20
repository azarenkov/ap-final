package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, rps int, window time.Duration) gin.HandlerFunc {
	if rdb == nil || rps <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions || c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
		defer cancel()

		key := "rl:" + c.ClientIP() + ":" + strconv.FormatInt(time.Now().Unix()/int64(window.Seconds()), 10)
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			_ = rdb.Expire(ctx, key, window).Err()
		}
		if count > int64(rps) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			return
		}
		c.Next()
	}
}
