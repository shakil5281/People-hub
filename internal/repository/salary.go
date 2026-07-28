package repository

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type SalaryRepository struct {
	db *gorm.DB
}

func NewSalaryRepository(db *gorm.DB) *SalaryRepository {
	return &SalaryRepository{db: db}
}

func (r *SalaryRepository) Create(salary *models.Salary) error {
	return r.db.Create(salary).Error
}

func (r *SalaryRepository) Upsert(salary *models.Salary) error {
	return r.db.Where("employee_id = ? AND month = ? AND year = ? AND company_id = ?",
		salary.EmployeeID, salary.Month, salary.Year, salary.CompanyID).
		Assign(salary).
		FirstOrCreate(salary).Error
}

func (r *SalaryRepository) FindByEmployeeMonth(employeeID string, month, year int) (*models.Salary, error) {
	var s models.Salary
	err := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").Where("employee_id = ? AND month = ? AND year = ? AND deleted_at IS NULL",
		employeeID, month, year).First(&s).Error
	return &s, err
}

func (r *SalaryRepository) ListByMonth(companyID string, month, year int, departmentID string, page, limit int) ([]models.Salary, int64, error) {
	base := r.db.Model(&models.Salary{}).Where("company_id = ? AND month = ? AND year = ? AND deleted_at IS NULL", companyID, month, year)
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT id FROM employees WHERE department_id = ?)", departmentID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var salaries []models.Salary
	err := base.Preload("Employee.Department").Preload("Employee.DesignationRef").Order("LENGTH(employee_id) ASC, employee_id ASC").Offset((page - 1) * limit).Limit(limit).Find(&salaries).Error
	return salaries, total, err
}

type SalaryFilter struct {
	CompanyID     string
	Month, Year   int
	DepartmentID  string
	SectionID     string
	DesignationID string
	LineID        string
	GroupID       string
	ShiftID       string
	EmployeeID    string
	AccountType   string
}

func (r *SalaryRepository) ListAllByMonth(companyID string, month, year int, departmentID string) ([]models.Salary, error) {
	return r.ListAllByMonthFiltered(SalaryFilter{
		CompanyID:    companyID,
		Month:        month,
		Year:         year,
		DepartmentID: departmentID,
	})
}

func (r *SalaryRepository) ListAllByMonthFiltered(f SalaryFilter) ([]models.Salary, error) {
	query := r.db.Preload("Employee.Department").
		Preload("Employee.SectionRef").
		Preload("Employee.DesignationRef").
		Preload("Employee.LineRef").
		Preload("Employee.GroupRef").
		Where("company_id = ? AND month = ? AND year = ? AND deleted_at IS NULL", f.CompanyID, f.Month, f.Year)
	if f.DepartmentID != "" {
		query = query.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", f.DepartmentID)
	}
	if f.SectionID != "" {
		query = query.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", f.SectionID)
	}
	if f.DesignationID != "" {
		query = query.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", f.DesignationID)
	}
	if f.LineID != "" {
		query = query.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", f.LineID)
	}
	if f.GroupID != "" {
		query = query.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", f.GroupID)
	}
	if f.ShiftID != "" {
		query = query.Where("employee_id IN (SELECT employee_id FROM employees WHERE shift_id = ?)", f.ShiftID)
	}
	if f.EmployeeID != "" {
		query = query.Where("employee_id LIKE ?", "%"+f.EmployeeID+"%")
	}
	if f.AccountType != "" {
		query = query.Where("employee_id IN (SELECT employee_id FROM employees WHERE account_type = ?)", f.AccountType)
	}
	var salaries []models.Salary
	err := query.Order("LENGTH(employee_id) ASC, employee_id ASC").Find(&salaries).Error
	return salaries, err
}

func (r *SalaryRepository) DeleteByMonth(companyID string, month, year int) error {
	return r.db.Unscoped().Where("company_id = ? AND month = ? AND year = ?", companyID, month, year).
		Delete(&models.Salary{}).Error
}

type DailySalaryRecord struct {
	ID             string  `json:"id"`
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Designation    string  `json:"designation"`
	DepartmentName string  `json:"department_name"`
	Date           string  `json:"date"`
	Status         string  `json:"status"`
	CheckIn        string  `json:"check_in"`
	CheckOut       string  `json:"check_out"`
	TotalHours     string  `json:"total_hours"`
	OverTime       string  `json:"over_time"`
	GrossSalary    float64 `json:"gross_salary"`
	DailyRate      float64 `json:"daily_rate"`
	OtHours        float64 `json:"ot_hours"`
	OtAmount       float64 `json:"ot_amount"`
	TotalPay       float64 `json:"total_pay"`
}

func (r *SalaryRepository) DailySheet(date, companyID, departmentID, sectionID, designationID, lineID, groupID string) ([]DailySalaryRecord, error) {
	daysInMonth := func() int {
		t, _ := time.Parse("2006-01-02", date)
		firstOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
		return lastOfMonth.Day()
	}()

	query := r.db.Table("attendances a").
		Select(`
			a.id,
			e.employee_id,
			COALESCE(e.name_en, '') as employee_name,
			COALESCE(d.name, '') as designation,
			COALESCE(dept.name, '') as department_name,
			a.date,
			a.status,
			COALESCE(to_char(a.check_in, 'HH24:MI'), '') as check_in,
			COALESCE(to_char(a.check_out, 'HH24:MI'), '') as check_out,
			COALESCE(a.total_hours, '') as total_hours,
			COALESCE(a.over_time, '') as over_time,
			e.gross_salary`).
		Joins("JOIN employees e ON e.employee_id = a.employee_id").
		Joins("LEFT JOIN designations d ON d.id = e.designation_id").
		Joins("LEFT JOIN departments dept ON dept.id = e.department_id").
		Where("a.date = ? AND a.deleted_at IS NULL", date)

	if companyID != "" {
		query = query.Where("a.company_id = ?", companyID)
	}

	if departmentID != "" {
		query = query.Where("e.department_id = ?", departmentID)
	}
	if sectionID != "" {
		query = query.Where("e.section_id = ?", sectionID)
	}
	if designationID != "" {
		query = query.Where("e.designation_id = ?", designationID)
	}
	if lineID != "" {
		query = query.Where("e.line_id = ?", lineID)
	}
	if groupID != "" {
		query = query.Where("e.group_id = ?", groupID)
	}

	query = query.Order("LENGTH(e.employee_id) ASC, e.employee_id ASC")

	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []DailySalaryRecord
	for rows.Next() {
		var r DailySalaryRecord
		var gross sql.NullFloat64
		err := rows.Scan(&r.ID, &r.EmployeeID, &r.EmployeeName, &r.Designation, &r.DepartmentName,
			&r.Date, &r.Status, &r.CheckIn, &r.CheckOut, &r.TotalHours, &r.OverTime, &gross)
		if err != nil {
			return nil, err
		}
		r.GrossSalary = gross.Float64
		if daysInMonth > 0 {
			r.DailyRate = gross.Float64 / float64(daysInMonth)
		}

		// Parse OT from HH:MM format
		otHours := parseOTHours(r.OverTime)
		r.OtHours = otHours
		r.OtAmount = calculateOTAmount(gross.Float64, otHours, daysInMonth)

		// Total pay = daily rate if present/on duty, else 0
		totalPay := float64(0)
		if r.Status == "present" || r.Status == "late" || r.Status == "half_day" {
			totalPay = r.DailyRate
		}
		if r.Status == "half_day" {
			totalPay = r.DailyRate * 0.5
		}
		totalPay += r.OtAmount
		r.TotalPay = totalPay

		records = append(records, r)
	}

	return records, nil
}

func parseOTHours(ot string) float64 {
	if ot == "" {
		return 0
	}
	parts := strings.Split(ot, ":")
	if len(parts) != 2 {
		return 0
	}
	h, _ := strconv.ParseFloat(parts[0], 64)
	m, _ := strconv.ParseFloat(parts[1], 64)
	return h + m/60.0
}

func calculateOTAmount(grossSalary, otHours float64, daysInMonth int) float64 {
	if daysInMonth == 0 || grossSalary == 0 || otHours == 0 {
		return 0
	}
	transport := 450.0
	food := 1250.0
	medical := 750.0
	core := grossSalary - transport - food - medical
	basic := core / 1.5
	otRate := basic / float64(daysInMonth) / 8
	return otHours * otRate
}
