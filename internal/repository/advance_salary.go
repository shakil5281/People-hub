package repository

import (
	"fmt"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type AdvanceSalaryRepository struct {
	db *gorm.DB
}

func NewAdvanceSalaryRepository(db *gorm.DB) *AdvanceSalaryRepository {
	return &AdvanceSalaryRepository{db: db}
}

type AdvanceFilter struct {
	CompanyID     string
	DepartmentID  string
	SectionID     string
	DesignationID string
	LineID        string
	GroupID       string
	Month         int
	Year          int
	Status        string
}

func (r *AdvanceSalaryRepository) List(f AdvanceFilter) ([]models.AdvanceSalary, error) {
	query := r.db.Preload("Employee.Department").
		Preload("Employee.DesignationRef").
		Where("advance_salaries.company_id = ? AND advance_salaries.deleted_at IS NULL", f.CompanyID)

	if f.Status != "" {
		query = query.Where("advance_salaries.status = ?", f.Status)
	}
	if f.DepartmentID != "" {
		query = query.Where("advance_salaries.employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", f.DepartmentID)
	}
	if f.SectionID != "" {
		query = query.Where("advance_salaries.employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", f.SectionID)
	}
	if f.DesignationID != "" {
		query = query.Where("advance_salaries.employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", f.DesignationID)
	}
	if f.LineID != "" {
		query = query.Where("advance_salaries.employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", f.LineID)
	}
	if f.GroupID != "" {
		query = query.Where("advance_salaries.employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", f.GroupID)
	}
	if f.Month > 0 && f.Year > 0 {
		monthStr := fmt.Sprintf("%02d", f.Month)
		yearStr := fmt.Sprintf("%d", f.Year)
		query = query.Where("(advance_salaries.advance_date::text LIKE ? OR advance_salaries.created_at::text LIKE ?)", yearStr+"-"+monthStr+"%", yearStr+"-"+monthStr+"%")
	}

	var advances []models.AdvanceSalary
	err := query.Order("advance_salaries.created_at DESC").Find(&advances).Error
	return advances, err
}

func (r *AdvanceSalaryRepository) Create(advance *models.AdvanceSalary) error {
	return r.db.Create(advance).Error
}

func (r *AdvanceSalaryRepository) CreateBatch(advances []models.AdvanceSalary) error {
	if len(advances) == 0 {
		return nil
	}
	return r.db.Create(&advances).Error
}

func (r *AdvanceSalaryRepository) UpdateStatus(id, status, updatedBy string) error {
	return r.db.Model(&models.AdvanceSalary{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_by": updatedBy,
	}).Error
}

func (r *AdvanceSalaryRepository) Delete(id string) error {
	// Only allow deleting pending advances
	return r.db.Where("id = ? AND status = 'pending'", id).Delete(&models.AdvanceSalary{}).Error
}

// MonthlyDeductions returns a map of employee_id -> total advance deduction amount for a specific month and year.
// It considers advances that are 'approved' or 'deducted' for the target month/year.
func (r *AdvanceSalaryRepository) MonthlyDeductions(companyID string, month, year int) (map[string]float64, error) {
	var results []struct {
		EmployeeID string
		Total      float64
	}

	err := r.db.Model(&models.AdvanceSalary{}).
		Select("employee_id, SUM(amount) as total").
		Where("company_id = ? AND deduction_month = ? AND deduction_year = ? AND status IN ('approved', 'deducted')", companyID, month, year).
		Group("employee_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	deductions := make(map[string]float64)
	for _, res := range results {
		deductions[res.EmployeeID] = res.Total
	}

	return deductions, nil
}

// MarkAsDeducted marks all 'approved' advances for a given month and year as 'deducted'.
func (r *AdvanceSalaryRepository) MarkAsDeducted(companyID string, month, year int) error {
	return r.db.Model(&models.AdvanceSalary{}).
		Where("company_id = ? AND deduction_month = ? AND deduction_year = ? AND status = 'approved'", companyID, month, year).
		Update("status", "deducted").Error
}
