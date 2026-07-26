package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type TiffinBillRepository struct {
	db *gorm.DB
}

func NewTiffinBillRepository(db *gorm.DB) *TiffinBillRepository {
	return &TiffinBillRepository{db: db}
}

func (r *TiffinBillRepository) List(companyID string) ([]models.TiffinBill, error) {
	var items []models.TiffinBill
	q := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").Where("deleted_at IS NULL")
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *TiffinBillRepository) ListByMonth(companyID string, month, year int) ([]models.TiffinBill, error) {
	var items []models.TiffinBill
	q := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").Where("month = ? AND year = ? AND deleted_at IS NULL", month, year)
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *TiffinBillRepository) Create(item *models.TiffinBill) error {
	return r.db.Create(item).Error
}

func (r *TiffinBillRepository) FindByID(id string) (*models.TiffinBill, error) {
	var item models.TiffinBill
	err := r.db.Preload("Employee.Department").
		Preload("Employee.DesignationRef").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error
	return &item, err
}

func (r *TiffinBillRepository) Update(item *models.TiffinBill) error {
	return r.db.Save(item).Error
}

func (r *TiffinBillRepository) Delete(item *models.TiffinBill) error {
	return r.db.Delete(item).Error
}
