package repository

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type SystemLogRepository struct {
	db *gorm.DB
}

func NewSystemLogRepository(db *gorm.DB) *SystemLogRepository {
	return &SystemLogRepository{db: db}
}

type SystemLogFilter struct {
	Level     string
	Source    string
	UserID    string
	StartDate string
	EndDate   string
	Page      int
	Limit     int
}

func (r *SystemLogRepository) Create(log *models.SystemLog) error {
	return r.db.Create(log).Error
}

func (r *SystemLogRepository) FindByID(id string) (*models.SystemLog, error) {
	var log models.SystemLog
	err := r.db.Preload("User").Preload("Company").
		Where("id = ?", id).First(&log).Error
	return &log, err
}

func (r *SystemLogRepository) List(filter SystemLogFilter) ([]models.SystemLog, int64, error) {
	base := r.db.Model(&models.SystemLog{})
	if filter.Level != "" {
		base = base.Where("level = ?", filter.Level)
	}
	if filter.Source != "" {
		base = base.Where("source = ?", filter.Source)
	}
	if filter.UserID != "" {
		base = base.Where("user_id = ?", filter.UserID)
	}
	if filter.StartDate != "" {
		base = base.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		base = base.Where("created_at <= ?", filter.EndDate+"T23:59:59Z")
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	var list []models.SystemLog
	err := base.Order("created_at DESC").
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Preload("User").
		Preload("Company").
		Find(&list).Error
	return list, total, err
}

func (r *SystemLogRepository) DeleteOlderThan(date time.Time) error {
	return r.db.Where("created_at < ?", date).Delete(&models.SystemLog{}).Error
}

func (r *SystemLogRepository) PurgeAll() error {
	return r.db.Where("1 = 1").Delete(&models.SystemLog{}).Error
}

type LogStats struct {
	Total    int64            `json:"total"`
	ByLevel []LevelCount     `json:"by_level"`
	BySource []SourceCount   `json:"by_source"`
}

type LevelCount struct {
	Level string `json:"level"`
	Count int64  `json:"count"`
}

type SourceCount struct {
	Source string `json:"source"`
	Count  int64  `json:"count"`
}

func (r *SystemLogRepository) Stats() (*LogStats, error) {
	var total int64
	if err := r.db.Model(&models.SystemLog{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var byLevel []LevelCount
	r.db.Model(&models.SystemLog{}).
		Select("level, COUNT(*) as count").
		Group("level").
		Order("count DESC").
		Find(&byLevel)

	var bySource []SourceCount
	r.db.Model(&models.SystemLog{}).
		Select("source, COUNT(*) as count").
		Group("source").
		Order("count DESC").
		Find(&bySource)

	return &LogStats{
		Total:    total,
		ByLevel:  byLevel,
		BySource: bySource,
	}, nil
}
