package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type AdvanceSalaryHandler struct {
	advanceRepo  *repository.AdvanceSalaryRepository
	employeeRepo *repository.EmployeeRepository
}

func NewAdvanceSalaryHandler(advanceRepo *repository.AdvanceSalaryRepository, employeeRepo *repository.EmployeeRepository) *AdvanceSalaryHandler {
	return &AdvanceSalaryHandler{advanceRepo: advanceRepo, employeeRepo: employeeRepo}
}

type AdvanceBulkApplyRequest struct {
	CompanyID      string   `json:"company_id" binding:"required"`
	EmployeeIDs    []string `json:"employee_ids" binding:"required"`
	Amount         float64  `json:"amount" binding:"required,gt=0"`
	AdvanceDate    string   `json:"advance_date" binding:"required"`
	DeductionMonth int      `json:"deduction_month" binding:"required"`
	DeductionYear  int      `json:"deduction_year" binding:"required"`
	Reason         string   `json:"reason"`
	SalaryType     string   `json:"salary_type"`
	IsOvertime     bool     `json:"is_overtime"`
}

// BulkApply godoc
// @Summary      Bulk apply advance salary
// @Description  Apply advance salary to multiple employees
// @Tags         Salary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body AdvanceBulkApplyRequest true "Bulk Apply Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/advances/bulk-apply [post]
func (h *AdvanceSalaryHandler) BulkApply(c *gin.Context) {
	var req AdvanceBulkApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetString("user_id")

	employees, err := h.employeeRepo.FindByEmployeeIDs(req.EmployeeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employees"})
		return
	}
	if len(employees) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No employees found for the given IDs"})
		return
	}

	var advances []models.AdvanceSalary
	for _, emp := range employees {
		advances = append(advances, models.AdvanceSalary{
			CompanyID:      req.CompanyID,
			EmployeeID:     emp.EmployeeID,
			Amount:         req.Amount,
			AdvanceDate:    req.AdvanceDate,
			DeductionMonth: req.DeductionMonth,
			DeductionYear:  req.DeductionYear,
			Reason:         req.Reason,
			SalaryType:     req.SalaryType,
			IsOvertime:     req.IsOvertime,
			Status:         "pending",
			CreatedBy:      userID,
		})
	}

	if err := h.advanceRepo.CreateBatch(advances); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply advances"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Advance applied successfully",
		"count":   len(advances),
	})
}

// List godoc
// @Summary      List advance salaries
// @Description  Get a filtered list of advance salaries
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        company_id     query string true  "Company ID"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        month          query int    false "Month (1-12)"
// @Param        year           query int    false "Year"
// @Param        status         query string false "Filter by status"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /salary/advances [get]
func (h *AdvanceSalaryHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	advances, err := h.advanceRepo.List(repository.AdvanceFilter{
		CompanyID:     companyID,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		Month:         month,
		Year:          year,
		Status:        c.Query("status"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"advances": advances,
		"total":    len(advances),
	})
}

// Approve godoc
// @Summary      Approve advance salary
// @Description  Approve pending advance salary
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Advance ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/advances/{id}/approve [put]
func (h *AdvanceSalaryHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.advanceRepo.UpdateStatus(id, "approved", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve advance"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Advance approved"})
}

// Reject godoc
// @Summary      Reject advance salary
// @Description  Reject pending advance salary
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Advance ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/advances/{id}/reject [put]
func (h *AdvanceSalaryHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.advanceRepo.UpdateStatus(id, "rejected", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject advance"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Advance rejected"})
}

// Delete godoc
// @Summary      Delete advance salary
// @Description  Delete pending advance salary
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Advance ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/advances/{id} [delete]
func (h *AdvanceSalaryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.advanceRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete advance"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Advance deleted"})
}
