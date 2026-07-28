package repository

import (
	"fmt"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type SalaryIncrementRepository struct {
	db *gorm.DB
}

func NewSalaryIncrementRepository(db *gorm.DB) *SalaryIncrementRepository {
	return &SalaryIncrementRepository{db: db}
}

type IncrementFilter struct {
	CompanyID     string
	DepartmentID  string
	SectionID     string
	DesignationID string
	LineID        string
	GroupID       string
	Month         int
	Year          int
}

func (r *SalaryIncrementRepository) List(f IncrementFilter) ([]models.SalaryIncrement, error) {
	query := r.db.Preload("Employee.Department").
		Preload("Employee.DesignationRef").
		Where("salary_increments.company_id = ? AND salary_increments.deleted_at IS NULL", f.CompanyID)

	if f.DepartmentID != "" {
		query = query.Where("salary_increments.employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", f.DepartmentID)
	}
	if f.SectionID != "" {
		query = query.Where("salary_increments.employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", f.SectionID)
	}
	if f.DesignationID != "" {
		query = query.Where("salary_increments.employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", f.DesignationID)
	}
	if f.LineID != "" {
		query = query.Where("salary_increments.employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", f.LineID)
	}
	if f.GroupID != "" {
		query = query.Where("salary_increments.employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", f.GroupID)
	}
	if f.Month > 0 && f.Year > 0 {
		monthStr := fmt.Sprintf("%02d", f.Month)
		yearStr := fmt.Sprintf("%d", f.Year)
		query = query.Where("(salary_increments.effective_date LIKE ? OR salary_increments.created_at::text LIKE ?)", yearStr+"-"+monthStr+"%", yearStr+"-"+monthStr+"%")
	}

	var incs []models.SalaryIncrement
	err := query.Order("salary_increments.created_at DESC").Find(&incs).Error
	return incs, err
}

func (r *SalaryIncrementRepository) Create(inc *models.SalaryIncrement) error {
	return r.db.Create(inc).Error
}

func (r *SalaryIncrementRepository) CreateBatch(incs []models.SalaryIncrement) error {
	if len(incs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(incs, 100).Error
}

func (r *SalaryIncrementRepository) FindByID(id string) (*models.SalaryIncrement, error) {
	var inc models.SalaryIncrement
	err := r.db.Preload("Employee.Department").
		Preload("Employee.DesignationRef").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&inc).Error
	return &inc, err
}

func (r *SalaryIncrementRepository) Update(inc *models.SalaryIncrement) error {
	return r.db.Save(inc).Error
}

func (r *SalaryIncrementRepository) FindEligibleEmployees(companyID, departmentID, sectionID, designationID, lineID, groupID string) ([]models.Employee, error) {
	query := r.db.Model(&models.Employee{}).
		Where("company_id = ? AND status = 'active' AND employee_type = 'regular' AND gross_salary > 0", companyID)
	if departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}
	if sectionID != "" {
		query = query.Where("section_id = ?", sectionID)
	}
	if designationID != "" {
		query = query.Where("designation_id = ?", designationID)
	}
	if lineID != "" {
		query = query.Where("line_id = ?", lineID)
	}
	if groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	var emps []models.Employee
	err := query.Order("employee_id ASC").Find(&emps).Error
	return emps, err
}
