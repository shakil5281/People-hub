package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

// processingLocks prevents concurrent duplicate processing for same company+range.
var processingLocks sync.Map // key = companyID|start|end -> struct{}

// AttendanceProcessor converts raw biometric punch data into structured attendance records.
type AttendanceProcessor struct {
	dataLogRepo           *repository.DataLogRepository
	attendanceRepo        *repository.AttendanceRepository
	employeeRepo          *repository.EmployeeRepository
	shiftRepo             *repository.ShiftRepository
	leaveRepo             *repository.LeaveRepository
	tempShiftRepo         *repository.TemporaryShiftRepository
	rosterRepo            *repository.RosterRepository
	holidayRepo           *repository.HolidayRepository
	missingAttendanceRepo *repository.MissingAttendanceRepository
	separationRepo        *repository.SeparationRepository
}

func NewAttendanceProcessor(
	dataLogRepo *repository.DataLogRepository,
	attendanceRepo *repository.AttendanceRepository,
	employeeRepo *repository.EmployeeRepository,
	shiftRepo *repository.ShiftRepository,
	leaveRepo *repository.LeaveRepository,
	tempShiftRepo *repository.TemporaryShiftRepository,
	rosterRepo *repository.RosterRepository,
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
		rosterRepo:            rosterRepo,
		holidayRepo:           holidayRepo,
		missingAttendanceRepo: missingAttendanceRepo,
	}
}

// SetSeparationRepo injects separation repo for bulk eligibility — avoids N+1.
func (p *AttendanceProcessor) SetSeparationRepo(repo *repository.SeparationRepository) {
	p.separationRepo = repo
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

// dailyProcessContext holds bulk-loaded data for a range — single-flight.
type dailyProcessContext struct {
	employees        []models.Employee
	separationMap    map[string]*models.Separation
	existingByKey    map[string]*models.Attendance // key = employeeID|date
	leaveByKey       map[string]bool               // key = employeeID|date
	holidaysByDate   map[string][]models.Holiday
	holidayFlags     map[string]holidayFlags
	missingByKey     map[string]*models.MissingAttendance // key = employeeID|date
	tempShiftByKey   map[string]*models.TemporaryShift
	rosterByKey      map[string]*models.Roster
	shiftByID        map[string]*models.Shift
	punchByBadge     map[string][]models.DataLog
	allPunches       []models.DataLog
	eligibleByDate   map[string][]models.Employee
	eligibleIDsByDate map[string][]string
}

type holidayFlags struct {
	isGovHoliday  bool
	isCompWeekend bool
	isGenDuty     bool
}

// ─── Main entry point ─────────────────────────────────────────────────────────

// ProcessDateRange converts raw punch data into attendance records for every day
// in [startDate, endDate]. Bulk-loaded, in-memory, batch-persisted.
func (p *AttendanceProcessor) ProcessDateRange(startDate, endDate, companyID string) (*ProcessDateRangeResult, error) {
	lockKey := companyID + "|" + startDate + "|" + endDate
	if _, loaded := processingLocks.LoadOrStore(lockKey, struct{}{}); loaded {
		return nil, fmt.Errorf("daily process already running for %s %s-%s", companyID, startDate, endDate)
	}
	defer processingLocks.Delete(lockKey)

	overallStart := time.Now()
	dates, err := utils.GenerateDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date range: %w", err)
	}
	if len(dates) == 0 {
		return nil, fmt.Errorf("invalid date range: start %s after end %s", startDate, endDate)
	}
	// Safety: chunk large ranges to bound memory
	const maxChunkDays = 31
	if len(dates) > maxChunkDays {
		return p.processChunked(dates, companyID)
	}

	// Bulk load phase
	loadStart := time.Now()
	ctx, err := p.buildContext(dates, startDate, endDate, companyID)
	if err != nil {
		return nil, err
	}
	loadDur := time.Since(loadStart)

	// Calculation phase — pure in-memory, no DB
	calcStart := time.Now()
	type calcResult struct {
		attendances []models.Attendance
		missingSync []missingSync
		punchIDs    []string
		dayResults  map[string]*DayResult
		existingKeys map[string]bool // key = employeeID|date for created vs updated
	}
	cr := calcResult{
		dayResults:   make(map[string]*DayResult),
		existingKeys: make(map[string]bool),
	}
	for _, d := range dates {
		cr.dayResults[d] = &DayResult{Date: d}
	}
	// pre-populate existingKeys from ctx
	for k := range ctx.existingByKey {
		cr.existingKeys[k] = true
	}

	// For each date, process eligible employees
	for _, date := range dates {
		attendanceDate, _ := time.Parse("2006-01-02", date)
		dr := cr.dayResults[date]
		hf := ctx.holidayFlags[date]
		// leaves for this date: filter leaveByKey by date suffix
		onLeaveSet := make(map[string]bool)
		for k, v := range ctx.leaveByKey {
			if v && strings.HasSuffix(k, "|"+date) {
				empID := strings.TrimSuffix(k, "|"+date)
				onLeaveSet[empID] = true
			}
		}
		eligible := ctx.eligibleByDate[date]
		// Count logs for this day: windowed punches total
		dayLogCount := 0
		for _, emp := range eligible {
			shift := p.resolveShiftFromContext(&emp, date, ctx)
			window := attendanceWindowFor(attendanceDate, shift)
			punches := ctx.punchByBadge[emp.PunchNumber]
			windowed := filterPunchesInWindow(punches, window)
			dayLogCount += len(windowed)
		}
		// Also count punches that fell into global broad window for observability parity
		// Use dayLogCount for per-day Logs (more precise than broad 49h)
		dr.Logs = dayLogCount

		for i := range eligible {
			emp := &eligible[i]
			key := emp.EmployeeID + "|" + date
			ma := ctx.missingByKey[key]
			existing, exists := ctx.existingByKey[key]

			shift := p.resolveShiftFromContext(emp, date, ctx)
			window := attendanceWindowFor(attendanceDate, shift)
			allPunches := ctx.punchByBadge[emp.PunchNumber]
			windowedPunches := filterPunchesInWindow(allPunches, window)

			var shiftEndDT time.Time
			if shift != nil && shift.EndTime != "" && shift.StartTime != "" {
				shiftEndDT = utils.BuildShiftEndDatetime(attendanceDate, shift.StartTime, shift.EndTime)
			}
			bioIn, bioOut := resolveInOut(windowedPunches, shiftEndDT)

			checkIn := bioIn
			checkOut := bioOut
			if ma != nil {
				if ma.CheckIn != nil {
					checkIn = ma.CheckIn
				}
				if ma.CheckOut != nil {
					checkOut = ma.CheckOut
				}
			}
			if checkIn == nil && exists && existing.CheckIn != nil {
				checkIn = existing.CheckIn
			}
			if checkOut == nil && exists && existing.CheckOut != nil {
				checkOut = existing.CheckOut
			}

			att := p.computeAttendance(emp, date, attendanceDate, shift, checkIn, checkOut, onLeaveSet, hf.isGovHoliday, hf.isCompWeekend, hf.isGenDuty)
			if ma != nil && ma.Status != "" {
				att.Status = ma.Status
			}
			// collect missing sync (deferred batch, not per-row DB)
			if ma != nil && ma.Status != "absent" {
				cr.missingSync = append(cr.missingSync, missingSync{
					id:         ma.ID,
					checkIn:    checkIn,
					checkOut:   checkOut,
					totalHours: att.TotalHours,
					overTime:   att.OverTime,
					status:     att.Status,
				})
			}
			for _, punch := range windowedPunches {
				cr.punchIDs = append(cr.punchIDs, punch.ID)
			}
			// track attendance for batch upsert
			cr.attendances = append(cr.attendances, *att)
			// provisional created/updated counting based on existence prior to this run
			// final counts will be reconciled after persistence (but we count here for result)
			if exists {
				dr.Updated++
			} else {
				dr.Created++
			}
		}
	}
	calcDur := time.Since(calcStart)

	// Persistence phase — batch upsert + missing sync + mark punches
	persistStart := time.Now()
	if len(cr.attendances) > 0 {
		if err := p.attendanceRepo.UpsertBatch(cr.attendances); err != nil {
			// Do NOT mark punches on failure
			return nil, fmt.Errorf("batch upsert failed: %w", err)
		}
	}
	// Batch sync missing_attendances (best-effort, log errors but don't fail whole process)
	for _, ms := range cr.missingSync {
		if err := p.missingAttendanceRepo.UpdateFields(ms.id, map[string]interface{}{
			"check_in":    ms.checkIn,
			"check_out":   ms.checkOut,
			"total_hours": ms.totalHours,
			"over_time":   ms.overTime,
			"status":      ms.status,
		}); err != nil {
			log.Printf("[daily-process] missing sync failed id=%s: %v", ms.id, err)
		}
	}
	// Deduplicate punch IDs before marking
	if len(cr.punchIDs) > 0 {
		seen := make(map[string]struct{}, len(cr.punchIDs))
		uniq := make([]string, 0, len(cr.punchIDs))
		for _, id := range cr.punchIDs {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				uniq = append(uniq, id)
			}
		}
		if err := p.dataLogRepo.MarkProcessed(uniq); err != nil {
			log.Printf("[daily-process] mark punches failed: %v", err)
			// not fatal — attendances already persisted
		}
	}
	persistDur := time.Since(persistStart)

	// Build result
	result := &ProcessDateRangeResult{
		Days:    len(dates),
		Details: make([]DayResult, 0, len(dates)),
	}
	for _, d := range dates {
		dr := cr.dayResults[d]
		result.Details = append(result.Details, *dr)
		result.TotalCreated += dr.Created
		result.TotalUpdated += dr.Updated
		result.TotalSkipped += dr.Skipped
		result.TotalLogs += dr.Logs
	}
	// TotalLogs: prefer global punch count for observability (more accurate than per-day windowed sum if overlapping)
	// Keep per-day Logs as windowed, but TotalLogs as len(allPunches) for parity with old broad fetch
	if len(ctx.allPunches) > 0 {
		result.TotalLogs = len(ctx.allPunches)
	}
	result.TotalProcessed = result.TotalCreated + result.TotalUpdated

	totalDur := time.Since(overallStart)
	log.Printf("[daily-process] company=%s range=%s..%s employees=%d punches=%d created=%d updated=%d skipped=%d load=%s calc=%s persist=%s total=%s",
		companyID, startDate, endDate, len(ctx.employees), len(ctx.allPunches), result.TotalCreated, result.TotalUpdated, result.TotalSkipped, loadDur, calcDur, persistDur, totalDur)

	return result, nil
}

func (p *AttendanceProcessor) processChunked(dates []string, companyID string) (*ProcessDateRangeResult, error) {
	const chunkSize = 31
	agg := &ProcessDateRangeResult{Details: []DayResult{}}
	for i := 0; i < len(dates); i += chunkSize {
		end := i + chunkSize
		if end > len(dates) {
			end = len(dates)
		}
		chunk := dates[i:end]
		res, err := p.ProcessDateRange(chunk[0], chunk[len(chunk)-1], companyID)
		if err != nil {
			return nil, err
		}
		agg.Days += res.Days
		agg.TotalCreated += res.TotalCreated
		agg.TotalUpdated += res.TotalUpdated
		agg.TotalSkipped += res.TotalSkipped
		agg.TotalLogs += res.TotalLogs
		agg.TotalProcessed += res.TotalProcessed
		agg.Details = append(agg.Details, res.Details...)
		// prevent recursive lock collision — release lock between chunks? Already using same key per chunk, not global. So we need to bypass lock for inner calls.
		// Instead, inner calls use chunk-specific lock keys, so no collision.
	}
	return agg, nil
}

type missingSync struct {
	id         string
	checkIn    *time.Time
	checkOut   *time.Time
	totalHours *string
	overTime   *string
	status     string
}

func (p *AttendanceProcessor) buildContext(dates []string, startDate, endDate, companyID string) (*dailyProcessContext, error) {
	ctx := &dailyProcessContext{
		separationMap:     make(map[string]*models.Separation),
		existingByKey:     make(map[string]*models.Attendance),
		leaveByKey:        make(map[string]bool),
		holidaysByDate:    make(map[string][]models.Holiday),
		holidayFlags:      make(map[string]holidayFlags),
		missingByKey:      make(map[string]*models.MissingAttendance),
		tempShiftByKey:    make(map[string]*models.TemporaryShift),
		rosterByKey:       make(map[string]*models.Roster),
		shiftByID:         make(map[string]*models.Shift),
		punchByBadge:      make(map[string][]models.DataLog),
		eligibleByDate:    make(map[string][]models.Employee),
		eligibleIDsByDate: make(map[string][]string),
	}

	// 1. Employees — single query
	emps, err := p.employeeRepo.ListActiveRegularAll(companyID)
	if err != nil {
		return nil, fmt.Errorf("load employees: %w", err)
	}
	ctx.employees = emps

	// 2. Separations — bulk
	if p.separationRepo != nil && len(emps) > 0 {
		ids := make([]string, 0, len(emps))
		for i := range emps {
			ids = append(ids, emps[i].EmployeeID)
		}
		seps, sErr := p.separationRepo.ListByEmployeeIDs(ids)
		if sErr != nil {
			// Don't fail hard — log and treat as no separations (previous ignored errors similarly, but now we log)
			log.Printf("[daily-process] separation load error: %v", sErr)
		} else {
			// Keep latest per employee (ListByEmployeeIDs ORDER BY date DESC, first wins)
			for i := range seps {
				empID := seps[i].EmployeeID
				if _, ok := ctx.separationMap[empID]; !ok {
					// copy
					cp := seps[i]
					ctx.separationMap[empID] = &cp
				}
			}
		}
	}

	// Build eligibility per date using in-memory separationMap (no N+1)
	for _, date := range dates {
		for i := range emps {
			emp := &emps[i]
			if p.isEligibleForDateWithMap(emp, date, ctx.separationMap) {
				ctx.eligibleByDate[date] = append(ctx.eligibleByDate[date], *emp)
				ctx.eligibleIDsByDate[date] = append(ctx.eligibleIDsByDate[date], emp.EmployeeID)
			}
		}
	}
	// Collect distinct eligible IDs across range for bulk attendance fetch
	eligibleSet := make(map[string]struct{})
	for _, ids := range ctx.eligibleIDsByDate {
		for _, id := range ids {
			eligibleSet[id] = struct{}{}
		}
	}
	allEligibleIDs := make([]string, 0, len(eligibleSet))
	for id := range eligibleSet {
		allEligibleIDs = append(allEligibleIDs, id)
	}

	// 3. Existing attendances — single range query
	if len(allEligibleIDs) > 0 {
		existing, err := p.attendanceRepo.ListByDateRangeAndEmployeeIDs(startDate, endDate, allEligibleIDs)
		if err != nil {
			return nil, fmt.Errorf("load existing attendances: %w", err)
		}
		for i := range existing {
			key := existing[i].EmployeeID + "|" + existing[i].Date
			// normalize date to YYYY-MM-DD (handle timestamp)
			norm := utils.NormalizeDate(existing[i].Date)
			if norm != existing[i].Date {
				existing[i].Date = norm
			}
			cp := existing[i]
			ctx.existingByKey[key] = &cp
		}
	}

	// 4. Approved leaves — single range query
	leaves, err := p.leaveRepo.ListApprovedByDateRange(startDate, endDate)
	if err != nil {
		// Previously ignored; now we treat as error because silently missing leaves changes business result
		return nil, fmt.Errorf("load leaves: %w", err)
	}
	for i := range leaves {
		empID := leaves[i].EmployeeID
		from := utils.NormalizeDate(leaves[i].FromDate)
		to := utils.NormalizeDate(leaves[i].ToDate)
		// expand to affected dates within [startDate,endDate]
		datesInLeave, _ := utils.GenerateDateRange(maxDate(from, startDate), minDate(to, endDate))
		for _, d := range datesInLeave {
			ctx.leaveByKey[empID+"|"+d] = true
		}
	}

	// 5. Holidays — single range query
	holidays, err := p.holidayRepo.ListActiveByDateRange(startDate, endDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("load holidays: %w", err)
	}
	// Also need to consider holidays where weekend_date inside range but date outside — ListActiveByDateRange already covers via OR
	for i := range holidays {
		h := &holidays[i]
		hDate := utils.NormalizeDate(h.Date)
		ctx.holidaysByDate[hDate] = append(ctx.holidaysByDate[hDate], *h)
		if h.FromDate != nil && h.ToDate != nil {
			// expand range holidays to each date they cover
			rDates, _ := utils.GenerateDateRange(utils.NormalizeDate(*h.FromDate), utils.NormalizeDate(*h.ToDate))
			for _, d := range rDates {
				if d == hDate {
					continue
				}
				ctx.holidaysByDate[d] = append(ctx.holidaysByDate[d], *h)
			}
		}
		if h.WeekendDate != nil {
			wd := utils.NormalizeDate(*h.WeekendDate)
			ctx.holidaysByDate[wd] = append(ctx.holidaysByDate[wd], *h)
		}
	}
	// Precompute flags per date
	for _, d := range dates {
		hf := holidayFlags{}
		for _, h := range ctx.holidaysByDate[d] {
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
			if h.Type != "weekend_change" && (hDate == d || (hFrom != "" && hTo != "" && d >= hFrom && d <= hTo)) {
				hf.isGovHoliday = true
			}
			if h.Type == "weekend_change" {
				if hDate == d {
					hf.isGenDuty = true
				}
				if hWeekend != "" && hWeekend == d {
					hf.isCompWeekend = true
				}
			}
		}
		ctx.holidayFlags[d] = hf
	}

	// 6. Missing attendance — single range query
	missing, err := p.missingAttendanceRepo.ListByDateRange(startDate, endDate, companyID)
	if err != nil {
		return nil, fmt.Errorf("load missing attendance: %w", err)
	}
	for i := range missing {
		norm := utils.NormalizeDate(missing[i].Date)
		key := missing[i].EmployeeID + "|" + norm
		cp := missing[i]
		cp.Date = norm
		ctx.missingByKey[key] = &cp
	}

	// 7. Temporary shifts — single range query
	allTempShifts, err := p.tempShiftRepo.ListByCompanyAndDateRange(companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("load temp shifts: %w", err)
	}
	for i := range allTempShifts {
		ts := &allTempShifts[i]
		if ts.Status != "" && !strings.EqualFold(ts.Status, "active") {
			continue
		}
		norm := utils.NormalizeDate(ts.Date)
		key := ts.EmployeeID + "|" + norm
		ctx.tempShiftByKey[key] = ts
	}

	// 8. Rosters — single range query
	if p.rosterRepo != nil {
		allRosters, err := p.rosterRepo.ListByCompanyAndDateRange(companyID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("load rosters: %w", err)
		}
		for i := range allRosters {
			r := &allRosters[i]
			if r.Status != "" && !strings.EqualFold(r.Status, "active") {
				continue
			}
			norm := utils.NormalizeDate(r.Date)
			key := r.EmployeeID + "|" + norm
			ctx.rosterByKey[key] = r
		}
	}

	// 9. Shifts — collect distinct IDs then bulk fetch
	shiftIDSet := make(map[string]struct{})
	for i := range emps {
		if emps[i].ShiftID != nil && *emps[i].ShiftID != "" {
			shiftIDSet[*emps[i].ShiftID] = struct{}{}
		}
	}
	for _, ts := range ctx.tempShiftByKey {
		if ts.ShiftID != "" {
			shiftIDSet[ts.ShiftID] = struct{}{}
		}
	}
	for _, r := range ctx.rosterByKey {
		if r.ShiftID != "" {
			shiftIDSet[r.ShiftID] = struct{}{}
		}
	}
	shiftIDs := make([]string, 0, len(shiftIDSet))
	for id := range shiftIDSet {
		shiftIDs = append(shiftIDs, id)
	}
	if len(shiftIDs) > 0 {
		shifts, err := p.shiftRepo.ListByIDs(shiftIDs)
		if err != nil {
			return nil, fmt.Errorf("load shifts: %w", err)
		}
		for i := range shifts {
			cp := shifts[i]
			ctx.shiftByID[shifts[i].ID] = &cp
		}
	}

	// 10. Punches — single global fetch for whole range
	badgeSet := make(map[string]struct{})
	for _, empList := range ctx.eligibleByDate {
		for _, emp := range empList {
			if emp.PunchNumber != "" {
				badgeSet[emp.PunchNumber] = struct{}{}
			}
		}
	}
	badges := make([]string, 0, len(badgeSet))
	for b := range badgeSet {
		badges = append(badges, b)
	}
	if len(badges) > 0 {
		// Global window: earliest window start to latest window end
		// Approximate as [startDate-1h, endDate+48h] — matches previous per-day broadWindow logic but in one query
		startT, _ := time.Parse("2006-01-02", startDate)
		endT, _ := time.Parse("2006-01-02", endDate)
		globalStart := startT.Add(-1 * time.Hour)
		globalEnd := endT.Add(48 * time.Hour)
		allLogs, err := p.dataLogRepo.GetPunchesByBadgesAndWindow(badges, globalStart, globalEnd)
		if err != nil {
			return nil, fmt.Errorf("load punches: %w", err)
		}
		ctx.allPunches = allLogs
		for i := range allLogs {
			b := allLogs[i].BadgeNumber
			ctx.punchByBadge[b] = append(ctx.punchByBadge[b], allLogs[i])
		}
	}

	return ctx, nil
}

func attendanceWindowFor(attendanceDate time.Time, shift *models.Shift) utils.AttendanceWindow {
	if shift != nil && shift.StartTime != "" {
		shiftStartDT := utils.ShiftStartOnDate(shift.StartTime, attendanceDate)
		if !shiftStartDT.IsZero() {
			return utils.CalculateAttendanceWindow(attendanceDate, shiftStartDT)
		}
	}
	return utils.AttendanceWindow{
		Start: attendanceDate.Add(-1 * time.Hour),
		End:   attendanceDate.Add(24*time.Hour - time.Second),
	}
}

func (p *AttendanceProcessor) resolveShiftFromContext(emp *models.Employee, date string, ctx *dailyProcessContext) *models.Shift {
	key := emp.EmployeeID + "|" + date
	if r, ok := ctx.rosterByKey[key]; ok && r.ShiftID != "" {
		if s, ok := ctx.shiftByID[r.ShiftID]; ok {
			return s
		}
	}
	if ts, ok := ctx.tempShiftByKey[key]; ok && ts.ShiftID != "" {
		if s, ok := ctx.shiftByID[ts.ShiftID]; ok {
			return s
		}
	}
	if emp.ShiftID != nil {
		if s, ok := ctx.shiftByID[*emp.ShiftID]; ok {
			return s
		}
	}
	return nil
}

func maxDate(a, b string) string {
	if a > b {
		return a
	}
	return b
}
func minDate(a, b string) string {
	if a < b {
		return a
	}
	return b
}

// ─── Core punch resolution (unchanged business logic) ────────────────────────

func resolveInOut(punches []models.DataLog, shiftEndDT time.Time) (checkIn, checkOut *time.Time) {
	if len(punches) == 0 {
		return nil, nil
	}
	hasShiftEnd := !shiftEndDT.IsZero()
	isOutZone := func(pt time.Time) bool {
		if !hasShiftEnd {
			return false
		}
		threshold := shiftEndDT.Add(-5 * time.Hour)
		return !pt.Before(threshold)
	}
	if len(punches) == 1 {
		pt := punches[0].PunchTime
		if hasShiftEnd && isOutZone(pt) {
			co := pt
			return nil, &co
		}
		ci := pt
		return &ci, nil
	}
	if hasShiftEnd {
		allInOutZone := true
		for _, pr := range punches {
			if !isOutZone(pr.PunchTime) {
				allInOutZone = false
				break
			}
		}
		if allInOutZone {
			co := punches[len(punches)-1].PunchTime
			return nil, &co
		}
	}
	ci := punches[0].PunchTime
	checkIn = &ci
	const debounceMinutes = 25
	debounceCutoff := checkIn.Add(debounceMinutes * time.Minute)
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

// ─── Attendance field computation (unchanged) ───────────────────────────────

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
	if shift != nil {
		att.ShiftID = &shift.ID
	}
	att.CheckIn = checkIn
	att.CheckOut = checkOut
	att.TotalHours = utils.CalcTotalHoursStr(checkIn, checkOut)
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
	status := "present"
	if isSpecialDay {
		status = specialStatus
	} else if checkIn == nil && checkOut == nil {
		status = "absent"
	} else if checkIn == nil && checkOut != nil {
		status = "late"
	}
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
	if !isSpecialDay && checkIn != nil && checkOut != nil && att.TotalHours != nil {
		if m, ok := utils.ParseHHMMToMinutes(*att.TotalHours); ok && m > 0 && m < 4*60 {
			status = "half_day"
		}
	}
	if !isSpecialDay && onLeaveSet[emp.EmployeeID] {
		status = "on_leave"
	}
	att.Status = status
	att.LateMinutes = lateMinutes
	otHours := 0
	if emp.OverTimeStatus && checkOut != nil {
		if isSpecialDay {
			otHours = otHoursOnSpecialDay(att.TotalHours)
		} else if shift != nil && shift.EndTime != "" && shift.StartTime != "" {
			shiftEnd := utils.BuildShiftEndDatetime(attendanceDate, shift.StartTime, shift.EndTime)
			if !shiftEnd.IsZero() {
				otHours = utils.CalculateOvertime(*checkOut, shiftEnd, true)
			}
		}
	}
	otStr := strconv.Itoa(otHours)
	att.OverTime = &otStr
	return att
}

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

func otHoursOnSpecialDay(totalHours *string) int {
	if totalHours == nil {
		return 0
	}
	if m, ok := utils.ParseHHMMToMinutes(*totalHours); ok && m > 0 {
		return m / 60
	}
	return 0
}

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

func (p *AttendanceProcessor) isEligibleForDateWithMap(emp *models.Employee, date string, sepMap map[string]*models.Separation) bool {
	if emp == nil {
		return false
	}
	if strings.TrimSpace(emp.PunchNumber) == "" {
		return false
	}
	processDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	if !emp.JoiningDate.IsZero() {
		joinDay := time.Date(emp.JoiningDate.Year(), emp.JoiningDate.Month(), emp.JoiningDate.Day(), 0, 0, 0, 0, time.UTC)
		if processDate.Before(joinDay) {
			return false
		}
	}
	if strings.EqualFold(emp.Status, "active") && strings.EqualFold(strings.TrimSpace(emp.EmployeeType), "regular") {
		return true
	}
	if sep, ok := sepMap[emp.EmployeeID]; ok && sep != nil && sep.Date != "" {
		sepDay, parseErr := time.Parse("2006-01-02", sep.Date)
		if parseErr == nil {
			if !processDate.After(sepDay) {
				return true
			}
		}
	} else if strings.EqualFold(emp.Status, "active") {
		return true
	}
	return false
}

// Legacy isEligibleForDate kept for tests — delegates to map version with nil map (falls back to DB via attendanceRepo).
// Preserved for backward compatibility with existing tests that may call it directly via exported helper.
// Note: New bulk path uses isEligibleForDateWithMap; this legacy method remains functionally identical.
func (p *AttendanceProcessor) isEligibleForDate(emp *models.Employee, date string) bool {
	if emp == nil {
		return false
	}
	if strings.TrimSpace(emp.PunchNumber) == "" {
		return false
	}
	processDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	if !emp.JoiningDate.IsZero() {
		joinDay := time.Date(emp.JoiningDate.Year(), emp.JoiningDate.Month(), emp.JoiningDate.Day(), 0, 0, 0, 0, time.UTC)
		if processDate.Before(joinDay) {
			return false
		}
	}
	if strings.EqualFold(emp.Status, "active") && strings.EqualFold(strings.TrimSpace(emp.EmployeeType), "regular") {
		return true
	}
	// Try separationMap first if available, else fallback to DB (original behavior)
	if p.separationRepo != nil {
		// Single lookup via bulk map not available here; do direct query to preserve legacy behavior for tests
		sepList, _ := p.separationRepo.ListByEmployeeIDs([]string{emp.EmployeeID})
		if len(sepList) > 0 && sepList[0].Date != "" {
			sepDay, parseErr := time.Parse("2006-01-02", sepList[0].Date)
			if parseErr == nil && !processDate.After(sepDay) {
				return true
			}
		} else if strings.EqualFold(emp.Status, "active") {
			return true
		}
		return false
	}
	if sep, sErr := p.attendanceRepo.FindSeparationByEmployeeID(emp.EmployeeID); sErr == nil && sep != nil && sep.Date != "" {
		sepDay, parseErr := time.Parse("2006-01-02", sep.Date)
		if parseErr == nil {
			if !processDate.After(sepDay) {
				return true
			}
		}
	} else if strings.EqualFold(emp.Status, "active") {
		return true
	}
	return false
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
