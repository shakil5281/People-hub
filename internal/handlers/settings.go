package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type SettingsHandler struct {
	repo *repository.SettingsRepository
}

func NewSettingsHandler(repo *repository.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

// ListSettings godoc
//
//	@Summary      List all system settings
//	@Description  Get all system settings as key-value pairs
//	@Tags         Settings
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /settings [get]
func (h *SettingsHandler) List(c *gin.Context) {
	settings, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// UpdateSettings godoc
//
//	@Summary      Update system settings
//	@Description  Bulk update system settings (key-value pairs)
//	@Tags         Settings
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body UpdateSettingsRequest true "Settings map"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /settings [put]
func (h *SettingsHandler) Update(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.BulkUpsert(req.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "settings saved"})
}
