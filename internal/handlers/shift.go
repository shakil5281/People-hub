package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type ShiftHandler struct {
	shiftRepo *repository.ShiftRepository
}

func NewShiftHandler(shiftRepo *repository.ShiftRepository) *ShiftHandler {
	return &ShiftHandler{shiftRepo: shiftRepo}
}

type CreateShiftRequest struct {
	Name             string `json:"name" binding:"required"`
	CompanyID        string `json:"company_id" binding:"required"`
	ShiftType        string `json:"shift_type"`
	StartTime        string `json:"start_time" binding:"required"`
	EndTime          string `json:"end_time" binding:"required"`
	LateGraceMinutes int    `json:"late_grace_minutes"`
	WeekendDays      string `json:"weekend_days"`
}

type UpdateShiftRequest struct {
	Name             string `json:"name" binding:"required"`
	CompanyID        string `json:"company_id" binding:"required"`
	ShiftType        string `json:"shift_type"`
	StartTime        string `json:"start_time" binding:"required"`
	EndTime          string `json:"end_time" binding:"required"`
	LateGraceMinutes int    `json:"late_grace_minutes"`
	WeekendDays      string `json:"weekend_days"`
}

// ListShifts godoc
//
// @Summary      List shifts
// @Description  Get all shifts
// @Tags         Shifts
// @Security     BearerAuth
// @Produce      json
// @Param        page   query int    false "Page number (default: 1)"
// @Param        limit  query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /shifts [get]
func (h *ShiftHandler) List(c *gin.Context) {
	p := utils.ParsePagination(c)
	shifts, total, err := h.shiftRepo.List(p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(shifts, total, p))
}

// GetShift godoc
//
// @Summary      Get shift by ID
// @Description  Get a shift by its ID
// @Tags         Shifts
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Shift ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /shifts/{id} [get]
func (h *ShiftHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	shift, err := h.shiftRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shift not found"})
		return
	}
	c.JSON(http.StatusOK, shift)
}

// CreateShift godoc
//
// @Summary      Create shift
// @Description  Create a new shift
// @Tags         Shifts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateShiftRequest true "Shift details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /shifts [post]
func (h *ShiftHandler) Create(c *gin.Context) {
	var req CreateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	shift := &models.Shift{
		CompanyID:        req.CompanyID,
		Name:             req.Name,
		ShiftType:        req.ShiftType,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		LateGraceMinutes: req.LateGraceMinutes,
		WeekendDays:      req.WeekendDays,
		Status:           "active",
		CreatedBy:        &userID,
	}

	if err := h.shiftRepo.Create(shift); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shift)
}

// UpdateShift godoc
//
// @Summary      Update shift
// @Description  Update an existing shift
// @Tags         Shifts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Shift ID"
// @Param        request body UpdateShiftRequest true "Updated shift details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /shifts/{id} [put]
func (h *ShiftHandler) Update(c *gin.Context) {
	id := c.Param("id")
	shift, err := h.shiftRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shift not found"})
		return
	}

	var req UpdateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	shift.Name = req.Name
	shift.CompanyID = req.CompanyID
	shift.ShiftType = req.ShiftType
	shift.StartTime = req.StartTime
	shift.EndTime = req.EndTime
	shift.LateGraceMinutes = req.LateGraceMinutes
	shift.WeekendDays = req.WeekendDays
	shift.UpdatedBy = &userID

	if err := h.shiftRepo.Update(shift); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, shift)
}

// DeleteShift godoc
//
// @Summary      Delete shift
// @Description  Soft delete a shift
// @Tags         Shifts
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Shift ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /shifts/{id} [delete]
func (h *ShiftHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	shift, err := h.shiftRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shift not found"})
		return
	}

	userID := c.GetString("user_id")
	shift.UpdatedBy = &userID

	if err := h.shiftRepo.Update(shift); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.shiftRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "shift deleted"})
}
