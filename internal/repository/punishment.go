package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type PunishmentRepository struct {
	db *gorm.DB
}

func NewPunishmentRepository(db *gorm.DB) *PunishmentRepository {
	return &PunishmentRepository{db: db}
}

func (r *PunishmentRepository) List(companyID string) ([]models.Punishment, error) {
	var items []models.Punishment
	q := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").Where("deleted_at IS NULL")
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *PunishmentRepository) ListPaginated(companyID string, page, limit int) ([]models.Punishment, int64, error) {
	base := r.db.Model(&models.Punishment{}).Preload("Employee.Department").Preload("Employee.DesignationRef").Where("deleted_at IS NULL")
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.Punishment
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *PunishmentRepository) Create(item *models.Punishment) error {
	return r.db.Create(item).Error
}

func (r *PunishmentRepository) FindByID(id string) (*models.Punishment, error) {
	var item models.Punishment
	err := r.db.Preload("Employee.Department").
		Preload("Employee.DesignationRef").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error
	return &item, err
}

func (r *PunishmentRepository) Update(item *models.Punishment) error {
	return r.db.Save(item).Error
}

func (r *PunishmentRepository) Delete(item *models.Punishment) error {
	return r.db.Delete(item).Error
}
