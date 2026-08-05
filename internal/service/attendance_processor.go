package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type AttendanceProcessor struct {
	dataLogRepo         *repository.DataLogRepository
	attendanceRepo      *repository.AttendanceRepository
	employeeRepo        *repository.EmployeeRepository
	shiftRepo           *repository.ShiftRepository
	leaveRepo           *repository.LeaveRepository
	tempShiftRepo       *repository.TemporaryShiftRepository
	holidayRepo         *repository.HolidayRepository
	missingAttendanceRepo *repository.MissingAttendanceRepository
}

func NewAttendanceProcessor(
	dataLogRepo *repository.DataLogRepository,
	attendanceRepo *repository.AttendanceRepository,
	employeeRepo *repository.EmployeeRepository,
	shiftRepo *repository.ShiftRepository,
	leaveRepo *repository.LeaveRepository,
	tempShiftRepo *repository.TemporaryShiftRepository,
	holidayRepo *repository.HolidayRepository,
	missingAttendanceRepo *repository.MissingAttendanceRepository,
) *AttendanceProcessor {
	return &AttendanceProcessor{
		dataLogRepo:         dataLogRepo,
		attendanceRepo:      attendanceRepo,
		employeeRepo:        employeeRepo,
		shiftRepo:           shiftRepo,
		leaveRepo:           leaveRepo,
		tempShiftRepo:       tempShiftRepo,
		holidayRepo:         holidayRepo,
		missingAttendanceRepo: missingAttendanceRepo,
	}
}

// DayResult holds per-day processing summary.
type DayResult struct {
	Date      string `json:"date"`
	Processed int    `json:"processed"`
	Logs      int    `json:"logs"`
}

// ProcessDateRangeResult holds the aggregated result of processing a date range.
type ProcessDateRangeResult struct {
	TotalProcessed int         `json:"total_processed"`
	TotalLogs      int         `json:"total_logs"`
	Days           int         `json:"days"`
	Details        []DayResult `json:"details"`
}

// ProcessDateRange converts unprocessed raw punch data into attendance records
// for each day in the given date range. Only active Regular employees are processed.
func (p *AttendanceProcessor) ProcessDateRange(startDate, endDate, companyID string) (*ProcessDateRangeResult, error) {
	dates, err := utils.GenerateDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := &ProcessDateRangeResult{
		Days:    len(dates),
		Details: make([]DayResult, 0, len(dates)),
	}

	// Pre-fetch temporary shifts keyed by business employee_id|date
	tempShiftByKey := make(map[string]*models.TemporaryShift)
	allTempShifts, err := p.tempShiftRepo.ListByCompanyAndDateRange(companyID, startDate, endDate)
	if err == nil {
		for i := range allTempShifts {
			if allTempShifts[i].Status != "" && !strings.EqualFold(allTempShifts[i].Status, "active") {
				continue
			}
			key := allTempShifts[i].EmployeeID + "|" + allTempShifts[i].Date
			tempShiftByKey[key] = &allTempShifts[i]
		}
	}

	shiftCache := make(map[string]*models.Shift)

	for _, date := range dates {
		dayProcessed, dayLogs, dayErr := p.processDay(date, companyID, tempShiftByKey, shiftCache)
		if dayErr != nil {
			return nil, fmt.Errorf("process date %s: %w", date, dayErr)
		}
		result.TotalProcessed += dayProcessed
		result.TotalLogs += dayLogs
		result.Details = append(result.Details, DayResult{
			Date:      date,
			Processed: dayProcessed,
			Logs:      dayLogs,
		})
	}

	return result, nil
}

func (p *AttendanceProcessor) processDay(
	date, companyID string,
	tempShiftByKey map[string]*models.TemporaryShift,
	shiftCache map[string]*models.Shift,
) (processedCount, logCount int, err error) {

	logs, err := p.dataLogRepo.ListByDate(date)
	if err != nil {
		return 0, 0, err
	}
	logCount = len(logs)

	// Badge → employee via punch_number only (active Regular)
	employeeByBadge, err := p.buildEmployeeByBadge(logs)
	if err != nil {
		return 0, logCount, err
	}

	// Only active Regular employees
	activeEmployees, err := p.employeeRepo.ListActiveRegularAll(companyID)
	if err != nil {
		return 0, logCount, err
	}

	eligible := make([]models.Employee, 0, len(activeEmployees))
	allEmployeeIDs := make([]string, 0, len(activeEmployees))
	for i := range activeEmployees {
		emp := &activeEmployees[i]
		if !p.isEligibleForDate(emp, date) {
			continue
		}
		eligible = append(eligible, activeEmployees[i])
		allEmployeeIDs = append(allEmployeeIDs, emp.EmployeeID)
	}

	existingAttByEmp := make(map[string]*models.Attendance)
	if len(allEmployeeIDs) > 0 {
		existingAttList, listErr := p.attendanceRepo.ListByDateAndEmployeeIDs(date, allEmployeeIDs)
		if listErr != nil {
			return 0, logCount, listErr
		}
		for i := range existingAttList {
			existingAttByEmp[existingAttList[i].EmployeeID] = &existingAttList[i]
		}
	}

	// Leave set keyed by business employee_id (varchar)
	onLeaveSet := make(map[string]bool)
	approvedLeaves, leaveErr := p.leaveRepo.ListApprovedByDate(date)
	if leaveErr != nil {
		return 0, logCount, leaveErr
	}
	for _, l := range approvedLeaves {
		onLeaveSet[l.EmployeeID] = true
	}

	// --- HIGHEST PRIORITY: Apply missing attendance overrides ---
	// Records in missing_attendances take absolute precedence over ZKTeco logs.
	// If a missing_attendance record exists for an employee+date, we create/update
	// the attendance record with the user-specified status/check_in/check_out and
	// mark that employee as "handled by missing attendance" so log processing is skipped.
	missingHandledByEmp := make(map[string]bool)
	if p.missingAttendanceRepo != nil {
		missingRecords, maErr := p.missingAttendanceRepo.ListByDateRange(date, date, companyID)
		if maErr == nil {
			for _, ma := range missingRecords {
				missingHandledByEmp[ma.EmployeeID] = true

				// Determine shift for this employee
				var shiftID *string
				for i := range eligible {
					if eligible[i].EmployeeID == ma.EmployeeID {
						if eligible[i].ShiftID != nil {
							shiftID = eligible[i].ShiftID
						}
						tempKey := ma.EmployeeID + "|" + date
						if ts, ok := tempShiftByKey[tempKey]; ok && ts.ShiftID != "" {
							shiftID = &ts.ShiftID
						}
						break
					}
				}

				if existing, exists := existingAttByEmp[ma.EmployeeID]; exists {
					existing.CheckIn = ma.CheckIn
					existing.CheckOut = ma.CheckOut
					existing.Status = ma.Status
					if shiftID != nil {
						existing.ShiftID = shiftID
					}
					if err := p.attendanceRepo.Update(existing); err == nil {
						processedCount++
					}
				} else {
					att := &models.Attendance{
						EmployeeID: ma.EmployeeID,
						CompanyID:  ma.CompanyID,
						Date:       date,
						CheckIn:    ma.CheckIn,
						CheckOut:   ma.CheckOut,
						Status:     ma.Status,
						ShiftID:    shiftID,
					}
					if err := p.attendanceRepo.Create(att); err == nil {
						processedCount++
						existingAttByEmp[ma.EmployeeID] = att
					}
				}
			}
		}
	}

	// Holiday sets
	isGovHoliday := false
	isCompWeekend := false
	isGenDuty := false
	holidays, holErr := p.holidayRepo.ListActiveByDate(date, companyID)
	if holErr == nil {
		for _, h := range holidays {
			hDate := utils.NormalizeDate(h.Date)
			var hFrom, hTo, hWeekend string
			if h.FromDate != nil {
				hFrom = utils.NormalizeDate(*h.FromDate)
			}
			if h.ToDate != nil {
				hTo = utils.NormalizeDate(*h.ToDate)
			}
			if h.WeekendDate != nil {
				hWeekend = utils.NormalizeDate(*h.WeekendDate)
			}
			if h.Type != "weekend_change" && (hDate == date || (hFrom != "" && hTo != "" && date >= hFrom && date <= hTo)) {
				isGovHoliday = true
			}
			if h.Type == "weekend_change" {
				if hDate == date {
					isGenDuty = true
				}
				if hWeekend != "" && hWeekend == date {
					isCompWeekend = true
				}
			}
		}
	}

	// --- Process punch logs ---
	if len(logs) > 0 {
		type empPunches struct {
			BadgeNumber string
			Punches     []models.DataLog
			LogIDs      []string
		}
		grouped := make(map[int]*empPunches)

		for _, log := range logs {
			if grouped[log.UserID] == nil {
				grouped[log.UserID] = &empPunches{
					BadgeNumber: log.BadgeNumber,
				}
			}
			grouped[log.UserID].Punches = append(grouped[log.UserID].Punches, log)
			grouped[log.UserID].LogIDs = append(grouped[log.UserID].LogIDs, log.ID)
		}

		var logIDsToMark []string

		for _, gp := range grouped {
			if len(gp.Punches) == 0 || gp.BadgeNumber == "" {
				continue
			}

			employee, ok := employeeByBadge[gp.BadgeNumber]
			if !ok {
				continue
			}
			// Skip employees already handled by missing attendance (highest priority)
			if missingHandledByEmp[employee.EmployeeID] {
				continue
			}
			if companyID != "" && employee.CompanyID != companyID {
				continue
			}
			if !p.isEligibleForDate(employee, date) {
				continue
			}

			checkIn, checkOut, totalHours, overTime, status, lateMinutes, shiftID := p.resolveAttendanceFromPunches(
				employee, date, gp.Punches, tempShiftByKey, shiftCache, onLeaveSet,
				isGovHoliday, isCompWeekend, isGenDuty,
			)

			var attErr error
			if existing, exists := existingAttByEmp[employee.EmployeeID]; exists {
				existing.CheckIn = checkIn
				existing.CheckOut = checkOut
				existing.TotalHours = totalHours
				existing.OverTime = overTime
				existing.Status = status
				existing.ShiftID = shiftID
				existing.LateMinutes = lateMinutes
				existing.PunchNumber = &gp.BadgeNumber
				attErr = p.attendanceRepo.Update(existing)
				if attErr == nil {
					processedCount++
				}
			} else {
				att := &models.Attendance{
					EmployeeID:  employee.EmployeeID,
					CompanyID:   employee.CompanyID,
					Date:        date,
					CheckIn:     checkIn,
					CheckOut:    checkOut,
					TotalHours:  totalHours,
					OverTime:    overTime,
					Status:      status,
					ShiftID:     shiftID,
					LateMinutes: lateMinutes,
					PunchNumber: &gp.BadgeNumber,
				}
				attErr = p.attendanceRepo.Create(att)
				if attErr == nil {
					processedCount++
					existingAttByEmp[employee.EmployeeID] = att
				}
			}
			if attErr == nil {
				logIDsToMark = append(logIDsToMark, gp.LogIDs...)
			}
		}

		if len(logIDsToMark) > 0 {
			if markErr := p.dataLogRepo.MarkProcessed(logIDsToMark); markErr != nil {
				return processedCount, logCount, markErr
			}
		}
	}

	// --- Absent / weekend / on_leave for eligible Regular employees with no punch record ---
	// Employees who already have check_in/check_out (created by punch processing) are skipped,
	// since the punch section above already set their correct status from logs.
	// Employees handled by missing attendance are also skipped.
	// Employees with existing absent/weekend/on_leave status ARE re-checked in case shift or leave changed.
	for i := range eligible {
		emp := &eligible[i]
		if strings.TrimSpace(emp.PunchNumber) == "" {
			continue
		}
		// Skip employees handled by missing attendance overrides
		if missingHandledByEmp[emp.EmployeeID] {
			continue
		}

		existing, hasExisting := existingAttByEmp[emp.EmployeeID]
		if hasExisting {
			// Skip employees whose attendance was set from actual punch logs (has check_in or check_out)
			if existing.CheckIn != nil || existing.CheckOut != nil {
				continue
			}
		}

		status := "absent"
		var shiftID *string
		var weekShift *models.Shift

		tempKey := emp.EmployeeID + "|" + date
		if ts, ok := tempShiftByKey[tempKey]; ok && ts.ShiftID != "" {
			shiftID = &ts.ShiftID
			weekShift = p.getShift(ts.ShiftID, shiftCache)
		} else if emp.ShiftID != nil {
			shiftID = emp.ShiftID
			weekShift = p.getShift(*emp.ShiftID, shiftCache)
		}

		if isGovHoliday {
			status = "holiday"
		} else if isCompWeekend {
			status = "weekend"
		} else if !isGenDuty && weekShift != nil && weekShift.WeekendDays != "" && utils.IsWeekend(date, weekShift.WeekendDays) {
			status = "weekend"
		}
		if onLeaveSet[emp.EmployeeID] {
			status = "on_leave"
		}

		if hasExisting {
			if existing.Status != status || (existing.ShiftID == nil && shiftID != nil) || (existing.ShiftID != nil && shiftID != nil && *existing.ShiftID != *shiftID) {
				existing.Status = status
				existing.ShiftID = shiftID
				if err := p.attendanceRepo.Update(existing); err == nil {
					processedCount++
				}
			}
		} else {
			att := &models.Attendance{
				EmployeeID: emp.EmployeeID,
				CompanyID:  emp.CompanyID,
				Date:       date,
				Status:     status,
				ShiftID:    shiftID,
			}
			if createErr := p.attendanceRepo.Create(att); createErr == nil {
				processedCount++
				existingAttByEmp[emp.EmployeeID] = att
			}
		}
	}

	return processedCount, logCount, nil
}

// isEligibleForDate: Regular is already filtered by repo; also skip not-yet-joined.
func (p *AttendanceProcessor) isEligibleForDate(emp *models.Employee, date string) bool {
	if emp == nil {
		return false
	}
	if !strings.EqualFold(emp.Status, "active") {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(emp.EmployeeType), "regular") {
		return false
	}
	if strings.TrimSpace(emp.PunchNumber) == "" {
		return false
	}
	if !emp.JoiningDate.IsZero() {
		processDate, err := time.Parse("2006-01-02", date)
		if err == nil {
			joinDay := time.Date(emp.JoiningDate.Year(), emp.JoiningDate.Month(), emp.JoiningDate.Day(), 0, 0, 0, 0, time.UTC)
			if processDate.Before(joinDay) {
				return false
			}
		}
	}
	return true
}

func (p *AttendanceProcessor) buildEmployeeByBadge(logs []models.DataLog) (map[string]*models.Employee, error) {
	badgeSet := make(map[string]bool)
	for _, l := range logs {
		if l.BadgeNumber != "" {
			badgeSet[l.BadgeNumber] = true
		}
	}
	if len(badgeSet) == 0 {
		return map[string]*models.Employee{}, nil
	}

	badgeNumbers := make([]string, 0, len(badgeSet))
	for b := range badgeSet {
		badgeNumbers = append(badgeNumbers, b)
	}

	employeeByBadge := make(map[string]*models.Employee)

	// Primary only: match by punch_number (active Regular)
	byPunch, err := p.employeeRepo.FindActiveRegularByPunchNumbers(badgeNumbers)
	if err != nil {
		return nil, err
	}
	for i := range byPunch {
		if byPunch[i].PunchNumber != "" {
			employeeByBadge[byPunch[i].PunchNumber] = &byPunch[i]
		}
	}

	return employeeByBadge, nil
}

func (p *AttendanceProcessor) resolveAttendanceFromPunches(
	employee *models.Employee,
	date string,
	punches []models.DataLog,
	tempShiftByKey map[string]*models.TemporaryShift,
	shiftCache map[string]*models.Shift,
	onLeaveSet map[string]bool,
	isGovHoliday bool,
	isCompWeekend bool,
	isGenDuty bool,
) (checkIn, checkOut *time.Time, totalHoursStr, overTime *string, status string, lateMinutes int, shiftID *string) {

	status = "present"

	// Resolve shift — key uses business employee_id
	var shift *models.Shift
	tempKey := employee.EmployeeID + "|" + date
	if ts, ok := tempShiftByKey[tempKey]; ok && ts.ShiftID != "" {
		shift = p.getShift(ts.ShiftID, shiftCache)
		if shift != nil {
			shiftID = &shift.ID
		}
	} else if employee.ShiftID != nil {
		shift = p.getShift(*employee.ShiftID, shiftCache)
		if shift != nil {
			shiftID = &shift.ID
		}
	}

	isSpecialDay := false
	if isGovHoliday {
		isSpecialDay = true
		status = "holiday"
	} else if isCompWeekend {
		isSpecialDay = true
		status = "weekend"
	} else if !isGenDuty && shift != nil && shift.WeekendDays != "" && utils.IsWeekend(date, shift.WeekendDays) {
		isSpecialDay = true
		status = "weekend"
	}

	var shiftStartTime, shiftOutTime time.Time
	isOvernight := false
	if shift != nil {
		if shift.StartTime != "" {
			shiftStartTime, _ = time.Parse("15:04", shift.StartTime)
		}
		if shift.EndTime != "" {
			shiftOutTime, _ = time.Parse("15:04", shift.EndTime)
		}
		if !shiftStartTime.IsZero() && !shiftOutTime.IsZero() && shiftStartTime.After(shiftOutTime) {
			isOvernight = true
		}
	}

	// Helper: get minutes-since-midnight from a time
	mins := func(t time.Time) int { return t.Hour()*60 + t.Minute() }

	// Shift times as minutes for comparison
	var shiftStartMins, shiftEndMins int
	hasShiftStart := !shiftStartTime.IsZero()
	hasShiftEnd := !shiftOutTime.IsZero()
	if hasShiftStart {
		shiftStartMins = mins(shiftStartTime)
	}
	if hasShiftEnd {
		shiftEndMins = mins(shiftOutTime)
	}

	// Check-in
	// Priority 1: first 'I' punch (valid side of shift)
	// Priority 2: earliest valid punch (devices often omit I/O type)
	for _, punch := range punches {
		pMins := mins(punch.PunchTime)

		if !isOvernight && hasShiftEnd && pMins >= shiftEndMins {
			continue
		}
		// Overnight: morning punches belong to check-out side
		if isOvernight && hasShiftEnd && pMins < shiftEndMins {
			continue
		}

		if punch.PunchType == "I" || punch.PunchType == "i" {
			checkIn = &punch.PunchTime
			break
		}

		if checkIn == nil {
			checkIn = &punch.PunchTime
		}
	}

	// No shift filter matched: first punch (but not for overnight with only morning punches)
	if checkIn == nil && len(punches) > 0 {
		if isOvernight && hasShiftEnd {
			if mins(punches[0].PunchTime) < shiftEndMins {
				// All punches are morning-side; no evening check-in exists → leave check-in nil
			} else {
				checkIn = &punches[0].PunchTime
			}
		} else if hasShiftEnd && mins(punches[0].PunchTime) >= shiftEndMins {
			// All punches after shift end → leave check-in nil (will be handled as outTime only)
		} else {
			checkIn = &punches[0].PunchTime
		}
	}

	// All punches at/after shift end → only last punch counts as outTime, inTime missing
	if checkIn == nil && len(punches) > 0 && hasShiftEnd {
		lastPunchMins := mins(punches[len(punches)-1].PunchTime)
		if isOvernight {
			if lastPunchMins >= shiftEndMins {
				checkOut = &punches[len(punches)-1].PunchTime
			}
		} else {
			if lastPunchMins >= shiftEndMins {
				checkOut = &punches[len(punches)-1].PunchTime
			}
		}
	}

	// Single punch within 1 hour before shift out (or at shift out) → count as outTime only
	if len(punches) == 1 && checkIn != nil && hasShiftEnd && !isOvernight {
		punchMins := mins(punches[0].PunchTime)
		oneHourBeforeEnd := shiftEndMins - 60
		if punchMins >= oneHourBeforeEnd && punchMins <= shiftEndMins {
			checkOut = checkIn
			checkIn = nil
		}
	}

	// 25-minute check-in window: punches within 25 min of check-in are ignored for outTime
	const checkInWindowMins = 25
	var checkInCutoffMins int
	hasCheckInCutoff := false
	if checkIn != nil {
		checkInCutoffMins = mins(*checkIn) + checkInWindowMins
		hasCheckInCutoff = true
	}

	isValidOutMins := func(pMins int) bool {
		if isOvernight {
			if hasShiftEnd && pMins <= shiftEndMins {
				return true
			}
			if hasShiftStart && pMins > shiftStartMins {
				return true
			}
			return false
		}
		return !hasShiftStart || pMins > shiftStartMins
	}

	// Check-out: last 'O' punch (skip punches within 25 min of check-in)
	for i := len(punches) - 1; i >= 0; i-- {
		pMins := mins(punches[i].PunchTime)
		if hasCheckInCutoff && pMins <= checkInCutoffMins {
			continue
		}
		if (punches[i].PunchType == "O" || punches[i].PunchType == "o") && isValidOutMins(pMins) {
			if isOvernight && hasShiftEnd && pMins > shiftEndMins && (hasShiftStart && pMins > shiftStartMins) {
				continue
			}
			checkOut = &punches[i].PunchTime
			break
		}
	}

	// Fallback check-out: last punch after shift end (day) or last morning punch (overnight)
	if checkOut == nil && hasShiftEnd {
		for i := len(punches) - 1; i >= 0; i-- {
			pMins := mins(punches[i].PunchTime)
			if hasCheckInCutoff && pMins <= checkInCutoffMins {
				continue
			}
			if isOvernight {
				if pMins <= shiftEndMins && isValidOutMins(pMins) {
					checkOut = &punches[i].PunchTime
					break
				}
			} else if pMins >= shiftEndMins && isValidOutMins(pMins) {
				checkOut = &punches[i].PunchTime
				break
			}
		}
	}

	// Final fallback: last punch after shift start (day) or last punch overall if >1
	if checkOut == nil && len(punches) > 1 {
		lastPunch := punches[len(punches)-1]
		lastPunchMins := mins(lastPunch.PunchTime)
		if hasCheckInCutoff && lastPunchMins <= checkInCutoffMins {
			// Skip — within 25 min of check-in
		} else if isValidOutMins(lastPunchMins) {
			if checkIn == nil || !checkIn.Equal(lastPunch.PunchTime) {
				checkOut = &lastPunch.PunchTime
			}
		}
	}

	// Status from punches
	if !isSpecialDay {
		if checkIn == nil && checkOut == nil {
			status = "absent"
		} else if checkIn == nil {
			// Has out only — incomplete arrival
			status = "late"
		}
	}

	// Late minutes + late status when check-in after grace
	if !isSpecialDay && shift != nil && checkIn != nil && hasShiftStart {
		grace := time.Duration(shift.LateGraceMinutes) * time.Minute
		deadlineMins := shiftStartMins + int(grace.Minutes())
		actualInMins := mins(*checkIn)
		if actualInMins > deadlineMins {
			lateMinutes = actualInMins - shiftStartMins
			if status == "present" {
				status = "late"
			}
		}
	}

	// Total hours
	totalHoursStr = calcTotalHours(checkIn, checkOut)

	// Half day: present/late with very short worked hours (< 4h) when both punches exist
	if !isSpecialDay && (status == "present" || status == "late") && totalHoursStr != nil {
		if m, ok := parseHHMMToMinutes(*totalHoursStr); ok && m > 0 && m < 4*60 {
			status = "half_day"
		}
	}

	// Leave override (business employee_id) — keeps punches but marks on_leave
	if !isSpecialDay && onLeaveSet[employee.EmployeeID] {
		status = "on_leave"
	}

	if checkOut != nil {
		if isSpecialDay && employee.OverTimeStatus && totalHoursStr != nil {
			if m, ok := parseHHMMToMinutes(*totalHoursStr); ok {
				otMin := m
				if checkOut.Hour()*60+checkOut.Minute() > 14*60 {
					otMin -= 60
				}
				if otMin < 0 {
					otMin = 0
				}
				s := strconv.Itoa(otMin / 60)
				overTime = &s
			}
		} else if !isSpecialDay && employee.OverTimeStatus && hasShiftEnd {
			checkOutMins := mins(*checkOut)
			var otMinutes int
			if isOvernight {
				if hasShiftStart && checkOutMins < shiftStartMins && checkOutMins > shiftEndMins {
					otMinutes = checkOutMins - shiftEndMins
				}
			} else {
				if hasShiftStart && checkOutMins > shiftEndMins {
					otMinutes = checkOutMins - shiftEndMins
				}
			}
			if otMinutes >= 45 {
				otHours := 1 + (otMinutes-45)/60
				s := strconv.Itoa(otHours)
				overTime = &s
			}
		}
	}

	return checkIn, checkOut, totalHoursStr, overTime, status, lateMinutes, shiftID
}

func calcTotalHours(checkIn, checkOut *time.Time) *string {
	if checkIn == nil || checkOut == nil {
		return nil
	}
	duration := checkOut.Sub(*checkIn)
	if duration < 0 {
		return nil
	}
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	s := fmt.Sprintf("%02d:%02d", hours, minutes)
	return &s
}

func parseHHMMToMinutes(hhmm string) (int, bool) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}

func (p *AttendanceProcessor) getShift(id string, cache map[string]*models.Shift) *models.Shift {
	if s, ok := cache[id]; ok {
		return s
	}
	s, err := p.shiftRepo.FindByID(id)
	if err != nil || s == nil {
		cache[id] = nil
		return nil
	}
	cache[id] = s
	return s
}
