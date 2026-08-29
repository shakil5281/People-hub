package repository

import (
	"fmt"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type SeparationRepository struct {
	db *gorm.DB
}

func NewSeparationRepository(db *gorm.DB) *SeparationRepository {
	return &SeparationRepository{db: db}
}

func (r *SeparationRepository) WithTx(tx *gorm.DB) *SeparationRepository {
	return &SeparationRepository{db: tx}
}

func (r *SeparationRepository) Create(sep *models.Separation) error {
	return r.db.Create(sep).Error
}

func (r *SeparationRepository) FindByID(id string) (*models.Separation, error) {
	var sep models.Separation
	err := r.db.Preload("Department").Where("id = ? AND deleted_at IS NULL", id).First(&sep).Error
	return &sep, err
}

func (r *SeparationRepository) List(page, limit int) ([]models.Separation, int64, error) {
	base := r.db.Model(&models.Separation{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var seps []models.Separation
	err := base.Preload("Department").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&seps).Error
	return seps, total, err
}

func (r *SeparationRepository) ListFiltered(employee, employeeID, departmentID, sepType, status, companyID, sectionID, designationID, lineID, groupID, dateFrom, dateTo string, page, limit int) ([]models.Separation, int64, error) {
	base := r.db.Model(&models.Separation{}).Where("separations.deleted_at IS NULL").Order("created_at DESC")

	hasOrgFilter := sectionID != "" || designationID != "" || lineID != "" || groupID != ""
	if hasOrgFilter {
		base = base.Joins("LEFT JOIN employees ON employees.employee_id = separations.employee_id")
	}

	if employee != "" {
		base = base.Where("separations.employee ILIKE ?", "%"+employee+"%")
	}
	if employeeID != "" {
		base = base.Where("separations.employee_id = ?", employeeID)
	}
	if departmentID != "" {
		base = base.Where("separations.department_id = ?", departmentID)
	}
	if sepType != "" {
		base = base.Where("separations.type = ?", sepType)
	}
	if status != "" {
		base = base.Where("separations.status = ?", status)
	}
	if companyID != "" {
		base = base.Where("separations.company_id = ?", companyID)
	}
	if sectionID != "" {
		base = base.Where("employees.section_id = ?", sectionID)
	}
	if designationID != "" {
		base = base.Where("employees.designation_id = ?", designationID)
	}
	if lineID != "" {
		base = base.Where("employees.line_id = ?", lineID)
	}
	if groupID != "" {
		base = base.Where("employees.group_id = ?", groupID)
	}
	if dateFrom != "" {
		base = base.Where("separations.date >= ?", dateFrom)
	}
	if dateTo != "" {
		base = base.Where("separations.date <= ?", dateTo)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var seps []models.Separation
	err := base.Preload("Department").Offset((page - 1) * limit).Limit(limit).Find(&seps).Error
	return seps, total, err
}

func (r *SeparationRepository) ListAllFiltered(employee, employeeID, departmentID, sepType, status, companyID, sectionID, designationID, lineID, groupID, dateFrom, dateTo string) ([]models.Separation, error) {
	base := r.db.Model(&models.Separation{}).Where("separations.deleted_at IS NULL").Order("separations.date DESC, separations.created_at DESC")

	base = base.Joins("LEFT JOIN employees ON employees.employee_id = separations.employee_id")

	if employee != "" {
		base = base.Where("separations.employee ILIKE ?", "%"+employee+"%")
	}
	if employeeID != "" {
		base = base.Where("separations.employee_id = ?", employeeID)
	}
	if departmentID != "" {
		base = base.Where("separations.department_id = ?", departmentID)
	}
	if sepType != "" {
		base = base.Where("separations.type = ?", sepType)
	}
	if status != "" {
		base = base.Where("separations.status = ?", status)
	}
	if companyID != "" {
		base = base.Where("separations.company_id = ?", companyID)
	}
	if sectionID != "" {
		base = base.Where("employees.section_id = ?", sectionID)
	}
	if designationID != "" {
		base = base.Where("employees.designation_id = ?", designationID)
	}
	if lineID != "" {
		base = base.Where("employees.line_id = ?", lineID)
	}
	if groupID != "" {
		base = base.Where("employees.group_id = ?", groupID)
	}
	if dateFrom != "" {
		base = base.Where("separations.date >= ?", dateFrom)
	}
	if dateTo != "" {
		base = base.Where("separations.date <= ?", dateTo)
	}

	var seps []models.Separation
	err := base.Preload("Department").Find(&seps).Error
	return seps, err
}

func (r *SeparationRepository) Update(sep *models.Separation) error {
	return r.db.Save(sep).Error
}

func (r *SeparationRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Separation{}).Error
}

// FindProcessedByEmployeeID returns the latest processed separation for an employee.
func (r *SeparationRepository) FindProcessedByEmployeeID(employeeID string) (*models.Separation, error) {
	var sep models.Separation
	err := r.db.Where("employee_id = ? AND status = ? AND deleted_at IS NULL", employeeID, "Processed").
		Order("date DESC").
		First(&sep).Error
	return &sep, err
}

// ExistsPendingOrApproved checks if employee has a non-final separation.
func (r *SeparationRepository) ExistsPendingOrApproved(employeeID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Separation{}).
		Where("employee_id = ? AND status IN ? AND deleted_at IS NULL", employeeID, []string{"Pending", "Approved"}).
		Count(&count).Error
	return count > 0, err
}

// FindPendingDue returns all pending/approved separations with date <= processDate.
func (r *SeparationRepository) FindPendingDue(processDate string) ([]models.Separation, error) {
	var list []models.Separation
	err := r.db.
		Where("status IN ? AND date <= ? AND deleted_at IS NULL", []string{"Pending", "Approved"}, processDate).
		Find(&list).Error
	return list, err
}

func (r *SeparationRepository) FindEmployeeByCode(empCode string) (*models.Employee, error) {
	var emp models.Employee
	err := r.db.Preload("Company").Preload("Department").Preload("DesignationRef").Preload("SectionRef").
		Where("employee_id = ? AND deleted_at IS NULL", empCode).First(&emp).Error
	return &emp, err
}

// GetSeparationDatesByMonth returns a map of employee_id -> separation_date (YYYY-MM-DD) for processed separations in that month.
func (r *SeparationRepository) GetSeparationDatesByMonth(companyID string, month, year int) (map[string]string, error) {
	startDate := fmt.Sprintf("%04d-%02d-01", year, month)
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := t.AddDate(0, 1, -1).Day()
	endDate := fmt.Sprintf("%04d-%02d-%02d", year, month, lastDay)

	var rows []struct {
		EmployeeID string `gorm:"column:employee_id"`
		Date       string `gorm:"column:date"`
	}

	q := r.db.Table("separations").
		Select("employee_id, MAX(date) as date").
		Where("deleted_at IS NULL AND LOWER(status) != 'cancelled' AND date >= ? AND date <= ?", startDate, endDate)
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}

	err := q.Group("employee_id").Find(&rows).Error
	if err != nil {
		return nil, err
	}

	res := make(map[string]string, len(rows))
	for _, row := range rows {
		res[row.EmployeeID] = row.Date
	}
	return res, nil
}
