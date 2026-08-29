package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type RosterHandler struct {
	repo         *repository.RosterRepository
	employeeRepo *repository.EmployeeRepository
}

func NewRosterHandler(repo *repository.RosterRepository, employeeRepo *repository.EmployeeRepository) *RosterHandler {
	return &RosterHandler{repo: repo, employeeRepo: employeeRepo}
}

type CreateRosterRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	ShiftID    string `json:"shift_id" binding:"required"`
	CompanyID  string `json:"company_id" binding:"required"`
	FromDate   string `json:"from_date" binding:"required"`
	ToDate     string `json:"to_date"`
	Reason     string `json:"reason"`
	Status     string `json:"status"`
}

type BulkRosterRequest struct {
	CompanyID    string   `json:"company_id" binding:"required"`
	ShiftID      string   `json:"shift_id" binding:"required"`
	FromDate     string   `json:"from_date" binding:"required"`
	ToDate       string   `json:"to_date"`
	EmployeeIDs  []string `json:"employee_ids"`
	DepartmentID string   `json:"department_id"`
	SectionID    string   `json:"section_id"`
	LineID       string   `json:"line_id"`
	GroupID      string   `json:"group_id"`
	Reason       string   `json:"reason"`
	Status       string   `json:"status"`
}

type UpdateRosterRequest struct {
	ShiftID string `json:"shift_id"`
	Date    string `json:"date"`
	Reason  string `json:"reason"`
	Status  string `json:"status"`
}

type BulkDeleteRosterRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// Create godoc
//
// @Summary      Create roster assignments
// @Description  Create roster entries for an employee over a date range. Upserts if roster exists for employee+date.
// @Tags         Rosters
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateRosterRequest true "Roster payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /rosters [post]
func (h *RosterHandler) Create(c *gin.Context) {
	var req CreateRosterRequest
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	userID := c.GetString("user_id")
	var created []models.Roster
	for _, date := range dates {
		existing, err := h.repo.FindByEmployeeAndDate(req.EmployeeID, date)
		if err == nil && existing != nil {
			existing.ShiftID = req.ShiftID
			existing.Reason = req.Reason
			existing.Status = status
			if userID != "" {
				existing.UpdatedBy = &userID
			}
			if err := h.repo.Update(existing); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			created = append(created, *existing)
			continue
		}

		roster := models.Roster{
			EmployeeID: req.EmployeeID,
			ShiftID:    req.ShiftID,
			CompanyID:  req.CompanyID,
			Date:       date,
			Reason:     req.Reason,
			Status:     status,
		}
		if userID != "" {
			roster.CreatedBy = &userID
		}
		if err := h.repo.Create(&roster); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = append(created, roster)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Roster(s) created successfully",
		"count":   len(created),
		"records": created,
	})
}

// BulkCreate godoc
//
// @Summary      Bulk create roster assignments
// @Description  Assign a shift to multiple employees (explicit IDs or org filter) over a date range.
// @Tags         Rosters
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkRosterRequest true "Bulk roster payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /rosters/bulk [post]
func (h *RosterHandler) BulkCreate(c *gin.Context) {
	var req BulkRosterRequest
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}
	userID := c.GetString("user_id")

	// Resolve employee IDs
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

	var created []models.Roster
	for _, empID := range employeeIDs {
		for _, date := range dates {
			existing, err := h.repo.FindByEmployeeAndDate(empID, date)
			if err == nil && existing != nil {
				existing.ShiftID = req.ShiftID
				existing.Reason = req.Reason
				existing.Status = status
				if userID != "" {
					existing.UpdatedBy = &userID
				}
				if err := h.repo.Update(existing); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				created = append(created, *existing)
				continue
			}
			roster := models.Roster{
				EmployeeID: empID,
				ShiftID:    req.ShiftID,
				CompanyID:  req.CompanyID,
				Date:       date,
				Reason:     req.Reason,
				Status:     status,
			}
			if userID != "" {
				roster.CreatedBy = &userID
			}
			if err := h.repo.Create(&roster); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			created = append(created, roster)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Bulk roster created successfully",
		"count":   len(created),
		"records": created,
	})
}

// List godoc
//
// @Summary      List rosters
// @Tags         Rosters
// @Security     BearerAuth
// @Produce      json
// @Param        company_id query string false "Filter by company"
// @Param        employee_id query string false "Filter by employee"
// @Param        shift_id query string false "Filter by shift"
// @Param        department_id query string false "Filter by department"
// @Param        section_id query string false "Filter by section"
// @Param        line_id query string false "Filter by line"
// @Param        group_id query string false "Filter by group"
// @Param        from_date query string false "From date YYYY-MM-DD"
// @Param        to_date query string false "To date YYYY-MM-DD"
// @Param        status query string false "Filter by status"
// @Param        search query string false "Search employee_id or reason"
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /rosters [get]
func (h *RosterHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	p := utils.ParsePagination(c)
	filters := map[string]string{
		"employee_id":   c.Query("employee_id"),
		"shift_id":      c.Query("shift_id"),
		"status":        c.Query("status"),
		"from_date":     c.Query("from_date"),
		"to_date":       c.Query("to_date"),
		"department_id": c.Query("department_id"),
		"section_id":    c.Query("section_id"),
		"line_id":       c.Query("line_id"),
		"group_id":      c.Query("group_id"),
		"search":        c.Query("search"),
	}
	list, total, err := h.repo.List(companyID, p.Page, p.Limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

// GetByID godoc
//
// @Summary      Get roster by ID
// @Tags         Rosters
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Roster ID"
// @Success      200  {object}  models.Roster
// @Failure      404  {object}  map[string]string
// @Router       /rosters/{id} [get]
func (h *RosterHandler) GetByID(c *gin.Context) {
	roster, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Roster not found"})
		return
	}
	c.JSON(http.StatusOK, roster)
}

// Update godoc
//
// @Summary      Update roster
// @Tags         Rosters
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Roster ID"
// @Param        request body UpdateRosterRequest true "Update payload"
// @Success      200  {object}  models.Roster
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /rosters/{id} [put]
func (h *RosterHandler) Update(c *gin.Context) {
	id := c.Param("id")
	roster, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Roster not found"})
		return
	}

	var req UpdateRosterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ShiftID != "" {
		roster.ShiftID = req.ShiftID
	}
	if req.Date != "" {
		if _, err := utils.ParseDate(req.Date); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
			return
		}
		roster.Date = req.Date
	}
	if req.Reason != "" {
		roster.Reason = req.Reason
	}
	if req.Status != "" {
		roster.Status = req.Status
	}

	if userID := c.GetString("user_id"); userID != "" {
		roster.UpdatedBy = &userID
	}

	if err := h.repo.Update(roster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roster)
}

// Delete godoc
//
// @Summary      Delete roster
// @Tags         Rosters
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Roster ID"
// @Success      200  {object}  map[string]string
// @Router       /rosters/{id} [delete]
func (h *RosterHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Roster deleted successfully"})
}

// BulkDelete godoc
//
// @Summary      Bulk delete rosters
// @Tags         Rosters
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkDeleteRosterRequest true "IDs to delete"
// @Success      200  {object}  map[string]string
// @Router       /rosters/bulk-delete [post]
func (h *RosterHandler) BulkDelete(c *gin.Context) {
	var req BulkDeleteRosterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.BulkDelete(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rosters deleted successfully", "count": len(req.IDs)})
}
