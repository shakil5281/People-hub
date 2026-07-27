package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type EidBonusRepository struct {
	db *gorm.DB
}

func NewEidBonusRepository(db *gorm.DB) *EidBonusRepository {
	return &EidBonusRepository{db: db}
}

func (r *EidBonusRepository) Create(bonus *models.EidBonus) error {
	return r.db.Create(bonus).Error
}

func (r *EidBonusRepository) Upsert(bonus *models.EidBonus) error {
	return r.db.Where("employee_id = ? AND year = ? AND company_id = ?",
		bonus.EmployeeID, bonus.Year, bonus.CompanyID).
		Assign(bonus).
		FirstOrCreate(bonus).Error
}

func (r *EidBonusRepository) FindByEmployeeYear(employeeID string, year int) (*models.EidBonus, error) {
	var b models.EidBonus
	err := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").Preload("Employee.LineRef").
		Where("employee_id = ? AND year = ? AND deleted_at IS NULL", employeeID, year).First(&b).Error
	return &b, err
}

type EidBonusFilter struct {
	CompanyID string
	Year      int
	BonusType string
}

func (r *EidBonusRepository) ListAllByYear(f EidBonusFilter) ([]models.EidBonus, error) {
	query := r.db.Preload("Employee.Department").
		Preload("Employee.SectionRef").
		Preload("Employee.DesignationRef").
		Preload("Employee.LineRef").
		Preload("Employee.GroupRef").
		Where("company_id = ? AND year = ? AND deleted_at IS NULL", f.CompanyID, f.Year)

	if f.BonusType != "" {
		query = query.Where("bonus_type = ?", f.BonusType)
	}

	var bonuses []models.EidBonus
	err := query.Order("LENGTH(employee_id) ASC, employee_id ASC").Find(&bonuses).Error
	return bonuses, err
}

func (r *EidBonusRepository) DeleteByYear(companyID string, year int) error {
	return r.db.Unscoped().Where("company_id = ? AND year = ?", companyID, year).
		Delete(&models.EidBonus{}).Error
}
