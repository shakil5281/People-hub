package service

import (
	"fmt"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

// Salary constants — fixed allowances
const (
	transportAllowance = 450
	foodAllowance      = 1250
	medicalAllowance   = 750
	attendanceBonus    = 500
)

type SalaryService struct {
	employeeRepo   *repository.EmployeeRepository
	attendanceRepo *repository.AttendanceRepository
	salaryRepo     *repository.SalaryRepository
}

func NewSalaryService(
	employeeRepo *repository.EmployeeRepository,
	attendanceRepo *repository.AttendanceRepository,
	salaryRepo *repository.SalaryRepository,
) *SalaryService {
	return &SalaryService{
		employeeRepo:   employeeRepo,
		attendanceRepo: attendanceRepo,
		salaryRepo:     salaryRepo,
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
//	OT Rate     = Basic / daysInMonth / 8
//	OT Amount   = OT Hours * OT Rate
//	Attendance Bonus = 500 if absent_days == 0 AND present_days > 0
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

	processed := 0

	for _, emp := range employees {
		salary := s.calculateEmployeeSalary(emp, attMap[emp.EmployeeID], otHoursMap[emp.EmployeeID], month, year, daysInMonth, userID)

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

	if att != nil {
		presentDays = toInt(att["present"])
		absentDays = toInt(att["absent"])
		lateDays = toInt(att["late"])
		leaveDays = toInt(att["leave"])
		holidayDays = toInt(att["holiday"])
		weekendDays = toInt(att["weekend"])
		totalDays = toInt(att["total_days"])
	}

	// Absent deduction
	absentDeduction := float64(0)
	if totalDays > 0 {
		perDaySalary := gross / float64(totalDays)
		absentDeduction = perDaySalary * float64(absentDays)
	}

	// Overtime
	otRate := float64(0)
	if daysInMonth > 0 {
		otRate = basic / float64(daysInMonth) / 8
	}
	otAmount := otHours * otRate

	// Attendance bonus
	attBonus := float64(0)
	if absentDays == 0 && presentDays > 0 {
		attBonus = attendanceBonus
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
