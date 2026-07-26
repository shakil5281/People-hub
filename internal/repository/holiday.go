package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type HolidayRepository struct {
	db *gorm.DB
}

func NewHolidayRepository(db *gorm.DB) *HolidayRepository {
	return &HolidayRepository{db: db}
}

func (r *HolidayRepository) WithTx(tx *gorm.DB) *HolidayRepository {
	return &HolidayRepository{db: tx}
}

func (r *HolidayRepository) Create(holiday *models.Holiday) error {
	return r.db.Create(holiday).Error
}

func (r *HolidayRepository) FindByID(id string) (*models.Holiday, error) {
	var holiday models.Holiday
	err := r.db.Preload("Company").Where("id = ? AND deleted_at IS NULL", id).First(&holiday).Error
	return &holiday, err
}

func (r *HolidayRepository) List(companyID string, page, limit int) ([]models.Holiday, int64, error) {
	base := r.db.Model(&models.Holiday{}).Where("deleted_at IS NULL")
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Holiday
	err := base.Preload("Company").Order("date DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *HolidayRepository) Update(holiday *models.Holiday) error {
	return r.db.Save(holiday).Error
}

func (r *HolidayRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Holiday{}).Error
}

func (r *HolidayRepository) ListActiveByDate(date, companyID string) ([]models.Holiday, error) {
	var list []models.Holiday
	query := r.db.Where("(date = ? OR weekend_date = ?) AND status = 'active' AND deleted_at IS NULL", date, date)
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *HolidayRepository) ListActiveByDateRange(startDate, endDate, companyID string) ([]models.Holiday, error) {
	var list []models.Holiday
	query := r.db.Where("(date BETWEEN ? AND ? OR weekend_date BETWEEN ? AND ?) AND status = 'active' AND deleted_at IS NULL",
		startDate, endDate, startDate, endDate)
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&list).Error
	return list, err
}
