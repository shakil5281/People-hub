package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type DailyScheduleHandler struct {
	scheduleRepo *repository.DailyScheduleRepository
}

func NewDailyScheduleHandler(scheduleRepo *repository.DailyScheduleRepository) *DailyScheduleHandler {
	return &DailyScheduleHandler{scheduleRepo: scheduleRepo}
}

type CreateDailyScheduleRequest struct {
	CompanyID    string `json:"company_id" binding:"required"`
	EmployeeID   string `json:"employee_id" binding:"required"`
	Date         string `json:"date" binding:"required"`
	ScheduleType string `json:"schedule_type" binding:"required"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Remarks      string `json:"remarks"`
}

// ListDailySchedules godoc
//
//	@Summary      List daily schedules
//	@Description  Get all daily schedules for a company, optionally filtered by date
//	@Tags         HR
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Param        date       query string false "Date (YYYY-MM-DD)"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /daily-schedules [get]
func (h *DailyScheduleHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}

	date := c.Query("date")
	var items []models.DailySchedule
	var err error

	if date != "" {
		items, err = h.scheduleRepo.ListByDate(companyID, date)
	} else {
		items, err = h.scheduleRepo.List(companyID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schedules": items,
		"total":     len(items),
	})
}

// CreateDailySchedule godoc
//
//	@Summary      Create daily schedule
//	@Description  Create a new daily schedule for an employee
//	@Tags         HR
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body CreateDailyScheduleRequest true "Schedule details"
//	@Success      201  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /daily-schedules [post]
func (h *DailyScheduleHandler) Create(c *gin.Context) {
	var req CreateDailyScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &models.DailySchedule{
		CompanyID:    req.CompanyID,
		EmployeeID:   req.EmployeeID,
		Date:         req.Date,
		ScheduleType: req.ScheduleType,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Remarks:      req.Remarks,
		Status:       "active",
		CreatedBy:    c.GetString("user_id"),
	}

	if err := h.scheduleRepo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateDailySchedule godoc
//
//	@Summary      Update daily schedule
//	@Description  Update an existing daily schedule
//	@Tags         HR
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        id      path string true "Schedule ID"
//	@Param        request body CreateDailyScheduleRequest false "Updated fields"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /daily-schedules/{id} [put]
func (h *DailyScheduleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	item, err := h.scheduleRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Daily schedule not found"})
		return
	}

	var req CreateDailyScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.Date = req.Date
	item.ScheduleType = req.ScheduleType
	item.StartTime = req.StartTime
	item.EndTime = req.EndTime
	item.Remarks = req.Remarks
	userID := c.GetString("user_id")
	item.UpdatedBy = &userID

	if err := h.scheduleRepo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteDailySchedule godoc
//
//	@Summary      Delete daily schedule
//	@Description  Soft delete a daily schedule
//	@Tags         HR
//	@Security     BearerAuth
//	@Produce      json
//	@Param        id path string true "Schedule ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /daily-schedules/{id} [delete]
func (h *DailyScheduleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	item, err := h.scheduleRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Daily schedule not found"})
		return
	}

	if err := h.scheduleRepo.Delete(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Daily schedule deleted successfully"})
}
