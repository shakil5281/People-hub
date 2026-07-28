package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *GroupRepository) FindByID(id string) (*models.Group, error) {
	var group models.Group
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&group).Error
	return &group, err
}

func (r *GroupRepository) List(page, limit int) ([]models.Group, int64, error) {
	base := r.db.Model(&models.Group{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var groups []models.Group
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&groups).Error
	return groups, total, err
}

func (r *GroupRepository) ListAll() ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Where("deleted_at IS NULL").Order("name ASC").Find(&groups).Error
	return groups, err
}

func (r *GroupRepository) Update(group *models.Group) error {
	return r.db.Save(group).Error
}

func (r *GroupRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Group{}).Error
}
