package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type DivisionRepository struct{ db *gorm.DB }

func NewDivisionRepository(db *gorm.DB) *DivisionRepository {
	return &DivisionRepository{db: db}
}

func (r *DivisionRepository) Create(m *models.Division) error { return r.db.Create(m).Error }

func (r *DivisionRepository) FindByID(id string) (*models.Division, error) {
	var m models.Division
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *DivisionRepository) List(page, limit int) ([]models.Division, int64, error) {
	base := r.db.Model(&models.Division{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Division
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *DivisionRepository) Update(m *models.Division) error { return r.db.Save(m).Error }

func (r *DivisionRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Division{}).Error
}

type DistrictRepository struct{ db *gorm.DB }

func NewDistrictRepository(db *gorm.DB) *DistrictRepository {
	return &DistrictRepository{db: db}
}

func (r *DistrictRepository) Create(m *models.District) error { return r.db.Create(m).Error }

func (r *DistrictRepository) FindByID(id string) (*models.District, error) {
	var m models.District
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *DistrictRepository) List(divisionID string, page, limit int) ([]models.District, int64, error) {
	base := r.db.Model(&models.District{}).Where("deleted_at IS NULL")
	if divisionID != "" {
		base = base.Where("division_id = ?", divisionID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.District
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *DistrictRepository) Update(m *models.District) error { return r.db.Save(m).Error }

func (r *DistrictRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.District{}).Error
}

type UpazilaRepository struct{ db *gorm.DB }

func NewUpazilaRepository(db *gorm.DB) *UpazilaRepository {
	return &UpazilaRepository{db: db}
}

func (r *UpazilaRepository) Create(m *models.Upazila) error { return r.db.Create(m).Error }

func (r *UpazilaRepository) FindByID(id string) (*models.Upazila, error) {
	var m models.Upazila
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *UpazilaRepository) List(districtID string, page, limit int) ([]models.Upazila, int64, error) {
	base := r.db.Model(&models.Upazila{}).Where("deleted_at IS NULL")
	if districtID != "" {
		base = base.Where("district_id = ?", districtID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Upazila
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *UpazilaRepository) Update(m *models.Upazila) error { return r.db.Save(m).Error }

func (r *UpazilaRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Upazila{}).Error
}

type UnionRepository struct{ db *gorm.DB }

func NewUnionRepository(db *gorm.DB) *UnionRepository {
	return &UnionRepository{db: db}
}

func (r *UnionRepository) Create(m *models.Union) error { return r.db.Create(m).Error }

func (r *UnionRepository) FindByID(id string) (*models.Union, error) {
	var m models.Union
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *UnionRepository) List(upazilaID string, page, limit int) ([]models.Union, int64, error) {
	base := r.db.Model(&models.Union{}).Where("deleted_at IS NULL")
	if upazilaID != "" {
		base = base.Where("upazila_id = ?", upazilaID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Union
	err := base.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *UnionRepository) Update(m *models.Union) error { return r.db.Save(m).Error }

func (r *UnionRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Union{}).Error
}

type PostOfficeRepository struct{ db *gorm.DB }

func NewPostOfficeRepository(db *gorm.DB) *PostOfficeRepository {
	return &PostOfficeRepository{db: db}
}

func (r *PostOfficeRepository) Create(m *models.PostOffice) error { return r.db.Create(m).Error }

func (r *PostOfficeRepository) FindByID(id string) (*models.PostOffice, error) {
	var m models.PostOffice
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PostOfficeRepository) FindByCode(code string) (*models.PostOffice, error) {
	var m models.PostOffice
	if err := r.db.Where("postal_code = ?", code).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PostOfficeRepository) List(upazilaID string, districtID string, page, limit int) ([]models.PostOffice, int64, error) {
	base := r.db.Model(&models.PostOffice{}).Where("deleted_at IS NULL")
	if upazilaID != "" {
		base = base.Where("upazila_id = ?", upazilaID)
	}
	if districtID != "" {
		base = base.Where("district_id = ?", districtID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.PostOffice
	err := base.Order("name ASC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *PostOfficeRepository) Update(m *models.PostOffice) error { return r.db.Save(m).Error }

func (r *PostOfficeRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.PostOffice{}).Error
}
