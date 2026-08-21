package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type TemporaryShiftHandler struct {
	repo         *repository.TemporaryShiftRepository
	employeeRepo *repository.EmployeeRepository
}

func NewTemporaryShiftHandler(repo *repository.TemporaryShiftRepository, employeeRepo *repository.EmployeeRepository) *TemporaryShiftHandler {
	return &TemporaryShiftHandler{repo: repo, employeeRepo: employeeRepo}
}

type CreateTemporaryShiftRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	ShiftID    string `json:"shift_id" binding:"required"`
	CompanyID  string `json:"company_id" binding:"required"`
	FromDate   string `json:"from_date" binding:"required"`
	ToDate     string `json:"to_date"`
	Reason     string `json:"reason"`
	Status     string `json:"status"`
}

type UpdateTemporaryShiftRequest struct {
	ShiftID string `json:"shift_id"`
	Date    string `json:"date"`
	Reason  string `json:"reason"`
	Status  string `json:"status"`
}

func (h *TemporaryShiftHandler) Create(c *gin.Context) {
	var req CreateTemporaryShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toDate := req.ToDate
	if toDate == "" {
		toDate = req.FromDate
	}

	dates, err := utils.GenerateDateRange(req.FromDate, toDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	var created []models.TemporaryShift
	for _, date := range dates {
		existing, err := h.repo.FindByEmployeeAndDate(req.EmployeeID, date)
		if err == nil && existing != nil {
			existing.ShiftID = req.ShiftID
			existing.Reason = req.Reason
			existing.Status = status
			h.repo.Update(existing)
			created = append(created, *existing)
			continue
		}

		ts := models.TemporaryShift{
			EmployeeID: req.EmployeeID,
			ShiftID:    req.ShiftID,
			CompanyID:  req.CompanyID,
			Date:       date,
			Reason:     req.Reason,
			Status:     status,
		}
		if err := h.repo.Create(&ts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = append(created, ts)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Temporary shift(s) created successfully",
		"count":   len(created),
		"records": created,
	})
}

// ListTemporaryShifts godoc
//
// @Summary      List temporary shifts
// @Tags         Temporary Shifts
// @Security     BearerAuth
// @Produce      json
// @Param        company_id query string false "Filter by company"
// @Param        page       query int    false "Page number (default: 1)"
// @Param        limit      query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /temporary-shifts [get]
func (h *TemporaryShiftHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	p := utils.ParsePagination(c)
	list, total, err := h.repo.List(companyID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

func (h *TemporaryShiftHandler) GetByID(c *gin.Context) {
	ts, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Temporary shift not found"})
		return
	}
	c.JSON(http.StatusOK, ts)
}

func (h *TemporaryShiftHandler) Update(c *gin.Context) {
	id := c.Param("id")
	ts, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Temporary shift not found"})
		return
	}

	var req UpdateTemporaryShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ShiftID != "" {
		ts.ShiftID = req.ShiftID
	}
	if req.Date != "" {
		ts.Date = req.Date
	}
	if req.Reason != "" {
		ts.Reason = req.Reason
	}
	if req.Status != "" {
		ts.Status = req.Status
	}

	if err := h.repo.Update(ts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ts)
}

func (h *TemporaryShiftHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Temporary shift deleted successfully"})
}
