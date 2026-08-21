package repository

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) SaveRefreshToken(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *AuthRepository) FindRefreshToken(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&token).Error
	return &token, err
}

func (r *AuthRepository) RevokeRefreshToken(tokenHash string) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked_at", &now).Error
}

func (r *AuthRepository) RevokeAllUserTokens(userID string) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

func (r *AuthRepository) RevokeAllUserTokensExcept(userID, tokenHash string) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND token_hash != ? AND revoked_at IS NULL", userID, tokenHash).
		Update("revoked_at", &now).Error
}

func (r *AuthRepository) SaveLoginHistory(record *models.LoginHistory) error {
	return r.db.Create(record).Error
}

func (r *AuthRepository) CountRecentFailedLogins(email string, minutes int) int64 {
	var count int64
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	r.db.Model(&models.LoginHistory{}).
		Where("email = ? AND status = 'failed' AND created_at > ?", email, since).
		Count(&count)
	return count
}

func (r *AuthRepository) CreateAuditLog(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuthRepository) SaveSession(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *AuthRepository) DeleteSession(sessionID string) error {
	return r.db.Where("id = ?", sessionID).Delete(&models.Session{}).Error
}

func (r *AuthRepository) DeleteAllUserSessions(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

func (r *AuthRepository) GetActiveSessions(userID string) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.Where("user_id = ? AND expires_at > ? AND deleted_at IS NULL", userID, time.Now()).
		Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

func (r *AuthRepository) GetActiveSessionCount(userID string) int64 {
	var count int64
	r.db.Model(&models.Session{}).
		Where("user_id = ? AND expires_at > ? AND deleted_at IS NULL", userID, time.Now()).
		Count(&count)
	return count
}

func (r *AuthRepository) SavePasswordReset(record *models.PasswordReset) error {
	return r.db.Create(record).Error
}

func (r *AuthRepository) FindPasswordReset(tokenHash string) (*models.PasswordReset, error) {
	var record models.PasswordReset
	err := r.db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ? AND deleted_at IS NULL", tokenHash, time.Now()).
		First(&record).Error
	return &record, err
}

func (r *AuthRepository) MarkPasswordResetUsed(recordID string) error {
	now := time.Now()
	return r.db.Model(&models.PasswordReset{}).Where("id = ?", recordID).
		Update("used_at", &now).Error
}

func (r *AuthRepository) SaveEmailVerification(record *models.EmailVerification) error {
	return r.db.Create(record).Error
}

func (r *AuthRepository) FindEmailVerification(tokenHash, purpose string) (*models.EmailVerification, error) {
	var record models.EmailVerification
	err := r.db.Where("token_hash = ? AND purpose = ? AND used_at IS NULL AND expires_at > ? AND deleted_at IS NULL",
		tokenHash, purpose, time.Now()).First(&record).Error
	return &record, err
}

func (r *AuthRepository) MarkEmailVerificationUsed(recordID string) error {
	now := time.Now()
	return r.db.Model(&models.EmailVerification{}).Where("id = ?", recordID).
		Update("used_at", &now).Error
}

func (r *AuthRepository) FindByRefreshTokenHash(hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	return &token, err
}
