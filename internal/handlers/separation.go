package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type SeparationHandler struct {
	repo    *repository.SeparationRepository
	service *service.SeparationService
}

func NewSeparationHandler(repo *repository.SeparationRepository, svc *service.SeparationService) *SeparationHandler {
	return &SeparationHandler{repo: repo, service: svc}
}

type CreateSeparationRequest struct {
	EmployeeID   string `json:"employee_id" binding:"required"`
	DepartmentID string `json:"department_id"`
	Type         string `json:"type" binding:"required"`
	Date         string `json:"date"`
	Reason       string `json:"reason"`
}

// ListSeparations godoc
//
// @Summary      List separations
// @Tags         Separations
// @Security     BearerAuth
// @Produce      json
// @Param        employee       query string false "Filter by employee name"
// @Param        employee_id    query string false "Filter by employee ID"
// @Param        department_id  query string false "Filter by department"
// @Param        type           query string false "Filter by type"
// @Param        status         query string false "Filter by status"
// @Param        company_id     query string false "Filter by company"
// @Param        section_id     query string false "Filter by section (via employee)"
// @Param        designation_id query string false "Filter by designation (via employee)"
// @Param        line_id        query string false "Filter by line (via employee)"
// @Param        group_id       query string false "Filter by group (via employee)"
// @Param        date_from      query string false "Filter by separation start date (YYYY-MM-DD)"
// @Param        date_to        query string false "Filter by separation end date (YYYY-MM-DD)"
// @Param        page           query int    false "Page number (default: 1)"
// @Param        limit          query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /separations [get]
func (h *SeparationHandler) List(c *gin.Context) {
	employee := c.Query("employee")
	employeeID := c.Query("employee_id")
	departmentID := c.Query("department_id")
	sepType := c.Query("type")
	status := c.Query("status")
	companyID := c.Query("company_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	p := utils.ParsePagination(c)

	hasFilters := employee != "" || employeeID != "" || departmentID != "" || sepType != "" || status != "" ||
		companyID != "" || sectionID != "" || designationID != "" || lineID != "" || groupID != "" ||
		dateFrom != "" || dateTo != ""

	if hasFilters {
		items, total, err := h.repo.ListFiltered(employee, employeeID, departmentID, sepType, status, companyID, sectionID, designationID, lineID, groupID, dateFrom, dateTo, p.Page, p.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, utils.NewPaginatedResponse(items, total, p))
		return
	}

	items, total, err := h.repo.List(p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(items, total, p))
}

func (h *SeparationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "separation not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateSeparation godoc
//
// @Summary      Create separation
// @Description  Create a separation record. Auto-processes if date <= today.
// @Tags         Separations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateSeparationRequest true "Separation data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /separations [post]
func (h *SeparationHandler) Create(c *gin.Context) {
	var req CreateSeparationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.CreateSeparationInput{
		EmployeeID:   req.EmployeeID,
		DepartmentID: req.DepartmentID,
		SepType:      req.Type,
		Date:         req.Date,
		Reason:       req.Reason,
	}

	sep, result, err := h.service.Create(input)
	if err != nil {
		switch err {
		case service.ErrEmployeeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found or not eligible (must be active and Regular)"})
		case service.ErrDuplicateSeparation:
			c.JSON(http.StatusConflict, gin.H{"error": "employee already has a pending or approved separation"})
		case service.ErrInvalidSeparationType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "type must be Resign, Lefty, or Close"})
		case service.ErrSeparationBeforeJoining:
			c.JSON(http.StatusBadRequest, gin.H{"error": "separation date cannot be before joining date"})
		case service.ErrDepartmentRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": "employee has no department assigned — department is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	autoProcessed := result.NewType != ""
	response := gin.H{
		"id":                 sep.ID,
		"message":            "Separation created",
		"employee_id":        result.EmployeeID,
		"employee_name":      result.EmployeeName,
		"auto_processed":     autoProcessed,
		"new_employee_type":  result.NewType,
		"employee_status":    result.NewStatus,
		"attendance_deleted": result.AttendanceDeleted,
	}
	c.JSON(http.StatusCreated, response)
}

func (h *SeparationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "separation not found"})
		return
	}
	if item.Status == "Processed" || item.Status == "Cancelled" {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot edit processed or cancelled separation"})
		return
	}

	var req CreateSeparationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type != "" {
		if _, ok := map[string]bool{"Resign": true, "Lefty": true, "Close": true}[req.Type]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type must be Resign, Lefty, or Close"})
			return
		}
		item.Type = req.Type
	}
	item.Date = req.Date
	item.Reason = req.Reason
	if req.DepartmentID != "" {
		item.DepartmentID = req.DepartmentID
	}

	if err := h.repo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, _ := h.repo.FindByID(item.ID)
	c.JSON(http.StatusOK, result)
}

func (h *SeparationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "separation not found"})
		return
	}
	if item.Status == "Processed" {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete a processed separation — cancel it or reactivate the employee instead"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "separation deleted"})
}

// ProcessSeparations godoc
//
// @Summary      Batch process separations
// @Description  Process all pending/approved separations with date <= given date
// @Tags         Separations
// @Security     BearerAuth
// @Produce      json
// @Param        date query string false "Process separations due on or before this date (default: today)"
// @Success      200  {object}  map[string]interface{}
// @Router       /separations/process [post]
func (h *SeparationHandler) ProcessBatch(c *gin.Context) {
	dateStr := c.DefaultQuery("date", "")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	results, err := h.service.ProcessBatch(dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type Detail struct {
		EmployeeID   string `json:"employee_id"`
		EmployeeName string `json:"employee_name"`
		NewType      string `json:"new_type"`
		AttnDeleted  int64  `json:"attendance_deleted"`
	}
	details := make([]Detail, len(results))
	for i, r := range results {
		details[i] = Detail{
			EmployeeID:   r.EmployeeID,
			EmployeeName: r.EmployeeName,
			NewType:      r.NewType,
			AttnDeleted:  r.AttendanceDeleted,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Processed %d separations for date <= %s", len(results), dateStr),
		"date":      dateStr,
		"processed": len(results),
		"details":   details,
	})
}

// ProcessOne godoc
//
// @Summary      Process a single separation
// @Description  Apply employee changes for a specific separation
// @Tags         Separations
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Separation ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /separations/{id}/process [post]
func (h *SeparationHandler) ProcessOne(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.ProcessOne(id)
	if err != nil {
		switch err {
		case service.ErrAlreadyProcessed:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case service.ErrEmployeeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Separation processed",
		"employee_id":        result.EmployeeID,
		"employee_name":      result.EmployeeName,
		"new_type":           result.NewType,
		"employee_status":    result.NewStatus,
		"attendance_deleted": result.AttendanceDeleted,
	})
}

// CancelSeparation godoc
//
// @Summary      Cancel a separation
// @Description  Mark a pending or approved separation as cancelled
// @Tags         Separations
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Separation ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /separations/{id}/cancel [post]
func (h *SeparationHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Cancel(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "separation cancelled"})
}

// ReactivateSeparation godoc
//
// @Summary      Reactivate a processed separation
// @Description  Revert a processed separation, restoring employee to active/Regular status
// @Tags         Separations
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Separation ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /separations/{id}/reactivate [post]
func (h *SeparationHandler) Reactivate(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Reactivate(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "separation reactivated, employee restored"})
}
