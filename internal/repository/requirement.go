package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type SectionSummaryRow struct {
	SectionName string `json:"section_name"`
	GroupType   string `json:"group_type"`
	Designation string `json:"designation"`
	Required    int64  `json:"required"`
	Present     int64  `json:"present"`
	Net         int64  `json:"net"`
}

type RequirementRepository struct {
	db *gorm.DB
}

func NewRequirementRepository(db *gorm.DB) *RequirementRepository {
	return &RequirementRepository{db: db}
}

func (r *RequirementRepository) Create(req *models.Requirement) error {
	return r.db.Create(req).Error
}

func (r *RequirementRepository) FindByID(id string) (*models.Requirement, error) {
	var req models.Requirement
	err := r.db.Preload("Department").Preload("Section").Preload("Designation").Where("id = ? AND deleted_at IS NULL", id).First(&req).Error
	return &req, err
}

func (r *RequirementRepository) List(page, limit int) ([]models.Requirement, int64, error) {
	base := r.db.Model(&models.Requirement{}).Where("deleted_at IS NULL")
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var reqs []models.Requirement
	err := base.Preload("Department").Preload("Section").Preload("Designation").Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&reqs).Error
	return reqs, total, err
}

func (r *RequirementRepository) ListFiltered(departmentID, status, priority, position string, page, limit int) ([]models.Requirement, int64, error) {
	base := r.db.Model(&models.Requirement{}).Where("deleted_at IS NULL").Order("created_at DESC")
	if departmentID != "" {
		base = base.Where("department_id = ?", departmentID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if priority != "" {
		base = base.Where("priority = ?", priority)
	}
	if position != "" {
		base = base.Where("position ILIKE ?", "%"+position+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var reqs []models.Requirement
	err := base.Preload("Department").Preload("Section").Preload("Designation").Offset((page - 1) * limit).Limit(limit).Find(&reqs).Error
	return reqs, total, err
}

func (r *RequirementRepository) ListBySection() ([]SectionSummaryRow, error) {
	sql := `
		SELECT
			COALESCE(sections.name, 'General') as section_name,
			COALESCE(req.group_type, 'Worker') as group_type,
			COALESCE(designations.name, req.position) as designation,
			SUM(req.vacancies) as required,
			COALESCE(MAX(emp.present_count), 0) as present,
			SUM(req.vacancies) - COALESCE(MAX(emp.present_count), 0) as net
		FROM requirements req
		LEFT JOIN sections ON sections.id = req.section_id
		LEFT JOIN designations ON designations.id = req.designation_id
		LEFT JOIN (
			SELECT designation_id, COUNT(*) as present_count
			FROM employees
			WHERE deleted_at IS NULL AND status = 'active'
			GROUP BY designation_id
		) emp ON emp.designation_id = req.designation_id
		WHERE req.deleted_at IS NULL AND req.status = 'Open'
		GROUP BY sections.name, req.group_type, req.position, designations.name
		ORDER BY sections.name, req.group_type, req.position
	`
	var results []SectionSummaryRow
	if err := r.db.Raw(sql).Scan(&results).Error; err != nil {
		return nil, err
	}
	if results == nil {
		results = []SectionSummaryRow{}
	}
	return results, nil
}

func (r *RequirementRepository) Update(req *models.Requirement) error {
	return r.db.Save(req).Error
}

func (r *RequirementRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Requirement{}).Error
}
