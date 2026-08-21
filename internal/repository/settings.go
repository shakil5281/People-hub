package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetAll() ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	err := r.db.Order("key ASC").Find(&settings).Error
	return settings, err
}

func (r *SettingsRepository) Upsert(key, value string) error {
	var existing models.SystemSetting
	result := r.db.Where("key = ?", key).First(&existing)
	if result.Error != nil {
		return r.db.Create(&models.SystemSetting{Key: key, Value: value}).Error
	}
	return r.db.Model(&existing).Update("value", value).Error
}

func (r *SettingsRepository) BulkUpsert(settings map[string]string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			var existing models.SystemSetting
			result := tx.Where("key = ?", key).First(&existing)
			if result.Error != nil {
				if err := tx.Create(&models.SystemSetting{Key: key, Value: value}).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&existing).Update("value", value).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
