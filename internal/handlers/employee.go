package handlers

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type EmployeeHandler struct{}

func NewEmployeeHandler() *EmployeeHandler {
	return &EmployeeHandler{}
}

type EmployeeRow struct {
	ID            string  `json:"id"`
	EmployeeID    string  `json:"employee_id"`
	PunchNumber   string  `json:"punch_number"`
	NameEn        string  `json:"name_en"`
	NameBn        string  `json:"name_bn"`
	Phone         string  `json:"phone"`
	NID           string  `json:"nid"`
	DateOfBirth   string  `json:"date_of_birth"`
	Designation   string  `json:"designation"`
	Department    string  `json:"department"`
	Section       string  `json:"section"`
	Line          string  `json:"line"`
	Group         string  `json:"group"`
	Floor         string  `json:"floor"`
	JoiningDate   string  `json:"joining_date"`
	GrossSalary        float64 `json:"gross_salary"`
	BasicSalary        float64 `json:"basic_salary"`
	HouseRent          float64 `json:"house_rent"`
	MedicalAllowance   float64 `json:"medical_allowance"`
	TransportAllowance float64 `json:"transport_allowance"`
	FoodAllowance      float64 `json:"food_allowance"`
	Status        string  `json:"status"`
	EmployeeType  string  `json:"employee_type"`
	Gender        string  `json:"gender"`
	AccountType   string  `json:"account_type"`
	AccountNumber string  `json:"account_number"`
	CompanyID     string  `json:"company_id"`
	ShiftID       *string `json:"shift_id"`
	DepartmentID  *string `json:"department_id"`
	SectionID     *string `json:"section_id"`
	DesignationID *string `json:"designation_id"`
	LineID        *string `json:"line_id"`
	GroupID       *string `json:"group_id"`
	FloorID       *string `json:"floor_id"`
	ImageURL      string  `json:"image_url"`
	SignatureURL  string  `json:"signature_url"`

	PresentAddress     string `json:"present_address"`
	PresentAddressBn   string `json:"present_address_bn"`
	PresentPostOffice  string `json:"present_post_office"`
	PresentPostCode    string `json:"present_post_code"`
	PermanentAddress   string `json:"permanent_address"`
	PermanentAddressBn string `json:"permanent_address_bn"`
	PermanentPostOffice string `json:"permanent_post_office"`
	PermanentPostCode  string `json:"permanent_post_code"`

	PresentDivisionName   string `json:"present_division_name"`
	PresentDistrictName   string `json:"present_district_name"`
	PermanentDivisionName string `json:"permanent_division_name"`
	PermanentDistrictName string `json:"permanent_district_name"`
}

func toEmployeeRow(e models.Employee) EmployeeRow {
	basic := e.BasicSalary
	house := e.HouseRent
	medical := e.MedicalAllowance
	transport := e.TransportAllowance
	food := e.FoodAllowance
	if medical == 0 {
		medical = 750
	}
	if transport == 0 {
		transport = 450
	}
	if food == 0 {
		food = 1250
	}
	if (basic <= 0 || house <= 0) && e.GrossSalary > 0 {
		core := e.GrossSalary - medical - transport - food
		if core > 0 {
			basic = math.Round(core / 1.5)
			house = core - basic
		}
	}

	r := EmployeeRow{
		ID:                 e.ID,
		EmployeeID:         e.EmployeeID,
		PunchNumber:        e.PunchNumber,
		NameEn:             e.NameEn,
		NameBn:             e.NameBn,
		Phone:              e.Phone,
		NID:                e.NID,
		DateOfBirth:        e.DateOfBirth,
		GrossSalary:        e.GrossSalary,
		BasicSalary:        basic,
		HouseRent:          house,
		MedicalAllowance:   medical,
		TransportAllowance: transport,
		FoodAllowance:      food,
		Status:             e.Status,
		EmployeeType:       e.EmployeeType,
		Gender:             e.Gender,
		AccountType:        e.AccountType,
		AccountNumber:      e.AccountNumber,
		CompanyID:          e.CompanyID,
		ShiftID:            e.ShiftID,
		DepartmentID:       e.DepartmentID,
		SectionID:          e.SectionID,
		DesignationID:      e.DesignationID,
		LineID:             e.LineID,
		GroupID:            e.GroupID,
		FloorID:            e.FloorID,
		ImageURL:           e.ImageURL,
		SignatureURL:       e.SignatureURL,
		PresentAddress:     e.PresentAddress,
		PresentAddressBn:   e.PresentAddressBn,
		PermanentAddress:   e.PermanentAddress,
		PermanentAddressBn: e.PermanentAddressBn,
	}
	if e.PresentPostOffice != nil {
		r.PresentPostOffice = *e.PresentPostOffice
	}
	if e.PresentPostCode != nil {
		r.PresentPostCode = *e.PresentPostCode
	}
	if e.PermanentPostOffice != nil {
		r.PermanentPostOffice = *e.PermanentPostOffice
	}
	if e.PermanentPostCode != nil {
		r.PermanentPostCode = *e.PermanentPostCode
	}
	if e.DesignationRef != nil {
		r.Designation = e.DesignationRef.Name
	}
	if e.Department != nil {
		r.Department = e.Department.Name
	}
	if e.SectionRef != nil {
		r.Section = e.SectionRef.Name
	}
	if e.LineRef != nil {
		r.Line = e.LineRef.Name
	}
	if e.GroupRef != nil {
		r.Group = e.GroupRef.Name
	}
	if e.FloorRef != nil {
		r.Floor = e.FloorRef.Name
	}
	if !e.JoiningDate.IsZero() {
		r.JoiningDate = e.JoiningDate.Format("2006-01-02")
	}
	if e.PresentDivision != nil {
		r.PresentDivisionName = e.PresentDivision.Name
	}
	if e.PresentDistrict != nil {
		r.PresentDistrictName = e.PresentDistrict.Name
	}
	if e.PermanentDivision != nil {
		r.PermanentDivisionName = e.PermanentDivision.Name
	}
	if e.PermanentDistrict != nil {
		r.PermanentDistrictName = e.PermanentDistrict.Name
	}
	return r
}

func toEmployeeRows(list []models.Employee) []EmployeeRow {
	res := make([]EmployeeRow, len(list))
	for i, e := range list {
		res[i] = toEmployeeRow(e)
	}
	return res
}

type CreateEmployeeRequest struct {
	// Personal
	NameEn           string `json:"name_en"`
	NameBn           string `json:"name_bn"`
	FatherName       string `json:"father_name"`
	MotherName       string `json:"mother_name"`
	DateOfBirth      string `json:"date_of_birth"`
	Gender           string `json:"gender"`
	BloodGroup       string `json:"blood_group"`
	MaritalStatus    string `json:"marital_status"`
	Religion         string `json:"religion"`
	Nationality      string `json:"nationality"`
	NID              string `json:"nid"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	PresentAddress     string `json:"present_address"`
	PresentAddressBn   string `json:"present_address_bn"`
	PermanentAddress   string `json:"permanent_address"`
	PermanentAddressBn string `json:"permanent_address_bn"`

	// Family
	SpouseName         string `json:"spouse_name"`
	EmergencyContact   string `json:"emergency_contact"`
	EmergencyPhone     string `json:"emergency_phone"`
	NumberOfDependents int    `json:"number_of_dependents"`

	// Office
	CompanyID     string `json:"company_id" binding:"required"`
	DepartmentID  string `json:"department_id"`
	SectionID     string `json:"section_id"`
	DesignationID string `json:"designation_id"`
	LineID        string `json:"line_id"`
	GroupID       string `json:"group_id"`
	FloorID       string `json:"floor_id"`
	EmployeeID    string `json:"employee_id" binding:"required"`
	PunchNumber   string `json:"punch_number" binding:"required"`
	EmployeeType  string `json:"employee_type"`
	Grade         string `json:"grade"`
	JoiningDate   string `json:"joining_date" binding:"required"`
	ShiftID       string `json:"shift_id"`
	ReportsTo     string `json:"reports_to"`

	// Address (present)
	PresentDivisionID string `json:"present_division_id"`
	PresentDistrictID string `json:"present_district_id"`
	PresentUpazilaID  string `json:"present_upazila_id"`
	PresentUnionID    string `json:"present_union_id"`
	PresentPostOffice string `json:"present_post_office"`
	PresentPostCode   string `json:"present_post_code"`
	// Address (permanent)
	PermanentDivisionID string `json:"permanent_division_id"`
	PermanentDistrictID string `json:"permanent_district_id"`
	PermanentUpazilaID  string `json:"permanent_upazila_id"`
	PermanentUnionID    string `json:"permanent_union_id"`
	PermanentPostOffice string `json:"permanent_post_office"`
	PermanentPostCode   string `json:"permanent_post_code"`

	// Salary
	GrossSalary        float64 `json:"gross_salary"`
	BasicSalary        float64 `json:"basic_salary"`
	HouseRent          float64 `json:"house_rent"`
	TransportAllowance float64 `json:"transport_allowance"`
	FoodAllowance      float64 `json:"food_allowance"`
	MedicalAllowance   float64 `json:"medical_allowance"`
	OtherAllowance     float64 `json:"other_allowance"`
	// Account
	AccountType   string `json:"account_type"`
	AccountNumber string `json:"account_number"`

	// Status
	Status         string `json:"status"`
	OverTimeStatus bool   `json:"over_time_status"`

	// Media
	ImageURL     string `json:"image_url"`
	SignatureURL string `json:"signature_url"`
}

func bindEmployeeFields(req *CreateEmployeeRequest, emp *models.Employee) {
	// Personal
	emp.NameEn = req.NameEn
	emp.NameBn = req.NameBn
	emp.FatherName = req.FatherName
	emp.MotherName = req.MotherName
	emp.DateOfBirth = req.DateOfBirth
	emp.Gender = req.Gender
	emp.BloodGroup = req.BloodGroup
	emp.MaritalStatus = req.MaritalStatus
	emp.Religion = req.Religion
	if req.Nationality != "" {
		emp.Nationality = req.Nationality
	} else {
		emp.Nationality = "Bangladeshi"
	}
	emp.NID = req.NID
	emp.Phone = req.Phone
	emp.Email = req.Email
	emp.PresentAddress = req.PresentAddress
	emp.PresentAddressBn = req.PresentAddressBn
	emp.PermanentAddress = req.PermanentAddress
	emp.PermanentAddressBn = req.PermanentAddressBn

	// Family
	emp.SpouseName = req.SpouseName
	emp.EmergencyContact = req.EmergencyContact
	emp.EmergencyPhone = req.EmergencyPhone
	emp.NumberOfDependents = req.NumberOfDependents

	// Office
	emp.CompanyID = req.CompanyID
	setPtr := func(val string) *string {
		if val == "" {
			return nil
		}
		return &val
	}
	emp.DepartmentID = setPtr(req.DepartmentID)
	emp.SectionID = setPtr(req.SectionID)
	emp.DesignationID = setPtr(req.DesignationID)
	emp.LineID = setPtr(req.LineID)
	emp.GroupID = setPtr(req.GroupID)
	emp.FloorID = setPtr(req.FloorID)
	emp.EmployeeID = req.EmployeeID
	emp.PunchNumber = req.PunchNumber
	emp.EmployeeType = req.EmployeeType
	emp.Grade = req.Grade
	emp.ShiftID = setPtr(req.ShiftID)
	emp.ReportsTo = setPtr(req.ReportsTo)
	emp.PresentDivisionID = setPtr(req.PresentDivisionID)
	emp.PresentDistrictID = setPtr(req.PresentDistrictID)
	emp.PresentUpazilaID = setPtr(req.PresentUpazilaID)
	emp.PresentUnionID = setPtr(req.PresentUnionID)
	emp.PresentPostOffice = setPtr(req.PresentPostOffice)
	emp.PresentPostCode = setPtr(req.PresentPostCode)
	emp.PermanentDivisionID = setPtr(req.PermanentDivisionID)
	emp.PermanentDistrictID = setPtr(req.PermanentDistrictID)
	emp.PermanentUpazilaID = setPtr(req.PermanentUpazilaID)
	emp.PermanentUnionID = setPtr(req.PermanentUnionID)
	emp.PermanentPostOffice = setPtr(req.PermanentPostOffice)
	emp.PermanentPostCode = setPtr(req.PermanentPostCode)

	// Salary
	emp.GrossSalary = req.GrossSalary
	emp.BasicSalary = req.BasicSalary
	emp.HouseRent = req.HouseRent
	emp.TransportAllowance = req.TransportAllowance
	emp.FoodAllowance = req.FoodAllowance
	emp.MedicalAllowance = req.MedicalAllowance
	emp.OtherAllowance = req.OtherAllowance

	// Account
	emp.AccountType = req.AccountType
	emp.AccountNumber = req.AccountNumber

	// Status
	if req.Status != "" {
		emp.Status = req.Status
	}
	emp.OverTimeStatus = req.OverTimeStatus

	// Media
	emp.ImageURL = req.ImageURL
	emp.SignatureURL = req.SignatureURL
}

func validateAccount(accountType, accountNumber string) string {
	if accountType == "" && accountNumber == "" {
		return ""
	}
	if accountType == "" {
		return "account_type is required when account_number is provided"
	}
	if accountNumber == "" {
		return "account_number is required when account_type is provided"
	}
	if accountType != "mCash" && accountType != "Card" {
		return "account_type must be mCash or Card"
	}
	digitRegex := regexp.MustCompile(`^\d+$`)
	if !digitRegex.MatchString(accountNumber) {
		return "account_number must contain only digits"
	}
	if accountType == "mCash" && len(accountNumber) != 12 {
		return "account_number must be exactly 12 digits for mCash"
	}
	if accountType == "Card" && len(accountNumber) != 17 {
		return "account_number must be exactly 17 digits for Card"
	}
	return ""
}

// GetEmployees godoc
//
// @Summary      List employees
// @Description  Get all employees with optional filters
// @Tags         Employees
// @Security     BearerAuth
// @Produce      json
// @Param        company_id      query string false "Filter by company ID"
// @Param        department_id   query string false "Filter by department ID"
// @Param        section_id      query string false "Filter by section ID"
// @Param        designation_id  query string false "Filter by designation ID"
// @Param        line_id         query string false "Filter by line ID"
// @Param        shift_id        query string false "Filter by shift ID"
// @Param        group_id        query string false "Filter by group ID"
// @Param        floor_id        query string false "Filter by floor ID"
// @Param        status          query string false "Filter by status (active/inactive)"
// @Param        employee_id     query string false "Filter by employee ID (partial match)"
// @Param        gender          query string false "Filter by gender"
// @Param        blood_group     query string false "Filter by blood group"
// @Param        employee_type   query string false "Filter by employee type"
// @Param        min_salary      query string false "Minimum gross salary"
// @Param        max_salary      query string false "Maximum gross salary"
// @Param        page            query int    false "Page number (default 1)"
// @Param        limit           query int    false "Page size (default 20, max 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /employees [get]
func (h *EmployeeHandler) GetEmployees(c *gin.Context) {
	var employees []models.Employee
	query := database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef")
	query = query.Preload("PresentDivision").Preload("PresentDistrict").Preload("PresentUpazila").Preload("PresentUnion")
	query = query.Preload("PermanentDivision").Preload("PermanentDistrict").Preload("PermanentUpazila").Preload("PermanentUnion")

	if v := c.Query("company_id"); v != "" {
		query = query.Where("company_id = ?", v)
	}
	if v := c.Query("department_id"); v != "" {
		query = query.Where("department_id = ?", v)
	}
	if v := c.Query("section_id"); v != "" {
		query = query.Where("section_id = ?", v)
	}
	if v := c.Query("designation_id"); v != "" {
		query = query.Where("designation_id = ?", v)
	}
	if v := c.Query("line_id"); v != "" {
		query = query.Where("line_id = ?", v)
	}
	if v := c.Query("shift_id"); v != "" {
		query = query.Where("shift_id = ?", v)
	}
	if v := c.Query("group_id"); v != "" {
		query = query.Where("group_id = ?", v)
	}
	if v := c.Query("floor_id"); v != "" {
		query = query.Where("floor_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("employee_id"); v != "" {
		query = query.Where("employee_id ILIKE ?", "%"+v+"%")
	}
	if v := c.Query("gender"); v != "" {
		query = query.Where("gender = ?", v)
	}
	if v := c.Query("blood_group"); v != "" {
		query = query.Where("blood_group = ?", v)
	}
	if v := c.Query("employee_type"); v != "" {
		query = query.Where("employee_type = ?", v)
	}
	if v := c.Query("account_type"); v != "" {
		if strings.EqualFold(v, "none") || strings.EqualFold(v, "unassigned") || strings.EqualFold(v, "hold") {
			query = query.Where("account_type IS NULL OR TRIM(account_type) = '' OR LOWER(account_type) IN ('none', 'hold')")
		} else {
			query = query.Where("LOWER(account_type) = LOWER(?)", v)
		}
	}
	if v := c.Query("min_salary"); v != "" {
		query = query.Where("gross_salary >= ?", v)
	}
	if v := c.Query("max_salary"); v != "" {
		query = query.Where("gross_salary <= ?", v)
	}
	if v := c.Query("joining_date_from"); v != "" {
		query = query.Where("joining_date >= ?", v)
	} else if v := c.Query("joining_from"); v != "" {
		query = query.Where("joining_date >= ?", v)
	} else if v := c.Query("start_date"); v != "" {
		query = query.Where("joining_date >= ?", v)
	}
	if v := c.Query("joining_date_to"); v != "" {
		query = query.Where("joining_date <= ?", v)
	} else if v := c.Query("joining_to"); v != "" {
		query = query.Where("joining_date <= ?", v)
	} else if v := c.Query("end_date"); v != "" {
		query = query.Where("joining_date <= ?", v)
	}

	// Filter by joining month anniversary (e.g. for Govt policy increment: joined after 2024 matching selected month)
	// Example: month=8 (Aug), year=2026 -> finds employees with joining_date in August for years >= 2024 and < 2026
	if monthStr := c.Query("joining_month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			targetYear := time.Now().Year()
			if yrStr := c.Query("joining_year"); yrStr != "" {
				if yr, err := strconv.Atoi(yrStr); err == nil && yr > 2000 {
					targetYear = yr
				}
			} else if yrStr := c.Query("target_year"); yrStr != "" {
				if yr, err := strconv.Atoi(yrStr); err == nil && yr > 2000 {
					targetYear = yr
				}
			} else if yrStr := c.Query("year"); yrStr != "" {
				if yr, err := strconv.Atoi(yrStr); err == nil && yr > 2000 {
					targetYear = yr
				}
			}

			startYear := 2024
			if afterYrStr := c.Query("joining_after_year"); afterYrStr != "" {
				if ay, err := strconv.Atoi(afterYrStr); err == nil {
					startYear = ay
				}
			}

			query = query.Where("EXTRACT(MONTH FROM joining_date) = ? AND joining_date >= ? AND EXTRACT(YEAR FROM joining_date) < ?",
				m, fmt.Sprintf("%04d-01-01", startYear), targetYear)
		}
	} else if monthStr := c.Query("increment_month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			targetYear := time.Now().Year()
			if yrStr := c.Query("increment_year"); yrStr != "" {
				if yr, err := strconv.Atoi(yrStr); err == nil && yr > 2000 {
					targetYear = yr
				}
			} else if yrStr := c.Query("year"); yrStr != "" {
				if yr, err := strconv.Atoi(yrStr); err == nil && yr > 2000 {
					targetYear = yr
				}
			}

			startYear := 2024
			if afterYrStr := c.Query("joining_after_year"); afterYrStr != "" {
				if ay, err := strconv.Atoi(afterYrStr); err == nil {
					startYear = ay
				}
			}

			query = query.Where("EXTRACT(MONTH FROM joining_date) = ? AND joining_date >= ? AND EXTRACT(YEAR FROM joining_date) < ?",
				m, fmt.Sprintf("%04d-01-01", startYear), targetYear)
		}
	}

	var total int64
	if err := query.Model(&models.Employee{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	p := utils.ParsePagination(c)
	offset := (p.Page - 1) * p.Limit
	if err := query.Order("LENGTH(employee_id) ASC, employee_id ASC").Offset(offset).Limit(p.Limit).Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(toEmployeeRows(employees), total, p))
}

// CreateEmployee godoc
//
// @Summary      Create employee
// @Description  Create a new employee record
// @Tags         Employees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateEmployeeRequest true "Employee details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /employees [post]
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateAccount(req.AccountType, req.AccountNumber); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	jd, err := time.Parse("2006-01-02", req.JoiningDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid joining_date format, use YYYY-MM-DD"})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	userID := c.GetString("user_id")

	var existing models.Employee
	if err := database.DB.Where("(employee_id = ? OR punch_number = ?) AND company_id = ?", req.EmployeeID, req.PunchNumber, req.CompanyID).First(&existing).Error; err == nil {
		if existing.EmployeeID == req.EmployeeID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id already exists"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "punch_number already exists"})
		}
		return
	}

	employee := models.Employee{
		EmployeeID:  req.EmployeeID,
		PunchNumber: req.PunchNumber,
		CompanyID:   req.CompanyID,
		JoiningDate: jd,
		Status:      status,
		CreatedBy:   &userID,
	}

	bindEmployeeFields(&req, &employee)

	if err := database.DB.Create(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").First(&employee, "id = ?", employee.ID)
	c.JSON(http.StatusCreated, employee)
}

// UpdateEmployee godoc
//
// @Summary      Update employee
// @Description  Update an existing employee record
// @Tags         Employees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Employee ID"
// @Param        request body CreateEmployeeRequest true "Employee details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /employees/{id} [put]
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateAccount(req.AccountType, req.AccountNumber); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if req.JoiningDate != "" {
		jd, err := time.Parse("2006-01-02", req.JoiningDate)
		if err == nil {
			emp.JoiningDate = jd
		}
	}

	bindEmployeeFields(&req, &emp)

	userID := c.GetString("user_id")
	emp.UpdatedBy = &userID

	if err := database.DB.Save(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").
		Preload("PresentDivision").Preload("PresentDistrict").Preload("PresentUpazila").Preload("PresentUnion").
		Preload("PermanentDivision").Preload("PermanentDistrict").Preload("PermanentUpazila").Preload("PermanentUnion").
		First(&emp, "id = ?", emp.ID)
	c.JSON(http.StatusOK, emp)
}

// GetEmployee godoc
//
// @Summary      Get employee by ID
// @Description  Get a single employee by ID
// @Tags         Employees
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Employee ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /employees/{id} [get]
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	var emp models.Employee
	if err := database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").
		Preload("PresentDivision").Preload("PresentDistrict").Preload("PresentUpazila").Preload("PresentUnion").
		Preload("PermanentDivision").Preload("PermanentDistrict").Preload("PermanentUpazila").Preload("PermanentUnion").
		First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}
	c.JSON(http.StatusOK, emp)
}

// GetEmployeeByCode godoc
//
// @Summary      Get employee by employee code
// @Description  Get a single employee by their business employee_id (e.g. "2857")
// @Tags         Employees
// @Security     BearerAuth
// @Produce      json
// @Param        code   path     string true "Employee business code"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /employees/by-code/{code} [get]
func (h *EmployeeHandler) GetEmployeeByCode(c *gin.Context) {
	code := c.Param("code")
	var emp models.Employee
	query := database.DB.Preload("Department").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").Preload("Shift").Preload("Company").
		Preload("PresentDivision").Preload("PresentDistrict").Preload("PresentUpazila").Preload("PresentUnion").
		Preload("PermanentDivision").Preload("PermanentDistrict").Preload("PermanentUpazila").Preload("PermanentUnion")
	if err := query.Where("employee_id = ?", code).First(&emp).Error; err != nil {
		if err2 := query.Where("punch_number = ?", code).First(&emp).Error; err2 != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
	}
	c.JSON(http.StatusOK, emp)
}

type AttendanceCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// GetEmployeeProfile godoc
//
// @Summary      Get employee profile with attendance and salary
// @Description  Get full employee details, current month attendance summary, and latest salary
// @Tags         Employees
// @Security     BearerAuth
// @Produce      json
// @Param        id       path string true "Employee ID"
// @Param        month    query int    false "Month (default current)"
// @Param        year     query int    false "Year (default current)"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /employees/{id}/profile [get]
func (h *EmployeeHandler) GetEmployeeProfile(c *gin.Context) {
	id := c.Param("id")

	var emp models.Employee
	if err := database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").
		Preload("PresentDivision").Preload("PresentDistrict").Preload("PresentUpazila").Preload("PresentUnion").
		Preload("PermanentDivision").Preload("PermanentDistrict").Preload("PermanentUpazila").Preload("PermanentUnion").
		First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	now := time.Now()
	month, _ := parseIntParam(c.Query("month"), int(now.Month()))
	year, _ := parseIntParam(c.Query("year"), now.Year())

	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	endDate := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	var attendanceCounts []AttendanceCount
	database.DB.Model(&models.Attendance{}).
		Select("status, count(*) as count").
		Where("employee_id = ? AND date BETWEEN ? AND ? AND deleted_at IS NULL", emp.EmployeeID, startDate, endDate).
		Group("status").
		Find(&attendanceCounts)

	var salary models.Salary
	salaryErr := database.DB.Where("employee_id = ? AND month = ? AND year = ?", emp.EmployeeID, month, year).First(&salary).Error

	// Generate all months from joining date to current month
	startYear := now.Year()
	startMonth := int(now.Month())
	if !emp.JoiningDate.IsZero() {
		startYear = emp.JoiningDate.Year()
		startMonth = int(emp.JoiningDate.Month())
	}
	if startYear > now.Year() || (startYear == now.Year() && startMonth > int(now.Month())) {
		startYear = now.Year()
		startMonth = int(now.Month())
	}
	if startYear < 1990 {
		startYear = 2020
		startMonth = 1
	}

	type ymPair struct {
		Year  int
		Month int
	}
	var allMonths []ymPair
	currY, currM := now.Year(), int(now.Month())
	for {
		allMonths = append(allMonths, ymPair{Year: currY, Month: currM})
		if currY == startYear && currM == startMonth {
			break
		}
		currM--
		if currM < 1 {
			currM = 12
			currY--
		}
		if currY < startYear {
			break
		}
	}

	// 1. All Salary History
	var allSalaries []models.Salary
	database.DB.Where("employee_id = ?", emp.EmployeeID).Find(&allSalaries)
	salaryMap := make(map[string]models.Salary)
	for _, s := range allSalaries {
		k := fmt.Sprintf("%d-%d", s.Year, s.Month)
		salaryMap[k] = s
	}

	var salaryHistory []gin.H
	for _, m := range allMonths {
		k := fmt.Sprintf("%d-%d", m.Year, m.Month)
		mName := time.Month(m.Month).String()
		item := gin.H{
			"month":      m.Month,
			"year":       m.Year,
			"month_name": mName,
			"salary":     nil,
		}
		if s, ok := salaryMap[k]; ok {
			item["salary"] = s
		}
		salaryHistory = append(salaryHistory, item)
	}

	// 2. All Attendance History
	type attCountRow struct {
		Year   int    `gorm:"column:year"`
		Month  int    `gorm:"column:month"`
		Status string `gorm:"column:status"`
		Count  int    `gorm:"column:count"`
	}
	var attCountRows []attCountRow
	database.DB.Model(&models.Attendance{}).
		Select("EXTRACT(YEAR FROM date::date)::int as year, EXTRACT(MONTH FROM date::date)::int as month, status, count(*) as count").
		Where("employee_id = ? AND deleted_at IS NULL", emp.EmployeeID).
		Group("EXTRACT(YEAR FROM date::date), EXTRACT(MONTH FROM date::date), status").
		Find(&attCountRows)

	attMap := make(map[string]map[string]int)
	for _, ac := range attCountRows {
		k := fmt.Sprintf("%d-%d", ac.Year, ac.Month)
		if _, ok := attMap[k]; !ok {
			attMap[k] = make(map[string]int)
		}
		attMap[k][ac.Status] = ac.Count
	}

	var attendanceHistory []gin.H
	for _, m := range allMonths {
		k := fmt.Sprintf("%d-%d", m.Year, m.Month)
		mName := time.Month(m.Month).String()
		statusMap := attMap[k]
		if statusMap == nil {
			statusMap = make(map[string]int)
		}

		present := statusMap["present"]
		absent := statusMap["absent"]
		late := statusMap["late"]
		leave := statusMap["on_leave"] + statusMap["leave"]
		weekend := statusMap["weekend"]
		halfDay := statusMap["half_day"]
		total := 0
		var breakdown []AttendanceCount
		for s, c := range statusMap {
			total += c
			breakdown = append(breakdown, AttendanceCount{Status: s, Count: c})
		}

		attendanceHistory = append(attendanceHistory, gin.H{
			"month":        m.Month,
			"year":         m.Year,
			"month_name":   mName,
			"total_days":   total,
			"present_days": present,
			"absent_days":  absent,
			"late_days":    late,
			"leave_days":   leave,
			"weekend_days": weekend,
			"half_days":    halfDay,
			"breakdown":    breakdown,
		})
	}

	response := gin.H{
		"employee":           emp,
		"attendance":         attendanceCounts,
		"salary_history":     salaryHistory,
		"attendance_history": attendanceHistory,
	}

	if salaryErr == nil {
		response["salary"] = salary
	}

	c.JSON(http.StatusOK, response)
}

func parseIntParam(val string, defaultVal int) (int, error) {
	if val == "" {
		return defaultVal, nil
	}
	var n int
	_, err := fmt.Sscanf(val, "%d", &n)
	if err != nil {
		return defaultVal, err
	}
	return n, nil
}

// DeleteEmployee godoc
//
// @Summary      Delete employee
// @Description  Soft delete an employee
// @Tags         Employees
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Employee ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /employees/{id} [delete]
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	var emp models.Employee
	if err := database.DB.First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}
	if err := database.DB.Delete(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "employee deleted"})
}

type UpdateSalaryAccountRequest struct {
	AccountType   string `json:"account_type"`
	AccountNumber string `json:"account_number"`
}

// UpdateSalaryAccount godoc
//
// @Summary      Update employee salary account
// @Description  Update only the account_type and account_number for an employee
// @Tags         Employees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Employee ID or UUID"
// @Param        request body     UpdateSalaryAccountRequest true "Salary account details"
// @Success      200     {object} map[string]interface{}
// @Failure      400     {object} map[string]string
// @Failure      404     {object} map[string]string
// @Failure      500     {object} map[string]string
// @Router       /employees/{id}/salary-account [put]
func (h *EmployeeHandler) UpdateSalaryAccount(c *gin.Context) {
	id := c.Param("id")
	var emp models.Employee
	if err := database.DB.Where("id = ? OR employee_id = ?", id, id).First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	var req UpdateSalaryAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accType := strings.TrimSpace(req.AccountType)
	accNum := strings.TrimSpace(req.AccountNumber)

	if msg := validateAccount(accType, accNum); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	userID := c.GetString("user_id")
	updates := map[string]interface{}{
		"account_type":   accType,
		"account_number": accNum,
		"updated_at":     time.Now(),
	}
	if userID != "" {
		updates["updated_by"] = userID
	}

	if err := database.DB.Model(&models.Employee{}).Where("id = ?", emp.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Salary account updated successfully",
		"employee_id":    emp.EmployeeID,
		"account_type":   accType,
		"account_number": accNum,
	})
}
