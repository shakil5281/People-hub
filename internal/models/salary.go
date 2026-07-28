package models

import (
	"time"

	"gorm.io/gorm"
)

type Salary struct {
	ID         string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  string `json:"company_id" gorm:"type:uuid;not null;uniqueIndex:idx_salary_emp_month"`
	EmployeeID string `json:"employee_id" gorm:"type:varchar(50);not null;uniqueIndex:idx_salary_emp_month"`
	Month      int    `json:"month" gorm:"not null;uniqueIndex:idx_salary_emp_month"`
	Year       int    `json:"year" gorm:"not null;uniqueIndex:idx_salary_emp_month"`

	BasicSalary        float64 `json:"basic_salary" gorm:"type:decimal(12,2);default:0"`
	HouseRent          float64 `json:"house_rent" gorm:"type:decimal(12,2);default:0"`
	MedicalAllowance   float64 `json:"medical_allowance" gorm:"type:decimal(12,2);default:0"`
	TransportAllowance float64 `json:"transport_allowance" gorm:"type:decimal(12,2);default:0"`
	FoodAllowance      float64 `json:"food_allowance" gorm:"type:decimal(12,2);default:0"`
	OtherAllowance     float64 `json:"other_allowance" gorm:"type:decimal(12,2);default:0"`
	GrossSalary        float64 `json:"gross_salary" gorm:"type:decimal(12,2);default:0"`

	ProvidentFund   float64 `json:"provident_fund" gorm:"type:decimal(12,2);default:0"`
	Tax             float64 `json:"tax" gorm:"type:decimal(12,2);default:0"`
	AbsentDeduction float64 `json:"absent_deduction" gorm:"type:decimal(12,2);default:0"`
	OtherDeduction  float64 `json:"other_deduction" gorm:"type:decimal(12,2);default:0"`
	TotalDeductions float64 `json:"total_deductions" gorm:"type:decimal(12,2);default:0"`

	OvertimeHours   float64 `json:"overtime_hours" gorm:"type:decimal(6,2);default:0"`
	OvertimeRate    float64 `json:"overtime_rate" gorm:"type:decimal(12,2);default:0"`
	OvertimeAmount  float64 `json:"overtime_amount" gorm:"type:decimal(12,2);default:0"`
	AttendanceBonus float64 `json:"attendance_bonus" gorm:"type:decimal(12,2);default:0"`
	NetSalary       float64 `json:"net_salary" gorm:"type:decimal(12,2);default:0"`

	PresentDays  int `json:"present_days" gorm:"default:0"`
	AbsentDays   int `json:"absent_days" gorm:"default:0"`
	LateDays     int `json:"late_days" gorm:"default:0"`
	LeaveDays    int `json:"leave_days" gorm:"default:0"`
	HolidayDays  int `json:"holiday_days" gorm:"default:0"`
	WeekendDays  int `json:"weekend_days" gorm:"default:0"`
	TotalDays    int `json:"total_days" gorm:"default:0"`

	Status    string         `json:"status" gorm:"type:varchar(20);default:processed"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy *string        `json:"created_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
