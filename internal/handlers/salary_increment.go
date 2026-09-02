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
	EmployeeIDSearch string   `json:"employee_id_search"`
	IncrementType    string   `json:"increment_type" binding:"required"`
	CalculationType  string   `json:"calculation_type"`
	NewDesignationID string   `json:"new_designation_id"`
	PromoDepartmentID string   `json:"promo_department_id"`
	PromoSectionID    string   `json:"promo_section_id"`
	PromoLineID       string   `json:"promo_line_id"`
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

// GetIncrementDetails godoc
//
// @Summary      Get employee increment details and history
// @Description  Get comprehensive salary increment details and progression for an employee
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        employee_id   query string false "Employee ID, Punch Number, or UUID"
// @Param        year          query int    false "Filter by year"
// @Param        month         query int    false "Filter by month (1-12)"
// @Param        company_id    query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Param        section_id    query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments/details [get]
func (h *SalaryIncrementHandler) GetIncrementDetails(c *gin.Context) {
	empCode := strings.TrimSpace(c.Query("employee_id"))
	yearStr := c.Query("year")
	monthStr := c.Query("month")
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)
	if month < 0 || month > 12 {
		month = 0
	}

	var emp *models.Employee
	if empCode != "" {
		emp, _ = h.employeeRepo.FindWithDetails(empCode)
	}

	if emp == nil {
		filter := repository.EmployeeFilter{
			CompanyID:     companyID,
			DepartmentID:  departmentID,
			SectionID:     sectionID,
			DesignationID: designationID,
			EmployeeID:    empCode,
		}
		emp, _ = h.employeeRepo.FindFirstFilteredWithDetails(filter)
	}

	if emp == nil {
		c.JSON(http.StatusOK, gin.H{
			"employee": nil,
			"summary": gin.H{
				"current_gross":          0,
				"current_basic":          0,
				"total_increment_amount": 0,
				"total_increment_count":  0,
				"total_promotions_count": 0,
				"last_increment_date":    "",
				"last_increment_amount":  0,
				"initial_gross":          0,
				"total_records":          0,
			},
			"increments": []interface{}{},
		})
		return
	}

	// Fetch all increments for the employee (all-time for complete summary calculations)
	allIncrements, _ := h.incrementRepo.ListByEmployee(emp.EmployeeID, 0, 0)

	// Fetch filtered increments for the specific year/month period
	var filteredIncrements []models.SalaryIncrement
	if year > 0 || month > 0 {
		filteredIncrements, _ = h.incrementRepo.ListByEmployee(emp.EmployeeID, year, month)
	} else {
		filteredIncrements = allIncrements
	}

	// Calculate summary metrics
	totalIncrementAmount := 0.0
	totalApprovedCount := 0
	totalPromotionsCount := 0
	lastIncrementDate := ""
	lastIncrementAmount := 0.0

	for _, inc := range allIncrements {
		if strings.EqualFold(inc.Status, "approved") {
			totalIncrementAmount += inc.IncrementAmount
			totalApprovedCount++
			if lastIncrementDate == "" {
				lastIncrementDate = inc.EffectiveDate
				if lastIncrementDate == "" {
					lastIncrementDate = inc.IncrementDate
				}
				lastIncrementAmount = inc.IncrementAmount
			}
			if strings.Contains(strings.ToLower(inc.IncrementType), "promotion") {
				totalPromotionsCount++
			}
		}
	}

	initialGross := emp.GrossSalary - totalIncrementAmount
	if initialGross < 0 {
		initialGross = emp.GrossSalary
	}

	// Format employee profile
	companyName := ""
	if emp.Company.CompanyNameEn != "" {
		companyName = emp.Company.CompanyNameEn
	}
	deptName := ""
	if emp.Department != nil && emp.Department.Name != "" {
		deptName = emp.Department.Name
	}
	desigName := ""
	if emp.DesignationRef != nil {
		desigName = emp.DesignationRef.Name
	}
	secName := ""
	if emp.SectionRef != nil {
		secName = emp.SectionRef.Name
	}
	lineName := ""
	if emp.LineRef != nil {
		lineName = emp.LineRef.Name
	}
	groupName := ""
	if emp.GroupRef != nil {
		groupName = emp.GroupRef.Name
	}
	shiftName := ""
	if emp.Shift != nil {
		shiftName = emp.Shift.Name
	}
	joiningDateStr := ""
	if !emp.JoiningDate.IsZero() {
		joiningDateStr = emp.JoiningDate.Format("2006-01-02")
	}

	employeeData := gin.H{
		"id":                  emp.ID,
		"employee_id":         emp.EmployeeID,
		"punch_number":        emp.PunchNumber,
		"name_en":             emp.NameEn,
		"name_bn":             emp.NameBn,
		"phone":               emp.Phone,
		"gender":              emp.Gender,
		"status":              emp.Status,
		"employee_type":       emp.EmployeeType,
		"joining_date":        joiningDateStr,
		"gross_salary":        emp.GrossSalary,
		"basic_salary":        emp.BasicSalary,
		"house_rent":          emp.HouseRent,
		"medical_allowance":   emp.MedicalAllowance,
		"transport_allowance": emp.TransportAllowance,
		"food_allowance":      emp.FoodAllowance,
		"company_id":          emp.CompanyID,
		"company_name":        companyName,
		"department_id":       emp.DepartmentID,
		"department_name":     deptName,
		"designation_id":      emp.DesignationID,
		"designation_name":    desigName,
		"section_id":          emp.SectionID,
		"section_name":        secName,
		"line_id":             emp.LineID,
		"line_name":           lineName,
		"group_id":            emp.GroupID,
		"group_name":          groupName,
		"shift_id":            emp.ShiftID,
		"shift_name":          shiftName,
		"photo_url":           emp.ImageURL,
	}

	c.JSON(http.StatusOK, gin.H{
		"employee": employeeData,
		"summary": gin.H{
			"current_gross":          emp.GrossSalary,
			"current_basic":          emp.BasicSalary,
			"total_increment_amount": totalIncrementAmount,
			"total_increment_count":  totalApprovedCount,
			"total_promotions_count": totalPromotionsCount,
			"last_increment_date":    lastIncrementDate,
			"last_increment_amount":  lastIncrementAmount,
			"initial_gross":          initialGross,
			"total_records":          len(filteredIncrements),
		},
		"increments": filteredIncrements,
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

	// Validate dates: increment_date and effective_date must be YYYY-MM-DD and effective >= increment
	if req.IncrementDate != "" {
		if _, err := time.Parse("2006-01-02", req.IncrementDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid increment_date, expected YYYY-MM-DD"})
			return
		}
	}
	if req.EffectiveDate != "" {
		if _, err := time.Parse("2006-01-02", req.EffectiveDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid effective_date, expected YYYY-MM-DD"})
			return
		}
	}
	if req.IncrementDate != "" && req.EffectiveDate != "" {
		incD, _ := time.Parse("2006-01-02", req.IncrementDate)
		effD, _ := time.Parse("2006-01-02", req.EffectiveDate)
		if effD.Before(incD) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "effective_date cannot be before increment_date"})
			return
		}
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
	// Priority 1: exact employee_id search — ignores all other filters (requested behavior)
	if strings.TrimSpace(req.EmployeeIDSearch) != "" {
		code := strings.TrimSpace(req.EmployeeIDSearch)
		var emp *models.Employee
		emp, err = h.employeeRepo.FindByEmployeeID(code)
		if err == nil && emp != nil && emp.CompanyID == req.CompanyID {
			// Enforce eligibility for search path as well (active + gross>0)
			if emp.Status == "active" && emp.GrossSalary > 0 {
				emps = []models.Employee{*emp}
			} else {
				emps = []models.Employee{}
			}
		} else {
			// Fallback: try punch_number exact
			emp2, err2 := h.employeeRepo.FindByPunchNumber(code)
			if err2 == nil && emp2 != nil && emp2.CompanyID == req.CompanyID && emp2.Status == "active" && emp2.GrossSalary > 0 {
				emps = []models.Employee{*emp2}
				err = nil
			} else if err == nil {
				// keep original err (not found) to return 404 below
				emps = []models.Employee{}
			}
		}
	} else if len(req.EmployeeIDs) > 0 {
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
		var promoDeptID *string = nil
		var promoSecID *string = nil
		var promoLineID *string = nil
		if incType == "promotion" || incType == "promotion_with_increment" {
			if v := strings.TrimSpace(req.NewDesignationID); v != "" {
				newDesigID = &v
			}
			if v := strings.TrimSpace(req.PromoDepartmentID); v != "" {
				promoDeptID = &v
			}
			if v := strings.TrimSpace(req.PromoSectionID); v != "" {
				promoSecID = &v
			}
			if v := strings.TrimSpace(req.PromoLineID); v != "" {
				promoLineID = &v
			}
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
			PromoDepartmentID:     promoDeptID,
			PromoSectionID:        promoSecID,
			PromoLineID:           promoLineID,
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

	// Temporal gate: only mutate live employee if effective_date <= today.
	// Future increments are approved but not yet applied to employees — salary ProcessMonth will pick them up when effective month arrives.
	// This ensures "salary process effective month to after next increment" and old salary for before months.
	isFuture := false
	if inc.EffectiveDate != "" {
		if eff, err := time.Parse("2006-01-02", inc.EffectiveDate); err == nil {
			todayStr := time.Now().Format("2006-01-02")
			today, _ := time.Parse("2006-01-02", todayStr)
			if eff.After(today) {
				isFuture = true
			}
		}
	}
	if !isFuture {
		emp.GrossSalary = inc.NewGross
		emp.BasicSalary = inc.NewBasic
		emp.HouseRent = inc.NewHouse
		emp.MedicalAllowance = inc.NewMedical

		if inc.PromoDepartmentID != nil && *inc.PromoDepartmentID != "" {
			emp.DepartmentID = inc.PromoDepartmentID
		}
		if inc.PromoSectionID != nil && *inc.PromoSectionID != "" {
			emp.SectionID = inc.PromoSectionID
		}
		if inc.NewDesignationID != nil && *inc.NewDesignationID != "" {
			emp.DesignationID = inc.NewDesignationID
		}
		if inc.PromoLineID != nil && *inc.PromoLineID != "" {
			emp.LineID = inc.PromoLineID
		}

		if err := h.employeeRepo.UpdateWithPromotion(emp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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
