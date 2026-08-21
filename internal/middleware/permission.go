package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "no permissions found"})
			c.Abort()
			return
		}

		perms, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid permissions"})
			c.Abort()
			return
		}

		if !slices.Contains(perms, permission) && !slices.Contains(perms, "*") {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "no roles found"})
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid roles"})
			c.Abort()
			return
		}

		if !slices.Contains(roleList, role) &&
			!slices.Contains(roleList, "superadmin") &&
			!slices.Contains(roleList, "super_admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}
