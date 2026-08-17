package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type leaveFormLabels struct {
	SystemTitle string
	FormTitle   string

	Company string
	Branch  string
	AppNo   string
	AppDate string

	EmployeeInfo string
	EmployeeID   string
	CardNo       string
	Name         string
	Mobile       string
	Department   string
	Section      string
	Designation  string
	Grade        string
	Shift        string
	JoiningDate  string
	ReportsTo    string
	EmpType      string

	LeaveDetails   string
	LeaveTypeLabel string
	FromDate       string
	ToDate         string
	TotalDays      string
	HalfDay        string
	HalfDayYes     string
	HalfDayNo      string
	AddressDuring  string
	EmergencyPhone string
	Reason         string

	LeaveBalance string
	BalLeaveType string
	BalEntitled  string
	BalUsed      string
	BalRemaining string

	Handover        string
	HandoverTo      string
	HandoverDept    string
	HandoverDesig   string
	HandoverDetails string

	Approval       string
	EmployeeRole   string
	SupervisorRole string
	DeptHeadRole   string
	HRRole         string
	FinalRole      string
	Approved       string
	Rejected       string
	Remarks        string
	Signature      string
	Date           string
	BalanceUpdated string
	Yes            string
	No             string
	HRManagerSig   string

	Notes string
	Note1 string
	Note2 string
	Note3 string

	GeneratedBy string
	DocumentNo  string
	PrintDate   string
}

var leaveEnLabels = leaveFormLabels{
	SystemTitle: "HR & PAYROLL MANAGEMENT SYSTEM",
	FormTitle:   "EMPLOYEE LEAVE APPLICATION FORM",

	Company: "Company",
	Branch:  "Branch",
	AppNo:   "Application No",
	AppDate: "Application Date",

	EmployeeInfo: "EMPLOYEE INFORMATION",
	EmployeeID:   "Employee ID",
	CardNo:       "Card No",
	Name:         "Employee Name",
	Mobile:       "Mobile",
	Department:   "Department",
	Section:      "Section",
	Designation:  "Designation",
	Grade:        "Grade",
	Shift:        "Shift",
	JoiningDate:  "Joining Date",
	ReportsTo:    "Supervisor",
	EmpType:      "Employment Type",

	LeaveDetails:   "LEAVE DETAILS",
	LeaveTypeLabel: "Leave Type",
	FromDate:       "Leave From",
	ToDate:         "Leave To",
	TotalDays:      "Total Days",
	HalfDay:        "Half Day",
	HalfDayYes:     "Yes",
	HalfDayNo:      "No",
	AddressDuring:  "Address During Leave",
	EmergencyPhone: "Emergency Contact",
	Reason:         "Reason for Leave",

	LeaveBalance: "LEAVE BALANCE",
	BalLeaveType: "Leave Type",
	BalEntitled:  "Entitled",
	BalUsed:      "Used",
	BalRemaining: "Remaining",

	Handover:        "WORK HANDOVER INFORMATION",
	HandoverTo:      "Assigned Employee",
	HandoverDept:    "Department",
	HandoverDesig:   "Designation",
	HandoverDetails: "Pending Work Details",

	Approval:       "APPROVAL WORKFLOW",
	EmployeeRole:   "Employee",
	SupervisorRole: "Supervisor",
	DeptHeadRole:   "Department Head",
	HRRole:         "HR Department",
	FinalRole:      "Final Approval",
	Approved:       "Approved",
	Rejected:       "Rejected",
	Remarks:        "Remarks",
	Signature:      "Signature",
	Date:           "Date",
	BalanceUpdated: "Balance Updated",
	Yes:            "Yes",
	No:             "No",
	HRManagerSig:   "HR Manager Sig",

	Notes: "Notes",
	Note1: "Leave must be applied before scheduled date unless emergency.",
	Note2: "Medical certificate mandatory for Sick Leave exceeding policy.",
	Note3: "Leave approval is subject to company rules & requirements.",

	GeneratedBy: "System Generated",
	DocumentNo:  "Document No",
	PrintDate:   "Print Date",
}

var leaveBnLabels = leaveFormLabels{
	SystemTitle: "GBPAvi A¨vÛ ceIivj wm‡÷g",
	FormTitle:   "QzwUi Av‡e`bcÎ (LEAVE APPLICATION)",

	Company: "tKv¤úvbxi bvg",
	Branch:  "kvLv/KviLvbv",
	AppNo:   "Av‡eb b¤^vi",
	AppDate: "Av‡eb ZvwiL",

	EmployeeInfo: "Kg©Pvixi Z_¨",
	EmployeeID:   "Kg©x AvBWw",
	CardNo:       "KviW b¤^vi",
	Name:         "bvg",
	Mobile:       "gvevBj",
	Department:   "wefvM",
	Section:      "tmKkb",
	Designation:  "c`ex",
	Grade:        "tMÖW",
	Shift:        "wkdU",
	JoiningDate:  "PvKzix‡Z tvevMv‡b",
	ReportsTo:    "wicwU©s Avwdmvi",
	EmpType:      "PvKzixi aiY",

	LeaveDetails:   "QywUi weeiY",
	LeaveTypeLabel: "QywUi aiY",
	FromDate:       "QywU kyiei",
	ToDate:         "tkl ZvwiL",
	TotalDays:      "tgvU QywU",
	HalfDay:        "AR©wjevm",
	HalfDayYes:     "n¨uv",
	HalfDayNo:      "bv",
	AddressDuring:  "QywUKvjxb wVKvbv",
	EmergencyPhone: "jeiyix gvevBj",
	Reason:         "QywUi KviY",

	LeaveBalance: "QywUi wnmei",
	BalLeaveType: "QywUi aiY",
	BalEntitled:  "cÖvc¨",
	BalUsed:      "eeüZ",
	BalRemaining: "Aewkó",

	Handover:        "QywUi c~‡e© `vwqZ¡ n¯ÍvbÍi",
	HandoverTo:      "n¯ÍvbÍiKZ©v Kg©Pvix",
	HandoverDept:    "wefvM",
	HandoverDesig:   "c`ex",
	HandoverDetails: "eeKvqv Kv‡ji weeiY",

	Approval:       "Abzgveb cÖwµqv",
	EmployeeRole:   "Kg©Pvix",
	SupervisorRole: "mycvifvBRvi",
	DeptHeadRole:   "wefvMxq cÖavb",
	HRRole:         "GBPAvi wefvM",
	FinalRole:      "P~ovšÍ Abzgveb",
	Approved:       "AbzgvweZ",
	Rejected:       "bvKoc",
	Remarks:        "gšÍee¨",
	Signature:      "¯^v¶i",
	Date:           "ZvwiL",
	BalanceUpdated: "wnmei Avcbve‡UW",
	Yes:            "n¨uv",
	No:             "bv",
	HRManagerSig:   "GBPAvi g¨vtbRvi ¯^v¶i",

	Notes: "we‡kl `ªóee¨",
	Note1: "jeiyix KviY eeZxZ QywU kyiei c~‡e© Av‡eb Ki‡Z n‡e|",
	Note2: "wPwKrmv QywUi Rb¨ Wv³vix mvwd©wd‡KU jvgvewb|",
	Note3: "QywU Abzgveb tKv¤úvbxi wbqgvfjx mvcxb|",

	GeneratedBy: "System Generated",
	DocumentNo:  "bw_ b¤^vi",
	PrintDate:   "wcÖ‡›Ui ZvwiL",
}

type leaveBalanceRow struct {
	LeaveType string
	Entitled  string
	Used      string
	Remaining string
}

type leaveFormData struct {
	BrandName string

	Company string
	Branch  string
	AppNo   string
	AppDate string

	EmployeeID  string
	CardNo      string
	Name        string
	Mobile      string
	Department  string
	Section     string
	Designation string
	Grade       string
	Shift       string
	JoiningDate string
	ReportsTo   string
	EmpType     string

	LeaveType        string
	LeaveTypeOptions []string
	LeaveTypeActive  int
	FromDate         string
	ToDate           string
	TotalDays        string
	AddressDuring    string
	EmergencyPhone   string
	Reason           string

	Balances []leaveBalanceRow

	HandoverTo      string
	HandoverDept    string
	HandoverDesig   string
	HandoverDetails string

	Status      string
	ApprovedYes bool
	RejectedYes bool

	Notes []string

	GeneratedBy string
	DocumentNo  string
	PrintDate   string
}

func bnDigits(s string) string {
	digits := map[rune]string{
		'0': "০", '1': "১", '2': "২", '3': "৩", '4': "৪",
		'5': "৫", '6': "৬", '7': "৭", '8': "৮", '9': "৯",
	}
	var b strings.Builder
	for _, r := range s {
		if d, ok := digits[r]; ok {
			b.WriteString(d)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatLeaveFormDate(s, lang string) string {
	if s == "" || s == "-" {
		return "-"
	}
	if len(s) >= 10 {
		sub := s[:10]
		if t, err := time.Parse("2006-01-02", sub); err == nil {
			return t.Format("02/01/2006")
		}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format("02/01/2006")
	}
	return s
}

func leaveAppNo(id, fromDate string) string {
	yearMonth := ""
	if len(fromDate) >= 7 {
		yearMonth = fromDate[:4] + fromDate[5:7]
	}
	hex := strings.ReplaceAll(id, "-", "")
	var n uint64
	if len(hex) >= 8 {
		n, _ = strconv.ParseUint(hex[len(hex)-8:], 16, 64)
	} else {
		n, _ = strconv.ParseUint(hex, 16, 64)
	}
	return fmt.Sprintf("LA-%s-%06d", yearMonth, n%1000000)
}

func buildLeaveFormData(leave *models.Leave, lang string, labels leaveFormLabels, allocs []models.LeaveAllocation) leaveFormData {
	emp := &leave.Employee

	d := leaveFormData{
		BrandName: "",
		Company:   companyNameFor(lang, leave.Company),
		Branch:    companyAddress(leave.Company, lang),
		AppNo:     leaveAppNo(leave.ID, leave.FromDate),
		AppDate:   formatLeaveFormDate(time.Now().Format("2006-01-02"), lang),

		EmployeeID:  orDash(emp.EmployeeID),
		CardNo:      orDash(emp.PunchNumber),
		Name:        orDash(employeeNameFor(lang, emp)),
		Mobile:      orDash(emp.Phone),
		Department:  orDash(departmentName(emp.Department, lang)),
		Section:     orDash(sectionName(emp.SectionRef, lang)),
		Designation: orDash(designationName(emp.DesignationRef, lang)),
		Grade:       orDash(emp.Grade),
		Shift:       orDash(shiftName(emp.Shift, lang)),
		JoiningDate: orDash(formatLeaveFormDate(emp.JoiningDate.Format("2006-01-02"), lang)),
		EmpType:     orDash(emp.EmployeeType),
		Status:      leave.Status,

		LeaveType:      orDash(leave.LeaveType.Name),
		FromDate:       formatLeaveFormDate(leave.FromDate, lang),
		ToDate:         formatLeaveFormDate(leave.ToDate, lang),
		TotalDays:      fmt.Sprint(leave.TotalDays),
		AddressDuring:  orDash(emp.PresentAddress),
		EmergencyPhone: orDash(emp.EmergencyPhone),
		Reason:         leave.Reason,

		Notes: []string{labels.Note1, labels.Note2, labels.Note3},

		GeneratedBy: labels.GeneratedBy,
		DocumentNo:  leaveAppNo(leave.ID, leave.FromDate),
		PrintDate:   formatLeaveFormDate(time.Now().Format("2006-01-02"), lang),
	}

	if emp.Manager != nil {
		d.ReportsTo = orDash(employeeNameFor(lang, emp.Manager))
	} else {
		d.ReportsTo = "-"
	}

	if lang == "bn" {
		d.Company = utils.UnicodeToBijoy(d.Company)
		d.Branch = utils.UnicodeToBijoy(d.Branch)
		d.Name = utils.UnicodeToBijoy(d.Name)
		d.Department = utils.UnicodeToBijoy(d.Department)
		d.Section = utils.UnicodeToBijoy(d.Section)
		d.Designation = utils.UnicodeToBijoy(d.Designation)
		d.Grade = utils.UnicodeToBijoy(d.Grade)
		d.Shift = utils.UnicodeToBijoy(d.Shift)
		d.EmpType = utils.UnicodeToBijoy(d.EmpType)
		d.ReportsTo = utils.UnicodeToBijoy(d.ReportsTo)
		d.LeaveType = utils.UnicodeToBijoy(d.LeaveType)
		d.AddressDuring = utils.UnicodeToBijoy(d.AddressDuring)
		d.Reason = utils.UnicodeToBijoy(d.Reason)
		d.LeaveTypeOptions = []string{"tbmwgvwZK (Casual)", "AjyeZv (Sick)", "evrmwiK (Annual)", "gvZ„Z¡Kvjxb", "wcZ„Z¡Kvjxb", "tveZbewnxb", "jeiyix", "Ab¨vb¨"}
		for i := range d.Balances {
			d.Balances[i].LeaveType = utils.UnicodeToBijoy(d.Balances[i].LeaveType)
		}
	} else {
		d.LeaveTypeOptions = []string{"Casual", "Sick", "Annual", "Maternity", "Without Pay"}
	}
	d.LeaveTypeActive = -1
	lt := strings.ToLower(leave.LeaveType.Name)
	for i, opt := range d.LeaveTypeOptions {
		key := strings.ToLower(opt)
		if strings.Contains(key, "casual") && strings.Contains(lt, "casual") {
			d.LeaveTypeActive = i
			break
		}
		if strings.Contains(key, "sick") && strings.Contains(lt, "sick") {
			d.LeaveTypeActive = i
			break
		}
		if strings.Contains(key, "annual") && (strings.Contains(lt, "annual") || strings.Contains(lt, "বেতন")) && i == 2 {
			d.LeaveTypeActive = i
			break
		}
		if strings.Contains(key, "maternity") && strings.Contains(lt, "maternity") {
			d.LeaveTypeActive = i
			break
		}
		if strings.Contains(key, "paternity") && strings.Contains(lt, "paternity") {
			d.LeaveTypeActive = i
			break
		}
		if strings.Contains(key, "without pay") && strings.Contains(lt, "without pay") {
			d.LeaveTypeActive = i
			break
		}
	}

	for _, a := range allocs {
		remaining := a.TotalDays - a.UsedDays - a.PendingDays
		d.Balances = append(d.Balances, leaveBalanceRow{
			LeaveType: orDash(a.LeaveType.Name),
			Entitled:  fmt.Sprint(a.TotalDays),
			Used:      fmt.Sprint(a.UsedDays),
			Remaining: fmt.Sprint(remaining),
		})
	}
	if len(d.Balances) == 0 {
		d.Balances = []leaveBalanceRow{
			{"Annual Leave", "14", "0", "14"},
			{"Casual Leave", "10", "0", "10"},
			{"Sick Leave", "14", "0", "14"},
		}
		if lang == "bn" {
			d.Balances = []leaveBalanceRow{
				{"বাৎসরিক ছুটি", "14", "0", "14"},
				{"নৈমিত্তিক ছুটি", "10", "0", "10"},
				{"অসুস্থতার ছুটি", "14", "0", "14"},
			}
		}
	}

	d.ApprovedYes = leave.Status == "approved"
	d.RejectedYes = leave.Status == "rejected"

	return d
}

func companyNameFor(lang string, c models.Company) string {
	if lang == "bn" && c.CompanyNameBn != "" {
		return c.CompanyNameBn
	}
	if c.CompanyNameEn != "" {
		return c.CompanyNameEn
	}
	return "-"
}

func employeeNameFor(lang string, e *models.Employee) string {
	if e == nil {
		return ""
	}
	if lang == "bn" && e.NameBn != "" {
		return e.NameBn
	}
	return e.NameEn
}

func leaveFormFont(pdf *gofpdf.Fpdf, lang string) string {
	if lang == "bn" {
		fn := loadBanglaFont(pdf)
		if fn != "" && fn != "Arial" {
			return fn
		}
	}
	return "Arial"
}

func drawDottedUnderline(pdf *gofpdf.Fpdf, x, y, width float64) {
	pdf.SetDrawColor(120, 120, 120)
	pdf.SetLineWidth(0.3)
	pdf.SetDashPattern([]float64{0.5, 0.7}, 0)
	pdf.Line(x, y, x+width, y)
	pdf.SetDashPattern([]float64{}, 0)
}

func renderLeaveFormPDFPage(pdf *gofpdf.Fpdf, font string, lang string, data leaveFormData, labels leaveFormLabels) {
	isBn := lang == "bn"
	x := 12.0
	w := 186.0
	curY := 10.0

	pdf.SetDrawColor(80, 80, 80)
	pdf.SetLineWidth(0.3)

	// Helper for text formatting
	font = leaveFormFont(pdf, lang)

	// Clean helper for English font to avoid encoding crashes
	clean := func(s string) string {
		if isBn {
			return s
		}
		var b strings.Builder
		for _, r := range s {
			if r <= 255 {
				b.WriteRune(r)
			}
		}
		return b.String()
	}

	// ---- 1. TOP HEADER ----
	pdf.SetTextColor(30, 30, 30)

	// Company Title
	pdf.SetFont(font, "B", 18.0)
	pdf.SetXY(x, curY)
	pdf.CellFormat(w, 6.0, clean(data.Company), "", 0, "C", false, 0, "")
	curY += 7.5

	// Company Address
	pdf.SetFont(font, "", 12.5)
	pdf.SetXY(x, curY)
	pdf.CellFormat(w, 4.5, clean(data.Branch), "", 0, "C", false, 0, "")
	curY += 6.5

	// Form Title
	pdf.SetFont(font, "B", 16.5)
	formTitle := "QzwUi Av‡e`bcÎ"
	if !isBn {
		formTitle = "LEAVE APPLICATION FORM"
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(w, 5.5, formTitle, "", 0, "C", false, 0, "")

	// Application Date (Right Side)
	pdf.SetFont(font, "", 12.5)
	dateLabel := "ZvwiL t " + data.AppDate
	if !isBn {
		dateLabel = "Date : " + data.AppDate
	}
	pdf.SetXY(x+w-65.0, curY+1.0)
	pdf.CellFormat(65.0, 4.5, dateLabel, "", 0, "R", false, 0, "")
	curY += 15.0

	// ---- 2. APPLICANT DETAILS & LEAVE REQUEST (SECTION 1) ----
	pdf.SetFont(font, "", 13.0)

	// Row 1: Name & Designation
	lbl1 := "bvg t "
	val1 := clean(data.Name)
	lbl2 := "c`ex t "
	val2 := clean(data.Designation)
	lbl1W := 11.0
	lbl2W := 14.0
	val1W := 96.5 - 11.0 // 85.5 mm
	val2X := x + 97.0 + 14.0
	if !isBn {
		lbl1 = "Name : "
		lbl2 = "Designation : "
		lbl1W = 17.0
		lbl2W = 30.0
		val1W = 96.5 - 17.0 // 79.5 mm
		val2X = x + 97.0 + 30.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lbl1W, 5.0, lbl1, "", 0, "L", false, 0, "")
	pdf.CellFormat(val1W, 5.0, val1, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lbl1W, curY+5.8, val1W)

	// Center gap: 0.5 mm (underline ends at x+96.5 mm, lbl2 starts at x+97.0 mm)
	pdf.SetXY(x+97.0, curY)
	pdf.CellFormat(lbl2W, 5.0, lbl2, "", 0, "L", false, 0, "")
	pdf.SetXY(val2X, curY)
	val2W := (x + w) - val2X
	pdf.CellFormat(val2W, 5.0, val2, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, val2X, curY+5.8, val2W)
	curY += 9.0

	// Row 2: Section/Line & Card No
	lbl3 := "‡mKkb / jvBb t "
	val3 := clean(data.Section)
	if data.Shift != "-" && data.Shift != "" {
		val3 += " / " + clean(data.Shift)
	}
	lbl4 := "KvW© bs t "
	val4 := clean(data.EmployeeID)
	lbl3W := 28.0
	lbl4W := 16.0
	val3W := 96.5 - 28.0 // 68.5 mm
	val4X := x + 97.0 + 16.0
	if !isBn {
		lbl3 = "Section / Line : "
		lbl4 = "Card No : "
		lbl3W = 34.0
		lbl4W = 20.0
		val3W = 96.5 - 34.0 // 62.5 mm
		val4X = x + 97.0 + 20.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lbl3W, 5.0, lbl3, "", 0, "L", false, 0, "")
	pdf.CellFormat(val3W, 5.0, val3, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lbl3W, curY+5.8, val3W)

	// Center gap: 0.5 mm (underline ends at x+96.5 mm, lbl4 starts at x+97.0 mm)
	pdf.SetXY(x+97.0, curY)
	pdf.CellFormat(lbl4W, 5.0, lbl4, "", 0, "L", false, 0, "")
	pdf.SetXY(val4X, curY)
	val4W := (x + w) - val4X
	pdf.CellFormat(val4W, 5.0, val4, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, val4X, curY+5.8, val4W)
	curY += 9.0
	lbl5 := "QywUi KviY t "
	val5 := clean(data.Reason)

	// Row 3: Reason for Leave & Leave Period (Single Inline Line)
	if !isBn {
		lbl5 = "Reason for Leave : "
		lbl5W := 38.0
		pdf.SetXY(x, curY)
		pdf.CellFormat(lbl5W, 5.0, lbl5, "", 0, "L", false, 0, "")
		val5W := 44.0
		pdf.CellFormat(val5W, 5.0, val5, "", 0, "L", false, 0, "")
		drawDottedUnderline(pdf, x+lbl5W, curY+5.8, val5W)

		pdf.SetXY(x+82.5, curY)
		pdf.CellFormat(28.0, 5.0, "Leave Period : ", "", 0, "L", false, 0, "")
		pdf.CellFormat(13.0, 5.0, "From : ", "", 0, "L", false, 0, "")
		pdf.CellFormat(24.0, 5.0, data.FromDate, "", 0, "C", false, 0, "")
		drawDottedUnderline(pdf, x+123.5, curY+5.8, 24.0)

		pdf.SetXY(x+148.0, curY)
		pdf.CellFormat(11.0, 5.0, "To : ", "", 0, "L", false, 0, "")
		pdf.CellFormat(24.0, 5.0, data.ToDate, "", 0, "C", false, 0, "")
		drawDottedUnderline(pdf, x+159.0, curY+5.8, 24.0)
	} else {
		lbl5W := 20.0
		pdf.SetXY(x, curY)
		pdf.CellFormat(lbl5W, 5.0, lbl5, "", 0, "L", false, 0, "")
		val5W := 66.0
		pdf.CellFormat(val5W, 5.0, val5, "", 0, "L", false, 0, "")
		drawDottedUnderline(pdf, x+lbl5W, curY+5.8, val5W)

		pdf.SetXY(x+86.5, curY)
		pdf.CellFormat(20.0, 5.0, "QywUi ZvwiL t ", "", 0, "L", false, 0, "")
		pdf.CellFormat(24.0, 5.0, data.FromDate, "", 0, "C", false, 0, "")
		drawDottedUnderline(pdf, x+106.5, curY+5.8, 24.0)

		pdf.SetXY(x+131.0, curY)
		pdf.CellFormat(11.0, 5.0, "‡_‡K t ", "", 0, "C", false, 0, "")
		pdf.CellFormat(24.0, 5.0, data.ToDate, "", 0, "C", false, 0, "")
		drawDottedUnderline(pdf, x+142.5, curY+5.8, 24.0)

		pdf.SetXY(x+167.0, curY)
		pdf.CellFormat(19.0, 5.0, "ch©šÍ", "", 0, "L", false, 0, "")
	}
	curY += 9.0

	// Row 5: Total Days Note
	lbl9 := "‡gvU t "
	lbl10 := "w`b-Gi QywU gbRyim~PK Av‡eb Kwi‡ZwQ|"
	lbl9W := 11.0
	if !isBn {
		lbl9 = "Total : "
		lbl10 = "days leave approval is humbly requested."
		lbl9W = 16.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lbl9W, 5.0, lbl9, "", 0, "L", false, 0, "")
	pdf.CellFormat(20.0, 5.0, data.TotalDays, "", 0, "C", false, 0, "")
	drawDottedUnderline(pdf, x+lbl9W, curY+5.8, 20.0)

	pdf.SetXY(x+lbl9W+20.5, curY)
	lbl10W := w - (lbl9W + 20.5)
	pdf.CellFormat(lbl10W, 5.0, lbl10, "", 0, "L", false, 0, "")
	curY += 9.0

	// Row 6: Address During Leave
	lbl11 := "QywUKvjxb wVKvbv t "
	val11 := clean(data.AddressDuring)
	lbl11W := 30.0
	if !isBn {
		lbl11 = "Address During Leave : "
		lbl11W = 48.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lbl11W, 5.0, lbl11, "", 0, "L", false, 0, "")
	val11W := 80.0
	pdf.CellFormat(val11W, 5.0, val11, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lbl11W, curY+5.8, val11W)
	curY += 9.0

	// Row 7: Phone & Applicant Signature (With Line Height Gap & Upper Underline)
	lbl12 := "‡dvo t "
	val12 := clean(data.EmergencyPhone)
	lblSig1 := "Av‡ebKvixi ¯^v¶i"
	lbl12W := 11.0
	if !isBn {
		lbl12 = "Phone : "
		lblSig1 = "Applicant's Signature"
		lbl12W = 17.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lbl12W, 5.0, lbl12, "", 0, "L", false, 0, "")
	pdf.CellFormat(64.0, 5.0, val12, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lbl12W, curY+5.8, 64.0)

	// Upper Underline for Employee Signature & Label placed below
	drawDottedUnderline(pdf, x+126.0, curY+4.0, 60.0)
	pdf.SetXY(x+126.0, curY+5.8)
	pdf.CellFormat(60.0, 4.0, lblSig1, "", 0, "C", false, 0, "")
	curY += 14.0

	// ---- 3. OFFICE USE SECTION DIVIDER ----
	pdf.SetDrawColor(120, 120, 120)
	pdf.SetLineWidth(0.3)
	pdf.Line(x, curY, x+w, curY)
	curY += 4.0

	officeNote := "------------------- GB Ask Awdm KZ©„K c~iY Kiv n‡e -------------------"
	if !isBn {
		officeNote = "------------------- THIS PORTION TO BE FILLED BY OFFICE -------------------"
	}
	pdf.SetFont(font, "", 12.5)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetXY(x, curY)
	pdf.CellFormat(w, 4.0, officeNote, "", 0, "C", false, 0, "")
	curY += 7.5

	// ---- 4. JOINING DATE & LEAVE PERIOD METADATA ----
	pdf.SetFont(font, "", 12.5)
	pdf.SetTextColor(30, 30, 30)

	lblJoin := "PvKix‡Z ‡hvM`v‡bi ZvwiLt"
	lblCalc := "QywUi wnmeiKvjt"
	lblHw := "nB‡Z"
	lblJoinW := 44.0
	box1X := x + 44.5
	box1W := 28.0
	lblCalcX := x + 73.5
	lblCalcW := 28.0
	box2X := x + 102.0
	box2W := 28.0
	lblHwX := x + 131.0
	lblHwW := 12.0
	box3X := x + 144.0
	box3W := 42.0

	if !isBn {
		lblJoin = "Joining Date :"
		lblCalc = "Leave Period :"
		lblHw = "To"
		lblJoinW = 28.0
		box1X = x + 28.5
		box1W = 28.0
		lblCalcX = x + 57.5
		lblCalcW = 28.0
		box2X = x + 86.0
		box2W = 28.0
		lblHwX = x + 115.0
		lblHwW = 10.0
		box3X = x + 126.0
		box3W = 60.0
	}

	// Extract Leave Calculation Year bounds (01/01/YYYY to 31/12/YYYY)
	calcFrom := "01/01/2026"
	calcTo := "31/12/2026"
	refDate := data.FromDate
	if refDate == "" {
		refDate = data.AppDate
	}
	for _, part := range strings.FieldsFunc(refDate, func(r rune) bool { return r == '/' || r == '-' || r == ' ' }) {
		if len(part) == 4 {
			calcFrom = "01/01/" + part
			calcTo = "31/12/" + part
			break
		}
	}

	pdf.SetXY(x, curY)
	pdf.CellFormat(lblJoinW, 6.5, lblJoin, "", 0, "L", false, 0, "")
	pdf.Rect(box1X, curY, box1W, 6.5, "D")
	pdf.SetXY(box1X, curY+1.0)
	pdf.CellFormat(box1W, 4.5, data.JoiningDate, "", 0, "C", false, 0, "")

	pdf.SetXY(lblCalcX, curY)
	pdf.CellFormat(lblCalcW, 6.5, lblCalc, "", 0, "L", false, 0, "")
	pdf.Rect(box2X, curY, box2W, 6.5, "D")
	pdf.SetXY(box2X, curY+1.0)
	pdf.CellFormat(box2W, 4.5, calcFrom, "", 0, "C", false, 0, "")

	pdf.SetXY(lblHwX, curY)
	pdf.CellFormat(lblHwW, 6.5, lblHw, "", 0, "C", false, 0, "")
	pdf.Rect(box3X, curY, box3W, 6.5, "D")
	pdf.SetXY(box3X, curY+1.0)
	pdf.CellFormat(box3W, 4.5, calcTo, "", 0, "C", false, 0, "")
	curY += 9.5

	// ---- 5. LEAVE BALANCE TABLE (4 TYPES x 3 ROWS) ----
	tableLeft := x
	tableW := w
	lblColW := 40.0
	valColW := (tableW - lblColW) / 4.0 // 36.5 mm each

	// Header Row (Leave Types)
	typeHeaders := []string{
		"‰bwgwËK QywU",
		"cxov-QywU",
		"AR©j QywU",
		"gvZ„Z¡RwbZ QywU",
	}
	if !isBn {
		typeHeaders = []string{"Casual Leave", "Sick Leave", "Earned Leave", "Maternity Leave"}
	}

	pdf.SetFont(font, "B", 12.5)
	lblTable := "QywUi weeiY   :"
	if !isBn {
		lblTable = "Leave Details :"
	}
	pdf.SetXY(tableLeft, curY)
	pdf.CellFormat(lblColW, 6.0, lblTable, "", 0, "L", false, 0, "")

	for i, th := range typeHeaders {
		px := tableLeft + lblColW + float64(i)*valColW
		pdf.Rect(px, curY, valColW, 6.0, "D")
		pdf.SetXY(px, curY+0.8)
		pdf.CellFormat(valColW, 4.4, th, "", 0, "C", false, 0, "")
	}
	curY += 7.5

	// Data Rows (Entitled, Used, Remaining)
	rowLabels := []string{
		"cÖvc¨ QywU   :",
		"‡fvMKyZ QywU   :",
		"Aewkó QywU   :",
	}
	if !isBn {
		rowLabels = []string{"Entitled Leave :", "Used Leave :", "Remaining Leave :"}
	}

	// Extract balance values per type (Casual, Sick, Earned, Maternity)
	balMap := make(map[string][3]string) // key: type, val: [entitled, used, remaining]
	for _, b := range data.Balances {
		k := strings.ToLower(b.LeaveType)
		balMap[k] = [3]string{b.Entitled, b.Used, b.Remaining}
	}

	findBal := func(kw string) [3]string {
		for k, v := range balMap {
			if strings.Contains(k, kw) {
				return v
			}
		}
		return [3]string{"-", "-", "-"}
	}

	cBal := findBal("casual")
	if cBal[0] == "-" || cBal[0] == "" {
		cBal = findBal("নৈমিত্তিক")
	}
	if cBal[0] == "-" || cBal[0] == "" {
		cBal[0] = "10"
		usedVal := 0
		if cBal[1] != "-" && cBal[1] != "" {
			fmt.Sscanf(cBal[1], "%d", &usedVal)
		} else {
			cBal[1] = "0"
		}
		cBal[2] = strconv.Itoa(10 - usedVal)
	}

	sBal := findBal("sick")
	if sBal[0] == "-" || sBal[0] == "" {
		sBal = findBal("পীড়া")
	}
	if sBal[0] == "-" || sBal[0] == "" {
		sBal[0] = "14" // Sick leave default entitled balance is 14 days
		usedVal := 0
		if sBal[1] != "-" && sBal[1] != "" {
			fmt.Sscanf(sBal[1], "%d", &usedVal)
		} else {
			sBal[1] = "0"
		}
		sBal[2] = strconv.Itoa(14 - usedVal)
	}

	eBal := findBal("earned")
	if eBal[0] == "-" || eBal[0] == "" {
		eBal = findBal("annual")
	}
	if eBal[0] == "-" || eBal[0] == "" {
		eBal = findBal("বাৎসরিক")
	}

	mBal := findBal("maternity")
	if mBal[0] == "-" || mBal[0] == "" {
		mBal = findBal("মাতৃত্ব")
	}

	typeBals := [4][3]string{cBal, sBal, eBal, mBal}

	pdf.SetFont(font, "", 12.5)
	for rIdx, rLbl := range rowLabels {
		pdf.SetXY(tableLeft, curY)
		pdf.CellFormat(lblColW, 6.0, rLbl, "", 0, "L", false, 0, "")

		for cIdx := 0; cIdx < 4; cIdx++ {
			px := tableLeft + lblColW + float64(cIdx)*valColW
			valStr := typeBals[cIdx][rIdx]
			if valStr == "" {
				valStr = "-"
			}
			pdf.Rect(px, curY, valColW, 6.0, "D")
			pdf.SetXY(px, curY+0.8)
			pdf.CellFormat(valColW, 4.4, valStr, "", 0, "C", false, 0, "")
		}
		curY += 7.5
	}
	curY += 3.0

	// ---- 6. APPROVAL GRANT NOTE ----
	pdf.SetFont(font, "", 12.5)
	pdf.Rect(x+30.0, curY, 35.0, 6.5, "D")
	pdf.SetXY(x+30.0, curY+1.0)
	pdf.CellFormat(35.0, 4.5, data.TotalDays, "", 0, "C", false, 0, "")

	grantNote := "w`bi ‰bwgwËK/ cxov/ AR©j/ gvZ„Z¡ RwbZ QywU gbRyi Kiv nBj|"
	if !isBn {
		grantNote = "days Casual / Sick / Earned / Maternity Leave granted."
	}
	pdf.SetXY(x+67.0, curY+1.0)
	pdf.CellFormat(119.0, 4.5, grantNote, "", 0, "L", false, 0, "")
	curY += 21.0

	// ---- 7. 5-COLUMN APPROVAL SIGNATURES ROW ----
	sigCols := []string{
		"GBP. Avi.",
		"BbPvR©",
		"‡cÖvWvKkb g¨v‡bRvi",
		"G¨vWwgb (G.wR.Gg)",
		"G. wR. Gg.",
	}
	if !isBn {
		sigCols = []string{
			"H.R.",
			"Incharge",
			"Production Manager",
			"Admin (A.G.M)",
			"A.G.M.",
		}
	}

	colW5 := w / 5.0 // 37.2 mm each
	pdf.SetFont(font, "", 11.0)
	pdf.SetDrawColor(120, 120, 120)
	pdf.SetLineWidth(0.3)

	for i, sc := range sigCols {
		px := x + float64(i)*colW5
		pdf.Line(px+2.5, curY, px+colW5-2.5, curY)
		pdf.SetXY(px, curY+1.0)
		pdf.CellFormat(colW5, 3.5, sc, "", 0, "C", false, 0, "")
	}
	curY += 10.0

	// ---- 8. JOINING REPORT AFTER LEAVE (BOTTOM CUT SECTION) ----
	pdf.SetDrawColor(120, 120, 120)
	pdf.SetLineWidth(0.3)
	pdf.Line(x, curY, x+w, curY)
	curY += 5.0

	// Company Title & Report Title
	pdf.SetFont(font, "B", 16.0)
	pdf.SetXY(x, curY)
	pdf.CellFormat(w, 5.0, clean(data.Company), "", 0, "C", false, 0, "")
	curY += 6.5

	pdf.SetFont(font, "", 12.5)
	reportTitle := "QywU ‡k‡l Kv‡R ‡hvM`v‡bi cÖwZ‡e`b"
	if !isBn {
		reportTitle = "REPORT OF JOINING WORK AFTER LEAVE"
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(w, 4.0, reportTitle, "", 0, "C", false, 0, "")
	curY += 7.5

	// Row 1: Name, Card No, Issue Date
	lblJ1 := "bvg t "
	lblJ2 := "KvW© bs : "
	lblJ3 := "Bmmyi ZvwiL t "
	lblJ1W := 11.0
	lblJ2W := 16.0
	lblJ3W := 24.0
	if !isBn {
		lblJ1 = "Name : "
		lblJ2 = "Card No : "
		lblJ3 = "Issue Date : "
		lblJ1W = 17.0
		lblJ2W = 20.0
		lblJ3W = 26.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lblJ1W, 5.0, lblJ1, "", 0, "L", false, 0, "")
	valJ1W := 52.0
	if !isBn {
		valJ1W = 46.0
	}
	pdf.CellFormat(valJ1W, 5.0, clean(data.Name), "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lblJ1W, curY+5.8, valJ1W)

	pdf.SetXY(x+lblJ1W+valJ1W+0.5, curY)
	pdf.CellFormat(lblJ2W, 5.0, lblJ2, "", 0, "L", false, 0, "")
	pdf.CellFormat(34.0, 5.0, clean(data.EmployeeID), "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lblJ1W+valJ1W+0.5+lblJ2W, curY+5.8, 34.0)

	pdf.SetXY(x+114.0, curY)
	pdf.CellFormat(lblJ3W, 5.0, lblJ3, "", 0, "L", false, 0, "")
	valJ3W := (x + w) - (x + 114.0 + lblJ3W)
	pdf.CellFormat(valJ3W, 5.0, data.AppDate, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+114.0+lblJ3W, curY+5.8, valJ3W)
	curY += 9.0

	// Row 2: Joining Date per Approved Leave
	lblJ4 := "gbRyiK…Z QywU Abymv‡i ‡hvM`v‡bi ZvwiL  :"
	lblJ4W := 64.0
	if !isBn {
		lblJ4 = "Joining Date per Approved Leave : "
		lblJ4W = 72.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lblJ4W, 5.0, lblJ4, "", 0, "L", false, 0, "")
	valJ4W := w - lblJ4W
	pdf.CellFormat(valJ4W, 5.0, data.ToDate, "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lblJ4W, curY+5.8, valJ4W)
	curY += 9.0

	// Row 3: Actual Joining Date
	lblJ5 := "‡hvM`v‡bi cÖK…Z ZvwiL  :"
	lblJ5W := 40.0
	if !isBn {
		lblJ5 = "Actual Joining Date : "
		lblJ5W = 46.0
	}
	pdf.SetXY(x, curY)
	pdf.CellFormat(lblJ5W, 5.0, lblJ5, "", 0, "L", false, 0, "")
	valJ5W := w - lblJ5W
	pdf.CellFormat(valJ5W, 5.0, "", "", 0, "L", false, 0, "")
	drawDottedUnderline(pdf, x+lblJ5W, curY+5.8, valJ5W)
	curY += 11.0

	// Bottom Footer Notes & Signatures
	lblSigApp := "Av‡ebKvixi ¯^v¶i"
	bottomNote := "GB AskwU QywU ‡k‡l Kv‡R ‡hvM`v‡bi mgq cÖkvmb kvLvq rgvgw‡Z n‡e|"
	lblSigHR := "GBPAvi kvLv"
	if !isBn {
		lblSigApp = "Applicant's Signature"
		bottomNote = "This portion must be submitted to Admin Dept upon joining work after leave."
		lblSigHR = "HR Dept"
	}

	pdf.SetFont(font, "", 11.5)
	pdf.SetXY(x, curY)
	pdf.CellFormat(45.0, 4.0, lblSigApp, "", 0, "L", false, 0, "")

	pdf.SetFont(font, "", 11.0)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetXY(x+45.0, curY)
	pdf.CellFormat(96.0, 4.0, bottomNote, "", 0, "C", false, 0, "")

	pdf.SetFont(font, "", 11.5)
	pdf.SetTextColor(30, 30, 30)
	pdf.SetXY(x+141.0, curY)
	pdf.CellFormat(45.0, 4.0, lblSigHR, "", 0, "R", false, 0, "")
}

func checkBoxText(label string, checked bool) string {
	mark := "[ ]"
	if checked {
		mark = "[x]"
	}
	return mark + " " + label
}
