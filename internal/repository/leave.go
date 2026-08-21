package repository

import (
	"errors"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type LeaveRepository struct {
	db *gorm.DB
}

func NewLeaveRepository(db *gorm.DB) *LeaveRepository {
	return &LeaveRepository{db: db}
}

// WithTx returns a new LeaveRepository using the provided transaction
func (r *LeaveRepository) WithTx(tx *gorm.DB) *LeaveRepository {
	return &LeaveRepository{db: tx}
}

// --- Leave Types ---

func (r *LeaveRepository) CreateLeaveType(lt *models.LeaveType) error {
	return r.db.Create(lt).Error
}

func (r *LeaveRepository) FindLeaveTypeByID(id string) (*models.LeaveType, error) {
	var lt models.LeaveType
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&lt).Error
	return &lt, err
}

func (r *LeaveRepository) ListLeaveTypes(companyID string, page, limit int) ([]models.LeaveType, int64, error) {
	base := r.db.Model(&models.LeaveType{}).Where("deleted_at IS NULL")
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.LeaveType
	err := base.Order("name ASC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *LeaveRepository) UpdateLeaveType(lt *models.LeaveType) error {
	return r.db.Save(lt).Error
}

func (r *LeaveRepository) DeleteLeaveType(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.LeaveType{}).Error
}

// --- Leave Applications ---

func (r *LeaveRepository) CreateLeave(l *models.Leave) error {
	return r.db.Create(l).Error
}

func (r *LeaveRepository) FindLeaveByID(id string) (*models.Leave, error) {
	var l models.Leave
	err := r.db.Preload("Employee").Preload("LeaveType").Preload("Company").
		Where("id = ? AND deleted_at IS NULL", id).First(&l).Error
	return &l, err
}

func (r *LeaveRepository) ListLeaves(companyID, departmentID, employeeID, status, fromDate, toDate string, page, limit int) ([]models.Leave, int64, error) {
	base := r.db.Model(&models.Leave{}).Where("leaves.deleted_at IS NULL")
	if companyID != "" {
		base = base.Where("leaves.company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("leaves.employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if employeeID != "" {
		base = base.Where("leaves.employee_id = ?", employeeID)
	}
	if status != "" {
		base = base.Where("leaves.status = ?", status)
	}
	if fromDate != "" {
		base = base.Where("leaves.from_date >= ?", fromDate)
	}
	if toDate != "" {
		base = base.Where("leaves.to_date <= ?", toDate)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Leave
	err := base.Preload("Employee").Preload("LeaveType").Order("leaves.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *LeaveRepository) FindApprovedByEmployeeAndDate(employeeID, date string) (*models.Leave, error) {
	var leave models.Leave
	err := r.db.Where("employee_id = ? AND status = 'approved' AND from_date <= ? AND to_date >= ? AND deleted_at IS NULL",
		employeeID, date, date).First(&leave).Error
	return &leave, err
}

func (r *LeaveRepository) ListApprovedByDate(date string) ([]models.Leave, error) {
	var leaves []models.Leave
	err := r.db.Where("status = 'approved' AND from_date <= ? AND to_date >= ? AND deleted_at IS NULL",
		date, date).Find(&leaves).Error
	return leaves, err
}

func (r *LeaveRepository) UpdateLeave(l *models.Leave) error {
	return r.db.Save(l).Error
}

func (r *LeaveRepository) DeleteLeave(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Leave{}).Error
}

func (r *LeaveRepository) HardDeleteLeave(id string) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.Leave{}).Error
}

// --- Leave Allocations / Balance ---

func (r *LeaveRepository) UpsertAllocation(a *models.LeaveAllocation) error {
	var existing models.LeaveAllocation
	err := r.db.Where("employee_id = ? AND leave_type_id = ? AND year = ? AND deleted_at IS NULL", a.EmployeeID, a.LeaveTypeID, a.Year).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return r.db.Create(a).Error
	}
	a.ID = existing.ID
	a.CreatedAt = existing.CreatedAt
	return r.db.Save(a).Error
}

func (r *LeaveRepository) FindAllocation(employeeID, leaveTypeID string, year int) (*models.LeaveAllocation, error) {
	var a models.LeaveAllocation
	err := r.db.Where("employee_id = ? AND leave_type_id = ? AND year = ? AND deleted_at IS NULL", employeeID, leaveTypeID, year).First(&a).Error
	return &a, err
}

func (r *LeaveRepository) ListAllocations(employeeID string, year, page, limit int) ([]models.LeaveAllocation, int64, error) {
	base := r.db.Model(&models.LeaveAllocation{}).Where("deleted_at IS NULL")
	if employeeID != "" {
		base = base.Where("employee_id = ?", employeeID)
	}
	if year > 0 {
		base = base.Where("year = ?", year)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.LeaveAllocation
	err := base.Preload("LeaveType").Order("created_at ASC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *LeaveRepository) ListAllocationsByEmployees(employeeIDs []string, year int) ([]models.LeaveAllocation, error) {
	var list []models.LeaveAllocation
	q := r.db.Preload("Employee").Preload("LeaveType").Where("deleted_at IS NULL")
	if len(employeeIDs) > 0 {
		q = q.Where("employee_id IN ?", employeeIDs)
	}
	if year > 0 {
		q = q.Where("year = ?", year)
	}
	err := q.Order("created_at ASC").Find(&list).Error
	return list, err
}

// --- Monthly Report ---

func (r *LeaveRepository) MonthlyReport(month, year int, companyID, departmentID string) ([]map[string]interface{}, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	q := r.db.Table("leaves").
		Select("departments.id as department_id, departments.name as department_name, COUNT(*) as total, SUM(CASE WHEN leaves.status = 'approved' THEN 1 ELSE 0 END) as approved, SUM(CASE WHEN leaves.status = 'rejected' THEN 1 ELSE 0 END) as rejected, SUM(CASE WHEN leaves.status = 'pending' THEN 1 ELSE 0 END) as pending").
		Joins("JOIN employees ON employees.employee_id = leaves.employee_id").
		Joins("JOIN departments ON departments.id = employees.department_id").
		Where("leaves.deleted_at IS NULL AND leaves.from_date >= ? AND leaves.from_date <= ?", startStr, endStr)
	if companyID != "" {
		q = q.Where("leaves.company_id = ?", companyID)
	}
	if departmentID != "" {
		q = q.Where("employees.department_id = ?", departmentID)
	}
	var results []map[string]interface{}
	err := q.Group("departments.id, departments.name").Order("departments.name ASC").Find(&results).Error
	return results, err
}
