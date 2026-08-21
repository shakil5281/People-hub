package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type FloorRepository struct {
	db *gorm.DB
}

func NewFloorRepository(db *gorm.DB) *FloorRepository {
	return &FloorRepository{db: db}
}

func (r *FloorRepository) Create(floor *models.Floor) error {
	return r.db.Create(floor).Error
}

func (r *FloorRepository) FindByID(id string) (*models.Floor, error) {
	var floor models.Floor
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&floor).Error
	return &floor, err
}

func (r *FloorRepository) List(page, limit int) ([]models.Floor, int64, error) {
	base := r.db.Model(&models.Floor{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var floors []models.Floor
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&floors).Error
	return floors, total, err
}

func (r *FloorRepository) Update(floor *models.Floor) error {
	return r.db.Save(floor).Error
}

func (r *FloorRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Floor{}).Error
}
