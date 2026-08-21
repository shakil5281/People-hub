package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/service"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/health" || strings.HasPrefix(path, "/swagger") {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		level := "info"
		if status >= 500 {
			level = "error"
		} else if status >= 400 {
			level = "warn"
		}

		userID, _ := c.Get("user_id")
		companyID, _ := c.Get("company_id")
		var uid, cid *string
		if id, ok := userID.(string); ok && id != "" {
			uid = &id
		}
		if id, ok := companyID.(string); ok && id != "" {
			cid = &id
		}

		latMs := latency.Milliseconds()
		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		msg := fmt.Sprintf("%s %s | %d | %v", method, path, status, latency)

		service.WriteRequestLog(level, "api", msg, ip, ua, path, method, uid, cid, status, latMs)
	}
}
