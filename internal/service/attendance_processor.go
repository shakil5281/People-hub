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

// AttendanceProcessor converts raw biometric punch data into structured attendance records.
type AttendanceProcessor struct {
	dataLogRepo           *repository.DataLogRepository
	attendanceRepo        *repository.AttendanceRepository
	employeeRepo          *repository.EmployeeRepository
	shiftRepo             *repository.ShiftRepository
	leaveRepo             *repository.LeaveRepository
	tempShiftRepo         *repository.TemporaryShiftRepository
	holidayRepo           *repository.HolidayRepository
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
		dataLogRepo:           dataLogRepo,
		attendanceRepo:        attendanceRepo,
		employeeRepo:          employeeRepo,
		shiftRepo:             shiftRepo,
		leaveRepo:             leaveRepo,
		tempShiftRepo:         tempShiftRepo,
		holidayRepo:           holidayRepo,
		missingAttendanceRepo: missingAttendanceRepo,
	}
}

// ─── Result types ─────────────────────────────────────────────────────────────

// DayResult holds per-day processing summary.
type DayResult struct {
	Date    string `json:"date"`
	Created int    `json:"created"`
	Updated int    `json:"updated"`
	Skipped int    `json:"skipped"`
	Logs    int    `json:"logs"`
}

// ProcessDateRangeResult holds the aggregated result of processing a date range.
type ProcessDateRangeResult struct {
	TotalProcessed int         `json:"total_processed"`
	TotalCreated   int         `json:"total_created"`
	TotalUpdated   int         `json:"total_updated"`
	TotalSkipped   int         `json:"total_skipped"`
	TotalLogs      int         `json:"total_logs"`
	Days           int         `json:"days"`
	Details        []DayResult `json:"details"`
}

// ─── Main entry point ─────────────────────────────────────────────────────────

// ProcessDateRange converts raw punch data into attendance records for every day
// in [startDate, endDate]. Only active Regular employees are processed.
// Running this function multiple times is idempotent: existing records are
// updated, not duplicated.
func (p *AttendanceProcessor) ProcessDateRange(startDate, endDate, companyID string) (*ProcessDateRangeResult, error) {
	dates, err := utils.GenerateDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := &ProcessDateRangeResult{
		Days:    len(dates),
		Details: make([]DayResult, 0, len(dates)),
	}

	// Pre-fetch temporary shifts for the whole range once.
	tempShiftByKey := make(map[string]*models.TemporaryShift)
	allTempShifts, tempErr := p.tempShiftRepo.ListByCompanyAndDateRange(companyID, startDate, endDate)
	if tempErr == nil {
		for i := range allTempShifts {
			ts := &allTempShifts[i]
			if ts.Status != "" && !strings.EqualFold(ts.Status, "active") {
				continue
			}
			key := ts.EmployeeID + "|" + ts.Date
			tempShiftByKey[key] = ts
		}
	}

	shiftCache := make(map[string]*models.Shift)

	for _, date := range dates {
		dr, dayErr := p.processDay(date, companyID, tempShiftByKey, shiftCache)
		if dayErr != nil {
			return nil, fmt.Errorf("process date %s: %w", date, dayErr)
		}
		result.TotalCreated += dr.Created
		result.TotalUpdated += dr.Updated
		result.TotalSkipped += dr.Skipped
		result.TotalLogs += dr.Logs
		result.TotalProcessed += dr.Created + dr.Updated
		result.Details = append(result.Details, dr)
	}

	return result, nil
}

// ─── Per-day processing ───────────────────────────────────────────────────────

func (p *AttendanceProcessor) processDay(
	date, companyID string,
	tempShiftByKey map[string]*models.TemporaryShift,
	shiftCache map[string]*models.Shift,
) (DayResult, error) {

	dr := DayResult{Date: date}

	attendanceDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return dr, fmt.Errorf("parse attendance date: %w", err)
	}

	// ── Load eligible employees ───────────────────────────────────────────────
	activeEmployees, err := p.employeeRepo.ListActiveRegularAll(companyID)
	if err != nil {
		return dr, err
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

	if len(eligible) == 0 {
		return dr, nil
	}

	// ── Load existing attendance records for upsert ───────────────────────────
	existingAttByEmp := make(map[string]*models.Attendance)
	if len(allEmployeeIDs) > 0 {
		existing, listErr := p.attendanceRepo.ListByDateAndEmployeeIDs(date, allEmployeeIDs)
		if listErr != nil {
			return dr, listErr
		}
		for i := range existing {
			existingAttByEmp[existing[i].EmployeeID] = &existing[i]
		}
	}

	// ── Approved leaves ───────────────────────────────────────────────────────
	onLeaveSet := make(map[string]bool)
	approvedLeaves, leaveErr := p.leaveRepo.ListApprovedByDate(date)
	if leaveErr == nil {
		for _, l := range approvedLeaves {
			onLeaveSet[l.EmployeeID] = true
		}
	}

	// ── Holiday / weekend-change flags ────────────────────────────────────────
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

	// ── HIGHEST PRIORITY: Missing Attendance Overrides ────────────────────────
	// If a missing_attendance record exists for an employee on this date, use
	// it directly and skip biometric punch processing for that employee.
	missingHandledByEmp := make(map[string]bool)
	if p.missingAttendanceRepo != nil {
		missingRecords, maErr := p.missingAttendanceRepo.ListByDateRange(date, date, companyID)
		if maErr == nil {
			for _, ma := range missingRecords {
				missingHandledByEmp[ma.EmployeeID] = true

				// Find the shift for this employee on this date.
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

				// ma.CheckIn/CheckOut are *time.Time — use them directly.
				checkIn := ma.CheckIn
				checkOut := ma.CheckOut
				totalHours := utils.CalcTotalHoursStr(checkIn, checkOut)
				zeroOT := "0"

				if existing, exists := existingAttByEmp[ma.EmployeeID]; exists {
					existing.CheckIn = checkIn
					existing.CheckOut = checkOut
					existing.TotalHours = totalHours
					existing.Status = ma.Status
					existing.LateMinutes = 0
					existing.OverTime = &zeroOT
					if shiftID != nil {
						existing.ShiftID = shiftID
					}
					if err := p.attendanceRepo.UpdateFields(existing.ID, map[string]interface{}{
						"check_in":     checkIn,
						"check_out":    checkOut,
						"total_hours":  totalHours,
						"over_time":    zeroOT,
						"status":       ma.Status,
						"late_minutes": 0,
						"shift_id":     shiftID,
					}); err == nil {
						dr.Updated++
					}
				} else {
					att := &models.Attendance{
						EmployeeID: ma.EmployeeID,
						CompanyID:  ma.CompanyID,
						Date:       date,
						CheckIn:    checkIn,
						CheckOut:   checkOut,
						TotalHours: totalHours,
						OverTime:   &zeroOT,
						Status:     ma.Status,
						ShiftID:    shiftID,
					}
					if err := p.attendanceRepo.Create(att); err == nil {
						dr.Created++
						existingAttByEmp[ma.EmployeeID] = att
					}
				}
			}
		}
	}

	// ── Collect all badge numbers for a batch punch query ────────────────────
	badgeNumbers := make([]string, 0, len(eligible))
	for i := range eligible {
		if eligible[i].PunchNumber != "" {
			badgeNumbers = append(badgeNumbers, eligible[i].PunchNumber)
		}
	}

	// ── Compute the broadest possible 24-hour window for this date ────────────
	//
	// For any shift starting at HH:mm on attendanceDate:
	//   windowStart = (attendanceDate @ HH:mm) - 1h
	//   windowEnd   = windowStart + 24h - 1s
	//
	// Earliest possible windowStart: shift=00:00 → 23:00 previous day
	//   = attendanceDate - 1h
	// Latest possible windowEnd:     shift=23:59 → next day 22:58:59
	//   = attendanceDate + 47h (rounded up to 48h for safety)
	//
	// We fetch all punches in this broad window once, then per-employee we
	// filter using the employee's actual computed window.
	broadWindowStart := attendanceDate.Add(-1 * time.Hour)
	broadWindowEnd := attendanceDate.Add(48 * time.Hour)

	batchLogs, batchErr := p.dataLogRepo.GetPunchesByBadgesAndWindow(badgeNumbers, broadWindowStart, broadWindowEnd)
	if batchErr != nil {
		return dr, batchErr
	}
	dr.Logs = len(batchLogs)

	// Index batch logs by badge number (already sorted ASC by punch_time from repo).
	punchByBadge := make(map[string][]models.DataLog)
	for i := range batchLogs {
		b := batchLogs[i].BadgeNumber
		if b != "" {
			punchByBadge[b] = append(punchByBadge[b], batchLogs[i])
		}
	}

	// Collect IDs to mark as processed.
	var logIDsToMark []string

	// ── Process each eligible employee ───────────────────────────────────────
	for i := range eligible {
		emp := &eligible[i]

		// Skip employees handled by missing attendance override.
		if missingHandledByEmp[emp.EmployeeID] {
			continue
		}

		// 1. Resolve shift (temporary shift takes priority).
		shift := p.resolveShift(emp, date, tempShiftByKey, shiftCache)

		// 2. Calculate the 24-hour attendance window.
		//    windowStart = shiftStart - 1h
		//    windowEnd   = windowStart + 24h - 1s
		var window utils.AttendanceWindow
		if shift != nil && shift.StartTime != "" {
			shiftStartDT := utils.ShiftStartOnDate(shift.StartTime, attendanceDate)
			if !shiftStartDT.IsZero() {
				window = utils.CalculateAttendanceWindow(attendanceDate, shiftStartDT)
			}
		}
		// Fallback: no shift → whole-day window centered on midnight.
		if window.Start.IsZero() {
			window = utils.AttendanceWindow{
				Start: attendanceDate.Add(-1 * time.Hour),
				End:   attendanceDate.Add(24*time.Hour - time.Second),
			}
		}

		// 3. Filter punches inside this employee's exact window.
		allPunches := punchByBadge[emp.PunchNumber]
		windowedPunches := filterPunchesInWindow(allPunches, window)

		// 4. Compute shift-end datetime for single-punch / all-outzone rules.
		var shiftEndDT time.Time
		if shift != nil && shift.EndTime != "" && shift.StartTime != "" {
			shiftEndDT = utils.BuildShiftEndDatetime(attendanceDate, shift.StartTime, shift.EndTime)
		}

		// 5. Determine check-in / check-out from windowed punches.
		checkIn, checkOut := resolveInOut(windowedPunches, shiftEndDT)

		// 6. Compute all attendance fields.
		att := p.computeAttendance(
			emp, date, attendanceDate,
			shift, checkIn, checkOut,
			onLeaveSet, isGovHoliday, isCompWeekend, isGenDuty,
		)

		// 7. Collect punch IDs to mark processed.
		for _, punch := range windowedPunches {
			logIDsToMark = append(logIDsToMark, punch.ID)
		}

		// 6. Upsert.
		if existing, exists := existingAttByEmp[emp.EmployeeID]; exists {
			updates := map[string]interface{}{
				"shift_id":     att.ShiftID,
				"check_in":     att.CheckIn,
				"check_out":    att.CheckOut,
				"total_hours":  att.TotalHours,
				"over_time":    att.OverTime,
				"status":       att.Status,
				"late_minutes": att.LateMinutes,
			}
			if emp.PunchNumber != "" {
				updates["punch_number"] = emp.PunchNumber
			}
			if err := p.attendanceRepo.UpdateFields(existing.ID, updates); err == nil {
				dr.Updated++
			} else {
				dr.Skipped++
			}
		} else {
			if emp.PunchNumber != "" {
				att.PunchNumber = &emp.PunchNumber
			}
			if err := p.attendanceRepo.Create(att); err == nil {
				dr.Created++
				existingAttByEmp[emp.EmployeeID] = att
			} else {
				dr.Skipped++
			}
		}
	}

	// Mark punch logs as processed.
	if len(logIDsToMark) > 0 {
		_ = p.dataLogRepo.MarkProcessed(logIDsToMark)
	}

	return dr, nil
}

// ─── Core punch resolution ────────────────────────────────────────────────────

// resolveInOut selects check-in and check-out from punches already filtered
// into the 24-hour attendance window (sorted ASC by punch_time).
//
// # Single-punch rules (when shiftEndDT is known)
//
// Threshold = shiftEndDT - 5 hours
//
//   - punch < threshold              → checkOut = punch, checkIn = nil
//     (employee clocked out early-morning; no in-punch captured)
//   - punch >= shiftEndDT            → checkOut = punch, checkIn = nil
//     (employee clocked out after shift end; no in-punch captured)
//   - threshold <= punch < shiftEnd  → checkIn  = punch, checkOut = nil
//     (employee punched in during normal working hours; no out captured)
//
// # Multi-punch "all outTime zone" rule
//
// If ALL punches in the window fall in the outTime zone
// (i.e. every punch < threshold OR every punch >= shiftEndDT),
// treat the LAST punch as checkOut and checkIn remains nil.
//
// # Normal rule (multi-punch, at least one in working-hours zone)
//
//  1. First punch  → checkIn  (In Time).
//  2. 25-minute debounce after checkIn: punches ≤ checkIn+25min are ignored
//     for checkOut selection (raw data is never deleted).
//  3. Last punch after the debounce cutoff → checkOut (Out Time).
func resolveInOut(punches []models.DataLog, shiftEndDT time.Time) (checkIn, checkOut *time.Time) {
	if len(punches) == 0 {
		return nil, nil
	}

	hasShiftEnd := !shiftEndDT.IsZero()

	// Helper: returns true when the punch time is in the "outTime zone"
	// (at or after shiftEnd - 5h).
	isOutZone := func(pt time.Time) bool {
		if !hasShiftEnd {
			return false
		}
		threshold := shiftEndDT.Add(-5 * time.Hour)
		return !pt.Before(threshold) // pt >= shiftEnd - 5h
	}

	// ── Single punch ──────────────────────────────────────────────────────────
	if len(punches) == 1 {
		pt := punches[0].PunchTime
		if hasShiftEnd && isOutZone(pt) {
			// Lone punch in outTime zone → count as checkOut, no checkIn.
			co := pt
			return nil, &co
		}
		// Lone punch in working-hours zone → count as checkIn, no checkOut.
		ci := pt
		return &ci, nil
	}

	// ── Multiple punches: check if ALL are in the outTime zone ────────────────
	if hasShiftEnd {
		allInOutZone := true
		for _, p := range punches {
			if !isOutZone(p.PunchTime) {
				allInOutZone = false
				break
			}
		}
		if allInOutZone {
			// All punches are checkout-zone punches.
			// Use the last punch as checkOut; no checkIn.
			co := punches[len(punches)-1].PunchTime
			return nil, &co
		}
	}

	// ── Normal multi-punch rule ───────────────────────────────────────────────
	// First punch → checkIn.
	ci := punches[0].PunchTime
	checkIn = &ci

	// 25-minute debounce: ignore punches within 25 min of checkIn for checkOut.
	const debounceMinutes = 25
	debounceCutoff := checkIn.Add(debounceMinutes * time.Minute)

	// Last punch strictly after the debounce cutoff → checkOut.
	for i := len(punches) - 1; i >= 1; i-- {
		pt := punches[i].PunchTime
		if pt.After(debounceCutoff) {
			co := pt
			checkOut = &co
			break
		}
	}

	return checkIn, checkOut
}

// ─── Attendance field computation ─────────────────────────────────────────────

// computeAttendance builds the full Attendance model fields for one employee on
// one date given check-in/check-out times and contextual flags.
func (p *AttendanceProcessor) computeAttendance(
	emp *models.Employee,
	date string,
	attendanceDate time.Time,
	shift *models.Shift,
	checkIn, checkOut *time.Time,
	onLeaveSet map[string]bool,
	isGovHoliday, isCompWeekend, isGenDuty bool,
) *models.Attendance {

	att := &models.Attendance{
		EmployeeID: emp.EmployeeID,
		CompanyID:  emp.CompanyID,
		Date:       date,
	}

	// Attach shift ID.
	if shift != nil {
		att.ShiftID = &shift.ID
	}

	att.CheckIn = checkIn
	att.CheckOut = checkOut
	att.TotalHours = utils.CalcTotalHoursStr(checkIn, checkOut)

	// ── Determine special-day status ──────────────────────────────────────────
	isSpecialDay := false
	specialStatus := ""
	if isGovHoliday {
		isSpecialDay = true
		specialStatus = "holiday"
	} else if isCompWeekend {
		isSpecialDay = true
		specialStatus = "weekend"
	} else if !isGenDuty && shift != nil && shift.WeekendDays != "" && utils.IsWeekend(date, shift.WeekendDays) {
		isSpecialDay = true
		specialStatus = "weekend"
	}

	// ── Status ────────────────────────────────────────────────────────────────
	status := "present"

	if isSpecialDay {
		status = specialStatus
	} else if checkIn == nil && checkOut == nil {
		status = "absent"
	} else if checkIn == nil && checkOut != nil {
		status = "late"
	}

	// ── Late minutes ──────────────────────────────────────────────────────────
	lateMinutes := 0
	if !isSpecialDay && checkIn != nil && shift != nil && shift.StartTime != "" {
		shiftStartDT := utils.ShiftStartOnDate(shift.StartTime, attendanceDate)
		if !shiftStartDT.IsZero() {
			grace := time.Duration(shift.LateGraceMinutes) * time.Minute
			deadline := shiftStartDT.Add(grace)
			if checkIn.After(deadline) {
				lateMinutes = int(checkIn.Sub(shiftStartDT).Minutes())
				if status == "present" {
					status = "late"
				}
			}
		}
	}

	// ── Half-day: worked hours > 0 and < 4h ──────────────────────────────────
	if !isSpecialDay && checkIn != nil && checkOut != nil && att.TotalHours != nil {
		if m, ok := utils.ParseHHMMToMinutes(*att.TotalHours); ok && m > 0 && m < 4*60 {
			status = "half_day"
		}
	}

	// ── Leave overrides present/late/absent (not holiday/weekend) ────────────
	if !isSpecialDay && onLeaveSet[emp.EmployeeID] {
		status = "on_leave"
	}

	att.Status = status
	att.LateMinutes = lateMinutes

	// ── Overtime ──────────────────────────────────────────────────────────────
	// Rules:
	//   - over_time_status = false → always 0, regardless of hours worked.
	//   - OT < 45 minutes         → 0 OT hours.
	//   - OT >= 45 minutes        → 1 + floor((OT_min - 45) / 60) hours.
	//   - Special days (holiday/weekend): OT = total worked time (if enabled).
	otHours := 0

	if emp.OverTimeStatus && checkOut != nil {
		if isSpecialDay {
			// On special days, all worked time counts as OT (after 1h break deduction
			// if worked past 14:00).
			if att.TotalHours != nil {
				if m, ok := utils.ParseHHMMToMinutes(*att.TotalHours); ok && m > 0 {
					otMin := m
					if checkOut.Hour()*60+checkOut.Minute() > 14*60 {
						otMin -= 60 // deduct 1h lunch break
					}
					if otMin < 0 {
						otMin = 0
					}
					otHours = calcOTHours(otMin)
				}
			}
		} else if !isSpecialDay && shift != nil && shift.EndTime != "" && shift.StartTime != "" {
			// Regular day: OT is time worked beyond shift end.
			shiftEnd := utils.BuildShiftEndDatetime(attendanceDate, shift.StartTime, shift.EndTime)
			if !shiftEnd.IsZero() {
				otHours = utils.CalculateOvertime(*checkOut, shiftEnd, true)
			}
		}
	}
	// If emp.OverTimeStatus == false, otHours stays 0.

	otStr := strconv.Itoa(otHours)
	att.OverTime = &otStr

	return att
}

// calcOTHours applies the 45-minute threshold rule to a raw number of OT minutes.
// Special rule: if calculated OT hours equals 7, add +1 (7 -> 8).
//
//	< 45  → 0
//	>= 45 → 1 + floor((min - 45) / 60)
//	== 7  → 8
func calcOTHours(otMin int) int {
	if otMin < 45 {
		return 0
	}
	h := 1 + (otMin-45)/60
	if h == 7 {
		h = 8
	}
	return h
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// filterPunchesInWindow returns only punches within [window.Start, window.End].
// Punches outside this window are silently discarded (not from the DB).
// Input must be sorted ASC (guaranteed by the repository query).
func filterPunchesInWindow(punches []models.DataLog, window utils.AttendanceWindow) []models.DataLog {
	if window.Start.IsZero() {
		return punches
	}
	result := make([]models.DataLog, 0, len(punches))
	for i := range punches {
		pt := punches[i].PunchTime
		if (pt.Equal(window.Start) || pt.After(window.Start)) &&
			(pt.Equal(window.End) || pt.Before(window.End)) {
			result = append(result, punches[i])
		}
	}
	return result
}

// resolveShift returns the shift for an employee on a given date.
// Priority: active temporary_shift → employee's default shift.
func (p *AttendanceProcessor) resolveShift(
	emp *models.Employee,
	date string,
	tempShiftByKey map[string]*models.TemporaryShift,
	shiftCache map[string]*models.Shift,
) *models.Shift {
	tempKey := emp.EmployeeID + "|" + date
	if ts, ok := tempShiftByKey[tempKey]; ok && ts.ShiftID != "" {
		return p.getShift(ts.ShiftID, shiftCache)
	}
	if emp.ShiftID != nil {
		return p.getShift(*emp.ShiftID, shiftCache)
	}
	return nil
}

// isEligibleForDate returns true when the employee should be processed.
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

// getShift fetches a shift by ID using an in-memory cache.
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
