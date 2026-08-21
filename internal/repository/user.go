package repository

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	return &user, err
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) IncrementFailedAttempts(userID string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		UpdateColumn("failed_attempts", gorm.Expr("failed_attempts + 1")).Error
}

func (r *UserRepository) ResetFailedAttempts(userID string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		UpdateColumn("failed_attempts", 0).Error
}

func (r *UserRepository) LockAccount(userID string, lockedBy string) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{
			"locked_at": &now,
			"locked_by": lockedBy,
			"status":    "locked",
		}).Error
}

func (r *UserRepository) UnlockAccount(userID string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{
			"locked_at":       nil,
			"locked_by":       nil,
			"failed_attempts": 0,
			"status":          "active",
		}).Error
}

func (r *UserRepository) UpdateLastLogin(userID, ip string) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{
			"last_login_at": &now,
			"last_login_ip": ip,
		}).Error
}

func (r *UserRepository) UpdatePassword(userID, passwordHash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

func (r *UserRepository) SetForcePasswordChange(userID string, value bool) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("force_password_change", value).Error
}

func (r *UserRepository) GetPasswordHistory(userID string, limit int) ([]models.PasswordHistory, error) {
	var history []models.PasswordHistory
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").Limit(limit).Find(&history).Error
	return history, err
}

func (r *UserRepository) AddPasswordHistory(userID, passwordHash string) error {
	record := models.PasswordHistory{
		UserID:       userID,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	return r.db.Create(&record).Error
}

func (r *UserRepository) GetUserRoles(userID string) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	return roles, err
}

func (r *UserRepository) GetUserPermissions(userID string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Distinct().
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&permissions).Error
	return permissions, err
}

func (r *UserRepository) GetUserCompany(userID string) (*string, error) {
	var employee models.Employee
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).First(&employee).Error
	if err != nil {
		return nil, err
	}
	return &employee.CompanyID, nil
}

func (r *UserRepository) ListFiltered(page, limit int, status, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{}).Where("deleted_at IS NULL")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("email ILIKE ? OR name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *UserRepository) CountByRole(roleName string) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", roleName).
		Count(&count).Error
	return count, err
}

func (r *UserRepository) ReplaceUserRoles(userID string, roleIDs []string, createdBy *string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			ur := models.UserRole{UserID: userID, RoleID: roleID, CreatedBy: createdBy}
			if err := tx.Create(&ur).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) AssignRole(userID, roleID, createdBy string) error {
	userRole := models.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		CreatedBy: &createdBy,
	}
	return r.db.Create(&userRole).Error
}

func (r *UserRepository) RemoveRole(userID, roleID string) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error
}
