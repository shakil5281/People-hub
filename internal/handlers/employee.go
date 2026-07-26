package handlers

import (
	"fmt"
	"net/http"
	"regexp"
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
	Designation   string  `json:"designation"`
	Department    string  `json:"department"`
	Section       string  `json:"section"`
	Line          string  `json:"line"`
	Group         string  `json:"group"`
	Floor         string  `json:"floor"`
	JoiningDate   string  `json:"joining_date"`
	GrossSalary   float64 `json:"gross_salary"`
	Status        string  `json:"status"`
	EmployeeType  string  `json:"employee_type"`
	Gender        string  `json:"gender"`
	CompanyID     string  `json:"company_id"`
	ShiftID       *string `json:"shift_id"`
	DepartmentID  *string `json:"department_id"`
	SectionID     *string `json:"section_id"`
	DesignationID *string `json:"designation_id"`
	LineID        *string `json:"line_id"`
	GroupID       *string `json:"group_id"`
	FloorID       *string `json:"floor_id"`
}

func toEmployeeRow(e models.Employee) EmployeeRow {
	r := EmployeeRow{
		ID:            e.ID,
		EmployeeID:    e.EmployeeID,
		PunchNumber:   e.PunchNumber,
		NameEn:        e.NameEn,
		NameBn:        e.NameBn,
		Phone:         e.Phone,
		GrossSalary:   e.GrossSalary,
		Status:        e.Status,
		EmployeeType:  e.EmployeeType,
		Gender:        e.Gender,
		CompanyID:     e.CompanyID,
		ShiftID:       e.ShiftID,
		DepartmentID:  e.DepartmentID,
		SectionID:     e.SectionID,
		DesignationID: e.DesignationID,
		LineID:        e.LineID,
		GroupID:       e.GroupID,
		FloorID:       e.FloorID,
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
	PresentAddress   string `json:"present_address"`
	PermanentAddress string `json:"permanent_address"`

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
	// Address (permanent)
	PermanentDivisionID string `json:"permanent_division_id"`
	PermanentDistrictID string `json:"permanent_district_id"`
	PermanentUpazilaID  string `json:"permanent_upazila_id"`
	PermanentUnionID    string `json:"permanent_union_id"`

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
	emp.PermanentAddress = req.PermanentAddress

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
	emp.PermanentDivisionID = setPtr(req.PermanentDivisionID)
	emp.PermanentDistrictID = setPtr(req.PermanentDistrictID)
	emp.PermanentUpazilaID = setPtr(req.PermanentUpazilaID)
	emp.PermanentUnionID = setPtr(req.PermanentUnionID)

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
	if v := c.Query("min_salary"); v != "" {
		query = query.Where("gross_salary >= ?", v)
	}
	if v := c.Query("max_salary"); v != "" {
		query = query.Where("gross_salary <= ?", v)
	}

	var total int64
	if err := query.Model(&models.Employee{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	p := utils.ParsePagination(c)
	offset := (p.Page - 1) * p.Limit
	if err := query.Offset(offset).Limit(p.Limit).Find(&employees).Error; err != nil {
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

	database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").First(&emp, "id = ?", emp.ID)
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
	if err := database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").First(&emp, "id = ?", id).Error; err != nil {
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
	if err := database.DB.Preload("Department").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").Preload("Shift").Preload("Company").Where("employee_id = ?", code).First(&emp).Error; err != nil {
		if err2 := database.DB.Preload("Department").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").Preload("Shift").Preload("Company").Where("punch_number = ?", code).First(&emp).Error; err2 != nil {
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
	if err := database.DB.Preload("User").Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").First(&emp, "id = ?", id).Error; err != nil {
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

	response := gin.H{
		"employee":   emp,
		"attendance": attendanceCounts,
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
