package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type NightBillRepository struct {
	db *gorm.DB
}

func NewNightBillRepository(db *gorm.DB) *NightBillRepository {
	return &NightBillRepository{db: db}
}

type NightBillFilter struct {
	Date          string
	StartDate     string
	EndDate       string
	CompanyID     string
	DepartmentID  string
	SectionID     string
	DesignationID string
	LineID        string
	GroupID       string
	BillType      string
	EmployeeID    string
}

func (r *NightBillRepository) Create(nb *models.NightBill) error {
	return r.db.Create(nb).Error
}

func (r *NightBillRepository) Upsert(nb *models.NightBill) error {
	var existing models.NightBill
	err := r.db.Where("employee_id = ? AND attendance_date = ? AND bill_type = ? AND deleted_at IS NULL", nb.EmployeeID, nb.AttendanceDate, nb.BillType).First(&existing).Error
	if err == nil && existing.ID != "" {
		nb.ID = existing.ID
		return r.db.Save(nb).Error
	}
	return r.db.Create(nb).Error
}

// Exists checks whether a night bill already exists for the given employee, date and bill type.
func (r *NightBillRepository) Exists(employeeID, attendanceDate, billType string) (bool, error) {
	var count int64
	err := r.db.Model(&models.NightBill{}).
		Where("employee_id = ? AND attendance_date = ? AND bill_type = ? AND deleted_at IS NULL", employeeID, attendanceDate, billType).
		Count(&count).Error
	return count > 0, err
}

func (r *NightBillRepository) FindByID(id string) (*models.NightBill, error) {
	var nb models.NightBill
	err := r.db.Preload("Employee.DesignationRef").
		Preload("Employee.DepartmentRef").
		Preload("Employee.SectionRef").
		Preload("Company").
		Preload("Shift").
		Preload("Attendance").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&nb).Error
	return &nb, err
}

func (r *NightBillRepository) Update(nb *models.NightBill) error {
	return r.db.Save(nb).Error
}

func (r *NightBillRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.NightBill{}).Error
}

func (r *NightBillRepository) DeleteBulk(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&models.NightBill{}).Error
}

func (r *NightBillRepository) applyFilter(query *gorm.DB, f NightBillFilter) *gorm.DB {
	query = query.Where("night_bills.deleted_at IS NULL")

	if f.StartDate != "" && f.EndDate != "" {
		query = query.Where("night_bills.attendance_date BETWEEN ? AND ?", f.StartDate, f.EndDate)
	} else if f.Date != "" {
		query = query.Where("night_bills.attendance_date = ?", f.Date)
	} else if f.StartDate != "" {
		query = query.Where("night_bills.attendance_date >= ?", f.StartDate)
	} else if f.EndDate != "" {
		query = query.Where("night_bills.attendance_date <= ?", f.EndDate)
	}

	if f.CompanyID != "" {
		query = query.Where("night_bills.company_id = ?", f.CompanyID)
	}
	if f.BillType != "" {
		query = query.Where("night_bills.bill_type = ?", f.BillType)
	}
	if f.EmployeeID != "" {
		query = query.Where("night_bills.employee_id = ?", f.EmployeeID)
	}

	if f.DepartmentID != "" || f.SectionID != "" || f.DesignationID != "" || f.LineID != "" || f.GroupID != "" {
		subQuery := r.db.Table("employees").Select("employee_id").Where("deleted_at IS NULL")
		if f.DepartmentID != "" {
			subQuery = subQuery.Where("department_id = ?", f.DepartmentID)
		}
		if f.SectionID != "" {
			subQuery = subQuery.Where("section_id = ?", f.SectionID)
		}
		if f.DesignationID != "" {
			subQuery = subQuery.Where("designation_id = ?", f.DesignationID)
		}
		if f.LineID != "" {
			subQuery = subQuery.Where("line_id = ?", f.LineID)
		}
		if f.GroupID != "" {
			subQuery = subQuery.Where("group_id = ?", f.GroupID)
		}
		query = query.Where("night_bills.employee_id IN (?)", subQuery)
	}

	return query
}

func (r *NightBillRepository) ListFiltered(f NightBillFilter, page, limit int) ([]models.NightBill, int64, error) {
	base := r.applyFilter(r.db.Model(&models.NightBill{}), f)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []models.NightBill
	err := r.applyFilter(r.db.Model(&models.NightBill{}), f).
		Preload("Employee.DesignationRef").
		Preload("Employee.DepartmentRef").
		Preload("Employee.SectionRef").
		Preload("Company").
		Preload("Shift").
		Preload("Attendance").
		Order("night_bills.attendance_date DESC, LENGTH(night_bills.employee_id) ASC, night_bills.employee_id ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

func (r *NightBillRepository) ListAllFiltered(f NightBillFilter) ([]models.NightBill, error) {
	var list []models.NightBill
	err := r.applyFilter(r.db.Model(&models.NightBill{}), f).
		Preload("Employee.DesignationRef").
		Preload("Employee.DepartmentRef").
		Preload("Employee.SectionRef").
		Preload("Company").
		Preload("Shift").
		Preload("Attendance").
		Order("night_bills.attendance_date DESC, LENGTH(night_bills.employee_id) ASC, night_bills.employee_id ASC").
		Find(&list).Error
	return list, err
}
