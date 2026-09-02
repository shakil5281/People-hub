package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Salary constants — fixed allowances
const (
	transportAllowance = 450
	foodAllowance      = 1250
	medicalAllowance   = 750
)

type SalaryService struct {
	employeeRepo       *repository.EmployeeRepository
	attendanceRepo     *repository.AttendanceRepository
	salaryRepo         *repository.SalaryRepository
	groupRepo          *repository.GroupRepository
	otEarlyExitRepo    *repository.OtEarlyExitRepository
	otEarlyExitService *OtEarlyExitService
	advanceRepo        *repository.AdvanceSalaryRepository
	separationRepo     *repository.SeparationRepository
	incrementRepo      *repository.SalaryIncrementRepository
}

func NewSalaryService(
	employeeRepo *repository.EmployeeRepository,
	attendanceRepo *repository.AttendanceRepository,
	salaryRepo *repository.SalaryRepository,
	groupRepo *repository.GroupRepository,
	otEarlyExitRepo *repository.OtEarlyExitRepository,
	otEarlyExitService *OtEarlyExitService,
	advanceRepo *repository.AdvanceSalaryRepository,
	separationRepo *repository.SeparationRepository,
) *SalaryService {
	return &SalaryService{
		employeeRepo:       employeeRepo,
		attendanceRepo:     attendanceRepo,
		salaryRepo:         salaryRepo,
		groupRepo:          groupRepo,
		otEarlyExitRepo:    otEarlyExitRepo,
		otEarlyExitService: otEarlyExitService,
		advanceRepo:        advanceRepo,
		separationRepo:     separationRepo,
	}
}

// SetIncrementRepo injects increment repo for temporal salary (effective month logic)
func (s *SalaryService) SetIncrementRepo(repo *repository.SalaryIncrementRepository) {
	s.incrementRepo = repo
}

// MonthResult holds the aggregated result of processing a month
type MonthResult struct {
	Processed int
	Total     int
	Month     int
	Year      int
}

// ProcessMonth calculates and upserts salaries for all active employees.
// If deductEarlyExit is true (default), early-exit shortfall hours are deducted from monthly OT.
// If deductEarlyExit is false, shortfalls are NOT deducted from OT ("do not pay less" formula).
func (s *SalaryService) ProcessMonth(companyID string, month, year int, userID string, deductEarlyExit bool) (*MonthResult, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")
	daysInMonth := endDate.Day()

	employees, err := s.employeeRepo.ListForSalaryProcessing(companyID, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch employees: %w", err)
	}

	if len(employees) == 0 {
		return &MonthResult{Processed: 0, Total: 0, Month: month, Year: year}, nil
	}

	groupNameByID := make(map[string]string)
	if s.groupRepo != nil {
		groups, gErr := s.groupRepo.ListAll()
		if gErr == nil {
			for _, g := range groups {
				groupNameByID[g.ID] = g.Name
			}
		}
	}

	// Build employee ID list for scoped queries
	empIDs := make([]string, len(employees))
	for i, e := range employees {
		empIDs[i] = e.EmployeeID
	}

	// Temporal salary: effective increment as of month end (effective month → next increment)
	// If increments exist with effective_date <= endStr, use their NewGross/Basic/House/Medical for that month.
	// Before increment effective month → old salary; from effective month onward → new salary.
	// If no effective ≤ month but future exists (live already mutated), use future's Previous* as old salary.
	effectiveMap := make(map[string]*models.SalaryIncrement)
	futureMap := make(map[string]*models.SalaryIncrement)
	if s.incrementRepo != nil {
		if effMap, effErr := s.incrementRepo.BatchEffectiveSalaries(companyID, empIDs, endStr); effErr == nil {
			effectiveMap = effMap
		}
		if futMap, futErr := s.incrementRepo.BatchEarliestFutureIncrements(companyID, empIDs, endStr); futErr == nil {
			futureMap = futMap
		}
	}

	// Parallel fetch: Combined attendance+OT summary, AdvanceDeductions, and Early-Exit recompute
	var salarySummary []repository.AttendanceSalarySummary
	var advanceMap map[string]float64
	var attErr, advErr error

	doneEarlyExit := make(chan struct{})
	go func() {
		if s.otEarlyExitService != nil {
			_, _ = s.otEarlyExitService.ComputeEarlyExitDeductions(companyID, month, year, userID)
		}
		close(doneEarlyExit)
	}()

	// Parallelize: one combined attendance+OT query replaces two separate heavy scans
	var wg2 sync.WaitGroup
	wg2.Add(2)
	go func() {
		defer wg2.Done()
		salarySummary, attErr = s.attendanceRepo.GetSalarySummary(companyID, startStr, endStr, empIDs)
	}()
	go func() {
		defer wg2.Done()
		if s.advanceRepo != nil {
			var adv map[string]float64
			adv, advErr = s.advanceRepo.MonthlyDeductions(companyID, month, year)
			if advErr == nil {
				advanceMap = adv
			} else {
				advanceMap = make(map[string]float64)
			}
		} else {
			advanceMap = make(map[string]float64)
		}
	}()
	wg2.Wait()
	<-doneEarlyExit

	if attErr != nil {
		return nil, fmt.Errorf("fetch attendance: %w", attErr)
	}
	if advanceMap == nil {
		advanceMap = make(map[string]float64)
	}

	// Build attendance map and OT map from the combined summary
	attMap := make(map[string]map[string]interface{})
	otHoursMap := make(map[string]float64)
	for _, s := range salarySummary {
		attMap[s.EmployeeID] = map[string]interface{}{
			"present": s.Present,
			"absent":  s.Absent,
			"late":    s.Late,
			"leave":   s.Leave,
			"weekend": s.Weekend,
			"holiday": s.Holiday,
		}
		otHoursMap[s.EmployeeID] = s.OvertimeHours
	}

	// 2. Early-exit shortfall deduction: net OT = raw OT - shortfall (if deductEarlyExit is true).
	shortfallMap := make(map[string]float64)
	if s.otEarlyExitRepo != nil && deductEarlyExit {
		if shortfalls, sfErr := s.otEarlyExitRepo.MonthlyShortfallTotals(companyID, month, year); sfErr == nil {
			shortfallMap = shortfalls
		}
	}

	// 3. Separation dates for this month (to process salary until separation date)
	sepDates := make(map[string]string)
	if s.separationRepo != nil {
		if sd, sErr := s.separationRepo.GetSeparationDatesByMonth(companyID, month, year); sErr == nil {
			sepDates = sd
		}
	}

	var toUpsert []*models.Salary
	var toDeleteIDs []string
	for _, emp := range employees {
		groupName := ""
		if emp.GroupID != nil {
			groupName = groupNameByID[*emp.GroupID]
		}
		if groupName == "" && emp.GroupRef != nil {
			groupName = emp.GroupRef.Name
		}
		netOt := otHoursMap[emp.EmployeeID]
		if deductEarlyExit {
			netOt = netOt - shortfallMap[emp.EmployeeID]
		}
		if netOt < 0 {
			netOt = 0
		}
		advDeduction := advanceMap[emp.EmployeeID]
		sepDate := sepDates[emp.EmployeeID]
		if sepDate == "" && emp.ResignDate != nil && !emp.ResignDate.IsZero() {
			if emp.ResignDate.Year() == year && int(emp.ResignDate.Month()) == month {
				sepDate = emp.ResignDate.Format("2006-01-02")
			}
		}
		if sepDate == "" && s.attendanceRepo != nil && (strings.EqualFold(emp.Status, "inactive") || !strings.EqualFold(strings.TrimSpace(emp.EmployeeType), "regular")) {
			if sep, sErr := s.attendanceRepo.FindSeparationByEmployeeID(emp.EmployeeID); sErr == nil && sep != nil && sep.Date != "" {
				sepDate = sep.Date
			}
		}
		// Temporal effective salary: from effective month onward use New*; before first effective use Previous* (old salary)
		effectiveEmp := emp
		if eff, ok := effectiveMap[emp.EmployeeID]; ok && eff != nil {
			effectiveEmp.GrossSalary = eff.NewGross
			effectiveEmp.BasicSalary = eff.NewBasic
			effectiveEmp.HouseRent = eff.NewHouse
			effectiveEmp.MedicalAllowance = eff.NewMedical
		} else if fut, ok := futureMap[emp.EmployeeID]; ok && fut != nil {
			// No effective ≤ month but future exists → live already holds new, need old for this month
			effectiveEmp.GrossSalary = fut.PreviousGross
			effectiveEmp.BasicSalary = fut.PreviousBasic
			effectiveEmp.HouseRent = fut.PreviousHouse
			effectiveEmp.MedicalAllowance = fut.PreviousMedical
			// Fallback if Previous* are zero (should not happen): keep live
			if effectiveEmp.GrossSalary == 0 {
				effectiveEmp.GrossSalary = emp.GrossSalary
			}
		}
		salary := s.calculateEmployeeSalary(effectiveEmp, groupName, attMap[emp.EmployeeID], netOt, advDeduction, month, year, daysInMonth, sepDate, userID)
		if salary.NetSalary <= 1000 {
			toDeleteIDs = append(toDeleteIDs, salary.EmployeeID)
			continue
		}
		toUpsert = append(toUpsert, salary)
	}

	// Optimized batch — single transaction, chunked upserts (100 rows per statement) to avoid giant SQL parse
	db := s.salaryRepo.DB()
	err = db.Transaction(func(tx *gorm.DB) error {
		if len(toDeleteIDs) > 0 {
			if err := tx.Unscoped().Where("company_id = ? AND employee_id IN ? AND month = ? AND year = ?", companyID, toDeleteIDs, month, year).Delete(&models.Salary{}).Error; err != nil {
				return err
			}
		}
		if len(toUpsert) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "company_id"}, {Name: "employee_id"}, {Name: "month"}, {Name: "year"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"basic_salary", "house_rent", "medical_allowance", "transport_allowance", "food_allowance", "other_allowance",
					"gross_salary", "provident_fund", "tax", "loan_deduction", "advance_deduction", "absent_deduction", "other_deduction", "total_deductions",
					"overtime_hours", "overtime_rate", "overtime_amount", "attendance_bonus", "net_salary",
					"present_days", "absent_days", "late_days", "leave_days", "holiday_days", "weekend_days", "total_days",
					"status", "updated_at",
				}),
			}).CreateInBatches(toUpsert, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("bulk upsert salaries: %w", err)
	}
	processed := len(toUpsert)

	// 4. Mark processed advances as deducted
	if s.advanceRepo != nil {
		_ = s.advanceRepo.MarkAsDeducted(companyID, month, year)
	}

	return &MonthResult{
		Processed: processed,
		Total:     len(employees),
		Month:     month,
		Year:      year,
	}, nil
}

// calculateEmployeeSalary contains ALL business rules — isolated and unit-testable.
// When an employee has separated, salary is processed until their separation date.
func (s *SalaryService) calculateEmployeeSalary(
	emp models.Employee,
	groupName string,
	att map[string]interface{},
	otHours float64,
	advanceDeduction float64,
	month, year, daysInMonth int,
	sepDateStr string,
	userID string,
) *models.Salary {
	gross := emp.GrossSalary

	// Fixed allowances
	transport := float64(transportAllowance)
	food := float64(foodAllowance)
	medical := float64(medicalAllowance)
	other := emp.OtherAllowance

	// Core = Gross - fixed allowances (OtherAllowance is kept separate)
	core := gross - transport - food - medical
	basic := core / 1.5
	houseRent := core - basic

	// Attendance breakdown
	presentDays := 0
	absentDays := 0
	lateDays := 0
	leaveDays := 0
	holidayDays := 0
	weekendDays := 0
	totalDays := 0

	// Determine active working window in this month:
	// Start day: from 1, or joining date if joined in this month
	startDay := 1
	if !emp.JoiningDate.IsZero() && emp.JoiningDate.Year() == year && int(emp.JoiningDate.Month()) == month {
		startDay = emp.JoiningDate.Day()
	}

	// Cutoff day: end of month (daysInMonth), or separation date if separated in this month
	cutoffDay := daysInMonth
	if sepDateStr != "" {
		if sepTime, err := time.Parse("2006-01-02", sepDateStr); err == nil {
			if sepTime.Year() == year && int(sepTime.Month()) == month {
				cutoffDay = sepTime.Day()
			}
		}
	} else if emp.ResignDate != nil && !emp.ResignDate.IsZero() {
		if emp.ResignDate.Year() == year && int(emp.ResignDate.Month()) == month {
			cutoffDay = emp.ResignDate.Day()
		}
	}

	// Total expected/eligible days for this employee in the month (e.g. up to separation date)
	effectiveDays := cutoffDay - startDay + 1
	if effectiveDays < 0 {
		effectiveDays = 0
	}
	if effectiveDays > daysInMonth {
		effectiveDays = daysInMonth
	}

	totalDays = effectiveDays

	if att != nil {
		presentDays = toInt(att["present"])
		absentDays = toInt(att["absent"])
		lateDays = toInt(att["late"])
		leaveDays = toInt(att["leave"])
		holidayDays = toInt(att["holiday"])
		weekendDays = toInt(att["weekend"])
	}

	// Paid days in month = Present + Late + Weekend + Leave + Holiday.
	paidDays := presentDays + lateDays + weekendDays + leaveDays + holidayDays
	if totalDays > 0 && paidDays < totalDays {
		calcAbsent := totalDays - paidDays
		if calcAbsent > absentDays {
			absentDays = calcAbsent
		}
	}

	// Absent deduction:
	// Daily rate is calculated based on calendar days in month: gross / daysInMonth
	// Total unworked days = absent days within working window + unworked days outside window (before joining / after separation)
	unworkedDays := (daysInMonth - totalDays) + absentDays
	absentDeduction := float64(0)
	if daysInMonth > 0 {
		perDaySalary := gross / float64(daysInMonth)
		absentDeduction = perDaySalary * float64(unworkedDays)
	}

	// Late attendance salary deduction:
	// - Every 3 late days = 1 day salary deduction (e.g. 3 late = 1 day, 6 late = 2 days)
	// - 4 or more late days = 1 day salary deduction + Attendance Bonus = 0
	// - Deduction is stored in OtherDeduction (absent_days / absent_deduction remain unchanged)
	lateDeductionDays := lateDays / 3
	lateDeductionAmount := float64(0)
	if daysInMonth > 0 && lateDeductionDays > 0 {
		perDaySalary := gross / float64(daysInMonth)
		lateDeductionAmount = perDaySalary * float64(lateDeductionDays)
	}
	otherDeduction := lateDeductionAmount

	// Overtime — only when employee over_time_status is enabled
	otRate := float64(0)
	if emp.OverTimeStatus && daysInMonth > 0 {
		otRate = (basic / 208) * 2
	}
	otAmount := otHours * otRate

	// Attendance bonus by employee group
	// Disqualified if absentDays > 0 OR lateDays >= 4
	attBonus := float64(0)
	if absentDays == 0 && lateDays < 4 && presentDays > 0 {
		switch {
		case strings.EqualFold(groupName, "worker"):
			attBonus = 725
		case strings.EqualFold(groupName, "staff"):
			attBonus = 300
		default:
			attBonus = 0
		}
	}

	totalDeductions := absentDeduction + otherDeduction + advanceDeduction
	netSalary := gross - totalDeductions + otAmount + attBonus
	if netSalary < 0 {
		netSalary = 0
	}

	return &models.Salary{
		CompanyID:          emp.CompanyID,
		EmployeeID:         emp.EmployeeID,
		Month:              month,
		Year:               year,
		BasicSalary:        basic,
		HouseRent:          houseRent,
		MedicalAllowance:   medical,
		TransportAllowance: transport,
		FoodAllowance:      food,
		OtherAllowance:     other,
		GrossSalary:        gross,
		ProvidentFund:      0,
		Tax:                0,
		LoanDeduction:      0,
		AdvanceDeduction:   advanceDeduction,
		AbsentDeduction:    absentDeduction,
		OtherDeduction:     otherDeduction,
		TotalDeductions:    totalDeductions,
		OvertimeHours:      otHours,
		OvertimeRate:       otRate,
		OvertimeAmount:     otAmount,
		AttendanceBonus:    attBonus,
		NetSalary:          netSalary,
		PresentDays:        presentDays,
		AbsentDays:         absentDays,
		LateDays:           lateDays,
		LeaveDays:          leaveDays,
		HolidayDays:        holidayDays,
		WeekendDays:        weekendDays,
		TotalDays:          totalDays,
		Status:             "processed",
		CreatedBy:          &userID,
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int64:
		return int(val)
	case float64:
		return int(val)
	case int:
		return val
	default:
		return 0
	}
}
