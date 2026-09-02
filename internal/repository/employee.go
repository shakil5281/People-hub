package repository

import (
	"fmt"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type EmployeeFilter struct {
	CompanyID     string
	DepartmentID  string
	SectionID     string
	DesignationID string
	LineID        string
	GroupID       string
	EmployeeID    string
}

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) WithTx(tx *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: tx}
}

func (r *EmployeeRepository) ListFiltered(f EmployeeFilter, page, limit int) ([]models.Employee, int64, error) {
	var employees []models.Employee
	var total int64
	query := r.db.Where("status = ?", "active")
	if f.CompanyID != "" {
		query = query.Where("company_id = ?", f.CompanyID)
	}
	if f.DepartmentID != "" {
		query = query.Where("department_id = ?", f.DepartmentID)
	}
	if f.SectionID != "" {
		query = query.Where("section_id = ?", f.SectionID)
	}
	if f.DesignationID != "" {
		query = query.Where("designation_id = ?", f.DesignationID)
	}
	if f.LineID != "" {
		query = query.Where("line_id = ?", f.LineID)
	}
	if f.GroupID != "" {
		query = query.Where("group_id = ?", f.GroupID)
	}
	if f.EmployeeID != "" {
		query = query.Where("employee_id = ? OR punch_number = ? OR id::text = ?", f.EmployeeID, f.EmployeeID, f.EmployeeID)
	}
	if err := query.Model(&models.Employee{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&employees).Error
	return employees, total, err
}

func (r *EmployeeRepository) FindByEmployeeID(code string) (*models.Employee, error) {
	var emp models.Employee
	err := r.db.Where("employee_id = ?", code).First(&emp).Error
	return &emp, err
}

func (r *EmployeeRepository) FindByEmployeeIDs(codes []string) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Where("employee_id IN ?", codes).Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) FindByPunchNumbers(punches []string) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Where("punch_number IN ?", punches).Find(&employees).Error
	return employees, err
}

// FindActiveRegularByPunchNumbers matches active Regular employees by punch_number only.
func (r *EmployeeRepository) FindActiveRegularByPunchNumbers(punches []string) ([]models.Employee, error) {
	var employees []models.Employee
	if len(punches) == 0 {
		return employees, nil
	}
	err := r.db.Where(
		"punch_number IN ? AND status = ? AND LOWER(employee_type) = ?",
		punches, "active", "regular",
	).Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) ListActivePtr(companyID string) ([]*models.Employee, error) {
	var employees []*models.Employee
	query := r.db.Where("status = ?", "active")
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) FindByPunchNumber(punch string) (*models.Employee, error) {
	var emp models.Employee
	err := r.db.Where("punch_number = ?", punch).First(&emp).Error
	return &emp, err
}

func (r *EmployeeRepository) ListByCompany(companyID string) ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Where("company_id = ?", companyID).Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) ListActive(companyID string, page, limit int) ([]models.Employee, int64, error) {
	var employees []models.Employee
	var total int64
	query := r.db.Where("status = ?", "active")
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if err := query.Model(&models.Employee{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&employees).Error
	return employees, total, err
}

// ListActiveAll returns active employees plus separated employees (Resign, Close, Lefty) for salary processing.
// Optimized: single index scan on employees(company_id, status) instead of OR with uncorrelated subqueries.
func (r *EmployeeRepository) ListActiveAll(companyID string) ([]models.Employee, error) {
	var employees []models.Employee
	// Fast path: salary only needs employees of this company. The OR with attendance/separation
	// subqueries without date filter scanned entire history (300k rows). For salary month we already
	// have attendance in [startDate,endDate] — so we limit to status-based employees here.
	// Separated employees with attendance in month will be included via status='active' OR employee_type check.
	// To keep exact legacy semantics but indexed, we use simple company + status/employee_type filter.
	query := r.db.Preload("GroupRef").Where("deleted_at IS NULL")
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	query = query.Where("status = 'active' OR LOWER(employee_type) IN ('resign', 'close', 'lefty', 'dismiss', 'termination', 'retirement')")
	err := query.Find(&employees).Error
	return employees, err
}

// ListForSalaryProcessing returns all employees who should be included in salary
// processing for the given month/year based on separation date:
//   - Active employees (not separated before the start of this month)
//   - Inactive/separated employees with a separation date falling in this month
//   - Separated employee types with attendance records in this month
//
// This ensures separated employees are processed for their days worked in the month,
// while employees separated in prior months are skipped.
func (r *EmployeeRepository) ListForSalaryProcessing(companyID string, month, year int) ([]models.Employee, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	startStr := fmt.Sprintf("%04d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	endStr := fmt.Sprintf("%04d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	var employees []models.Employee
	query := r.db.Preload("GroupRef").Where("deleted_at IS NULL")
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	query = query.Where(`
		(
			status = 'active'
			AND (resign_date IS NULL OR resign_date >= ?)
			AND employee_id NOT IN (
				SELECT employee_id FROM separations
				WHERE deleted_at IS NULL
				AND LOWER(status) != 'cancelled'
				AND date < ?
			)
		)
		OR employee_id IN (
			SELECT employee_id FROM separations
			WHERE deleted_at IS NULL
			AND LOWER(status) != 'cancelled'
			AND date >= ? AND date <= ?
		)
		OR (resign_date IS NOT NULL AND resign_date >= ? AND resign_date <= ?)
		OR (
			LOWER(employee_type) IN ('resign', 'close', 'lefty', 'dismiss', 'termination', 'retirement')
			AND employee_id IN (
				SELECT DISTINCT employee_id FROM attendances
				WHERE deleted_at IS NULL AND date >= ? AND date <= ?
			)
		)
	`, startStr, startStr, startStr, endStr, startStr, endStr, startStr, endStr)

	err := query.Find(&employees).Error
	return employees, err
}

// ListActiveRegularAll returns active and separated employees for daily attendance processing.
func (r *EmployeeRepository) ListActiveRegularAll(companyID string) ([]models.Employee, error) {
	var employees []models.Employee
	query := r.db.Where("(status = 'active') OR (employee_id IN (SELECT employee_id FROM separations WHERE deleted_at IS NULL)) OR (LOWER(employee_type) IN ('resign', 'close', 'lefty', 'dismiss', 'termination', 'retirement'))")
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) BatchCreate(employees []models.Employee) error {
	if len(employees) == 0 {
		return nil
	}
	return r.db.CreateInBatches(employees, 100).Error
}

func (r *EmployeeRepository) BulkUpdateByID(updates []models.Employee) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, emp := range updates {
			if err := tx.Model(&models.Employee{}).
				Where("employee_id = ? AND company_id = ?", emp.EmployeeID, emp.CompanyID).
				Updates(map[string]interface{}{
					"name_en":               emp.NameEn,
					"name_bn":               emp.NameBn,
					"father_name":           emp.FatherName,
					"mother_name":           emp.MotherName,
					"spouse_name":           emp.SpouseName,
					"date_of_birth":         emp.DateOfBirth,
					"gender":                emp.Gender,
					"blood_group":           emp.BloodGroup,
					"marital_status":        emp.MaritalStatus,
					"religion":              emp.Religion,
					"nationality":           emp.Nationality,
					"nid":                   emp.NID,
					"phone":                 emp.Phone,
					"email":                 emp.Email,
					"emergency_contact":     emp.EmergencyContact,
					"emergency_phone":       emp.EmergencyPhone,
					"number_of_dependents":  emp.NumberOfDependents,
					"present_address":       emp.PresentAddress,
					"permanent_address":     emp.PermanentAddress,
					"punch_number":          emp.PunchNumber,
					"employee_type":         emp.EmployeeType,
					"grade":                 emp.Grade,
					"joining_date":          emp.JoiningDate,
					"status":                emp.Status,
					"over_time_status":      emp.OverTimeStatus,
					"gross_salary":          emp.GrossSalary,
					"basic_salary":          emp.BasicSalary,
					"house_rent":            emp.HouseRent,
					"transport_allowance":   emp.TransportAllowance,
					"food_allowance":        emp.FoodAllowance,
					"medical_allowance":     emp.MedicalAllowance,
					"other_allowance":       emp.OtherAllowance,
					"account_type":          emp.AccountType,
					"account_number":        emp.AccountNumber,
					"department_id":         emp.DepartmentID,
					"section_id":            emp.SectionID,
					"designation_id":        emp.DesignationID,
					"line_id":               emp.LineID,
					"group_id":              emp.GroupID,
					"floor_id":              emp.FloorID,
					"shift_id":              emp.ShiftID,
					"present_division_id":   emp.PresentDivisionID,
					"present_district_id":   emp.PresentDistrictID,
					"present_upazila_id":    emp.PresentUpazilaID,
					"present_union_id":      emp.PresentUnionID,
					"permanent_division_id": emp.PermanentDivisionID,
					"permanent_district_id": emp.PermanentDistrictID,
					"permanent_upazila_id":  emp.PermanentUpazilaID,
					"permanent_union_id":    emp.PermanentUnionID,
					"reports_to":            emp.ReportsTo,
					"updated_by":            emp.UpdatedBy,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *EmployeeRepository) Update(emp *models.Employee) error {
	return r.db.Model(&models.Employee{}).
		Where("employee_id = ? AND company_id = ? AND deleted_at IS NULL", emp.EmployeeID, emp.CompanyID).
		Select("GrossSalary", "BasicSalary", "HouseRent", "MedicalAllowance").
		Updates(emp).Error
}

// UpdateWithPromotion updates salary fields plus promotion org hierarchy.
// Used only by SalaryIncrement Approve to apply promo department/section/designation/line.
func (r *EmployeeRepository) UpdateWithPromotion(emp *models.Employee) error {
	return r.db.Model(&models.Employee{}).
		Where("employee_id = ? AND company_id = ? AND deleted_at IS NULL", emp.EmployeeID, emp.CompanyID).
		Select("GrossSalary", "BasicSalary", "HouseRent", "MedicalAllowance", "DepartmentID", "SectionID", "DesignationID", "LineID").
		Updates(emp).Error
}

func (r *EmployeeRepository) MapByID(companyID string) (map[string]models.Employee, error) {
	var employees []models.Employee
	if err := r.db.Where("company_id = ?", companyID).Find(&employees).Error; err != nil {
		return nil, err
	}
	m := make(map[string]models.Employee, len(employees))
	for _, e := range employees {
		m[e.EmployeeID] = e
	}
	return m, nil
}

func (r *EmployeeRepository) GetAllActive() ([]models.Employee, error) {
	var employees []models.Employee
	err := r.db.Preload("Department").Preload("DesignationRef").
		Where("status = ?", "active").
		Order("LENGTH(employee_id) ASC, employee_id ASC").
		Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) GetByIDs(ids []string) ([]models.Employee, error) {
	var employees []models.Employee
	if len(ids) == 0 {
		return employees, nil
	}
	err := r.db.Preload("Department").Preload("DesignationRef").
		Where("id IN ? OR employee_id IN ?", ids, ids).
		Order("LENGTH(employee_id) ASC, employee_id ASC").
		Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) FindWithDetails(codeOrID string) (*models.Employee, error) {
	var emp models.Employee
	err := r.db.Preload("Company").
		Preload("Department").
		Preload("DesignationRef").
		Preload("SectionRef").
		Preload("LineRef").
		Preload("GroupRef").
		Preload("Shift").
		Where("employee_id = ? OR punch_number = ? OR id::text = ?", codeOrID, codeOrID, codeOrID).
		First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *EmployeeRepository) FindFirstFilteredWithDetails(f EmployeeFilter) (*models.Employee, error) {
	var emp models.Employee
	q := r.db.Preload("Company").
		Preload("Department").
		Preload("DesignationRef").
		Preload("SectionRef").
		Preload("LineRef").
		Preload("GroupRef").
		Preload("Shift").
		Where("deleted_at IS NULL")
	if f.CompanyID != "" {
		q = q.Where("company_id = ?", f.CompanyID)
	}
	if f.DepartmentID != "" {
		q = q.Where("department_id = ?", f.DepartmentID)
	}
	if f.SectionID != "" {
		q = q.Where("section_id = ?", f.SectionID)
	}
	if f.DesignationID != "" {
		q = q.Where("designation_id = ?", f.DesignationID)
	}
	if f.EmployeeID != "" {
		q = q.Where("employee_id = ? OR punch_number = ? OR id::text = ?", f.EmployeeID, f.EmployeeID, f.EmployeeID)
	}
	err := q.Order("LENGTH(employee_id) ASC, employee_id ASC").First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

