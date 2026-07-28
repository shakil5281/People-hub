package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type SalaryIncrementHandler struct {
	incrementRepo *repository.SalaryIncrementRepository
	employeeRepo  *repository.EmployeeRepository
}

func NewSalaryIncrementHandler(incrementRepo *repository.SalaryIncrementRepository, employeeRepo *repository.EmployeeRepository) *SalaryIncrementHandler {
	return &SalaryIncrementHandler{incrementRepo: incrementRepo, employeeRepo: employeeRepo}
}

type BulkApplyRequest struct {
	CompanyID     string  `json:"company_id" binding:"required"`
	DepartmentID  string  `json:"department_id"`
	SectionID     string  `json:"section_id"`
	DesignationID string  `json:"designation_id"`
	LineID        string  `json:"line_id"`
	GroupID       string  `json:"group_id"`
	IncrementType string  `json:"increment_type" binding:"required,oneof=percentage fixed"`
	IncrementDate string  `json:"increment_date" binding:"required"`
	EffectiveDate string  `json:"effective_date" binding:"required"`
	Value         float64 `json:"value" binding:"required,min=1"`
}

// ListIncrements godoc
//
// @Summary      List salary increments
// @Description  Get salary increments with filters
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        company_id     query string true  "Company ID"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        month          query int    false "Filter by month (1-12)"
// @Param        year           query int    false "Filter by year"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments [get]
func (h *SalaryIncrementHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	increments, err := h.incrementRepo.List(repository.IncrementFilter{
		CompanyID:     companyID,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		Month:         month,
		Year:          year,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"increments": increments,
		"total":      len(increments),
	})
}

// BulkApplyIncrement godoc
//
// @Summary      Bulk apply salary increments
// @Description  Apply salary increment to all eligible employees matching filters
// @Tags         Salary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkApplyRequest true "Bulk increment details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments/bulk-apply [post]
func (h *SalaryIncrementHandler) BulkApply(c *gin.Context) {
	var req BulkApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emps, err := h.incrementRepo.FindEligibleEmployees(req.CompanyID, req.DepartmentID, req.SectionID, req.DesignationID, req.LineID, req.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(emps) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No eligible employees found matching the filters"})
		return
	}

	userID := c.GetString("user_id")

	var incs []models.SalaryIncrement
	for _, emp := range emps {
		var incAmount float64
		if req.IncrementType == "percentage" {
			incAmount = emp.GrossSalary * req.Value / 100
		} else {
			incAmount = req.Value
		}

		newGross := emp.GrossSalary + incAmount

		incs = append(incs, models.SalaryIncrement{
			CompanyID:       req.CompanyID,
			EmployeeID:      emp.EmployeeID,
			PreviousGross:   emp.GrossSalary,
			PreviousBasic:   emp.BasicSalary,
			PreviousHouse:   emp.HouseRent,
			PreviousMedical: emp.MedicalAllowance,
			IncrementAmount: incAmount,
			NewGross:        newGross,
			NewBasic:        newGross * 0.5,
			NewHouse:        newGross * 0.25,
			NewMedical:      newGross * 0.1,
			EffectiveDate:   req.EffectiveDate,
			Status:          "pending",
			CreatedBy:       userID,
		})
	}

	if err := h.incrementRepo.CreateBatch(incs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Salary increment applied to %d employees", len(incs)),
		"applied":  len(incs),
		"total":    len(emps),
	})
}

// ApproveIncrement godoc
//
// @Summary      Approve salary increment
// @Description  Approve a pending increment and update employee salary
// @Tags         Salary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string true "Increment ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments/{id}/approve [put]
func (h *SalaryIncrementHandler) Approve(c *gin.Context) {
	id := c.Param("id")

	inc, err := h.incrementRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Increment not found"})
		return
	}

	if inc.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending increments can be approved"})
		return
	}

	emp, err := h.employeeRepo.FindByEmployeeID(inc.EmployeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	emp.GrossSalary = inc.NewGross
	emp.BasicSalary = inc.NewBasic
	emp.HouseRent = inc.NewHouse
	emp.MedicalAllowance = inc.NewMedical

	if err := h.employeeRepo.Update(emp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	userID := c.GetString("user_id")
	inc.Status = "approved"
	inc.ApprovedBy = &userID
	inc.ApprovedAt = &now

	if err := h.incrementRepo.Update(inc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inc)
}

// RejectIncrement godoc
//
// @Summary      Reject salary increment
// @Description  Reject a pending increment request
// @Tags         Salary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path string true "Increment ID"
// @Param        request body map[string]string false "Rejection reason"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments/{id}/reject [put]
func (h *SalaryIncrementHandler) Reject(c *gin.Context) {
	id := c.Param("id")

	inc, err := h.incrementRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Increment not found"})
		return
	}

	if inc.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending increments can be rejected"})
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&body)

	now := time.Now()
	userID := c.GetString("user_id")
	inc.Status = "rejected"
	inc.RejectedBy = &userID
	inc.RejectedAt = &now
	inc.RejectionReason = body.Reason

	if err := h.incrementRepo.Update(inc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inc)
}
