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

type BulkDeleteTemporaryShiftRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// fallbackNote explains shift resolution after a temporary shift is removed:
// roster shift applies first, otherwise the employee's default shift.
const tempShiftFallbackNote = "Employees revert to roster shift, otherwise default shift"

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
// @Param        employee_id query string false "Filter by employee code"
// @Param        shift_id query string false "Filter by shift"
// @Param        department_id query string false "Filter by department"
// @Param        section_id query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id query string false "Filter by line"
// @Param        group_id query string false "Filter by group"
// @Param        status query string false "Filter by status"
// @Param        from_date query string false "From date YYYY-MM-DD"
// @Param        to_date query string false "To date YYYY-MM-DD"
// @Param        search query string false "Search employee code or reason"
// @Param        page       query int    false "Page number (default: 1)"
// @Param        limit      query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /temporary-shifts [get]
func (h *TemporaryShiftHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	p := utils.ParsePagination(c)
	filters := map[string]string{
		"employee_id":    c.Query("employee_id"),
		"shift_id":       c.Query("shift_id"),
		"department_id":  c.Query("department_id"),
		"section_id":     c.Query("section_id"),
		"designation_id": c.Query("designation_id"),
		"line_id":        c.Query("line_id"),
		"group_id":       c.Query("group_id"),
		"status":         c.Query("status"),
		"from_date":      c.Query("from_date"),
		"to_date":        c.Query("to_date"),
		"search":         c.Query("search"),
	}
	list, total, err := h.repo.List(companyID, p.Page, p.Limit, filters)
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
	c.JSON(http.StatusOK, gin.H{"message": "Temporary shift deleted successfully", "fallback": tempShiftFallbackNote})
}

// BulkDelete godoc
//
// @Summary      Bulk delete temporary shifts
// @Description  Delete multiple temporary shifts. Affected employees automatically revert to roster shift, otherwise default shift.
// @Tags         Temporary Shifts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkDeleteTemporaryShiftRequest true "IDs to delete"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /temporary-shifts/bulk-delete [post]
func (h *TemporaryShiftHandler) BulkDelete(c *gin.Context) {
	var req BulkDeleteTemporaryShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.BulkDelete(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Temporary shifts deleted successfully", "count": len(req.IDs), "fallback": tempShiftFallbackNote})
}
