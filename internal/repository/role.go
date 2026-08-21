package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByID(id string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	return &role, err
}

func (r *RoleRepository) FindByName(companyID, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("company_id = ? AND name = ? AND deleted_at IS NULL", companyID, name).First(&role).Error
	return &role, err
}

func (r *RoleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Role{}).Error
}

func (r *RoleRepository) GetRolePermissions(roleID string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *RoleRepository) AssignPermission(roleID, permissionID string) error {
	rp := models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	return r.db.Create(&rp).Error
}

func (r *RoleRepository) RemovePermission(roleID, permissionID string) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{}).Error
}

func (r *RoleRepository) ListByCompany(companyID string) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) ListAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Order("resource ASC, action ASC").Find(&permissions).Error
	return permissions, err
}

func (r *RoleRepository) ReplaceRolePermissions(roleID string, permissionIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		for _, pid := range permissionIDs {
			rp := models.RolePermission{RoleID: roleID, PermissionID: pid}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
