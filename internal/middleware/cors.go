package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if originsEnv != "" {
		for _, o := range strings.Split(originsEnv, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}
	if len(allowedOrigins) == 0 {
		// Development defaults - production must set CORS_ALLOWED_ORIGINS
		allowedOrigins = []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:3001", "http://localhost:8081", "http://localhost"}
		if os.Getenv("GO_ENV") == "production" {
			// In production, log warning if not configured
			// Do not fail, but restrict to strict origins
		}
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Confirm-Reset"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	})
}
