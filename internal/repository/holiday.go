package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"gorm.io/gorm"
)

type HolidayRepository struct {
	db *gorm.DB
}

func NewHolidayRepository(db *gorm.DB) *HolidayRepository {
	return &HolidayRepository{db: db}
}

func (r *HolidayRepository) WithTx(tx *gorm.DB) *HolidayRepository {
	return &HolidayRepository{db: tx}
}

func normalizeHolidayDates(h *models.Holiday) {
	if h == nil {
		return
	}
	h.Date = utils.NormalizeDate(h.Date)
	if h.FromDate != nil {
		n := utils.NormalizeDate(*h.FromDate)
		h.FromDate = &n
	}
	if h.ToDate != nil {
		n := utils.NormalizeDate(*h.ToDate)
		h.ToDate = &n
	}
	if h.WeekendDate != nil {
		n := utils.NormalizeDate(*h.WeekendDate)
		h.WeekendDate = &n
	}
}

func (r *HolidayRepository) Create(holiday *models.Holiday) error {
	if err := r.db.Create(holiday).Error; err != nil {
		return err
	}
	normalizeHolidayDates(holiday)
	return nil
}

func (r *HolidayRepository) FindByID(id string) (*models.Holiday, error) {
	var holiday models.Holiday
	err := r.db.Preload("Company").Where("id = ? AND deleted_at IS NULL", id).First(&holiday).Error
	if err != nil {
		return &holiday, err
	}
	normalizeHolidayDates(&holiday)
	return &holiday, nil
}

func (r *HolidayRepository) List(companyID string, page, limit int) ([]models.Holiday, int64, error) {
	base := r.db.Model(&models.Holiday{}).Where("deleted_at IS NULL")
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Holiday
	err := base.Preload("Company").Order("date DESC").Offset((page - 1) * limit).Limit(limit).Find(&list).Error
	if err != nil {
		return list, total, err
	}
	for i := range list {
		normalizeHolidayDates(&list[i])
	}
	return list, total, nil
}

func (r *HolidayRepository) Update(holiday *models.Holiday) error {
	return r.db.Save(holiday).Error
}

func (r *HolidayRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Holiday{}).Error
}

func (r *HolidayRepository) ListActiveByDate(date, companyID string) ([]models.Holiday, error) {
	var list []models.Holiday
	query := r.db.Where(`(date = ? OR weekend_date = ? OR (from_date IS NOT NULL AND to_date IS NOT NULL AND from_date <= ? AND to_date >= ?)) AND status = 'active' AND deleted_at IS NULL`,
		date, date, date, date)
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&list).Error
	if err != nil {
		return list, err
	}
	for i := range list {
		normalizeHolidayDates(&list[i])
	}
	return list, nil
}

// CountActiveWeekendChangeCollision counts active weekend_change records where the
// given date is already used as either the general duty date or the weekend date.
// When excludeID is non-empty, that record is ignored (used on update to allow
// editing the same row). Mirrors the dedup rule in the weekend-change validation.
func (r *HolidayRepository) CountActiveWeekendChangeCollision(date, companyID, excludeID string) (int64, error) {
	var count int64
	query := r.db.Model(&models.Holiday{}).
		Where(`(date = ? OR weekend_date = ?) AND type = 'weekend_change' AND status = 'active' AND deleted_at IS NULL`, date, date)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Count(&count).Error
	return count, err
}

// CountActiveHolidayOnDate counts active non-weekend-change holidays covering the given
// date (single date, weekend_date, or a from/to range). Used to reject a weekend change
// whose general duty date is already a government holiday.
func (r *HolidayRepository) CountActiveHolidayOnDate(date, companyID string) (int64, error) {
	var count int64
	query := r.db.Model(&models.Holiday{}).
		Where(`type != 'weekend_change' AND status = 'active' AND deleted_at IS NULL
			AND (date = ? OR weekend_date = ? OR (from_date IS NOT NULL AND to_date IS NOT NULL AND from_date <= ? AND to_date >= ?))`,
			date, date, date, date)
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *HolidayRepository) ListActiveByDateRange(startDate, endDate, companyID string) ([]models.Holiday, error) {
	var list []models.Holiday
	query := r.db.Where(`(date BETWEEN ? AND ? OR weekend_date BETWEEN ? AND ? OR (from_date IS NOT NULL AND to_date IS NOT NULL AND from_date <= ? AND to_date >= ?)) AND status = 'active' AND deleted_at IS NULL`,
		startDate, endDate, startDate, endDate, endDate, startDate)
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	err := query.Find(&list).Error
	if err != nil {
		return list, err
	}
	for i := range list {
		normalizeHolidayDates(&list[i])
	}
	return list, nil
}
