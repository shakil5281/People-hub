package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

// Salary constants — fixed allowances
const (
	transportAllowance = 450
	foodAllowance      = 1250
	medicalAllowance   = 750
)

type SalaryService struct {
	employeeRepo    *repository.EmployeeRepository
	attendanceRepo  *repository.AttendanceRepository
	salaryRepo      *repository.SalaryRepository
	groupRepo       *repository.GroupRepository
	otEarlyExitRepo *repository.OtEarlyExitRepository
}

func NewSalaryService(
	employeeRepo *repository.EmployeeRepository,
	attendanceRepo *repository.AttendanceRepository,
	salaryRepo *repository.SalaryRepository,
	groupRepo *repository.GroupRepository,
	otEarlyExitRepo *repository.OtEarlyExitRepository,
) *SalaryService {
	return &SalaryService{
		employeeRepo:    employeeRepo,
		attendanceRepo:  attendanceRepo,
		salaryRepo:      salaryRepo,
		groupRepo:       groupRepo,
		otEarlyExitRepo: otEarlyExitRepo,
	}
}

// MonthResult holds the aggregated result of processing a month
type MonthResult struct {
	Processed int
	Total     int
	Month     int
	Year      int
}

// ProcessMonth calculates and upserts salaries for all active employees using the new formula:
//
//	core        = Gross - Transport(450) - Food(1250) - Medical(750)
//	Basic       = core / 1.5
//	House Rent  = core - Basic
//	OT Rate     = (Basic / 208) * 2 if over_time_status else 0
//	OT Amount   = OT Hours * OT Rate
//	Attendance Bonus = 725 for Worker / 300 for Staff / 0 others if absent_days == 0 AND present_days > 0
//	Net Salary  = Gross - AbsentDeduction + OTAmount + AttendanceBonus
func (s *SalaryService) ProcessMonth(companyID string, month, year int, userID string) (*MonthResult, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")
	daysInMonth := endDate.Day()

	employees, err := s.employeeRepo.ListActiveAll(companyID)
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

	attendanceReport, err := s.attendanceRepo.MonthlyReport(startStr, endStr, companyID, "", "", "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("fetch attendance: %w", err)
	}

	attMap := make(map[string]map[string]interface{})
	for _, r := range attendanceReport {
		if empID, ok := r["employee_id"].(string); ok {
			attMap[empID] = r
		}
	}

	otHoursMap, err := s.attendanceRepo.GetMonthlyOvertimeHours(companyID, startStr, endStr)
	if err != nil {
		return nil, fmt.Errorf("fetch overtime: %w", err)
	}

	// Early-exit shortfall is deducted from the employee's monthly overtime.
	// net OT = raw OT - shortfall (clamped at 0).
	shortfallMap := make(map[string]float64)
	if s.otEarlyExitRepo != nil {
		if shortfalls, sfErr := s.otEarlyExitRepo.MonthlyShortfallTotals(companyID, month, year); sfErr == nil {
			shortfallMap = shortfalls
		}
	}

	processed := 0

	for _, emp := range employees {
		groupName := ""
		if emp.GroupID != nil {
			groupName = groupNameByID[*emp.GroupID]
		}
		if groupName == "" && emp.GroupRef != nil {
			groupName = emp.GroupRef.Name
		}
		netOt := otHoursMap[emp.EmployeeID] - shortfallMap[emp.EmployeeID]
		if netOt < 0 {
			netOt = 0
		}
		salary := s.calculateEmployeeSalary(emp, groupName, attMap[emp.EmployeeID], netOt, month, year, daysInMonth, userID)

		if err := s.salaryRepo.Upsert(salary); err != nil {
			continue
		}
		processed++
	}

	return &MonthResult{
		Processed: processed,
		Total:     len(employees),
		Month:     month,
		Year:      year,
	}, nil
}

// calculateEmployeeSalary contains ALL business rules — isolated and unit-testable.
func (s *SalaryService) calculateEmployeeSalary(
	emp models.Employee,
	groupName string,
	att map[string]interface{},
	otHours float64,
	month, year, daysInMonth int,
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

	// Use calendar month days for totalDays and per-day salary calculations
	// instead of att["total_days"] which may be inflated by duplicate rows.
	totalDays = daysInMonth

	if att != nil {
		presentDays = toInt(att["present"])
		absentDays = toInt(att["absent"])
		lateDays = toInt(att["late"])
		leaveDays = toInt(att["leave"])
		holidayDays = toInt(att["holiday"])
		weekendDays = toInt(att["weekend"])
	}

	// Absent deduction
	absentDeduction := float64(0)
	if totalDays > 0 {
		perDaySalary := gross / float64(totalDays)
		absentDeduction = perDaySalary * float64(absentDays)
	}

	// Overtime — only when employee over_time_status is enabled
	otRate := float64(0)
	if emp.OverTimeStatus && daysInMonth > 0 {
		otRate = (basic / 208) * 2
	}
	otAmount := otHours * otRate

	// Attendance bonus by employee group
	attBonus := float64(0)
	if absentDays == 0 && presentDays > 0 {
		switch {
		case strings.EqualFold(groupName, "worker"):
			attBonus = 725
		case strings.EqualFold(groupName, "staff"):
			attBonus = 300
		default:
			attBonus = 0
		}
	}

	totalDeductions := absentDeduction
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
		AbsentDeduction:    absentDeduction,
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
