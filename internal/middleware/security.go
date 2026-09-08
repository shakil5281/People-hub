package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds production security headers without breaking frontend functionality.
// CSP is permissive for Next.js which uses inline scripts; tighten if you add nonce/hashes.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		// Prevent clickjacking
		h.Set("X-Frame-Options", "DENY")
		// Prevent MIME sniffing
		h.Set("X-Content-Type-Options", "nosniff")
		// XSS filter (legacy but harmless)
		h.Set("X-XSS-Protection", "0")
		// Referrer policy
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Permissions policy - restrict sensitive features
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// HSTS - only if HTTPS, enable when behind TLS terminator
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		// Basic CSP allowing self, inline scripts for Next.js, and images
		// Adjust if you add external CDNs
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob: http: https:; font-src 'self' data:; connect-src 'self' http: https:; frame-ancestors 'none'")
		c.Next()
	}
}
