package models

import (
	"time"

	"gorm.io/gorm"
)

// EarnedLeavePolicy — configurable per company, date-based (§2)
type EarnedLeavePolicy struct {
	ID        string  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID string  `json:"company_id" gorm:"type:uuid;not null;index:idx_el_policy_company"`
	Name      string  `json:"name" gorm:"type:varchar(100);not null"`
	// Accrual
	AccrualRate      float64 `json:"accrual_rate" gorm:"type:decimal(5,2);default:1"`
	AccrualFrequency string  `json:"accrual_frequency" gorm:"type:varchar(20);default:MONTHLY"` // MONTHLY, YEARLY
	// Balance
	MaxBalance         float64 `json:"max_balance" gorm:"type:decimal(6,2);default:40"`
	MinServiceMonths   int     `json:"min_service_months" gorm:"default:0"`
	CarryForwardAllowed bool    `json:"carry_forward_allowed" gorm:"default:true"`
	CarryForwardLimit   float64 `json:"carry_forward_limit" gorm:"type:decimal(6,2);default:40"`
	// Encashment
	EncashmentAllowed bool   `json:"encashment_allowed" gorm:"default:true"`
	EncashmentBasis   string `json:"encashment_basis" gorm:"type:varchar(30);default:BASIC"` // BASIC, GROSS, BASIC_PLUS_ALLOWANCES, AVERAGE_3M
	EncashmentDivisor int    `json:"encashment_divisor" gorm:"default:30"`                    // 30 or 26
	RoundingRule      string `json:"rounding_rule" gorm:"type:varchar(20);default:ROUND"`    // FLOOR, ROUND, CEIL
	EffectiveFrom     string `json:"effective_from" gorm:"type:date;not null"`
	EffectiveTo       *string `json:"effective_to" gorm:"type:date"`
	Status            string `json:"status" gorm:"type:varchar(20);default:active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`

	Company Company `json:"company" gorm:"foreignKey:CompanyID"`
}

// EarnedLeaveLedger — append-only audit trail (§6)
type EarnedLeaveLedger struct {
	ID              string  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID       string  `json:"company_id" gorm:"type:uuid;not null;index:idx_el_ledger_company"`
	EmployeeID      string  `json:"employee_id" gorm:"type:varchar(50);not null;index:idx_el_ledger_emp_date"`
	TransactionDate string  `json:"transaction_date" gorm:"type:date;not null;index:idx_el_ledger_emp_date"`
	Period          string  `json:"period" gorm:"type:varchar(7);index:idx_el_ledger_period"` // YYYY-MM
	TransactionType string  `json:"transaction_type" gorm:"type:varchar(30);not null;index:idx_el_ledger_type"` // ACCRUAL, LEAVE_USED, ADJUSTMENT_ADD, ADJUSTMENT_DEDUCT, ENCASHMENT, EXPIRY, OPENING_BALANCE, REVERSAL, CARRY_FORWARD

	OpeningBalance float64 `json:"opening_balance" gorm:"type:decimal(6,2);default:0"`
	Accrued        float64 `json:"accrued" gorm:"type:decimal(6,2);default:0"`
	Used           float64 `json:"used" gorm:"type:decimal(6,2);default:0"`
	Adjusted       float64 `json:"adjusted" gorm:"type:decimal(6,2);default:0"` // + for ADD, - for DEDUCT stored as signed
	Encashed       float64 `json:"encashed" gorm:"type:decimal(6,2);default:0"`
	Expired        float64 `json:"expired" gorm:"type:decimal(6,2);default:0"`
	ClosingBalance float64 `json:"closing_balance" gorm:"type:decimal(6,2);default:0"`

	ReferenceID   *string `json:"reference_id" gorm:"type:varchar(50)"`
	ReferenceType string  `json:"reference_type" gorm:"type:varchar(30)"`
	Remarks       string  `json:"remarks" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy *string        `json:"created_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}

// EarnedLeaveBalance — materialized per employee per year for fast reads (§7)
type EarnedLeaveBalance struct {
	ID         string  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  string  `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID string  `json:"employee_id" gorm:"type:varchar(50);not null;uniqueIndex:ux_el_balance_emp_year"`
	Year       int     `json:"year" gorm:"not null;uniqueIndex:ux_el_balance_emp_year"`
	Period     string  `json:"period" gorm:"type:varchar(4)"` // YYYY

	Opening  float64 `json:"opening" gorm:"type:decimal(6,2);default:0"`
	Accrued  float64 `json:"accrued" gorm:"type:decimal(6,2);default:0"`
	Used     float64 `json:"used" gorm:"type:decimal(6,2);default:0"`
	Adjusted float64 `json:"adjusted" gorm:"type:decimal(6,2);default:0"`
	Encashed float64 `json:"encashed" gorm:"type:decimal(6,2);default:0"`
	Expired  float64 `json:"expired" gorm:"type:decimal(6,2);default:0"`
	Closing  float64 `json:"closing" gorm:"type:decimal(6,2);default:0"`

	LastLedgerID *string `json:"last_ledger_id" gorm:"type:uuid"`

	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// EarnedLeaveSalarySheet — header (§15-16)
type EarnedLeaveSalarySheet struct {
	ID          string  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   string  `json:"company_id" gorm:"type:uuid;not null;uniqueIndex:ux_el_sheet_company_period"`
	PeriodMonth int     `json:"period_month" gorm:"not null;uniqueIndex:ux_el_sheet_company_period"`
	PeriodYear  int     `json:"period_year" gorm:"not null;uniqueIndex:ux_el_sheet_company_period"`
	Period      string  `json:"period" gorm:"type:varchar(7);not null;index:idx_el_sheet_period"` // YYYY-MM
	Status      string  `json:"status" gorm:"type:varchar(20);default:DRAFT"`                    // DRAFT, CALCULATED, VERIFIED, APPROVED, FINALIZED, REVERSED

	TotalEmployees int     `json:"total_employees"`
	TotalDays      float64 `json:"total_days" gorm:"type:decimal(8,2)"`
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(12,2)"`

	CreatedBy   *string    `json:"created_by" gorm:"type:uuid"`
	ApprovedBy  *string    `json:"approved_by" gorm:"type:uuid"`
	FinalizedBy *string    `json:"finalized_by" gorm:"type:uuid"`
	ApprovedAt  *time.Time `json:"approved_at"`
	FinalizedAt *time.Time `json:"finalized_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Company Company `json:"company" gorm:"foreignKey:CompanyID"`
}

// EarnedLeaveSalaryItem — per employee (§15)
type EarnedLeaveSalaryItem struct {
	ID        string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SheetID   string `json:"sheet_id" gorm:"type:uuid;not null;index:idx_el_item_sheet;uniqueIndex:ux_el_item_sheet_employee"`
	CompanyID string `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID string `json:"employee_id" gorm:"type:varchar(50);not null;uniqueIndex:ux_el_item_sheet_employee"`

	Opening float64 `json:"opening" gorm:"type:decimal(6,2)"`
	Accrued float64 `json:"accrued" gorm:"type:decimal(6,2)"`
	Used    float64 `json:"used" gorm:"type:decimal(6,2)"`
	Closing float64 `json:"closing" gorm:"type:decimal(6,2)"`

	EncashableDays float64 `json:"encashable_days" gorm:"type:decimal(6,2)"`
	SalaryBasis    float64 `json:"salary_basis" gorm:"type:decimal(12,2)"`
	Rate           float64 `json:"rate" gorm:"type:decimal(12,2)"`
	Amount         float64 `json:"amount" gorm:"type:decimal(12,2)"`

	Remarks string `json:"remarks" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Sheet    EarnedLeaveSalarySheet `json:"sheet" gorm:"foreignKey:SheetID"`
	Employee Employee               `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
}
