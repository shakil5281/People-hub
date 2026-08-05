package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type ShiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

func (r *ShiftRepository) Create(shift *models.Shift) error {
	return r.db.Create(shift).Error
}

func (r *ShiftRepository) FindByID(id string) (*models.Shift, error) {
	var shift models.Shift
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&shift).Error
	return &shift, err
}

func (r *ShiftRepository) List(page, limit int) ([]models.Shift, int64, error) {
	base := r.db.Model(&models.Shift{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var shifts []models.Shift
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&shifts).Error
	return shifts, total, err
}

func (r *ShiftRepository) Update(shift *models.Shift) error {
	return r.db.Save(shift).Error
}

func (r *ShiftRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Shift{}).Error
}

// ListActiveByCompany returns active shifts for a company. Used to determine the
// official weekend days (union of shift WeekendDays) when validating a weekend change.
func (r *ShiftRepository) ListActiveByCompany(companyID string) ([]models.Shift, error) {
	var shifts []models.Shift
	query := r.db.Where("status = 'active' AND deleted_at IS NULL")
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Order("created_at DESC").Find(&shifts).Error
	return shifts, err
}
