package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
//
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body service.LoginRequest true "Login credentials"
// @Success      200  {object}  service.LoginResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	resp, err := h.authService.Login(req, ip, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
//
// @Summary      Refresh access token
// @Description  Exchange refresh token for new access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body object true "Refresh token"
// @Success      200  {object}  service.LoginResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout godoc
//
// @Summary      Logout user
// @Description  Revoke refresh token to logout
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body object true "Refresh token to revoke"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// LogoutAll godoc
//
// @Summary      Logout all sessions
// @Description  Revoke all refresh tokens for the user
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := c.GetString("user_id")

	if err := h.authService.LogoutAll(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout all sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all sessions terminated"})
}

// ChangePassword godoc
//
// @Summary      Change password
// @Description  Change user password after verifying current password
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body object true "Password change request"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=12"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// GetProfile godoc
//
// @Summary      Get user profile
// @Description  Get current authenticated user profile
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	profile, err := h.authService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
//
// @Summary      Update user profile
// @Description  Update current authenticated user's profile (name)
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body service.UpdateProfileRequest true "Profile update"
// @Success      200  {object}  service.ProfileResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.authService.UpdateProfile(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetSessions godoc
//
// @Summary      Get active sessions
// @Description  List all active sessions for the user
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}    map[string]interface{}
// @Failure      401  {object}   map[string]string
// @Failure      500  {object}   map[string]string
// @Router       /auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	userID := c.GetString("user_id")

	sessions, err := h.authService.GetSessions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sessions"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}
