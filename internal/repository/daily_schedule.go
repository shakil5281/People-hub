package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type DailyScheduleRepository struct {
	db *gorm.DB
}

func NewDailyScheduleRepository(db *gorm.DB) *DailyScheduleRepository {
	return &DailyScheduleRepository{db: db}
}

func (r *DailyScheduleRepository) List(companyID string) ([]models.DailySchedule, error) {
	var items []models.DailySchedule
	q := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").Where("deleted_at IS NULL")
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *DailyScheduleRepository) ListByDate(companyID string, date string) ([]models.DailySchedule, error) {
	var items []models.DailySchedule
	q := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").Where("date = ? AND deleted_at IS NULL", date)
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *DailyScheduleRepository) Create(item *models.DailySchedule) error {
	return r.db.Create(item).Error
}

func (r *DailyScheduleRepository) FindByID(id string) (*models.DailySchedule, error) {
	var item models.DailySchedule
	err := r.db.Preload("Employee.Department").
		Preload("Employee.DesignationRef").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error
	return &item, err
}

func (r *DailyScheduleRepository) Update(item *models.DailySchedule) error {
	return r.db.Save(item).Error
}

func (r *DailyScheduleRepository) Delete(item *models.DailySchedule) error {
	return r.db.Delete(item).Error
}
