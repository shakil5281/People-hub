package repository

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type DataLogRepository struct {
	db *gorm.DB
}

func NewDataLogRepository(db *gorm.DB) *DataLogRepository {
	return &DataLogRepository{db: db}
}

func (r *DataLogRepository) BatchCreate(logs []models.DataLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(logs, 500).Error
}

func (r *DataLogRepository) ListByDateRange(start, end string) ([]models.DataLog, error) {
	var logs []models.DataLog
	err := r.db.Where("date >= ? AND date <= ? AND deleted_at IS NULL", start, end).
		Order("user_id ASC, punch_time ASC").Find(&logs).Error
	return logs, err
}

func (r *DataLogRepository) ListUnprocessed() ([]models.DataLog, error) {
	var logs []models.DataLog
	err := r.db.Where("processed = false AND deleted_at IS NULL").
		Order("user_id ASC, punch_time ASC").Find(&logs).Error
	return logs, err
}

func (r *DataLogRepository) ListUnprocessedByDate(date string) ([]models.DataLog, error) {
	var logs []models.DataLog
	err := r.db.Where("date = ? AND processed = false AND deleted_at IS NULL", date).
		Order("user_id ASC, punch_time ASC").Find(&logs).Error
	return logs, err
}

func (r *DataLogRepository) ListByDate(date string) ([]models.DataLog, error) {
	var logs []models.DataLog
	err := r.db.Where("date = ? AND deleted_at IS NULL", date).
		Order("user_id ASC, punch_time ASC").Find(&logs).Error
	return logs, err
}

func (r *DataLogRepository) ListUnprocessedByDateRange(start, end string) ([]models.DataLog, error) {
	var logs []models.DataLog
	err := r.db.Where("date >= ? AND date <= ? AND processed = false AND deleted_at IS NULL", start, end).
		Order("date ASC, user_id ASC, punch_time ASC").Find(&logs).Error
	return logs, err
}

func (r *DataLogRepository) ListUnprocessedDatesInRange(start, end string) ([]string, error) {
	var dates []string
	err := r.db.Model(&models.DataLog{}).
		Where("date >= ? AND date <= ? AND processed = false AND deleted_at IS NULL", start, end).
		Distinct("date").Order("date ASC").Pluck("date", &dates).Error
	return dates, err
}

func (r *DataLogRepository) MarkProcessed(ids []string) error {
	return r.db.Model(&models.DataLog{}).Where("id IN ?", ids).
		Update("processed", true).Error
}

func (r *DataLogRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.DataLog{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

func (r *DataLogRepository) CountByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&models.DataLog{}).Where("date = ? AND deleted_at IS NULL", date).Count(&count).Error
	return count, err
}

func (r *DataLogRepository) DeleteOlderThan(t time.Time) error {
	return r.db.Where("punch_time < ?", t).Delete(&models.DataLog{}).Error
}

func (r *DataLogRepository) DeleteAll() error {
	return r.db.Unscoped().Where("1 = 1").Delete(&models.DataLog{}).Error
}

func (r *DataLogRepository) ListByBadgeAndDateRange(badgeNumbers []string, start, end string) ([]models.DataLog, error) {
	var logs []models.DataLog
	err := r.db.Where("badge_number IN ? AND date >= ? AND date <= ? AND deleted_at IS NULL", badgeNumbers, start, end).
		Order("badge_number ASC, punch_time ASC").Find(&logs).Error
	return logs, err
}

func (r *DataLogRepository) ExistsByBadgeAndPunchTime(badgeNumber string, punchTime time.Time) bool {
	var count int64
	if err := r.db.Model(&models.DataLog{}).
		Where("badge_number = ? AND punch_time = ? AND deleted_at IS NULL", badgeNumber, punchTime).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}
