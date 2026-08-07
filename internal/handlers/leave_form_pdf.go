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
	FormTitle:   "QywUi Av‡ebbcÎ (LEAVE APPLICATION)",

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
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	if lang == "bn" {
		return fmt.Sprintf("%s %s %s", bnDigits(t.Format("02")), bnMonthNames[int(t.Month())], bnDigits(t.Format("2006")))
	}
	return t.Format("02 Jan 2006")
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
		return loadBanglaFont(pdf)
	}
	return "Helvetica"
}

const (
	leaveFormPageW = 210.0
	leaveFormPageH = 297.0
	leaveFormLeft  = 10.0
	leaveFormW     = 190.0
	leaveFormTop   = 8.0
)

func renderLeaveFormPDFPage(pdf *gofpdf.Fpdf, font string, lang string, data leaveFormData, labels leaveFormLabels) {
	s := 1.0
	x := leaveFormLeft
	w := leaveFormW
	curY := leaveFormTop

	// Set line width & colors
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.3)

	// ---- 1. Header Band ----
	pdf.SetFillColor(15, 23, 42) // Dark Slate Navy #0F172A
	pdf.Rect(x, curY, w, 18*s, "F")

	// Left: Company Name & Branch/Address
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 10*s)
	pdf.SetXY(x+3*s, curY+2.5*s)
	pdf.CellFormat(110*s, 5*s, data.Company, "", 0, "L", false, 0, "")

	pdf.SetFont(font, "", 5.5*s)
	pdf.SetTextColor(203, 213, 225)
	pdf.SetXY(x+3*s, curY+8.0*s)
	pdf.CellFormat(110*s, 3.5*s, data.Branch, "", 0, "L", false, 0, "")
	pdf.SetXY(x+3*s, curY+12.0*s)
	pdf.CellFormat(110*s, 3.5*s, labels.SystemTitle, "", 0, "L", false, 0, "")

	// Right: Form Title & App No / Date
	pdf.SetTextColor(245, 158, 11) // Amber #F59E0B
	pdf.SetFont(font, "B", 10*s)
	pdf.SetXY(x+115*s, curY+2.5*s)
	pdf.CellFormat(72*s, 5*s, labels.FormTitle, "", 0, "R", false, 0, "")

	pdf.SetFont(font, "", 5.5*s)
	pdf.SetTextColor(203, 213, 225)
	pdf.SetXY(x+115*s, curY+8.0*s)
	pdf.CellFormat(72*s, 3.5*s, labels.AppNo+": "+data.AppNo, "", 0, "R", false, 0, "")
	pdf.SetXY(x+115*s, curY+12.0*s)
	pdf.CellFormat(72*s, 3.5*s, labels.AppDate+": "+data.AppDate, "", 0, "R", false, 0, "")
	curY += 18*s + 2*s

	// ---- 2. Employee Information Grid (4 Columns) ----
	curY = drawLeaveSectionHeader(pdf, font, s, x, curY, w, labels.EmployeeInfo)
	empFields := []payslipField{
		{labels.EmployeeID, data.EmployeeID},
		{labels.Name, data.Name},
		{labels.Department, data.Department},
		{labels.Section, data.Section},
		{labels.Designation, data.Designation},
		{labels.Grade, data.Grade},
		{labels.Shift, data.Shift},
		{labels.JoiningDate, data.JoiningDate},
		{labels.ReportsTo, data.ReportsTo},
		{labels.EmpType, data.EmpType},
		{labels.CardNo, data.CardNo},
		{labels.Mobile, data.Mobile},
	}
	curY = drawLeaveFieldGrid(pdf, font, s, x, curY, w, empFields, 4)
	curY += 2.5 * s

	// ---- 3. Leave Details (Left) + Leave Balance (Right) Side-by-Side ----
	startY := curY
	halfW := (w - 3*s) / 2

	// Left: Leave Details
	leftY := drawLeaveSectionHeader(pdf, font, s, x, startY, halfW, labels.LeaveDetails)
	leftY = drawLeaveTypeCheckboxes(pdf, font, s, x, leftY, halfW, data.LeaveTypeOptions, data.LeaveTypeActive)
	detailFields := []payslipField{
		{labels.FromDate, data.FromDate},
		{labels.ToDate, data.ToDate},
		{labels.TotalDays, data.TotalDays},
		{labels.EmergencyPhone, data.EmergencyPhone},
		{labels.AddressDuring, data.AddressDuring},
	}
	leftY = drawLeaveFieldGrid(pdf, font, s, x, leftY, halfW, detailFields, 2)
	leftY = drawLeaveTextBox(pdf, font, s, x, leftY, halfW, labels.Reason, data.Reason, 12.0)

	// Right: Leave Balance Table
	rightY := drawLeaveSectionHeader(pdf, font, s, x+halfW+3*s, startY, halfW, labels.LeaveBalance)
	rightY = drawLeaveBalanceTable(pdf, font, s, x+halfW+3*s, rightY, halfW, labels, data.Balances)

	if leftY > rightY {
		curY = leftY
	} else {
		curY = rightY
	}
	curY += 2.5 * s

	// ---- 4. Work Handover Information ----
	curY = drawLeaveSectionHeader(pdf, font, s, x, curY, w, labels.Handover)
	handoverFields := []payslipField{
		{labels.HandoverTo, data.HandoverTo},
		{labels.HandoverDept, data.HandoverDept},
		{labels.HandoverDesig, data.HandoverDesig},
	}
	curY = drawLeaveFieldGrid(pdf, font, s, x, curY, w, handoverFields, 3)
	curY = drawLeaveTextBox(pdf, font, s, x, curY, w, labels.HandoverDetails, data.HandoverDetails, 9.0)
	curY += 2.5 * s

	// ---- 5. Approval Workflow (5 Box Side-by-Side Cards) ----
	curY = drawLeaveSectionHeader(pdf, font, s, x, curY, w, labels.Approval)
	curY = drawLeaveApprovalCards(pdf, font, s, x, curY, w, labels, data)
	curY += 2.5 * s

	// ---- 6. Notes & Footer ----
	pdf.SetFont(font, "", 4.5*s)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(x+1.5*s, curY)
	notesStr := fmt.Sprintf("%s: 1. %s  2. %s  3. %s", labels.Notes, labels.Note1, labels.Note2, labels.Note3)
	pdf.CellFormat(w-3*s, 2.5*s, notesStr, "", 0, "C", false, 0, "")
	curY += 3.5 * s

	pdf.SetFont(font, "", 4.5*s)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(x+1.5*s, curY)
	footerStr := fmt.Sprintf("%s    •    %s: %s    •    %s: %s", data.GeneratedBy, labels.DocumentNo, data.DocumentNo, labels.PrintDate, data.PrintDate)
	pdf.CellFormat(w-3*s, 2.5*s, footerStr, "", 0, "C", false, 0, "")
}

func drawLeaveSectionHeader(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title string) float64 {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFont(font, "B", 5.8*s)
	pdf.SetTextColor(30, 58, 138)
	pdf.Rect(x, y, w, 3.8*s, "DF")
	pdf.SetXY(x+1.5*s, y+0.6*s)
	pdf.CellFormat(w-3*s, 2.6*s, title, "", 0, "L", false, 0, "")
	return y + 3.8*s
}

func drawLeaveFieldGrid(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, fields []payslipField, cols int) float64 {
	rowH := 3.8 * s
	colW := w / float64(cols)
	for i, fld := range fields {
		col := i % cols
		row := i / cols
		px := x + float64(col)*colW
		py := y + float64(row)*rowH
		pdf.SetDrawColor(203, 213, 225)
		pdf.Rect(px, py, colW, rowH, "D")

		labelStr := fld.Label
		valStr := fld.Value

		lblFontSize := 4.8 * s
		if cols >= 3 && len(labelStr) > 10 {
			lblFontSize = 4.0 * s
		}
		pdf.SetFont(font, "B", lblFontSize)
		pdf.SetTextColor(100, 116, 139)

		lblW := colW * 0.44
		if cols >= 3 {
			lblW = colW * 0.58
		}
		pdf.SetXY(px+0.6*s, py+0.8*s)
		pdf.CellFormat(lblW-0.6*s, 2.2*s, labelStr, "", 0, "L", false, 0, "")

		valFontSize := 4.8 * s
		if len(valStr) > 20 {
			valFontSize = 4.0 * s
		}
		pdf.SetFont(font, "", valFontSize)
		pdf.SetTextColor(15, 23, 42)
		valW := colW - lblW
		pdf.SetXY(px+lblW, py+0.8*s)
		pdf.CellFormat(valW-0.4*s, 2.2*s, valStr, "", 0, "L", false, 0, "")
	}
	rows := (len(fields) + cols - 1) / cols
	return y + float64(rows)*rowH
}

func drawLeaveTypeCheckboxes(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, options []string, active int) float64 {
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFillColor(255, 255, 255)
	h := 5.0 * s
	pdf.Rect(x, y, w, h, "D")

	cbSize := 2.5 * s
	cx := x + 1.5*s
	cy := y + 1.25*s
	pdf.SetFont(font, "", 4.5*s)
	for i, opt := range options {
		if cx+cbSize+12*s > x+w {
			break
		}
		pdf.SetDrawColor(15, 23, 42)
		pdf.SetLineWidth(0.3)
		pdf.Rect(cx, cy, cbSize, cbSize, "D")
		if i == active {
			pdf.SetLineWidth(0.5)
			pdf.Line(cx+0.3*s, cy+0.3*s, cx+cbSize-0.3*s, cy+cbSize-0.3*s)
			pdf.Line(cx+cbSize-0.3*s, cy+0.3*s, cx+0.3*s, cy+cbSize-0.3*s)
		}
		pdf.SetTextColor(15, 23, 42)
		txtW := pdf.GetStringWidth(opt) + 2*s
		pdf.SetXY(cx+cbSize+0.8*s, cy-0.2*s)
		pdf.CellFormat(txtW, 2.6*s, opt, "", 0, "L", false, 0, "")
		cx += cbSize + txtW + 3*s
	}
	return y + h
}

func drawLeaveTextBox(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title, content string, h float64) float64 {
	pdf.SetDrawColor(203, 213, 225)
	pdf.Rect(x, y, w, h*s, "D")
	pdf.SetFont(font, "B", 4.8*s)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(x+1.5*s, y+0.8*s)
	pdf.CellFormat(w-3*s, 2.2*s, title+":", "", 0, "L", false, 0, "")

	if content != "" {
		pdf.SetFont(font, "", 4.6*s)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(x+1.5*s, y+3.2*s)
		pdf.CellFormat(w-3*s, 2.2*s, content, "", 0, "L", false, 0, "")
	}
	return y + h*s
}

func drawLeaveBalanceTable(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, labels leaveFormLabels, rows []leaveBalanceRow) float64 {
	headers := []string{labels.BalLeaveType, labels.BalEntitled, labels.BalUsed, labels.BalRemaining}
	colW := w / 4.0
	h := 3.8 * s

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFont(font, "B", 4.8*s)
	pdf.SetTextColor(30, 58, 138)
	for i, hdr := range headers {
		px := x + float64(i)*colW
		pdf.Rect(px, y, colW, h, "DF")
		pdf.SetXY(px+0.5*s, y+0.8*s)
		pdf.CellFormat(colW-1*s, 2.2*s, hdr, "", 0, "C", false, 0, "")
	}
	y += h

	pdf.SetFont(font, "", 4.6*s)
	pdf.SetTextColor(15, 23, 42)
	for _, r := range rows {
		vals := []string{r.LeaveType, r.Entitled, r.Used, r.Remaining}
		for i, v := range vals {
			px := x + float64(i)*colW
			pdf.SetDrawColor(203, 213, 225)
			pdf.Rect(px, y, colW, h, "D")
			pdf.SetXY(px+0.5*s, y+0.8*s)
			align := "C"
			if i == 0 {
				align = "L"
			}
			pdf.CellFormat(colW-1*s, 2.2*s, v, "", 0, align, false, 0, "")
		}
		y += h
	}
	return y
}

func drawLeaveApprovalCards(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, labels leaveFormLabels, data leaveFormData) float64 {
	cards := []struct {
		role string
		sig  string
	}{
		{labels.EmployeeRole, labels.Signature},
		{labels.SupervisorRole, labels.Signature},
		{labels.DeptHeadRole, labels.Signature},
		{labels.HRRole, labels.Signature},
		{labels.FinalRole, labels.HRManagerSig},
	}

	cardW := w / 5.0
	h := 18.0 * s

	for i, c := range cards {
		cx := x + float64(i)*cardW
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetLineWidth(0.3)
		pdf.Rect(cx, y, cardW, h, "D")

		// Header
		pdf.SetFillColor(248, 250, 252)
		pdf.Rect(cx, y, cardW, 3.8*s, "DF")
		pdf.SetFont(font, "B", 4.8*s)
		pdf.SetTextColor(30, 58, 138)
		pdf.SetXY(cx+0.5*s, y+0.8*s)
		pdf.CellFormat(cardW-1*s, 2.2*s, c.role, "", 0, "C", false, 0, "")

		// Status line
		pdf.SetFont(font, "", 4.2*s)
		pdf.SetTextColor(100, 116, 139)
		if i == 0 {
			pdf.SetXY(cx+0.5*s, y+5.0*s)
			pdf.CellFormat(cardW-1*s, 2.0*s, labels.Date+": _____", "", 0, "C", false, 0, "")
		} else {
			pdf.SetXY(cx+0.5*s, y+5.0*s)
			pdf.CellFormat(cardW-1*s, 2.0*s, checkBoxText(labels.Approved, data.ApprovedYes)+" "+checkBoxText(labels.Rejected, data.RejectedYes), "", 0, "C", false, 0, "")
		}

		// Signature line at bottom
		pdf.SetFont(font, "", 4.5*s)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(cx+0.5*s, y+h-3.2*s)
		pdf.CellFormat(cardW-1*s, 2.2*s, c.sig, "", 0, "C", false, 0, "")
	}
	return y + h
}

func checkBoxText(label string, checked bool) string {
	mark := "[ ]"
	if checked {
		mark = "[x]"
	}
	return mark + " " + label
}
