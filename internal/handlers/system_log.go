package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type SystemLogHandler struct {
	repo *repository.SystemLogRepository
}

func NewSystemLogHandler(repo *repository.SystemLogRepository) *SystemLogHandler {
	return &SystemLogHandler{repo: repo}
}

// ListSystemLogs godoc
//
// @Summary      List system logs
// @Description  Get all system logs with optional filters
// @Tags         System Logs
// @Security     BearerAuth
// @Produce      json
// @Param        level      query string false "Filter by level (info|warn|error|debug)"
// @Param        source     query string false "Filter by source (api|gateway|web)"
// @Param        user_id    query string false "Filter by user ID"
// @Param        start_date query string false "Start date (YYYY-MM-DD)"
// @Param        end_date   query string false "End date (YYYY-MM-DD)"
// @Param        page       query int    false "Page number (default: 1)"
// @Param        limit      query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /system-logs [get]
func (h *SystemLogHandler) List(c *gin.Context) {
	filter := repository.SystemLogFilter{
		Level:     c.Query("level"),
		Source:    c.Query("source"),
		UserID:    c.Query("user_id"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}
	p := utils.ParsePagination(c)
	filter.Page = p.Page
	filter.Limit = p.Limit

	list, total, err := h.repo.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

// GetSystemLog godoc
//
// @Summary      Get system log by ID
// @Tags         System Logs
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "System Log ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /system-logs/{id} [get]
func (h *SystemLogHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	log, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
		return
	}
	c.JSON(http.StatusOK, log)
}

// GetSystemLogStats godoc
//
// @Summary      Get system log statistics
// @Description  Get log count grouped by level and source
// @Tags         System Logs
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  repository.LogStats
// @Failure      500  {object}  map[string]string
// @Router       /system-logs/stats [get]
func (h *SystemLogHandler) Stats(c *gin.Context) {
	stats, err := h.repo.Stats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// DeleteOldSystemLogs godoc
//
// @Summary      Delete old system logs
// @Description  Delete system logs older than the specified date (admin only)
// @Tags         System Logs
// @Security     BearerAuth
// @Produce      json
// @Param        before query string true "Delete logs before this date (YYYY-MM-DD)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /system-logs [delete]
func (h *SystemLogHandler) Delete(c *gin.Context) {
	before := c.Query("before")
	if before == "" {
		before = c.Query("date")
	}
	if before == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "before or date query param required (YYYY-MM-DD)"})
		return
	}
	parsed, err := utils.ParseDate(before)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}
	if err := h.repo.DeleteOlderThan(parsed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "old logs deleted"})
}

// PurgeAllSystemLogs godoc
//
// @Summary      Delete all system logs
// @Description  Purge all system log entries (admin only)
// @Tags         System Logs
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /system-logs/purge [delete]
func (h *SystemLogHandler) Purge(c *gin.Context) {
	if err := h.repo.PurgeAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all logs purged"})
}
