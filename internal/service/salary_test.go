package service

import (
	"testing"

	"github.com/shakil5281/peoplehub-api/internal/models"
)

func TestCalculateEmployeeSalary_LateDeductions(t *testing.T) {
	s := &SalaryService{}
	emp := models.Employee{
		CompanyID:      "comp-1",
		EmployeeID:     "1603",
		GrossSalary:    30000,
		OverTimeStatus: true,
	}

	daysInMonth := 30

	tests := []struct {
		name                 string
		lateDays             int
		absentDays           int
		expectedOtherDeduct  float64
		expectedAbsentDeduct float64
		expectedAttBonus     float64
	}{
		{
			name:                 "0 Late Days -> No Deduction, Full Bonus",
			lateDays:             0,
			absentDays:           0,
			expectedOtherDeduct:  0,
			expectedAbsentDeduct: 0,
			expectedAttBonus:     725,
		},
		{
			name:                 "3 Late Days -> 1 Day Salary Deduction in OtherDeduction, Bonus Retained",
			lateDays:             3,
			absentDays:           0,
			expectedOtherDeduct:  1000,
			expectedAbsentDeduct: 0,
			expectedAttBonus:     725,
		},
		{
			name:                 "4 Late Days -> 1 Day Salary Deduction in OtherDeduction, Attendance Bonus = 0",
			lateDays:             4,
			absentDays:           0,
			expectedOtherDeduct:  1000,
			expectedAbsentDeduct: 0,
			expectedAttBonus:     0,
		},
		{
			name:                 "5 Late Days -> 1 Day Salary Deduction in OtherDeduction, Attendance Bonus = 0",
			lateDays:             5,
			absentDays:           0,
			expectedOtherDeduct:  1000,
			expectedAbsentDeduct: 0,
			expectedAttBonus:     0,
		},
		{
			name:                 "6 Late Days -> 2 Days Salary Deduction in OtherDeduction, Attendance Bonus = 0",
			lateDays:             6,
			absentDays:           0,
			expectedOtherDeduct:  2000,
			expectedAbsentDeduct: 0,
			expectedAttBonus:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			att := map[string]interface{}{
				"present": 30 - tt.absentDays,
				"absent":  tt.absentDays,
				"late":    tt.lateDays,
			}
			res := s.calculateEmployeeSalary(emp, "Worker", att, 0, 0, 8, 2026, daysInMonth, "", "user-1")

			if res.OtherDeduction != tt.expectedOtherDeduct {
				t.Errorf("OtherDeduction got = %v, want = %v", res.OtherDeduction, tt.expectedOtherDeduct)
			}
			if res.AbsentDeduction != tt.expectedAbsentDeduct {
				t.Errorf("AbsentDeduction got = %v, want = %v", res.AbsentDeduction, tt.expectedAbsentDeduct)
			}
			if res.AttendanceBonus != tt.expectedAttBonus {
				t.Errorf("AttendanceBonus got = %v, want = %v", res.AttendanceBonus, tt.expectedAttBonus)
			}
		})
	}
}

func TestCalculateEmployeeSalary_PartialMonthPaidDays(t *testing.T) {
	s := &SalaryService{}
	emp := models.Employee{
		CompanyID:      "comp-1",
		EmployeeID:     "1001",
		GrossSalary:    9875,
		OverTimeStatus: true,
	}

	daysInMonth := 31 // July 2026

	att := map[string]interface{}{
		"present": 12,
		"late":    1,
		"weekend": 1,
	}

	res := s.calculateEmployeeSalary(emp, "Worker", att, 0, 0, 7, 2026, daysInMonth, "", "user-1")

	expectedAbsentDays := 17 // 31 - (12 + 1 + 1)
	if res.AbsentDays != expectedAbsentDays {
		t.Errorf("AbsentDays got = %v, want = %v", res.AbsentDays, expectedAbsentDays)
	}

	expectedNetSalary := (9875.0 / 31.0) * 14.0 // (Gross / 31) * (Weekend + Late + Present)
	diff := res.NetSalary - expectedNetSalary
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("NetSalary got = %v, want = %v", res.NetSalary, expectedNetSalary)
	}
}

func TestCalculateEmployeeSalary_SeparationDate(t *testing.T) {
	s := &SalaryService{}
	emp := models.Employee{
		CompanyID:      "comp-1",
		EmployeeID:     "1002",
		GrossSalary:    31000,
		OverTimeStatus: true,
	}

	daysInMonth := 31 // August 2026
	sepDate := "2026-08-15" // Separated on August 15

	// Worked 13 present days + 2 weekend days = 15 paid days out of 15 days until separation
	att := map[string]interface{}{
		"present": 13,
		"weekend": 2,
		"late":    0,
		"absent":  0,
	}

	res := s.calculateEmployeeSalary(emp, "Worker", att, 0, 0, 8, 2026, daysInMonth, sepDate, "user-1")

	// Calendar Working Days is always daysInMonth (31) per new spec
	if res.TotalDays != 31 {
		t.Errorf("TotalDays got = %v, want = 31", res.TotalDays)
	}

	// Calendar absent = 31 - 15 paid = 16 (days after separation count as absent)
	if res.AbsentDays != 16 {
		t.Errorf("AbsentDays got = %v, want = 16", res.AbsentDays)
	}

	// Absent deduction = (31000 / 31) * 16 = 16000 (calendar)
	expectedAbsentDeduct := (31000.0 / 31.0) * 16.0
	if res.AbsentDeduction != expectedAbsentDeduct {
		t.Errorf("AbsentDeduction got = %v, want = %v", res.AbsentDeduction, expectedAbsentDeduct)
	}

	// Calendar bonus: absent 16 => no bonus (window bonus would be 725, but calendar has absent)
	if res.AttendanceBonus != 0 {
		t.Errorf("AttendanceBonus got = %v, want = 0", res.AttendanceBonus)
	}

	// NetSalary = Gross (31000) - 16000 + 0 = 15000 (calendar, no bonus)
	expectedNetSalary := 31000.0 - expectedAbsentDeduct
	if res.NetSalary != expectedNetSalary {
		t.Errorf("NetSalary got = %v, want = %v", res.NetSalary, expectedNetSalary)
	}
}
