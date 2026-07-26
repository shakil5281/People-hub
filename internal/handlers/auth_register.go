package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/auth"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12"`
	Name     string `json:"name" binding:"required"`
}

// Register godoc
//
// @Summary      Register user
// @Description  Register a new user account
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existing models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Status:       "active",
	}

	if err := database.DB.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user": gin.H{
			"id":     user.ID,
			"email":  user.Email,
			"name":   user.Name,
			"status": user.Status,
		},
	})
}
