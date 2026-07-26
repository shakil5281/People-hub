package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ListUsers godoc
//
// @Summary      List users
// @Description  List all users with pagination and filters
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        page     query     int     false  "Page number"
// @Param        limit    query     int     false  "Items per page"
// @Param        status   query     string  false  "Filter by status"
// @Param        search   query     string  false  "Search by email or name"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	p := utils.ParsePagination(c)
	status := c.Query("status")
	search := c.Query("search")

	users, total, err := h.userService.ListUsers(p.Page, p.Limit, status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(users, total, p))
}

// GetUser godoc
//
// @Summary      Get user by ID
// @Description  Get a single user with assigned roles
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, roles, err := h.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"status":     user.Status,
		"roles":      roles,
		"created_at": user.CreatedAt,
	})
}

// CreateUser godoc
//
// @Summary      Create user
// @Description  Create a new user and return the generated password
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body service.CreateUserRequest true "User creation data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdBy := c.GetString("user_id")
	user, password, err := h.userService.CreateUser(req, createdBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                    user.ID,
		"email":                 user.Email,
		"name":                  user.Name,
		"status":                user.Status,
		"generated_password":    password,
		"force_password_change": true,
		"message":               "User created. Share the generated password securely; user must change it on first login.",
	})
}

// UpdateUser godoc
//
// @Summary      Update user
// @Description  Update user name or status
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path  string                    true  "User ID"
// @Param        request  body  service.UpdateUserRequest  true  "Update data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(id, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"status":  user.Status,
		"message": "user updated successfully",
	})
}

// DeleteUser godoc
//
// @Summary      Delete user
// @Description  Soft-delete a user (status set to deleted)
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.userService.DeleteUser(id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// GetUserRoles godoc
//
// @Summary      Get user roles
// @Description  Get roles assigned to a user
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "User ID"
// @Success      200  {array}   models.Role
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id}/roles [get]
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	id := c.Param("id")
	roles, err := h.userService.GetUserRoles(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// AssignRoles godoc
//
// @Summary      Assign roles to user
// @Description  Replace all roles for a user
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path  string                    true  "User ID"
// @Param        request  body  service.AssignRolesRequest true  "Role IDs"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /users/{id}/roles [put]
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id := c.Param("id")
	var req service.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var currentRoles []string
	if r, ok := c.Get("roles"); ok {
		if roleList, ok := r.([]string); ok {
			currentRoles = roleList
		}
	}

	if err := h.userService.AssignRoles(id, req, currentRoles); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "roles assigned successfully"})
}

// AdminResetPassword godoc
//
// @Summary      Admin reset password
// @Description  Force reset a user's password and return the new one
// @Tags         Users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id}/reset-password [post]
func (h *UserHandler) AdminResetPassword(c *gin.Context) {
	id := c.Param("id")
	password, err := h.userService.AdminResetPassword(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"generated_password":    password,
		"force_password_change": true,
		"message":               "Password reset. Share the new password securely; user must change it on next login.",
	})
}

// ForgotPassword godoc
//
// @Summary      Forgot password
// @Description  Request a password reset token (public)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body service.ForgotPasswordRequest true "Email"
// @Success      200  {object}  map[string]interface{}
// @Router       /auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req service.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, _ := h.userService.ForgotPassword(req)
	// Token returned directly (no email configured). In production, email this instead.
	if token != "" {
		c.JSON(http.StatusOK, gin.H{
			"reset_token": token,
			"message":     "Use this token with POST /auth/reset-password to set a new password.",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset token has been generated."})
	}
}

// ResetPassword godoc
//
// @Summary      Reset password
// @Description  Reset password using a token (public)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body service.ResetPasswordRequest true "Token + new password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /auth/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ResetPassword(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}
