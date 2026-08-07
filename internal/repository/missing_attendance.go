package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type MissingAttendanceRepository struct {
	db *gorm.DB
}

func NewMissingAttendanceRepository(db *gorm.DB) *MissingAttendanceRepository {
	return &MissingAttendanceRepository{db: db}
}

func (r *MissingAttendanceRepository) Create(ma *models.MissingAttendance) error {
	return r.db.Create(ma).Error
}

func (r *MissingAttendanceRepository) Update(ma *models.MissingAttendance) error {
	return r.db.Save(ma).Error
}

func (r *MissingAttendanceRepository) Delete(id string) error {
	return r.db.Delete(&models.MissingAttendance{}, "id = ?", id).Error
}

// UpsertByEmployeeAndDate creates or updates a missing attendance record by employee_id + date.
// Handles soft-deleted records by checking for existing active record first.
func (r *MissingAttendanceRepository) UpsertByEmployeeAndDate(ma *models.MissingAttendance) error {
	var existing models.MissingAttendance
	err := r.db.Where("employee_id = ? AND date = ? AND deleted_at IS NULL", ma.EmployeeID, ma.Date).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(ma).Error
	}
	if err != nil {
		return err
	}
	existing.CompanyID = ma.CompanyID
	existing.CheckIn = ma.CheckIn
	existing.CheckOut = ma.CheckOut
	existing.Status = ma.Status
	existing.Notes = ma.Notes
	existing.CreatedBy = ma.CreatedBy
	return r.db.Save(&existing).Error
}

// BulkUpsert creates or updates multiple missing attendance records.
// Also updates the attendances table for each record so reports reflect changes immediately.
func (r *MissingAttendanceRepository) BulkUpsert(records []models.MissingAttendance, attRepo *AttendanceRepository) error {
	for i := range records {
		ma := &records[i]
		// Upsert into missing_attendances
		var existing models.MissingAttendance
		err := r.db.Where("employee_id = ? AND date = ? AND deleted_at IS NULL", ma.EmployeeID, ma.Date).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if createErr := r.db.Create(ma).Error; createErr != nil {
				continue
			}
		} else if err == nil {
			existing.CompanyID = ma.CompanyID
			existing.CheckIn = ma.CheckIn
			existing.CheckOut = ma.CheckOut
			existing.Status = ma.Status
			existing.Notes = ma.Notes
			existing.CreatedBy = ma.CreatedBy
			r.db.Save(&existing)
		} else {
			continue
		}

		// Also update attendances table immediately
		if attRepo != nil {
			dateStr := ma.Date
			if len(dateStr) > 10 {
				dateStr = dateStr[:10]
			}
			existingAtt, attErr := attRepo.FindByEmployeeAndDate(ma.EmployeeID, dateStr)
			if attErr == nil && existingAtt != nil {
				existingAtt.CheckIn = ma.CheckIn
				existingAtt.CheckOut = ma.CheckOut
				existingAtt.Status = ma.Status
				existingAtt.CalculateHours()
				attRepo.Update(existingAtt)
			} else {
				att := &models.Attendance{
					EmployeeID: ma.EmployeeID,
					CompanyID:  ma.CompanyID,
					Date:       dateStr,
					CheckIn:    ma.CheckIn,
					CheckOut:   ma.CheckOut,
					Status:     ma.Status,
				}
				att.CalculateHours()
				attRepo.Create(att)
			}
		}
	}
	return nil
}

func (r *MissingAttendanceRepository) FindByID(id string) (*models.MissingAttendance, error) {
	var ma models.MissingAttendance
	err := r.db.Where("id = ?", id).
		Preload("Employee").Preload("Company").
		First(&ma).Error
	return &ma, err
}

// FindByEmployeeAndDate returns the missing attendance override for a specific employee+date.
func (r *MissingAttendanceRepository) FindByEmployeeAndDate(employeeID, date string) (*models.MissingAttendance, error) {
	var ma models.MissingAttendance
	err := r.db.Where("employee_id = ? AND date = ?", employeeID, date).First(&ma).Error
	return &ma, err
}

// ListByDateRange returns all missing attendance override records for a date range.
// Used by the daily process to apply overrides with highest priority.
func (r *MissingAttendanceRepository) ListByDateRange(startDate, endDate, companyID string) ([]models.MissingAttendance, error) {
	q := r.db.Where("date BETWEEN ? AND ?", startDate, endDate)
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	var results []models.MissingAttendance
	err := q.Find(&results).Error
	return results, err
}

// List returns paginated missing attendance records with filters.
func (r *MissingAttendanceRepository) List(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID string, page, limit int) ([]map[string]interface{}, int64, error) {
	type row struct {
		models.MissingAttendance
		EmployeeName string `json:"employee_name" gorm:"column:employee_name"`
		EmpNameBn    string `json:"name_bn" gorm:"column:name_bn"`
		Designation  string `json:"designation" gorm:"column:designation_name"`
		ShiftName    string `json:"shift_name" gorm:"column:shift_name"`
	}

	query := r.db.Table("missing_attendances").
		Select(`
			missing_attendances.*,
			employees.name_en as employee_name,
			employees.name_bn,
			COALESCE(designations.name, '') as designation_name,
			COALESCE(shifts.name, '') as shift_name
		`).
		Joins("JOIN employees ON employees.employee_id = missing_attendances.employee_id").
		Joins("LEFT JOIN designations ON designations.id = employees.designation_id").
		Joins("LEFT JOIN shifts ON shifts.id = employees.shift_id").
		Where("missing_attendances.date BETWEEN ? AND ? AND missing_attendances.deleted_at IS NULL", startDate, endDate).
		Where("(missing_attendances.check_in IS NULL AND missing_attendances.check_out IS NOT NULL) OR (missing_attendances.check_in IS NOT NULL AND missing_attendances.check_out IS NULL)")

	if companyID != "" {
		query = query.Where("missing_attendances.company_id = ?", companyID)
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
	if status != "" {
		query = query.Where("missing_attendances.status = ?", status)
	}
	if employeeID != "" {
		query = query.Where("missing_attendances.employee_id ILIKE ?", "%"+employeeID+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []row
	err := query.Order("missing_attendances.date DESC, employees.name_en ASC").
		Offset((page - 1) * limit).Limit(limit).Find(&rows).Error

	results := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		results[i] = map[string]interface{}{
			"id":            r.ID,
			"employee_id":   r.EmployeeID,
			"employee_name": r.EmployeeName,
			"name_bn":       r.EmpNameBn,
			"designation":   r.Designation,
			"shift_name":    r.ShiftName,
			"date":          r.Date,
			"check_in":      r.CheckIn,
			"check_out":     r.CheckOut,
			"status":        r.Status,
			"notes":         r.Notes,
			"company_id":    r.CompanyID,
		}
	}
	return results, total, err
}
