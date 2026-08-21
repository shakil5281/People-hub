package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(company *models.Company) error {
	return r.db.Create(company).Error
}

func (r *CompanyRepository) FindByID(id string) (*models.Company, error) {
	var company models.Company
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&company).Error
	return &company, err
}

func (r *CompanyRepository) FindBySlug(slug string) (*models.Company, error) {
	var company models.Company
	err := r.db.Where("slug = ? AND deleted_at IS NULL", slug).First(&company).Error
	return &company, err
}

func (r *CompanyRepository) List(page, limit int) ([]models.Company, int64, error) {
	base := r.db.Model(&models.Company{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var companies []models.Company
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&companies).Error
	return companies, total, err
}

func (r *CompanyRepository) Update(company *models.Company) error {
	return r.db.Save(company).Error
}

func (r *CompanyRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Company{}).Error
}
