package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type DepartmentRepository struct{ db *gorm.DB }

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(m *models.Department) error { return r.db.Create(m).Error }

func (r *DepartmentRepository) FindByID(id string) (*models.Department, error) {
	var m models.Department
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *DepartmentRepository) List(page, limit int) ([]models.Department, int64, error) {
	base := r.db.Model(&models.Department{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Department
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *DepartmentRepository) Update(m *models.Department) error { return r.db.Save(m).Error }

func (r *DepartmentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Department{}).Error
}

type SectionRepository struct{ db *gorm.DB }

func NewSectionRepository(db *gorm.DB) *SectionRepository {
	return &SectionRepository{db: db}
}

func (r *SectionRepository) Create(m *models.Section) error { return r.db.Create(m).Error }

func (r *SectionRepository) FindByID(id string) (*models.Section, error) {
	var m models.Section
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *SectionRepository) List(departmentID string, page, limit int) ([]models.Section, int64, error) {
	base := r.db.Model(&models.Section{}).Where("deleted_at IS NULL")
	if departmentID != "" {
		base = base.Where("department_id = ?", departmentID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Section
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *SectionRepository) Update(m *models.Section) error { return r.db.Save(m).Error }

func (r *SectionRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Section{}).Error
}

type DesignationRepository struct{ db *gorm.DB }

func NewDesignationRepository(db *gorm.DB) *DesignationRepository {
	return &DesignationRepository{db: db}
}

func (r *DesignationRepository) Create(m *models.Designation) error { return r.db.Create(m).Error }

func (r *DesignationRepository) FindByID(id string) (*models.Designation, error) {
	var m models.Designation
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *DesignationRepository) List(sectionID string, page, limit int) ([]models.Designation, int64, error) {
	base := r.db.Model(&models.Designation{}).Where("deleted_at IS NULL")
	if sectionID != "" {
		base = base.Where("section_id = ?", sectionID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Designation
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *DesignationRepository) Update(m *models.Designation) error { return r.db.Save(m).Error }

func (r *DesignationRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Designation{}).Error
}

type LineRepository struct{ db *gorm.DB }

func NewLineRepository(db *gorm.DB) *LineRepository {
	return &LineRepository{db: db}
}

func (r *LineRepository) Create(m *models.Line) error { return r.db.Create(m).Error }

func (r *LineRepository) FindByID(id string) (*models.Line, error) {
	var m models.Line
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *LineRepository) List(sectionID string, page, limit int) ([]models.Line, int64, error) {
	base := r.db.Model(&models.Line{}).Where("deleted_at IS NULL")
	if sectionID != "" {
		base = base.Where("section_id = ?", sectionID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Line
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *LineRepository) Update(m *models.Line) error { return r.db.Save(m).Error }

func (r *LineRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Line{}).Error
}
