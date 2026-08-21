package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type IdCardRepository struct {
	db *gorm.DB
}

func NewIdCardRepository(db *gorm.DB) *IdCardRepository {
	return &IdCardRepository{db: db}
}

func (r *IdCardRepository) Create(card *models.IdCard) error {
	return r.db.Create(card).Error
}

func (r *IdCardRepository) FindByID(id string) (*models.IdCard, error) {
	var card models.IdCard
	err := r.db.Preload("Department").Preload("Designation").Where("id = ? AND deleted_at IS NULL", id).First(&card).Error
	return &card, err
}

func (r *IdCardRepository) List(page, limit int) ([]models.IdCard, int64, error) {
	base := r.db.Model(&models.IdCard{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var cards []models.IdCard
	err := base.Preload("Department").Preload("Designation").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&cards).Error
	return cards, total, err
}

func (r *IdCardRepository) ListFiltered(employee, employeeID, departmentID, designationID, status, cardNo string, page, limit int) ([]models.IdCard, int64, error) {
	base := r.db.Model(&models.IdCard{}).Where("deleted_at IS NULL").Order("created_at DESC")
	if employee != "" {
		base = base.Where("employee ILIKE ?", "%"+employee+"%")
	}
	if employeeID != "" {
		base = base.Where("employee_id ILIKE ?", "%"+employeeID+"%")
	}
	if departmentID != "" {
		base = base.Where("department_id = ?", departmentID)
	}
	if designationID != "" {
		base = base.Where("designation_id = ?", designationID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if cardNo != "" {
		base = base.Where("card_no ILIKE ?", "%"+cardNo+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var cards []models.IdCard
	err := base.Preload("Department").Preload("Designation").Offset((page - 1) * limit).Limit(limit).Find(&cards).Error
	return cards, total, err
}

func (r *IdCardRepository) Update(card *models.IdCard) error {
	return r.db.Save(card).Error
}

func (r *IdCardRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.IdCard{}).Error
}
