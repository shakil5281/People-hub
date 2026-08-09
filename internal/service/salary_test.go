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
			res := s.calculateEmployeeSalary(emp, "Worker", att, 0, 8, 2026, daysInMonth, "user-1")

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
