package models

import (
	"time"

	"gorm.io/gorm"
)

type Employee struct {
	ID        string  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    *string `json:"user_id" gorm:"type:uuid"`
	CompanyID string  `json:"company_id" gorm:"type:uuid;not null"`

	// Personal Information
	NameEn           string `json:"name_en" gorm:"type:varchar(255)"`
	NameBn           string `json:"name_bn" gorm:"type:varchar(255)"`
	FatherName       string `json:"father_name" gorm:"type:varchar(255)"`
	MotherName       string `json:"mother_name" gorm:"type:varchar(255)"`
	DateOfBirth      string `json:"date_of_birth" gorm:"type:varchar(10)"`
	Gender           string `json:"gender" gorm:"type:varchar(20)"`
	BloodGroup       string `json:"blood_group" gorm:"type:varchar(5)"`
	MaritalStatus    string `json:"marital_status" gorm:"type:varchar(20)"`
	Religion         string `json:"religion" gorm:"type:varchar(50)"`
	Nationality      string `json:"nationality" gorm:"type:varchar(100);default:'Bangladeshi'"`
	NID              string `json:"nid" gorm:"type:varchar(50)"`
	Phone            string `json:"phone" gorm:"type:varchar(20)"`
	Email            string `json:"email" gorm:"type:varchar(255)"`
	PresentAddress   string `json:"present_address" gorm:"type:text"`
	PermanentAddress string `json:"permanent_address" gorm:"type:text"`

	// Family Information
	SpouseName         string `json:"spouse_name" gorm:"type:varchar(255)"`
	EmergencyContact   string `json:"emergency_contact" gorm:"type:varchar(255)"`
	EmergencyPhone     string `json:"emergency_phone" gorm:"type:varchar(20)"`
	NumberOfDependents int    `json:"number_of_dependents" gorm:"default:0"`

	// Office Information
	DepartmentID  *string   `json:"department_id" gorm:"type:uuid"`
	SectionID     *string   `json:"section_id" gorm:"type:uuid"`
	DesignationID *string   `json:"designation_id" gorm:"type:uuid"`
	LineID        *string   `json:"line_id" gorm:"type:uuid"`
	GroupID       *string   `json:"group_id" gorm:"type:uuid"`
	FloorID       *string   `json:"floor_id" gorm:"type:uuid"`
	EmployeeID    string    `json:"employee_id" gorm:"column:employee_id;type:varchar(50);uniqueIndex;not null"`
	PunchNumber   string    `json:"punch_number" gorm:"type:varchar(50);uniqueIndex;not null"`
	EmployeeType  string    `json:"employee_type" gorm:"type:varchar(50)"`
	Grade         string    `json:"grade" gorm:"type:varchar(50)"`
	JoiningDate   time.Time `json:"joining_date" gorm:"not null"`
	ShiftID       *string   `json:"shift_id" gorm:"type:uuid"`
	ReportsTo     *string   `json:"reports_to" gorm:"type:uuid"`

	// Address (present)
	PresentDivisionID *string `json:"present_division_id" gorm:"type:uuid"`
	PresentDistrictID *string `json:"present_district_id" gorm:"type:uuid"`
	PresentUpazilaID  *string `json:"present_upazila_id" gorm:"type:uuid"`
	PresentUnionID    *string `json:"present_union_id" gorm:"type:uuid"`
	// Address (permanent)
	PermanentDivisionID *string `json:"permanent_division_id" gorm:"type:uuid"`
	PermanentDistrictID *string `json:"permanent_district_id" gorm:"type:uuid"`
	PermanentUpazilaID  *string `json:"permanent_upazila_id" gorm:"type:uuid"`
	PermanentUnionID    *string `json:"permanent_union_id" gorm:"type:uuid"`

	// Salary Information
	GrossSalary        float64 `json:"gross_salary" gorm:"type:decimal(12,2)"`
	BasicSalary        float64 `json:"basic_salary" gorm:"type:decimal(12,2)"`
	HouseRent          float64 `json:"house_rent" gorm:"type:decimal(12,2)"`
	TransportAllowance float64 `json:"transport_allowance" gorm:"type:decimal(12,2);default:450"`
	FoodAllowance      float64 `json:"food_allowance" gorm:"type:decimal(12,2);default:1250"`
	MedicalAllowance   float64 `json:"medical_allowance" gorm:"type:decimal(12,2);default:750"`
	OtherAllowance     float64 `json:"other_allowance" gorm:"type:decimal(12,2)"`
	// Account Information
	AccountType   string `json:"account_type" gorm:"type:varchar(20)"`
	AccountNumber string `json:"account_number" gorm:"type:varchar(20)"`

	// Status
	Status         string `json:"status" gorm:"type:varchar(20);default:active"`
	OverTimeStatus bool   `json:"over_time_status" gorm:"default:false"`

	// Signature & Image
	SignatureURL string `json:"signature_url" gorm:"type:varchar(500)"`
	ImageURL     string `json:"image_url" gorm:"type:varchar(500)"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`
	DeletedBy *string        `json:"deleted_by" gorm:"type:uuid"`

	// Relations
	User              *User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Company           Company           `json:"company" gorm:"foreignKey:CompanyID"`
	Department        *Department       `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	SectionRef        *Section          `json:"section_ref,omitempty" gorm:"foreignKey:SectionID"`
	DesignationRef    *Designation      `json:"designation_ref,omitempty" gorm:"foreignKey:DesignationID"`
	LineRef           *Line             `json:"line_ref,omitempty" gorm:"foreignKey:LineID"`
	GroupRef          *Group            `json:"group_ref,omitempty" gorm:"foreignKey:GroupID"`
	FloorRef          *Floor            `json:"floor_ref,omitempty" gorm:"foreignKey:FloorID"`
	Shift             *Shift            `json:"shift,omitempty" gorm:"foreignKey:ShiftID"`
	Manager           *Employee         `json:"manager,omitempty" gorm:"foreignKey:ReportsTo"`
	PresentDivision   *Division         `json:"present_division,omitempty" gorm:"foreignKey:PresentDivisionID"`
	PresentDistrict   *District         `json:"present_district,omitempty" gorm:"foreignKey:PresentDistrictID"`
	PresentUpazila    *Upazila          `json:"present_upazila,omitempty" gorm:"foreignKey:PresentUpazilaID"`
	PresentUnion      *Union            `json:"present_union,omitempty" gorm:"foreignKey:PresentUnionID"`
	PermanentDivision *Division         `json:"permanent_division,omitempty" gorm:"foreignKey:PermanentDivisionID"`
	PermanentDistrict *District         `json:"permanent_district,omitempty" gorm:"foreignKey:PermanentDistrictID"`
	PermanentUpazila  *Upazila          `json:"permanent_upazila,omitempty" gorm:"foreignKey:PermanentUpazilaID"`
	PermanentUnion    *Union            `json:"permanent_union,omitempty" gorm:"foreignKey:PermanentUnionID"`
	Attendances       []Attendance      `gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Leaves            []Leave           `gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	LeaveAllocations  []LeaveAllocation `gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Salaries          []Salary          `gorm:"foreignKey:EmployeeID;references:EmployeeID"`
}
