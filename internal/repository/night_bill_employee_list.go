package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type NightBillEmployeeListRepository struct {
	db *gorm.DB
}

func NewNightBillEmployeeListRepository(db *gorm.DB) *NightBillEmployeeListRepository {
	return &NightBillEmployeeListRepository{db: db}
}

func (r *NightBillEmployeeListRepository) preload(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Employee.DesignationRef").
		Preload("Employee.Department").
		Preload("Employee.DepartmentRef").
		Preload("Employee.SectionRef").
		Preload("Employee.LineRef").
		Preload("Employee.GroupRef").
		Preload("Company")
}

// Create inserts a new employee night bill list record.
func (r *NightBillEmployeeListRepository) Create(rec *models.NightBillEmployeeList) error {
	return r.db.Create(rec).Error
}

// BulkCreate inserts multiple records in one transaction.
func (r *NightBillEmployeeListRepository) BulkCreate(recs []*models.NightBillEmployeeList) error {
	if len(recs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, rec := range recs {
			if err := tx.Create(rec).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID retrieves a single record by primary key.
func (r *NightBillEmployeeListRepository) FindByID(id string) (*models.NightBillEmployeeList, error) {
	var rec models.NightBillEmployeeList
	err := r.preload(r.db).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&rec).Error
	return &rec, err
}

// FindByEmployeeID finds active record for a given employee.
func (r *NightBillEmployeeListRepository) FindByEmployeeID(employeeID string) (*models.NightBillEmployeeList, error) {
	var rec models.NightBillEmployeeList
	err := r.preload(r.db).
		Where("employee_id = ? AND deleted_at IS NULL", employeeID).
		First(&rec).Error
	return &rec, err
}

// ExistsForEmployee checks if an employee is already in the list.
func (r *NightBillEmployeeListRepository) ExistsForEmployee(employeeID string) bool {
	var count int64
	r.db.Model(&models.NightBillEmployeeList{}).
		Where("employee_id = ? AND deleted_at IS NULL", employeeID).
		Count(&count)
	return count > 0
}

// List returns all active records with optional company filter.
func (r *NightBillEmployeeListRepository) List(companyID string, page, limit int) ([]models.NightBillEmployeeList, int64, error) {
	q := r.db.Model(&models.NightBillEmployeeList{}).Where("deleted_at IS NULL")
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []models.NightBillEmployeeList
	err := r.preload(q).
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

// ListAll returns all active records (for export / bulk ops).
func (r *NightBillEmployeeListRepository) ListAll(companyID string) ([]models.NightBillEmployeeList, error) {
	q := r.db.Model(&models.NightBillEmployeeList{}).Where("deleted_at IS NULL")
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	var list []models.NightBillEmployeeList
	err := r.preload(q).Order("created_at DESC").Find(&list).Error
	return list, err
}

// ListEligible returns active (non-deleted, enabled) entries for night bill processing,
// optionally filtered by company.
func (r *NightBillEmployeeListRepository) ListEligible(companyID string) ([]models.NightBillEmployeeList, error) {
	q := r.db.Model(&models.NightBillEmployeeList{}).Where("deleted_at IS NULL AND is_active = ?", true)
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	var list []models.NightBillEmployeeList
	err := q.Find(&list).Error
	return list, err
}

// Update saves changes to an existing record.
func (r *NightBillEmployeeListRepository) Update(rec *models.NightBillEmployeeList) error {
	return r.db.Save(rec).Error
}

// Delete soft-deletes a record by ID.
func (r *NightBillEmployeeListRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.NightBillEmployeeList{}).Error
}

// DeleteByEmployeeID soft-deletes by employee ID.
func (r *NightBillEmployeeListRepository) DeleteByEmployeeID(employeeID string) error {
	return r.db.Where("employee_id = ?", employeeID).Delete(&models.NightBillEmployeeList{}).Error
}

// BulkDelete soft-deletes multiple records by IDs.
func (r *NightBillEmployeeListRepository) BulkDelete(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&models.NightBillEmployeeList{}).Error
}
