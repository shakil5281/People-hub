package handlers

type salarySheetLabels struct {
	Title string

	// Salary sheet column headers
	Sl             string
	EmployeeID     string
	Name           string
	Designation    string
	Department     string
	Section        string
	WorkingDays    string
	Present        string
	Absent         string
	Late           string
	Leave          string
	Holiday        string
	Weekend        string
	BasicSalary    string
	HouseRent      string
	Medical        string
	Transport      string
	Food           string
	OtherAllowance string
	Gross          string
	AbsentDed      string
	PF             string
	Tax            string
	TotalDeduction string
	OTHours        string
	OTRate         string
	OTAmount       string
	AttBonus       string
	NetSalary      string
	Status         string
	Signature      string

	// Summary export
	GrandTotal string
	TotalLabel string

	// Group-by labels
	GroupLabel string
	Employees  string
	GrossTotal string
	BasicTotal string
	HouseRentH string
	MedicalH   string
	TransportH string
	Deductions string
	NetTotal   string
}

var salarySheetEnLabels = salarySheetLabels{
	Title: "Salary Sheet",

	Sl:             "Sl",
	EmployeeID:     "Employee ID",
	Name:           "Name",
	Designation:    "Designation",
	Department:     "Department",
	Section:        "Section",
	WorkingDays:    "Working Days",
	Present:        "Present",
	Absent:         "Absent",
	Late:           "Late",
	Leave:          "Leave",
	Holiday:        "Holiday",
	Weekend:        "Weekend",
	BasicSalary:    "Basic Salary",
	HouseRent:      "House Rent",
	Medical:        "Medical",
	Transport:      "Transport",
	Food:           "Food",
	OtherAllowance: "Other",
	Gross:          "Gross",
	AbsentDed:      "Absent Deduction",
	PF:             "PF",
	Tax:            "Tax",
	TotalDeduction: "Total Deduction",
	OTHours:        "OT Hours",
	OTRate:         "OT Rate",
	OTAmount:       "OT Amount",
	AttBonus:       "Att. Bonus",
	NetSalary:      "Net Salary",
	Status:         "Status",
	Signature:      "Signature",

	GrandTotal: "Grand Total",
	TotalLabel: "Total",

	GroupLabel: "Group",
	Employees:  "Employees",
	GrossTotal: "Gross Total",
	BasicTotal: "Basic Total",
	HouseRentH: "House Rent",
	MedicalH:   "Medical",
	TransportH: "Transport",
	Deductions: "Deductions",
	NetTotal:   "Net Total",
}

var salarySheetBnLabels = salarySheetLabels{
	Title: "‡eZb ‡mvm©",

	Sl:             "µwgK bs",
	EmployeeID:     "AvBwW bv¤^vi",
	Name:           "bvg",
	Designation:    "c`ex",
	Department:     "wefvM",
	Section:        "‡mKkb",
	WorkingDays:    "‡gvU Kg©w`em",
	Present:        "Dcw¯'Z",
	Absent:         "Abycw¯'Z",
	Late:           "‡`wi",
	Leave:          "QzwU",
	Holiday:        "miKvwi QzwU",
	Weekend:        "mvßvwnK QzwU",
	BasicSalary:    "g~j †eZb",
	HouseRent:      "evwo fvov",
	Medical:        "wPwKrmv fvZv",
	Transport:      "hvZvhvZ fvZv",
	Food:           "Lvevi fvZv",
	OtherAllowance: "Ab¨vb¨ fvZv",
	Gross:          "‡gvU †eZb",
	AbsentDed:      "Abycw¯'Z KZ©b",
	PF:             "wcGd",
	Tax:            "Ki",
	TotalDeduction: "‡gvU KZ©b",
	OTHours:        "IwUf Uv",
	OTRate:         "IwU nvi",
	OTAmount:       "IwU cwigvY",
	AttBonus:       "Dcw¯'wZ †evbvm",
	NetSalary:      "wbU †eZb",
	Status:         "Ae¯'v",
	Signature:      "¯^v¶i",

	GrandTotal: "me©‡gvU",
	TotalLabel: "‡gvU",

	GroupLabel: "MÖæc",
	Employees:  "Kg©Pvix",
	GrossTotal: "‡gvU †eZb",
	BasicTotal: "g~j †eZb",
	HouseRentH: "evwo fvov",
	MedicalH:   "wPwKrmv",
	TransportH: "hvZvhvZ",
	Deductions: "KZ©b",
	NetTotal:   "wbU †eZb",
}

func getSalarySheetLabels(lang string) salarySheetLabels {
	if lang == "bn" {
		return salarySheetBnLabels
	}
	return salarySheetEnLabels
}

func bijoyLabels(l salarySheetLabels) salarySheetLabels {
	l.Title = toBijoy(l.Title)
	l.Sl = toBijoy(l.Sl)
	l.EmployeeID = toBijoy(l.EmployeeID)
	l.Name = toBijoy(l.Name)
	l.Designation = toBijoy(l.Designation)
	l.Department = toBijoy(l.Department)
	l.Section = toBijoy(l.Section)
	l.WorkingDays = toBijoy(l.WorkingDays)
	l.Present = toBijoy(l.Present)
	l.Absent = toBijoy(l.Absent)
	l.Late = toBijoy(l.Late)
	l.Leave = toBijoy(l.Leave)
	l.Holiday = toBijoy(l.Holiday)
	l.Weekend = toBijoy(l.Weekend)
	l.BasicSalary = toBijoy(l.BasicSalary)
	l.HouseRent = toBijoy(l.HouseRent)
	l.Medical = toBijoy(l.Medical)
	l.Transport = toBijoy(l.Transport)
	l.Food = toBijoy(l.Food)
	l.OtherAllowance = toBijoy(l.OtherAllowance)
	l.Gross = toBijoy(l.Gross)
	l.AbsentDed = toBijoy(l.AbsentDed)
	l.PF = toBijoy(l.PF)
	l.Tax = toBijoy(l.Tax)
	l.TotalDeduction = toBijoy(l.TotalDeduction)
	l.OTHours = toBijoy(l.OTHours)
	l.OTRate = toBijoy(l.OTRate)
	l.OTAmount = toBijoy(l.OTAmount)
	l.AttBonus = toBijoy(l.AttBonus)
	l.NetSalary = toBijoy(l.NetSalary)
	l.Status = toBijoy(l.Status)
	l.Signature = toBijoy(l.Signature)
	l.GrandTotal = toBijoy(l.GrandTotal)
	l.TotalLabel = toBijoy(l.TotalLabel)
	l.GroupLabel = toBijoy(l.GroupLabel)
	l.Employees = toBijoy(l.Employees)
	l.GrossTotal = toBijoy(l.GrossTotal)
	l.BasicTotal = toBijoy(l.BasicTotal)
	l.HouseRentH = toBijoy(l.HouseRentH)
	l.MedicalH = toBijoy(l.MedicalH)
	l.TransportH = toBijoy(l.TransportH)
	l.Deductions = toBijoy(l.Deductions)
	l.NetTotal = toBijoy(l.NetTotal)
	return l
}
