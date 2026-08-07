package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"gorm.io/gorm"
)

func writeLog(level, source, message string) {
	database.DB.Create(&models.SystemLog{
		Level:    level,
		Source:   source,
		Message:  message,
		CreatedAt: time.Now(),
	})
}

type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) WithTx(tx *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: tx}
}

func (r *AttendanceRepository) Create(attendance *models.Attendance) error {
	return r.db.Create(attendance).Error
}

func (r *AttendanceRepository) CountByDate(date string, filters ...string) (int64, error) {
	base := r.db.Model(&models.Attendance{}).
		Where("date = ? AND deleted_at IS NULL", date)
	// Optional: company_id as first filter
	if len(filters) > 0 && filters[0] != "" {
		base = base.Where("company_id = ?", filters[0])
	}
	var count int64
	err := base.Count(&count).Error
	return count, err
}

func (r *AttendanceRepository) CountRangeByDate(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID string) (int64, error) {
	base := r.db.Model(&models.Attendance{}).
		Where("date BETWEEN ? AND ? AND deleted_at IS NULL", startDate, endDate)
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if employeeID != "" {
		base = base.Where("employee_id = ?", employeeID)
	}
	var count int64
	err := base.Count(&count).Error
	return count, err
}

func (r *AttendanceRepository) CountFilteredByDate(date, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID string) (int64, error) {
	return r.CountFilteredByDateRange(date, date, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID)
}

func (r *AttendanceRepository) CountFilteredByDateRange(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID string) (int64, error) {
	base := r.db.Model(&models.Attendance{}).Where("deleted_at IS NULL")
	if startDate != "" && endDate != "" {
		base = base.Where("date BETWEEN ? AND ?", startDate, endDate)
	} else if startDate != "" {
		base = base.Where("date = ?", startDate)
	}
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if employeeID != "" {
		base = base.Where("employee_id = ?", employeeID)
	}
	var count int64
	err := base.Count(&count).Error
	return count, err
}

func (r *AttendanceRepository) FindByID(id string) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.Preload("Employee").Preload("Shift").Where("id = ? AND deleted_at IS NULL", id).First(&attendance).Error
	return &attendance, err
}

func (r *AttendanceRepository) DeleteBulk(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&models.Attendance{}).Error
}

func (r *AttendanceRepository) ListByDate(date string, page, limit int) ([]models.Attendance, int64, error) {
	base := r.db.Model(&models.Attendance{}).Where("date = ? AND deleted_at IS NULL", date)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var attendances []models.Attendance
	err := base.Preload("Employee.DesignationRef").Preload("Employee").Preload("Shift").Order("LENGTH(employee_id) ASC, employee_id ASC").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

func (r *AttendanceRepository) ListAllByDate(date string) ([]models.Attendance, error) {
	var attendances []models.Attendance
	err := r.db.Preload("Employee.DesignationRef").Preload("Employee").Preload("Shift").Where("date = ? AND deleted_at IS NULL", date).Order("created_at ASC").Find(&attendances).Error
	return attendances, err
}

func (r *AttendanceRepository) FindByEmployeeAndDate(employeeID, date string) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.Where("employee_id = ? AND date = ? AND deleted_at IS NULL", employeeID, date).First(&attendance).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

func (r *AttendanceRepository) ListByDateAndEmployeeIDs(date string, employeeIDs []string) ([]models.Attendance, error) {
	var attendances []models.Attendance
	err := r.db.Where("date = ? AND employee_id IN ? AND deleted_at IS NULL", date, employeeIDs).Find(&attendances).Error
	return attendances, err
}

func (r *AttendanceRepository) ListByDateFiltered(date, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID string, page, limit int) ([]models.Attendance, int64, error) {
	return r.ListByDateRangeFiltered(date, date, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID, page, limit)
}

func (r *AttendanceRepository) ListByDateRangeFiltered(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID string, page, limit int) ([]models.Attendance, int64, error) {
	base := r.db.Model(&models.Attendance{}).Where("deleted_at IS NULL")
	if startDate != "" && endDate != "" {
		base = base.Where("date BETWEEN ? AND ?", startDate, endDate)
	} else if startDate != "" {
		base = base.Where("date = ?", startDate)
	}
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if employeeID != "" {
		base = base.Where("employee_id = ?", employeeID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var attendances []models.Attendance
	err := base.Preload("Employee.DesignationRef").Preload("Employee").Preload("Shift").Order("date DESC, LENGTH(employee_id) ASC, employee_id ASC").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

func (r *AttendanceRepository) Summary(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, statusFilter, employeeID string) ([]map[string]interface{}, error) {
	query := r.db.Table("attendances").
		Select("date, company_id, status, COUNT(*) as count").
		Where("attendances.date BETWEEN ? AND ? AND attendances.deleted_at IS NULL", startDate, endDate)
	if companyID != "" {
		query = query.Where("attendances.company_id = ?", companyID)
	}
	if departmentID != "" {
		query = query.Where("attendances.employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		query = query.Where("attendances.employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		query = query.Where("attendances.employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		query = query.Where("attendances.employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		query = query.Where("attendances.employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		query = query.Where("attendances.shift_id = ?", shiftID)
	}
	if statusFilter != "" {
		query = query.Where("attendances.status = ?", statusFilter)
	}
	if employeeID != "" {
		query = query.Where("attendances.employee_id = ?", employeeID)
	}
	rows, err := query.Group("attendances.date, attendances.company_id, attendances.status").Order("attendances.date ASC").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaryMap := make(map[string]map[string]interface{})
	for rows.Next() {
		var date, companyID, status string
		var count int64
		rows.Scan(&date, &companyID, &status, &count)
		key := date + "|" + companyID
		if _, ok := summaryMap[key]; !ok {
			summaryMap[key] = map[string]interface{}{
				"date":       date,
				"company_id": companyID,
				"present":    int64(0), "late": int64(0), "absent": int64(0), "half_day": int64(0), "total": int64(0),
			}
		}
		entry := summaryMap[key]
		entry[status] = count
		entry["total"] = entry["total"].(int64) + count
	}

	var result []map[string]interface{}
	for _, v := range summaryMap {
		result = append(result, v)
	}
	return result, nil
}

func (r *AttendanceRepository) SummaryByGroup(startDate, endDate, groupBy, companyID, departmentID, sectionID, designationID, lineID, groupFilter, shiftID, statusFilter, employeeID string) ([]map[string]interface{}, error) {
	var entityCol, joinClause, groupCol, orderCol string
	switch groupBy {
	case "department":
		entityCol = "departments.id as entity_id, departments.name"
		joinClause = "LEFT JOIN departments ON departments.id = employees.department_id"
		groupCol = "departments.id, departments.name"
		orderCol = "departments.name"
	case "section":
		entityCol = "sections.id as entity_id, sections.name"
		joinClause = "LEFT JOIN sections ON sections.id = employees.section_id"
		groupCol = "sections.id, sections.name"
		orderCol = "sections.name"
	case "designation":
		entityCol = "designations.id as entity_id, designations.name"
		joinClause = "LEFT JOIN designations ON designations.id = employees.designation_id"
		groupCol = "designations.id, designations.name"
		orderCol = "designations.name"
	case "line":
		entityCol = "lines.id as entity_id, lines.name"
		joinClause = "LEFT JOIN lines ON lines.id = employees.line_id"
		groupCol = "lines.id, lines.name"
		orderCol = "lines.name"
	case "group":
		entityCol = "\"groups\".id as entity_id, \"groups\".name"
		joinClause = "LEFT JOIN \"groups\" ON \"groups\".id = employees.group_id"
		groupCol = "\"groups\".id, \"groups\".name"
		orderCol = "\"groups\".name"
	default:
		return nil, fmt.Errorf("invalid group_by: %s", groupBy)
	}

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "attendances.date BETWEEN ? AND ?", "attendances.deleted_at IS NULL")
	args = append(args, startDate, endDate)

	if companyID != "" {
		conditions = append(conditions, "attendances.company_id = ?")
		args = append(args, companyID)
	}
	if departmentID != "" {
		conditions = append(conditions, "employees.department_id = ?")
		args = append(args, departmentID)
	}
	if sectionID != "" {
		conditions = append(conditions, "employees.section_id = ?")
		args = append(args, sectionID)
	}
	if designationID != "" {
		conditions = append(conditions, "employees.designation_id = ?")
		args = append(args, designationID)
	}
	if lineID != "" {
		conditions = append(conditions, "employees.line_id = ?")
		args = append(args, lineID)
	}
	if groupFilter != "" {
		conditions = append(conditions, "employees.group_id = ?")
		args = append(args, groupFilter)
	}
	if shiftID != "" {
		conditions = append(conditions, "attendances.shift_id = ?")
		args = append(args, shiftID)
	}
	if statusFilter != "" {
		conditions = append(conditions, "attendances.status = ?")
		args = append(args, statusFilter)
	}
	if employeeID != "" {
		conditions = append(conditions, "attendances.employee_id = ?")
		args = append(args, employeeID)
	}

	whereClause := ""
	for i, c := range conditions {
		if i == 0 {
			whereClause = "WHERE " + c
		} else {
			whereClause += " AND " + c
		}
	}

	sql := fmt.Sprintf(`
		SELECT %s,
			SUM(CASE WHEN attendances.status = 'present' THEN 1 ELSE 0 END) as present,
			SUM(CASE WHEN attendances.status = 'late' THEN 1 ELSE 0 END) as late,
			SUM(CASE WHEN attendances.status = 'absent' THEN 1 ELSE 0 END) as absent,
			SUM(CASE WHEN attendances.status = 'half_day' THEN 1 ELSE 0 END) as half_day,
			SUM(CASE WHEN attendances.status = 'on_leave' THEN 1 ELSE 0 END) as on_leave,
			SUM(CASE WHEN attendances.status = 'weekend' THEN 1 ELSE 0 END) as weekend,
			COUNT(*) as total
		FROM attendances
		JOIN employees ON employees.employee_id = attendances.employee_id
		%s
		%s
		GROUP BY %s
		ORDER BY %s ASC`,
		entityCol, joinClause, whereClause, groupCol, orderCol)

	rows, err := r.db.Raw(sql, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		rows.Scan(valPtrs...)
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = vals[i]
		}
		results = append(results, row)
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	if len(results) == 0 {
		// Diagnostic: count matching employees without grouping
		diagSQL := fmt.Sprintf(`SELECT COUNT(DISTINCT employees.employee_id) FROM attendances JOIN employees ON employees.employee_id = attendances.employee_id %s %s`, joinClause, whereClause)
		var empCount int64
		if diagErr := r.db.Raw(diagSQL, args...).Scan(&empCount).Error; diagErr == nil {
			writeLog("debug", "attendance", fmt.Sprintf("[SummaryByGroup] group_by=%s matched_employees=%d", groupBy, empCount))
		}
		// Extra diagnostic: show distinct designation IDs present when filtering by designation
		if designationID != "" {
			var actualIDs []string
			idSQL := fmt.Sprintf(`SELECT DISTINCT employees.designation_id FROM attendances JOIN employees ON employees.employee_id = attendances.employee_id %s %s`, joinClause, whereClause)
			if idErr := r.db.Raw(idSQL, args...).Pluck("designation_id", &actualIDs).Error; idErr == nil {
				writeLog("debug", "attendance", fmt.Sprintf("[SummaryByGroup] requested_designation_id=%s actual_designation_ids_in_data=%v", designationID, actualIDs))
			}
		}
	}
	return results, nil
}

func (r *AttendanceRepository) Overtime(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, statusFilter string) ([]map[string]interface{}, error) {
	query := r.db.Table("attendances").
		Select("attendances.id, attendances.employee_id, attendances.date, attendances.check_in, attendances.check_out, attendances.over_time, employees.name_en as employee_name, employees.employee_id as emp_id, employees.over_time_status").
		Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
		Where("attendances.date BETWEEN ? AND ? AND attendances.deleted_at IS NULL AND attendances.check_in IS NOT NULL AND attendances.check_out IS NOT NULL", startDate, endDate)
	query = query.Where("employees.over_time_status = ?", true)
	query = query.Where("attendances.over_time IS NOT NULL AND attendances.over_time != '' AND CAST(attendances.over_time AS numeric) > 0")
	if companyID != "" {
		query = query.Where("attendances.company_id = ?", companyID)
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
		query = query.Where("attendances.shift_id = ?", shiftID)
	}
	if statusFilter != "" {
		query = query.Where("attendances.status = ?", statusFilter)
	}
	var results []map[string]interface{}
	err := query.Order("attendances.date ASC").Find(&results).Error
	return results, err
}

func (r *AttendanceRepository) OvertimeSummary(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, statusFilter string) ([]map[string]interface{}, error) {
	query := r.db.Table("attendances").
		Select("employees.department_id, departments.name as department_name, COUNT(DISTINCT attendances.employee_id) as employee_count").
		Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
		Joins("JOIN departments ON departments.id = employees.department_id").
		Where("attendances.date BETWEEN ? AND ? AND attendances.deleted_at IS NULL AND attendances.check_in IS NOT NULL AND attendances.check_out IS NOT NULL", startDate, endDate)
	if companyID != "" {
		query = query.Where("attendances.company_id = ?", companyID)
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
		query = query.Where("attendances.shift_id = ?", shiftID)
	}
	if statusFilter != "" {
		query = query.Where("attendances.status = ?", statusFilter)
	}
	var results []map[string]interface{}
	err := query.Group("employees.department_id, departments.name").Find(&results).Error
	return results, err
}

func (r *AttendanceRepository) OvertimeSummaryGrouped(startDate, endDate, companyID, groupBy string, departmentID, sectionID, designationID, lineID string) ([]map[string]interface{}, error) {
	var selectCols, joinClause, groupClause string
	switch groupBy {
	case "section":
		selectCols = "employees.section_id as id, sections.name as name, COUNT(DISTINCT attendances.employee_id) as employee_count, COALESCE(SUM(CAST(attendances.over_time AS numeric)), 0) as total_hours"
		joinClause = "JOIN sections ON sections.id = employees.section_id AND sections.deleted_at IS NULL"
		groupClause = "employees.section_id, sections.name"
	case "designation":
		selectCols = "employees.designation_id as id, designations.name as name, COUNT(DISTINCT attendances.employee_id) as employee_count, COALESCE(SUM(CAST(attendances.over_time AS numeric)), 0) as total_hours"
		joinClause = "JOIN designations ON designations.id = employees.designation_id AND designations.deleted_at IS NULL"
		groupClause = "employees.designation_id, designations.name"
	case "line":
		selectCols = "employees.line_id as id, lines.name as name, COUNT(DISTINCT attendances.employee_id) as employee_count, COALESCE(SUM(CAST(attendances.over_time AS numeric)), 0) as total_hours"
		joinClause = "JOIN lines ON lines.id = employees.line_id AND lines.deleted_at IS NULL"
		groupClause = "employees.line_id, lines.name"
	default:
		selectCols = "employees.department_id as id, departments.name as name, COUNT(DISTINCT attendances.employee_id) as employee_count, COALESCE(SUM(CAST(attendances.over_time AS numeric)), 0) as total_hours"
		joinClause = "JOIN departments ON departments.id = employees.department_id AND departments.deleted_at IS NULL"
		groupClause = "employees.department_id, departments.name"
	}

	query := r.db.Table("attendances").
		Select(selectCols).
		Joins("JOIN employees ON employees.employee_id = attendances.employee_id AND employees.deleted_at IS NULL").
		Joins(joinClause).
		Where("attendances.date BETWEEN ? AND ? AND attendances.deleted_at IS NULL AND attendances.check_in IS NOT NULL AND attendances.check_out IS NOT NULL", startDate, endDate)
	query = query.Where("employees.over_time_status = ?", true)
	query = query.Where("attendances.over_time IS NOT NULL AND attendances.over_time != '' AND CAST(attendances.over_time AS numeric) > 0")
	if companyID != "" {
		query = query.Where("attendances.company_id = ?", companyID)
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
	var results []map[string]interface{}
	err := query.Group(groupClause).Order("name ASC").Find(&results).Error
	return results, err
}

func (r *AttendanceRepository) Update(attendance *models.Attendance) error {
	return r.db.Save(attendance).Error
}

// BulkUpdateMissing updates many missing attendance records inside one
// PostgreSQL transaction. The original attendance date of each record is read
// from the database and combined with the entered In/Out times. Only the
// missing fields are updated; existing values are preserved. Alongside the
// attendances table, each record is upserted into missing_attendances so the
// daily process picks it up with highest priority. If any record fails
// validation or update, the whole transaction rolls back.
func (r *AttendanceRepository) BulkUpdateMissing(attendanceIDs []string, inTime, outTime, status, updatedBy string) (int, error) {
	if len(attendanceIDs) == 0 {
		return 0, nil
	}

	updated := 0
	now := time.Now()

	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range attendanceIDs {
			var att models.Attendance
			if err := tx.Where("id = ? AND deleted_at IS NULL", id).First(&att).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("attendance record %s not found", id)
				}
				return err
			}

			if inTime != "" {
				if att.CheckIn != nil {
					return fmt.Errorf("in time already exists for attendance %s", id)
				}
				t, err := utils.ParseDateTime(inTime, utils.NormalizeDate(att.Date))
				if err != nil {
					return fmt.Errorf("invalid in time %q for attendance %s: %v", inTime, id, err)
				}
				att.CheckIn = &t
			}

			if outTime != "" {
				if att.CheckOut != nil {
					return fmt.Errorf("out time already exists for attendance %s", id)
				}
				t, err := utils.ParseDateTime(outTime, utils.NormalizeDate(att.Date))
				if err != nil {
					return fmt.Errorf("invalid out time %q for attendance %s: %v", outTime, id, err)
				}
				att.CheckOut = &t
			}

			if status != "" {
				att.Status = status
			}
			att.UpdatedAt = now
			if updatedBy != "" {
				att.UpdatedBy = &updatedBy
			}

			if err := tx.Save(&att).Error; err != nil {
				return err
			}

			// Upsert the corresponding missing_attendance override so the daily
			// process applies it with highest priority (employee_id + date).
			if err := upsertMissingAttendance(tx, &att, updatedBy); err != nil {
				return err
			}

			updated++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return updated, nil
}

// upsertMissingAttendance creates or updates a missing_attendances record for the
// given attendance, keyed by employee_id + date. Runs inside the caller's transaction.
func upsertMissingAttendance(tx *gorm.DB, att *models.Attendance, by string) error {
	var ma models.MissingAttendance
	err := tx.Where("employee_id = ? AND date = ? AND deleted_at IS NULL", att.EmployeeID, att.Date).First(&ma).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	isNew := errors.Is(err, gorm.ErrRecordNotFound)

	ma.EmployeeID = att.EmployeeID
	ma.CompanyID = att.CompanyID
	ma.Date = att.Date
	ma.CheckIn = att.CheckIn
	ma.CheckOut = att.CheckOut
	ma.Status = att.Status
	if by != "" {
		ma.CreatedBy = &by
	}

	if isNew {
		return tx.Create(&ma).Error
	}
	return tx.Save(&ma).Error
}

func (r *AttendanceRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Attendance{}).Error
}

func (r *AttendanceRepository) DeleteAll() error {
	return r.db.Unscoped().Where("1 = 1").Delete(&models.Attendance{}).Error
}

func (r *AttendanceRepository) UpdateStatusByEmployeeAndDateRange(employeeID, fromDate, toDate, status string) error {
	return r.db.Model(&models.Attendance{}).
		Where("employee_id = ? AND date >= ? AND date <= ? AND deleted_at IS NULL", employeeID, fromDate, toDate).
		Update("status", status).Error
}

func (r *AttendanceRepository) ClearOnLeaveStatus(employeeID, fromDate, toDate string) error {
	return r.db.Model(&models.Attendance{}).
		Where("employee_id = ? AND date >= ? AND date <= ? AND status = 'on_leave' AND deleted_at IS NULL", employeeID, fromDate, toDate).
		Update("status", "").Error
}

func (r *AttendanceRepository) CountByDateOnly(date string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Attendance{}).Where("date = ? AND deleted_at IS NULL", date).Count(&count).Error
	return count, err
}

func (r *AttendanceRepository) ListJobCard(startDate, endDate, companyID, employeeID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status string, page, limit int) ([]models.Attendance, int64, error) {
	base := r.db.Model(&models.Attendance{}).Where("date BETWEEN ? AND ? AND deleted_at IS NULL", startDate, endDate)
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if employeeID != "" {
		base = base.Where("employee_id = ?", employeeID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var attendances []models.Attendance
	err := base.Preload("Employee.DesignationRef").Preload("Employee.Department").Preload("Employee.Company").Preload("Shift").Order("date ASC, LENGTH(employee_id) ASC, employee_id ASC").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

func (r *AttendanceRepository) ListJobCardEmployees(startDate, endDate, companyID, employeeID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status string) ([]models.Employee, error) {
	subQuery := r.db.Table("attendances a").
		Select("DISTINCT a.employee_id").
		Where("a.date BETWEEN ? AND ? AND a.deleted_at IS NULL", startDate, endDate)

	if companyID != "" {
		subQuery = subQuery.Where("a.company_id = ?", companyID)
	}
	if employeeID != "" {
		subQuery = subQuery.Where("a.employee_id = ?", employeeID)
	}
	if departmentID != "" {
		subQuery = subQuery.Where("a.employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		subQuery = subQuery.Where("a.employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		subQuery = subQuery.Where("a.employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		subQuery = subQuery.Where("a.employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		subQuery = subQuery.Where("a.employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		subQuery = subQuery.Where("a.shift_id = ?", shiftID)
	}
	if status != "" {
		subQuery = subQuery.Where("a.status = ?", status)
	}

	var employees []models.Employee
	err := r.db.Where("employee_id IN (?)", subQuery).
		Preload("DesignationRef").Preload("Department").Preload("Company").
		Order("LENGTH(employee_id) ASC, employee_id ASC").
		Find(&employees).Error
	return employees, err
}

func (r *AttendanceRepository) MonthlyReport(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, employeeID string) ([]map[string]interface{}, error) {
	// Compute calendar month days from the date range so total_days reflects the
	// actual number of days in the period rather than COUNT(*), which is inflated
	// when duplicate attendance rows exist for the same employee+date.
	startParsed, _ := time.Parse("2006-01-02", startDate)
	endParsed, _ := time.Parse("2006-01-02", endDate)
	calendarDays := int(endParsed.Sub(startParsed).Hours()/24) + 1

	// Inject the integer literal directly (derived from date math, not user input)
	// to avoid binding an int through pgx text-format encoding.
	totalDaysLiteral := fmt.Sprintf("%d", calendarDays)

	query := r.db.Table("attendances").
		Select(`
			attendances.employee_id,
			employees.employee_id as emp_id,
			employees.name_en as employee_name,
			COALESCE(designations.name, '') as designation_name,
			COALESCE(departments.name, '') as department_name,
			COALESCE(shifts.name, '') as shift_name,
			COALESCE(SUM(CASE WHEN attendances.status = 'present' THEN 1 ELSE 0 END), 0) as present,
			COALESCE(SUM(CASE WHEN attendances.status = 'absent' THEN 1 ELSE 0 END), 0) as absent,
			COALESCE(SUM(CASE WHEN attendances.status = 'late' THEN 1 ELSE 0 END), 0) as late,
			COALESCE(SUM(CASE WHEN attendances.status = 'on_leave' THEN 1 ELSE 0 END), 0) as leave,
			COALESCE(SUM(CASE WHEN attendances.status = 'weekend' THEN 1 ELSE 0 END), 0) as weekend,
			COALESCE(SUM(CASE WHEN attendances.status = 'half_day' THEN 1 ELSE 0 END), 0) as half_day,
			COALESCE(SUM(CASE WHEN attendances.status = 'holiday' THEN 1 ELSE 0 END), 0) as holiday,
			COALESCE(SUM(CAST(NULLIF(attendances.over_time, '') AS INTEGER)), 0) as over_time,
			`+totalDaysLiteral+` as total_days
		`).
		Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
		Joins("LEFT JOIN departments ON departments.id = employees.department_id").
		Joins("LEFT JOIN designations ON designations.id = employees.designation_id").
		Joins("LEFT JOIN shifts ON shifts.id = employees.shift_id").
		Where("attendances.date BETWEEN ? AND ? AND attendances.deleted_at IS NULL", startDate, endDate)
	if companyID != "" {
		query = query.Where("attendances.company_id = ?", companyID)
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
		query = query.Where("employees.employee_id LIKE ?", "%"+employeeID+"%")
	}
	var results []map[string]interface{}
	err := query.Group("attendances.employee_id, employees.employee_id, employees.name_en, designations.name, departments.name, shifts.name").
		Order("employees.name_en ASC").
		Find(&results).Error
	return results, err
}

func (r *AttendanceRepository) ListByStatus(startDate, endDate, status, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, employeeID string, page, limit int) ([]models.Attendance, int64, error) {
	base := r.db.Model(&models.Attendance{}).Where("date BETWEEN ? AND ? AND status = ? AND deleted_at IS NULL", startDate, endDate, status)
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if employeeID != "" {
		base = base.Where("employee_id = ?", employeeID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var attendances []models.Attendance
	err := base.Preload("Employee").Order("date ASC, created_at ASC").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

// CountAbsencesByEmployee returns the number of absent records per employee in a date range
// that also match the same org/status filters as the absent report rows.
func (r *AttendanceRepository) CountAbsencesByEmployee(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, employeeID string) (map[string]int64, error) {
	base := r.db.Model(&models.Attendance{}).
		Select("employee_id, COUNT(*) AS absent_count").
		Where("date BETWEEN ? AND ? AND status = 'absent' AND deleted_at IS NULL", startDate, endDate).
		Group("employee_id")
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if employeeID != "" {
		base = base.Where("employee_id = ?", employeeID)
	}

	var counts []struct {
		EmployeeID  string `gorm:"column:employee_id"`
		AbsentCount int64  `gorm:"column:absent_count"`
	}
	if err := base.Scan(&counts).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(counts))
	for _, c := range counts {
		result[c.EmployeeID] = c.AbsentCount
	}
	return result, nil
}

func (r *AttendanceRepository) GetMonthlyOvertimeHours(companyID, startDate, endDate string) (map[string]float64, error) {
	var records []struct {
		EmployeeID string  `gorm:"column:employee_id"`
		OtHours    float64 `gorm:"column:overtime_hours"`
	}
	err := r.db.Table("attendances a").
		Select(`
			a.employee_id,
			FLOOR(COALESCE(SUM(
				CASE
					WHEN e.over_time_status = false THEN 0
					WHEN a.check_out IS NOT NULL AND s.end_time IS NOT NULL THEN
						CASE
					WHEN s.start_time < s.end_time AND a.check_out::time > s.end_time::time
							THEN EXTRACT(EPOCH FROM (a.check_out::time - s.end_time::time)) / 3600
						WHEN s.start_time > s.end_time AND a.check_out::time < s.start_time::time AND a.check_out::time > s.end_time::time
								THEN EXTRACT(EPOCH FROM (a.check_out::time - s.end_time::time)) / 3600
							ELSE 0
						END
					ELSE 0
				END
			), 0)) as overtime_hours
		`).
		Joins("JOIN employees e ON e.employee_id = a.employee_id").
		Joins("LEFT JOIN shifts s ON s.id = COALESCE(a.shift_id, e.shift_id)").
		Where("a.company_id = ? AND a.date BETWEEN ? AND ? AND a.deleted_at IS NULL AND a.check_in IS NOT NULL AND a.check_out IS NOT NULL",
			companyID, startDate, endDate).
		Group("a.employee_id").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64)
	for _, r := range records {
		result[r.EmployeeID] = r.OtHours
	}
	return result, nil
}

// ListMissing finds attendance records where check_in OR check_out is null within a date range.
func (r *AttendanceRepository) ListMissing(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status string, page, limit int) ([]models.Attendance, int64, error) {
	base := r.db.Model(&models.Attendance{}).
		Where("(check_in IS NULL AND check_out IS NOT NULL OR check_in IS NOT NULL AND check_out IS NULL) AND date BETWEEN ? AND ? AND deleted_at IS NULL", startDate, endDate)
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var attendances []models.Attendance
	err := base.Preload("Employee.DesignationRef").Preload("Employee").Preload("Shift").Order("date ASC, created_at ASC").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

// DeleteAfterDate soft-deletes attendance records after the given date for an employee.
func (r *AttendanceRepository) DeleteAfterDate(employeeID, date string) (int64, error) {
	res := r.db.Where("employee_id = ? AND date > ? AND deleted_at IS NULL", employeeID, date).Delete(&models.Attendance{})
	return res.RowsAffected, res.Error
}

// ListCustom lists all attendance records in a date range (any status) with flexible filters.
func (r *AttendanceRepository) ListCustom(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID string, page, limit int) ([]models.Attendance, int64, error) {
	base := r.db.Model(&models.Attendance{}).
		Where("date BETWEEN ? AND ? AND deleted_at IS NULL", startDate, endDate)
	if companyID != "" {
		base = base.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE department_id = ?)", departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE section_id = ?)", sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE designation_id = ?)", designationID)
	}
	if lineID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE line_id = ?)", lineID)
	}
	if groupID != "" {
		base = base.Where("employee_id IN (SELECT employee_id FROM employees WHERE group_id = ?)", groupID)
	}
	if shiftID != "" {
		base = base.Where("shift_id = ?", shiftID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if employeeID != "" {
		base = base.Where("employee_id ILIKE ?", "%"+employeeID+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var attendances []models.Attendance
	err := base.Preload("Employee.DesignationRef").Preload("Employee").Preload("Shift").Order("date ASC, created_at ASC").Offset((page - 1) * limit).Limit(limit).Find(&attendances).Error
	return attendances, total, err
}

// CustomSummarySection queries attendance summary for a single report section with flexible filters.
// Filters are applied as AND conditions on employee org relations.
type CustomSectionFilter struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"` // section_group, section_line, department, etc.
	SectionNames     []string `json:"section_names"`
	DepartmentNames  []string `json:"department_names"`
	GroupNames       []string `json:"group_names"`
	DesignationNames []string `json:"designation_names"`
	LineNames        []string `json:"line_names"`
	GroupByLine      bool     `json:"group_by_line"`
}

func (r *AttendanceRepository) CustomSummarySection(companyID, startDate, endDate string, filter CustomSectionFilter) ([]map[string]interface{}, error) {
	selectCols := "SUM(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END) as present, SUM(CASE WHEN a.status = 'absent' THEN 1 ELSE 0 END) as absent, SUM(CASE WHEN a.status = 'on_leave' THEN 1 ELSE 0 END) as on_leave, SUM(CASE WHEN a.status IN ('late','half_day','weekend') THEN 1 ELSE 0 END) as others, COUNT(*) as total"
	joinClause := "JOIN employees e ON e.employee_id = a.employee_id"

	if filter.GroupByLine || len(filter.LineNames) > 0 {
		joinClause = "JOIN employees e ON e.employee_id = a.employee_id LEFT JOIN lines l ON l.id = e.line_id"
	}
	groupByLine := filter.GroupByLine

	query := r.db.Table("attendances a").
		Select(selectCols).
		Joins(joinClause).
		Where("a.company_id = ? AND a.date BETWEEN ? AND ? AND a.deleted_at IS NULL", companyID, startDate, endDate)

	if len(filter.SectionNames) > 0 {
		query = query.Where("e.section_id IN (SELECT id FROM sections WHERE name IN ?)", filter.SectionNames)
	}
	if len(filter.DepartmentNames) > 0 {
		query = query.Where("e.department_id IN (SELECT id FROM departments WHERE name IN ?)", filter.DepartmentNames)
	}
	if len(filter.GroupNames) > 0 {
		query = query.Where("e.group_id IN (SELECT id FROM \"groups\" WHERE name IN ?)", filter.GroupNames)
	}
	if len(filter.DesignationNames) > 0 {
		query = query.Where("e.designation_id IN (SELECT id FROM designations WHERE name IN ?)", filter.DesignationNames)
	}
	if len(filter.LineNames) > 0 {
		query = query.Where("l.name IN ?", filter.LineNames)
	}

	if groupByLine {
		query = query.Select("COALESCE(l.name, 'No Line') as name, " + selectCols).Group("l.name").Order("l.name ASC")
	}

	var results []map[string]interface{}
	err := query.Find(&results).Error

	if err == nil && !groupByLine {
		results = prefixName(results)
	}
	return results, err
}

func prefixName(results []map[string]interface{}) []map[string]interface{} {
	for _, r := range results {
		r["name"] = "Total"
	}
	return results
}
