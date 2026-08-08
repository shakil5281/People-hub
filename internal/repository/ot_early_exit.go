package repository

import (
	"math"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"gorm.io/gorm"
)

// OtEarlyExitRepository manages the early-exit shortfall ledger and the
// computation that feeds monthly overtime reductions.
type OtEarlyExitRepository struct {
	db *gorm.DB
}

func NewOtEarlyExitRepository(db *gorm.DB) *OtEarlyExitRepository {
	return &OtEarlyExitRepository{db: db}
}

// ShortfallRow is a single qualifying attendance day with its computed hours.
type ShortfallRow struct {
	EmployeeID     string
	CompanyID      string
	Date           string
	Status         string
	CheckIn        *time.Time
	CheckOut       *time.Time
	ShiftID        *string
	ShiftStartTime string
	ShiftEndTime   string
	WeekendDays    string
	ExpectedHours  float64
	WorkedHours    float64
	ShortfallHours float64
}

// ListShortfallRows scans attendance for a month and returns every day where an
// employee worked less than their expected shift hours. Shift resolution follows
// the attendance processor: temporary shift > attendance shift > employee shift.
// Weekend and on_leave days are excluded (they have no work baseline).
func (r *OtEarlyExitRepository) ListShortfallRows(companyID, startDate, endDate string) ([]ShortfallRow, error) {
	var rows []struct {
		EmployeeID     string     `gorm:"column:employee_id"`
		CompanyID      string     `gorm:"column:company_id"`
		Date           string     `gorm:"column:date"`
		Status         string     `gorm:"column:status"`
		CheckIn        *time.Time `gorm:"column:check_in"`
		CheckOut       *time.Time `gorm:"column:check_out"`
		ShiftID        *string    `gorm:"column:shift_id"`
		ShiftStartTime string     `gorm:"column:shift_start_time"`
		ShiftEndTime   string     `gorm:"column:shift_end_time"`
		WeekendDays    string     `gorm:"column:weekend_days"`
	}

	// Exclude weekend days and on_leave. Weekend detection needs the resolved
	// shift's weekend_days, so we join shifts through temp > attendance > employee.
	err := r.db.Raw(`
		SELECT
			a.employee_id,
			a.company_id,
			a.date,
			a.status,
			a.check_in,
			a.check_out,
			COALESCE(ts.shift_id, a.shift_id, e.shift_id) AS shift_id,
			COALESCE(s.start_time, '') AS shift_start_time,
			COALESCE(s.end_time, '') AS shift_end_time,
			COALESCE(s.weekend_days, '') AS weekend_days
		FROM attendances a
		JOIN employees e ON e.employee_id = a.employee_id AND e.deleted_at IS NULL
		LEFT JOIN temporary_shifts ts ON ts.employee_id = a.employee_id
			AND ts.date = a.date AND ts.deleted_at IS NULL AND ts.status = 'active'
		LEFT JOIN shifts s ON s.id = COALESCE(ts.shift_id, a.shift_id, e.shift_id) AND s.deleted_at IS NULL
		WHERE a.company_id = ?
			AND a.date BETWEEN ? AND ?
			AND a.deleted_at IS NULL
			AND a.check_in IS NOT NULL AND a.check_out IS NOT NULL
			AND a.status <> 'on_leave'
			AND e.over_time_status = true
		ORDER BY a.employee_id, a.date
	`, companyID, startDate, endDate).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]ShortfallRow, 0, len(rows))
	for i := range rows {
		rw := rows[i]
		if rw.CheckIn == nil || rw.CheckOut == nil {
			continue
		}

		worked := utils.CalcNetWorkHours(*rw.CheckIn, *rw.CheckOut)

		// Baseline expected daily working hours = 8 hours
		expectedWork := 8.0
		if worked >= expectedWork {
			continue
		}

		shortfall := expectedWork - worked
		shortfallOT := math.Round(shortfall)
		if shortfallOT <= 0 {
			continue
		}

		result = append(result, ShortfallRow{
			EmployeeID:     rw.EmployeeID,
			CompanyID:      rw.CompanyID,
			Date:           rw.Date,
			Status:         rw.Status,
			CheckIn:        rw.CheckIn,
			CheckOut:       rw.CheckOut,
			ShiftID:        rw.ShiftID,
			ShiftStartTime: rw.ShiftStartTime,
			ShiftEndTime:   rw.ShiftEndTime,
			WeekendDays:    rw.WeekendDays,
			ExpectedHours:  8.0,
			WorkedHours:    math.Round(worked),
			ShortfallHours: shortfallOT,
		})
	}
	return result, nil
}

// expectedShiftHours returns shift duration in hours, handling overnight shifts.
func expectedShiftHours(start, end time.Time) float64 {
	delta := end.Sub(start)
	if delta <= 0 {
		delta += 24 * time.Hour
	}
	return delta.Hours()
}

// calcShortfallOTHours converts short work duration into whole integer OT hours deducted.
// Uses integer floor matching Salary Sheet calculation (e.g. 2.93h -> 2h OT deducted).
func calcShortfallOTHours(diffHours float64) float64 {
	if diffHours < 1.0 {
		return 0
	}
	return float64(int(diffHours))
}

func round0(v float64) float64 {
	return float64(int64(v + 0.5))
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// Upsert replaces the ledger for a given month with the freshly computed rows.
// It runs in one transaction so a re-run always matches the latest attendance.
func (r *OtEarlyExitRepository) UpsertMonth(companyID string, month, year int, rows []ShortfallRow, createdBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ? AND month = ? AND year = ?", companyID, month, year).Delete(&models.OtEarlyExitDeduction{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		toInsert := make([]models.OtEarlyExitDeduction, len(rows))
		for i, r := range rows {
			toInsert[i] = models.OtEarlyExitDeduction{
				CompanyID:      companyID,
				EmployeeID:     r.EmployeeID,
				Month:          month,
				Year:           year,
				Date:           r.Date,
				Status:         r.Status,
				ShiftID:        r.ShiftID,
				ShiftStartTime: r.ShiftStartTime,
				ShiftEndTime:   r.ShiftEndTime,
				ExpectedHours:  r.ExpectedHours,
				WorkedHours:    r.WorkedHours,
				ShortfallHours: r.ShortfallHours,
				CreatedBy:      &createdBy,
			}
		}
		return tx.Create(&toInsert).Error
	})
}

// MonthlyShortfallTotals returns total shortfall hours per employee for a month.
func (r *OtEarlyExitRepository) MonthlyShortfallTotals(companyID string, month, year int) (map[string]float64, error) {
	var records []struct {
		EmployeeID string  `gorm:"column:employee_id"`
		Total      float64 `gorm:"column:total"`
	}
	q := r.db.Table("ot_early_exit_deductions o").
		Select("o.employee_id, SUM(o.shortfall_hours) as total").
		Joins("JOIN employees e ON e.employee_id = o.employee_id").
		Where("o.month = ? AND o.year = ? AND o.deleted_at IS NULL AND e.over_time_status = true", month, year)
	if companyID != "" {
		q = q.Where("o.company_id = ?", companyID)
	}
	err := q.Group("o.employee_id").Scan(&records).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64, len(records))
	for _, rec := range records {
		result[rec.EmployeeID] = rec.Total
	}
	return result, nil
}

// List returns paginated ledger rows with employee details joined in.
func (r *OtEarlyExitRepository) List(companyID string, month, year int, departmentID, sectionID, designationID, lineID, groupID, shiftID, employeeID string, page, limit int) ([]map[string]interface{}, int64, error) {
	type row struct {
		models.OtEarlyExitDeduction
		EmployeeName string `json:"employee_name" gorm:"column:employee_name"`
		Designation  string `json:"designation" gorm:"column:designation_name"`
		Department   string `json:"department" gorm:"column:department_name"`
	}

	query := r.db.Table("ot_early_exit_deductions o").
		Select(`
			o.*,
			employees.name_en as employee_name,
			COALESCE(designations.name, '') as designation_name,
			COALESCE(departments.name, '') as department_name
		`).
		Joins("JOIN employees ON employees.employee_id = o.employee_id").
		Joins("LEFT JOIN designations ON designations.id = employees.designation_id").
		Joins("LEFT JOIN departments ON departments.id = employees.department_id").
		Where("o.month = ? AND o.year = ? AND o.deleted_at IS NULL AND employees.over_time_status = true", month, year)

	if companyID != "" {
		query = query.Where("o.company_id = ?", companyID)
	}

	if departmentID != "" {
		query = query.Where("employees.department_id = ?", departmentID)
	}
	if sectionID != "" {
		query = query.Where("employees.section_id = ?", sectionID)
	}
	if designationID != "" {
		query = query.Where("employees.designation_id = ?", designationID)
	}
	if lineID != "" {
		query = query.Where("employees.line_id = ?", lineID)
	}
	if groupID != "" {
		query = query.Where("employees.group_id = ?", groupID)
	}
	if shiftID != "" {
		query = query.Where("employees.shift_id = ?", shiftID)
	}
	if employeeID != "" {
		query = query.Where("o.employee_id ILIKE ?", "%"+employeeID+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []row
	err := query.Order("o.date DESC, employees.name_en ASC").
		Offset((page - 1) * limit).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	results := make([]map[string]interface{}, len(rows))
	for i, rw := range rows {
		results[i] = map[string]interface{}{
			"id":             rw.ID,
			"employee_id":    rw.EmployeeID,
			"employee_name":  rw.EmployeeName,
			"designation":    rw.Designation,
			"department":     rw.Department,
			"date":           rw.Date,
			"status":         rw.Status,
			"shift_id":       rw.ShiftID,
			"shift_start":    rw.ShiftStartTime,
			"shift_end":      rw.ShiftEndTime,
			"expected_hours": rw.ExpectedHours,
			"worked_hours":   rw.WorkedHours,
			"shortfall_hours": rw.ShortfallHours,
			"company_id":     rw.CompanyID,
		}
	}
	return results, total, nil
}

// ListStats computes total shortfall hours and affected employee count for the given filter parameters.
func (r *OtEarlyExitRepository) ListStats(companyID string, month, year int, departmentID, sectionID, designationID, lineID, groupID, shiftID, employeeID string) (float64, int64, error) {
	type summary struct {
		TotalShortfall float64 `gorm:"column:total_shortfall"`
		Affected       int64   `gorm:"column:affected"`
	}
	query := r.db.Table("ot_early_exit_deductions o").
		Select(`
			COALESCE(SUM(o.shortfall_hours), 0) as total_shortfall,
			COUNT(DISTINCT o.employee_id) as affected
		`).
		Joins("JOIN employees ON employees.employee_id = o.employee_id").
		Where("o.month = ? AND o.year = ? AND o.deleted_at IS NULL AND employees.over_time_status = true", month, year)

	if companyID != "" {
		query = query.Where("o.company_id = ?", companyID)
	}
	if departmentID != "" {
		query = query.Where("employees.department_id = ?", departmentID)
	}
	if sectionID != "" {
		query = query.Where("employees.section_id = ?", sectionID)
	}
	if designationID != "" {
		query = query.Where("employees.designation_id = ?", designationID)
	}
	if lineID != "" {
		query = query.Where("employees.line_id = ?", lineID)
	}
	if groupID != "" {
		query = query.Where("employees.group_id = ?", groupID)
	}
	if shiftID != "" {
		query = query.Where("employees.shift_id = ?", shiftID)
	}
	if employeeID != "" {
		query = query.Where("o.employee_id ILIKE ?", "%"+employeeID+"%")
	}

	var sum summary
	err := query.Scan(&sum).Error
	return round2(sum.TotalShortfall), sum.Affected, err
}
