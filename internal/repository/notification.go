package repository

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) WithTx(tx *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: tx}
}

func (r *NotificationRepository) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepository) FindByID(id string) (*models.Notification, error) {
	var n models.Notification
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&n).Error
	return &n, err
}

func (r *NotificationRepository) List(userID string, isRead *bool, page, limit int) ([]models.Notification, int64, error) {
	base := r.db.Model(&models.Notification{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if isRead != nil {
		base = base.Where("is_read = ?", *isRead)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Notification
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *NotificationRepository) MarkAsRead(id, userID string) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": &now}).Error
}

func (r *NotificationRepository) MarkAllAsRead(userID string) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ? AND deleted_at IS NULL", userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": &now}).Error
}

func (r *NotificationRepository) GetUnreadCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ? AND deleted_at IS NULL", userID, false).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepository) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Notification{}).Error
}
