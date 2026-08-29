package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type RosterRepository struct {
	db *gorm.DB
}

func NewRosterRepository(db *gorm.DB) *RosterRepository {
	return &RosterRepository{db: db}
}

func (r *RosterRepository) WithTx(tx *gorm.DB) *RosterRepository {
	return &RosterRepository{db: tx}
}

func (r *RosterRepository) Create(roster *models.Roster) error {
	return r.db.Create(roster).Error
}

func (r *RosterRepository) BatchCreate(rosters []models.Roster) error {
	if len(rosters) == 0 {
		return nil
	}
	return r.db.Create(&rosters).Error
}

func (r *RosterRepository) FindByID(id string) (*models.Roster, error) {
	var roster models.Roster
	err := r.db.Preload("Employee").Preload("Shift").Preload("Company").Where("id = ?", id).First(&roster).Error
	if err != nil {
		return nil, err
	}
	return &roster, nil
}

func (r *RosterRepository) FindByEmployeeAndDate(employeeID, date string) (*models.Roster, error) {
	var roster models.Roster
	err := r.db.Preload("Shift").Where("employee_id = ? AND date = ?", employeeID, date).First(&roster).Error
	if err != nil {
		return nil, err
	}
	return &roster, nil
}

func (r *RosterRepository) List(companyID string, page, limit int, filters map[string]string) ([]models.Roster, int64, error) {
	base := r.db.Model(&models.Roster{})

	if companyID != "" {
		base = base.Where("rosters.company_id = ?", companyID)
	}
	if v := filters["employee_id"]; v != "" {
		base = base.Where("rosters.employee_id = ?", v)
	}
	if v := filters["shift_id"]; v != "" {
		base = base.Where("rosters.shift_id = ?", v)
	}
	if v := filters["status"]; v != "" {
		base = base.Where("rosters.status = ?", v)
	}
	if v := filters["from_date"]; v != "" {
		base = base.Where("rosters.date >= ?", v)
	}
	if v := filters["to_date"]; v != "" {
		base = base.Where("rosters.date <= ?", v)
	}
	if v := filters["department_id"]; v != "" {
		base = base.Where("rosters.employee_id IN (SELECT employee_id FROM employees WHERE department_id = ? AND deleted_at IS NULL)", v)
	}
	if v := filters["section_id"]; v != "" {
		base = base.Where("rosters.employee_id IN (SELECT employee_id FROM employees WHERE section_id = ? AND deleted_at IS NULL)", v)
	}
	if v := filters["line_id"]; v != "" {
		base = base.Where("rosters.employee_id IN (SELECT employee_id FROM employees WHERE line_id = ? AND deleted_at IS NULL)", v)
	}
	if v := filters["group_id"]; v != "" {
		base = base.Where("rosters.employee_id IN (SELECT employee_id FROM employees WHERE group_id = ? AND deleted_at IS NULL)", v)
	}
	if v := filters["search"]; v != "" {
		like := "%" + v + "%"
		base = base.Where("rosters.employee_id ILIKE ? OR rosters.reason ILIKE ?", like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []models.Roster
	err := base.Preload("Employee").Preload("Shift").
		Order("rosters.date DESC, rosters.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *RosterRepository) ListByEmployeeAndDateRange(employeeID, startDate, endDate string) ([]models.Roster, error) {
	var list []models.Roster
	err := r.db.Preload("Shift").Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, startDate, endDate).Find(&list).Error
	return list, err
}

func (r *RosterRepository) ListByCompanyAndDateRange(companyID, startDate, endDate string) ([]models.Roster, error) {
	var list []models.Roster
	q := r.db.Preload("Shift").Where("date BETWEEN ? AND ?", startDate, endDate)
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	err := q.Find(&list).Error
	return list, err
}

func (r *RosterRepository) Update(roster *models.Roster) error {
	return r.db.Save(roster).Error
}

func (r *RosterRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Roster{}).Error
}

func (r *RosterRepository) BulkDelete(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&models.Roster{}).Error
}
