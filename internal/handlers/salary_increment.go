package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
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
	CompanyID        string   `json:"company_id" binding:"required"`
	DepartmentID     string   `json:"department_id"`
	SectionID        string   `json:"section_id"`
	DesignationID    string   `json:"designation_id"`
	LineID           string   `json:"line_id"`
	GroupID          string   `json:"group_id"`
	EmployeeIDs      []string `json:"employee_ids"`
	IncrementType    string   `json:"increment_type" binding:"required"`
	CalculationType  string   `json:"calculation_type"`
	NewDesignationID string   `json:"new_designation_id"`
	IncrementDate    string   `json:"increment_date" binding:"required"`
	EffectiveDate    string   `json:"effective_date" binding:"required"`
	Value            float64  `json:"value"`
	Remarks          string   `json:"remarks"`
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
// @Param        increment_type query string false "Filter by increment type"
// @Param        month          query int    false "Filter by month (1-12)"
// @Param        year           query int    false "Filter by year"
// @Param        status         query string false "Filter by status"
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
		IncrementType: c.Query("increment_type"),
		Month:         month,
		Year:          year,
		Status:        c.Query("status"),
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

	// Normalize increment type
	incType := strings.ToLower(strings.TrimSpace(req.IncrementType))
	calcType := strings.ToLower(strings.TrimSpace(req.CalculationType))

	switch incType {
	case "gov_policy", "gov", "policy", "as per gov policy", "as_per_gov_policy":
		incType = "gov_policy"
		if req.Value <= 0 {
			req.Value = 9 // Standard 9% as per policy
		}
		calcType = "percentage"
	case "promotion", "prom":
		incType = "promotion"
	case "promotion_with_increment", "promotion_increment":
		incType = "promotion_with_increment"
	default:
		incType = "increment"
	}

	if (incType == "promotion" || incType == "promotion_with_increment") && strings.TrimSpace(req.NewDesignationID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New designation is required for promotions"})
		return
	}

	var emps []models.Employee
	var err error
	if len(req.EmployeeIDs) > 0 {
		emps, err = h.incrementRepo.FindEligibleEmployeesByIDs(req.CompanyID, req.EmployeeIDs)
	} else {
		emps, err = h.incrementRepo.FindEligibleEmployees(req.CompanyID, req.DepartmentID, req.SectionID, req.DesignationID, req.LineID, req.GroupID)
	}
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
		curBasic := emp.BasicSalary
		curHouse := emp.HouseRent
		medical := emp.MedicalAllowance
		transport := emp.TransportAllowance
		food := emp.FoodAllowance
		if medical == 0 {
			medical = 750
		}
		if transport == 0 {
			transport = 450
		}
		if food == 0 {
			food = 1250
		}

		if (curBasic <= 0 || curHouse <= 0) && emp.GrossSalary > 0 {
			core := emp.GrossSalary - medical - transport - food
			if core > 0 {
				curBasic = math.Round(core / 1.5)
				curHouse = core - curBasic
			}
		}

		coreBase := curBasic + curHouse

		var incAmount float64
		switch incType {
		case "gov_policy":
			rate := req.Value
			if rate <= 0 {
				rate = 9
			}
			incAmount = math.Round(coreBase * rate / 100)
		case "promotion":
			if calcType == "percentage" {
				incAmount = math.Round(coreBase * req.Value / 100)
			} else {
				incAmount = req.Value
			}
		case "promotion_with_increment":
			if calcType == "percentage" {
				incAmount = math.Round(coreBase * req.Value / 100)
			} else {
				incAmount = req.Value
			}
		default: // "increment"
			if calcType == "percentage" || req.IncrementType == "percentage" {
				incAmount = math.Round(coreBase * req.Value / 100)
			} else {
				incAmount = req.Value
			}
		}

		newGross := emp.GrossSalary + incAmount
		newCore := newGross - transport - food - medical
		newBasic := math.Round(newCore / 1.5)
		newHouse := newCore - newBasic

		var prevDesigID *string = emp.DesignationID
		var newDesigID *string = nil
		if (incType == "promotion" || incType == "promotion_with_increment") && strings.TrimSpace(req.NewDesignationID) != "" {
			targetDesig := strings.TrimSpace(req.NewDesignationID)
			newDesigID = &targetDesig
		}

		incs = append(incs, models.SalaryIncrement{
			CompanyID:             req.CompanyID,
			EmployeeID:            emp.EmployeeID,
			IncrementType:         incType,
			CalculationType:       calcType,
			CalculationValue:      req.Value,
			PreviousGross:         emp.GrossSalary,
			PreviousBasic:         emp.BasicSalary,
			PreviousHouse:         emp.HouseRent,
			PreviousMedical:       emp.MedicalAllowance,
			PreviousDesignationID: prevDesigID,
			NewDesignationID:      newDesigID,
			IncrementAmount:       incAmount,
			NewGross:              newGross,
			NewBasic:              newBasic,
			NewHouse:              newHouse,
			NewMedical:            medical,
			IncrementDate:         req.IncrementDate,
			EffectiveDate:         req.EffectiveDate,
			Status:                "pending",
			Remarks:               req.Remarks,
			CreatedBy:             userID,
		})
	}

	if err := h.incrementRepo.CreateBatch(incs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Salary increment applied to %d employees", len(incs)),
		"applied": len(incs),
		"total":   len(emps),
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

	if inc.NewDesignationID != nil && *inc.NewDesignationID != "" {
		emp.DesignationID = inc.NewDesignationID
	}

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

// Delete godoc
//
// @Summary      Delete salary increment
// @Description  Delete a pending increment request
// @Tags         Salary
// @Security     BearerAuth
// @Param        id      path string true "Increment ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments/{id} [delete]
func (h *SalaryIncrementHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	inc, err := h.incrementRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Increment not found"})
		return
	}

	if inc.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending increments can be deleted"})
		return
	}

	if err := h.incrementRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Increment deleted successfully"})
}
