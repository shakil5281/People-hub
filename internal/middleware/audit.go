package middleware

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

// AuditMiddleware logs mutating API requests to the audit_logs table.
// It captures POST, PUT, PATCH, DELETE operations with request/response metadata.
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip non-mutating methods and public health checks
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}
		if c.Request.URL.Path == "/health" || strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}

		// Capture request body (up to 16KB) for audit
		// Skip for multipart/form-data (file uploads) — body must remain intact for FormFile parsing
		var reqBody []byte
		if c.Request.Body != nil && !strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
			reqBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, 16*1024))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// Execute handler and capture response status
		c.Next()

		// Only audit successful mutating requests (2xx)
		if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			return
		}

		userID, _ := c.Get("user_id")
		companyID, _ := c.Get("company_id")

		action := method
		resource := extractResourceName(c.Request.URL.Path)
		resourceID := extractResourceID(c.Request.URL.Path)

		var uid, cid *string
		if id, ok := userID.(string); ok && id != "" {
			uid = &id
		}
		if id, ok := companyID.(string); ok && id != "" {
			cid = &id
		}

		// Truncate request body if too large
		oldValue := []byte("{}")
		newValue := []byte("{}")
		if len(reqBody) > 0 && len(reqBody) < 8*1024 {
			newValue = reqBody
		}

		audit := models.AuditLog{
			UserID:     uid,
			CompanyID:  cid,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			OldValue:   oldValue,
			NewValue:   newValue,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			CreatedAt:  time.Now(),
		}

		// Fire-and-forget: do not block response on audit write failure
		go func() {
			database.DB.Create(&audit)
		}()
	}
}

var resourcePattern = regexp.MustCompile(`^/api/v1/([^/]+)(?:/([^/]+))?.*$`)

func extractResourceName(path string) string {
	matches := resourcePattern.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return path
}

func extractResourceID(path string) string {
	matches := resourcePattern.FindStringSubmatch(path)
	if len(matches) > 2 && matches[2] != "" {
		// Check if second segment is a UUID or ID; skip if it looks like a sub-resource action
		if matches[2] == "approve" || matches[2] == "reject" || matches[2] == "template" || matches[2] == "import" || matches[2] == "export" || matches[2] == "process" || matches[2] == "stats" || matches[2] == "backup" || matches[2] == "backups" || matches[2] == "reset" || matches[2] == "sheet" || matches[2] == "payslip" || matches[2] == "list" || matches[2] == "summary" || matches[2] == "monthly" || matches[2] == "monthly-report" || matches[2] == "overtime" || matches[2] == "overtime-summary" || matches[2] == "job-card" || matches[2] == "missing" || matches[2] == "absent" {
			return ""
		}
		return matches[2]
	}
	return ""
}
