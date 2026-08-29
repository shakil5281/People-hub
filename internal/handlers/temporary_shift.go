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

type BulkTemporaryShiftRequest struct {
	CompanyID    string   `json:"company_id" binding:"required"`
	ShiftID      string   `json:"shift_id" binding:"required"`
	Date         string   `json:"date" binding:"required"`
	EmployeeIDs  []string `json:"employee_ids"`
	DepartmentID string   `json:"department_id"`
	SectionID    string   `json:"section_id"`
	LineID       string   `json:"line_id"`
	GroupID      string   `json:"group_id"`
	Reason       string   `json:"reason"`
	Status       string   `json:"status"`
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

	// Enforce single-day only — ToDate must be empty or equal to FromDate
	if req.ToDate != "" && req.ToDate != req.FromDate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Temporary shift only allows single day; to_date must be empty or same as from_date"})
		return
	}

	if _, err := utils.ParseDate(req.FromDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	date := req.FromDate
	existing, err := h.repo.FindByEmployeeAndDate(req.EmployeeID, date)
	if err == nil && existing != nil {
		existing.ShiftID = req.ShiftID
		existing.Reason = req.Reason
		existing.Status = status
		if err := h.repo.Update(existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Temporary shift created successfully",
			"count":   1,
			"records": []models.TemporaryShift{*existing},
		})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Temporary shift created successfully",
		"count":   1,
		"records": []models.TemporaryShift{ts},
	})
}

// BulkCreate godoc
//
// @Summary      Bulk create temporary shifts for single day
// @Description  Assign a shift to multiple employees on a single date. Supports explicit employee_ids or org filter (department/section/line/group).
// @Tags         Temporary Shifts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkTemporaryShiftRequest true "Bulk payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /temporary-shifts/bulk [post]
func (h *TemporaryShiftHandler) BulkCreate(c *gin.Context) {
	var req BulkTemporaryShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := utils.ParseDate(req.Date); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	var employeeIDs []string
	if len(req.EmployeeIDs) > 0 {
		employeeIDs = req.EmployeeIDs
	} else if req.DepartmentID != "" || req.SectionID != "" || req.LineID != "" || req.GroupID != "" {
		filter := repository.EmployeeFilter{
			CompanyID:    req.CompanyID,
			DepartmentID: req.DepartmentID,
			SectionID:    req.SectionID,
			LineID:       req.LineID,
			GroupID:      req.GroupID,
		}
		employees, _, err := h.employeeRepo.ListFiltered(filter, 1, 5000)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, emp := range employees {
			employeeIDs = append(employeeIDs, emp.EmployeeID)
		}
		if len(employeeIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No employees found for given filter"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_ids or one of department_id/section_id/line_id/group_id is required"})
		return
	}

	var created []models.TemporaryShift
	for _, empID := range employeeIDs {
		date := req.Date
		existing, err := h.repo.FindByEmployeeAndDate(empID, date)
		if err == nil && existing != nil {
			existing.ShiftID = req.ShiftID
			existing.Reason = req.Reason
			existing.Status = status
			if err := h.repo.Update(existing); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			created = append(created, *existing)
			continue
		}
		ts := models.TemporaryShift{
			EmployeeID: empID,
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
		"message": "Temporary shifts created successfully",
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
