package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

type leaveFormLabels struct {
	SystemTitle string
	FormTitle   string

	Company    string
	Branch     string
	AppNo      string
	AppDate    string

	EmployeeInfo  string
	EmployeeID    string
	CardNo        string
	Name          string
	Mobile        string
	Department    string
	Section       string
	Designation   string
	Grade         string
	Shift         string
	JoiningDate   string
	ReportsTo     string
	EmpType       string

	LeaveDetails    string
	LeaveTypeLabel  string
	FromDate        string
	ToDate          string
	TotalDays       string
	HalfDay         string
	HalfDayYes      string
	HalfDayNo       string
	AddressDuring   string
	EmergencyPhone  string
	Reason          string

	LeaveBalance    string
	BalLeaveType    string
	BalEntitled     string
	BalUsed         string
	BalRemaining    string

	Handover       string
	HandoverTo     string
	HandoverDept   string
	HandoverDesig  string
	HandoverDetails string

	Approval      string
	EmployeeRole  string
	SupervisorRole string
	DeptHeadRole  string
	HRRole        string
	FinalRole     string
	Approved      string
	Rejected      string
	Remarks       string
	Signature     string
	Date          string
	BalanceUpdated string
	Yes           string
	No            string
	HRManagerSig  string

	Notes      string
	Note1      string
	Note2      string
	Note3      string

	GeneratedBy string
	DocumentNo  string
	PrintDate   string
}

var leaveEnLabels = leaveFormLabels{
	SystemTitle: "PEOPLEHUB HR & PAYROLL MANAGEMENT SYSTEM",
	FormTitle:   "EMPLOYEE LEAVE APPLICATION FORM",

	Company:   "Company",
	Branch:    "Branch",
	AppNo:     "Application No",
	AppDate:   "Application Date",

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
	EmergencyPhone: "Emergency Contact Number",
	Reason:         "Reason for Leave",

	LeaveBalance: "LEAVE BALANCE",
	BalLeaveType: "Leave Type",
	BalEntitled:  "Entitled",
	BalUsed:      "Used",
	BalRemaining: "Remaining",

	Handover:        "WORK HANDOVER INFORMATION",
	HandoverTo:      "Assigned To Employee",
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
	BalanceUpdated: "Leave Balance Updated",
	Yes:            "Yes",
	No:             "No",
	HRManagerSig:   "HR Manager Signature",

	Notes: "Notes",
	Note1: "Leave must be applied before the scheduled leave date unless it is an emergency.",
	Note2: "Medical certificate is mandatory for Sick Leave exceeding company policy.",
	Note3: "Leave approval is subject to company rules and business requirements.",

	GeneratedBy: "PeopleHub HR & Payroll Management System",
	DocumentNo:  "Document No",
	PrintDate:   "Print Date",
}

var leaveBnLabels = leaveFormLabels{
	SystemTitle: "PEOPLEHUB HR & PAYROLL MANAGEMENT SYSTEM",
	FormTitle:   "ছুটির আবেদনপত্র (LEAVE APPLICATION)",

	Company:   "কোম্পানির নাম",
	Branch:    "শাখা/কারখানা",
	AppNo:     "আবেদন নং",
	AppDate:   "আবেদন তারিখ",

	EmployeeInfo: "কর্মচারীর তথ্য",
	EmployeeID:   "কর্মচারী আইডি",
	CardNo:       "কার্ড নং",
	Name:         "নাম",
	Mobile:       "মোবাইল",
	Department:   "বিভাগ",
	Section:      "সেকশন",
	Designation:  "পদবী",
	Grade:        "গ্রেড",
	Shift:        "শিফট",
	JoiningDate:  "চাকরিতে যোগদানের তারিখ",
	ReportsTo:    "রিপোর্টিং অফিসার",
	EmpType:      "চাকরির ধরন",

	LeaveDetails:   "ছুটির বিবরণ",
	LeaveTypeLabel: "ছুটির ধরন",
	FromDate:       "ছুটি শুরুর তারিখ",
	ToDate:         "শেষ তারিখ",
	TotalDays:      "মোট ছুটির দিন",
	HalfDay:        "অর্ধদিবস",
	HalfDayYes:     "হ্যাঁ",
	HalfDayNo:      "না",
	AddressDuring:  "ছুটিকালীন যোগাযোগের ঠিকানা",
	EmergencyPhone: "জরুরি মোবাইল নম্বর",
	Reason:         "ছুটির কারণ",

	LeaveBalance: "ছুটির হিসাব",
	BalLeaveType: "ছুটির ধরন",
	BalEntitled:  "প্রাপ্য",
	BalUsed:      "ব্যবহৃত",
	BalRemaining: "অবশিষ্ট",

	Handover:        "ছুটির পূর্বে দায়িত্ব হস্তান্তর",
	HandoverTo:      "যার নিকট দায়িত্ব অর্পণ করা হয়েছে",
	HandoverDept:    "বিভাগ",
	HandoverDesig:   "পদবী",
	HandoverDetails: "অসমাপ্ত কাজের বিবরণ",

	Approval:       "অনুমোদন অংশ",
	EmployeeRole:   "আবেদনকারী",
	SupervisorRole: "সুপারভাইজার",
	DeptHeadRole:   "বিভাগীয় প্রধান",
	HRRole:         "মানবসম্পদ বিভাগ (HR)",
	FinalRole:      "চূড়ান্ত অনুমোদন",
	Approved:       "অনুমোদিত",
	Rejected:       "অননুমোদিত",
	Remarks:        "মন্তব্য",
	Signature:      "স্বাক্ষর",
	Date:           "তারিখ",
	BalanceUpdated: "ছুটির হিসাব আপডেট",
	Yes:            "হ্যাঁ",
	No:             "না",
	HRManagerSig:   "এইচআর ম্যানেজারের স্বাক্ষর",

	Notes: "মন্তব্য",
	Note1: "জরুরি পরিস্থিতি ব্যতীত পূর্বেই ছুটির আবেদন করতে হবে।",
	Note2: "অসুস্থতার ছুটির ক্ষেত্রে প্রয়োজন হলে চিকিৎসা সনদ সংযুক্ত করতে হবে।",
	Note3: "প্রতিষ্ঠানের নীতিমালা অনুযায়ী ছুটি অনুমোদিত হবে।",

	GeneratedBy: "PeopleHub HR & Payroll Management System",
	DocumentNo:  "নথি নং",
	PrintDate:   "মুদ্রণের তারিখ",
}

type leaveBalanceRow struct {
	LeaveType string
	Entitled  string
	Used      string
	Remaining string
}

type leaveFormData struct {
	BrandName string

	Company   string
	Branch    string
	AppNo     string
	AppDate   string

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

	Status       string
	ApprovedYes  bool
	RejectedYes  bool

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
		BrandName:     "PeopleHub",
		Company:       companyNameFor(lang, leave.Company),
		Branch:        companyAddress(leave.Company, lang),
		AppNo:         leaveAppNo(leave.ID, leave.FromDate),
		AppDate:       formatLeaveFormDate(time.Now().Format("2006-01-02"), lang),

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
		d.LeaveTypeOptions = []string{"নৈমিত্তিক (Casual)", "অসুস্থতা (Sick)", "বাৎসরিক (Annual)", "মাতৃত্বকালীন", "পিতৃত্বকালীন", "বেতনবিহীন", "জরুরি", "অন্যান্য"}
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
			{"Annual Leave", "", "", ""},
			{"Casual Leave", "", "", ""},
			{"Sick Leave", "", "", ""},
		}
		if lang == "bn" {
			d.Balances = []leaveBalanceRow{
				{"বাৎসরিক ছুটি", "", "", ""},
				{"নৈমিত্তিক ছুটি", "", "", ""},
				{"অসুস্থতার ছুটি", "", "", ""},
				{"অন্যান্য", "", "", ""},
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
	leaveFormLeft  = 12.0
	leaveFormW     = 186.0
	leaveFormTop   = 10.0
)

type leaveFormDraw struct {
	pdf  *gofpdf.Fpdf
	font string
	lang string
	labels leaveFormLabels
	x     float64
	w     float64
	y     float64
}

func (d *leaveFormDraw) addPage() {
	d.pdf.AddPage()
	d.y = leaveFormTop
}

func (d *leaveFormDraw) need(h float64) {
	if d.y+h > leaveFormPageH-14 {
		d.addPage()
	}
}

func (d *leaveFormDraw) sectionTitle(title string) {
	d.need(10)
	d.pdf.SetFillColor(241, 245, 249)
	d.pdf.SetDrawColor(203, 213, 225)
	d.pdf.SetLineWidth(0.3)
	d.pdf.SetFont(d.font, "B", 8)
	d.pdf.SetTextColor(30, 58, 138)
	d.pdf.Rect(d.x, d.y, d.w, 8, "DF")
	d.pdf.SetXY(d.x+2, d.y+1.6)
	d.pdf.CellFormat(d.w-4, 5, title, "", 0, "L", false, 0, "")
	d.y += 8
}

func (d *leaveFormDraw) row(label, value string) {
	d.need(8)
	h := 7.0
	d.pdf.SetDrawColor(203, 213, 225)
	d.pdf.SetLineWidth(0.3)
	d.pdf.Rect(d.x, d.y, d.w, h, "D")
	d.pdf.SetFont(d.font, "B", 7.5)
	d.pdf.SetTextColor(71, 85, 105)
	d.pdf.SetXY(d.x+2, d.y+1.4)
	d.pdf.CellFormat(52, 4.5, label+":", "", 0, "L", false, 0, "")
	d.pdf.SetFont(d.font, "", 7.5)
	d.pdf.SetTextColor(15, 23, 42)
	d.pdf.SetXY(d.x+56, d.y+1.4)
	d.pdf.CellFormat(d.w-60, 4.5, value, "", 0, "L", false, 0, "")
	d.y += h
}

func (d *leaveFormDraw) pair(l1, v1, l2, v2 string) {
	d.need(8)
	h := 7.0
	half := d.w / 2
	d.pdf.SetDrawColor(203, 213, 225)
	d.pdf.SetLineWidth(0.3)
	d.pdf.Rect(d.x, d.y, half, h, "D")
	d.pdf.Rect(d.x+half, d.y, half, h, "D")
	d.pdf.SetFont(d.font, "B", 7.5)
	d.pdf.SetTextColor(71, 85, 105)
	d.pdf.SetXY(d.x+2, d.y+1.4)
	d.pdf.CellFormat(half*0.45, 4.5, l1+":", "", 0, "L", false, 0, "")
	d.pdf.SetFont(d.font, "", 7.5)
	d.pdf.SetTextColor(15, 23, 42)
	d.pdf.SetXY(d.x+2+half*0.45, d.y+1.4)
	d.pdf.CellFormat(half-half*0.45-2, 4.5, v1, "", 0, "L", false, 0, "")
	d.pdf.SetFont(d.font, "B", 7.5)
	d.pdf.SetTextColor(71, 85, 105)
	d.pdf.SetXY(d.x+half+2, d.y+1.4)
	d.pdf.CellFormat(half*0.45, 4.5, l2+":", "", 0, "L", false, 0, "")
	d.pdf.SetFont(d.font, "", 7.5)
	d.pdf.SetTextColor(15, 23, 42)
	d.pdf.SetXY(d.x+half+2+half*0.45, d.y+1.4)
	d.pdf.CellFormat(half-half*0.45-2, 4.5, v2, "", 0, "L", false, 0, "")
	d.y += h
}

func (d *leaveFormDraw) line(h float64) {
	d.need(h)
	d.pdf.SetDrawColor(203, 213, 225)
	d.pdf.Line(d.x, d.y, d.x+d.w, d.y)
	d.y += h
}

func (d *leaveFormDraw) renderLeaveTypeOptions(data leaveFormData) {
	cbSize := 3.2
	gap := 5.0
	lineH := 6.5
	d.pdf.SetFont(d.font, "", 7.5)

	cx := d.x + 2
	cy := d.y
	for i, opt := range data.LeaveTypeOptions {
		txtW := d.pdf.GetStringWidth(opt)
		itemW := cbSize + gap + txtW
		if cx+itemW > d.x+d.w-2 && cx > d.x+2 {
			cx = d.x + 2
			cy += lineH
		}
		d.pdf.SetDrawColor(15, 23, 42)
		d.pdf.SetLineWidth(0.4)
		d.pdf.Rect(cx, cy+1.2, cbSize, cbSize, "D")
		if i == data.LeaveTypeActive {
			d.pdf.SetLineWidth(0.6)
			d.pdf.Line(cx+0.3, cy+1.5, cx+cbSize-0.3, cy+1.2+cbSize-0.3)
			d.pdf.Line(cx+cbSize-0.3, cy+1.5, cx+0.3, cy+1.2+cbSize-0.3)
		}
		d.pdf.SetTextColor(15, 23, 42)
		d.pdf.SetXY(cx+cbSize+gap-1.5, cy+1)
		d.pdf.CellFormat(txtW+2, 4.5, opt, "", 0, "L", false, 0, "")
		cx += itemW + 5
	}
	rows := 1
	if len(data.LeaveTypeOptions) > 4 {
		rows = 2
	}
	d.y = cy + lineH
	_ = rows
}

func (d *leaveFormDraw) textLines(title string, value string, lines int) {
	d.need(8)
	h := 7.0
	d.pdf.SetDrawColor(203, 213, 225)
	d.pdf.SetLineWidth(0.3)
	d.pdf.Rect(d.x, d.y, d.w, h, "D")
	d.pdf.SetFont(d.font, "B", 7.5)
	d.pdf.SetTextColor(71, 85, 105)
	d.pdf.SetXY(d.x+2, d.y+1.4)
	d.pdf.CellFormat(d.w-4, 4.5, title+":", "", 0, "L", false, 0, "")
	d.y += h

	linesUsed := 0
	if strings.TrimSpace(value) != "" {
		for _, ln := range strings.Split(value, "\n") {
			if linesUsed >= lines {
				break
			}
			d.need(8)
			d.pdf.SetFont(d.font, "", 7.5)
			d.pdf.SetTextColor(15, 23, 42)
			d.pdf.SetXY(d.x+2, d.y+1.2)
			d.pdf.CellFormat(d.w-4, 4.5, ln, "", 0, "L", false, 0, "")
			d.y += 6.5
			linesUsed++
		}
	}
	for linesUsed < lines {
		d.need(8)
		d.pdf.SetDrawColor(203, 213, 225)
		d.pdf.Line(d.x+2, d.y+3.5, d.x+d.w-2, d.y+3.5)
		d.y += 6.5
		linesUsed++
	}
}

func (d *leaveFormDraw) table(headers []string, rows [][]string) {
	d.need(8)
	colCount := len(headers)
	colW := d.w / float64(colCount)
	h := 7.0
	d.pdf.SetFillColor(241, 245, 249)
	d.pdf.SetDrawColor(203, 213, 225)
	d.pdf.SetLineWidth(0.3)
	for i, hdr := range headers {
		d.pdf.SetFont(d.font, "B", 7.5)
		d.pdf.SetTextColor(30, 58, 138)
		d.pdf.Rect(d.x+float64(i)*colW, d.y, colW, h, "DF")
		d.pdf.SetXY(d.x+float64(i)*colW+1, d.y+1.4)
		d.pdf.CellFormat(colW-2, 4.5, hdr, "", 0, "C", false, 0, "")
	}
	d.y += h
	for _, r := range rows {
		d.need(8)
		for i, v := range r {
			d.pdf.SetFont(d.font, "", 7.5)
			d.pdf.SetTextColor(15, 23, 42)
			d.pdf.Rect(d.x+float64(i)*colW, d.y, colW, h, "D")
			d.pdf.SetXY(d.x+float64(i)*colW+2, d.y+1.4)
			d.pdf.CellFormat(colW-4, 4.5, v, "", 0, "L", false, 0, "")
		}
		d.y += h
	}
}

func renderLeaveFormPDF(d *leaveFormDraw, data leaveFormData, labels leaveFormLabels) {
	pdf := d.pdf
	s := 1.0

	// ---- Header band ----
	pdf.SetFillColor(15, 23, 42)
	pdf.Rect(d.x, d.y, d.w, 22*s, "F")
	pdf.SetFont(d.font, "B", 13*s)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(d.x+2, d.y+2)
	pdf.CellFormat(d.w-4, 7*s, data.BrandName, "", 0, "C", false, 0, "")
	pdf.SetFont(d.font, "", 7*s)
	pdf.SetTextColor(148, 163, 184)
	pdf.SetXY(d.x+2, d.y+9)
	pdf.CellFormat(d.w-4, 5*s, labels.SystemTitle, "", 0, "C", false, 0, "")
	pdf.SetFont(d.font, "B", 10*s)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(d.x+2, d.y+14.5)
	pdf.CellFormat(d.w-4, 6*s, labels.FormTitle, "", 0, "C", false, 0, "")
	d.y += 22*s + 3*s

	// ---- Company / App box ----
	d.pair(labels.Company, data.Company, labels.AppNo, data.AppNo)
	d.pair(labels.Branch, data.Branch, labels.AppDate, data.AppDate)
	d.y += 2

	// ---- Employee information ----
	d.sectionTitle(labels.EmployeeInfo)
	d.pair(labels.EmployeeID, data.EmployeeID, labels.Name, data.Name)
	d.pair(labels.CardNo, data.CardNo, labels.Department, data.Department)
	d.pair(labels.Section, data.Section, labels.Designation, data.Designation)
	d.pair(labels.Shift, data.Shift, labels.EmpType, data.EmpType)
	d.pair(labels.JoiningDate, data.JoiningDate, labels.ReportsTo, data.ReportsTo)
	d.y += 2

	// ---- Leave details ----
	d.sectionTitle(labels.LeaveDetails)
	d.need(8)
	pdf.SetFont(d.font, "B", 7.5)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetXY(d.x+2, d.y+1)
	pdf.CellFormat(d.w-4, 4.5, labels.LeaveTypeLabel+":", "", 0, "L", false, 0, "")
	d.y += 6
	d.renderLeaveTypeOptions(data)
	d.y += 1

	d.pair(labels.FromDate, data.FromDate, labels.ToDate, data.ToDate)
	d.pair(labels.TotalDays, data.TotalDays, labels.EmergencyPhone, data.EmergencyPhone)
	d.row(labels.AddressDuring, data.AddressDuring)
	d.textLines(labels.Reason, data.Reason, 3)
	d.y += 2

	// ---- Leave balance ----
	d.sectionTitle(labels.LeaveBalance)
	balanceRows := make([][]string, 0, len(data.Balances))
	for _, b := range data.Balances {
		balanceRows = append(balanceRows, []string{b.LeaveType, b.Entitled, b.Used, b.Remaining})
	}
	d.table([]string{labels.BalLeaveType, labels.BalEntitled, labels.BalUsed, labels.BalRemaining}, balanceRows)
	d.y += 2

	// ---- Work handover ----
	d.sectionTitle(labels.Handover)
	d.row(labels.HandoverTo, data.HandoverTo)
	d.pair(labels.HandoverDept, data.HandoverDept, labels.HandoverDesig, data.HandoverDesig)
	d.textLines(labels.HandoverDetails, "", 2)
	d.y += 2

	// ---- Approval workflow ----
	d.sectionTitle(labels.Approval)

	approvalRows := []struct {
		role  string
		lines []string
	}{
		{labels.EmployeeRole, []string{labels.Signature + ": ______________    " + labels.Date + ": ______________"}},
		{labels.SupervisorRole, []string{
			checkBoxText(labels.Approved, data.ApprovedYes) + "    " + checkBoxText(labels.Rejected, data.RejectedYes) + "    " + labels.Signature + ": ______________",
			labels.Remarks + ": ________________________________________",
		}},
		{labels.DeptHeadRole, []string{
			checkBoxText(labels.Approved, data.ApprovedYes) + "    " + checkBoxText(labels.Rejected, data.RejectedYes) + "    " + labels.Signature + ": ______________",
			labels.Remarks + ": ________________________________________",
		}},
		{labels.HRRole, []string{
			checkBoxText(labels.Approved, data.ApprovedYes) + "    " + checkBoxText(labels.Rejected, data.RejectedYes) + "    " + labels.Signature + ": ______________",
			labels.BalanceUpdated + ": " + checkBoxText(labels.Yes, data.ApprovedYes) + "  " + checkBoxText(labels.No, data.RejectedYes),
		}},
		{labels.FinalRole, []string{labels.HRManagerSig + ": ______________________"}},
	}

	roleW := d.w * 0.28
	contentW := d.w - roleW
	for _, r := range approvalRows {
		lineCount := len(r.lines)
		if lineCount < 1 {
			lineCount = 1
		}
		rowH := 6.5*float64(lineCount) + 3
		d.need(rowH + 2)
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetLineWidth(0.3)
		pdf.Rect(d.x, d.y, roleW, rowH, "D")
		pdf.Rect(d.x+roleW, d.y, contentW, rowH, "D")
		pdf.SetFont(d.font, "B", 7.5)
		pdf.SetTextColor(30, 58, 138)
		roleX := d.x + roleW/2 - pdf.GetStringWidth(r.role)/2
		pdf.SetXY(roleX, d.y+rowH/2-2)
		pdf.CellFormat(roleW, 4.5, r.role, "", 0, "C", false, 0, "")
		pdf.SetFont(d.font, "", 7.5)
		pdf.SetTextColor(15, 23, 42)
		for li, ln := range r.lines {
			pdf.SetXY(d.x+roleW+2, d.y+1+float64(li)*6.5)
			pdf.CellFormat(contentW-4, 4.5, ln, "", 0, "L", false, 0, "")
		}
		d.y += rowH
	}
	d.y += 2

	// ---- Notes ----
	d.sectionTitle(labels.Notes)
	for _, n := range data.Notes {
		d.need(7)
		pdf.SetFont(d.font, "", 7.5)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(d.x+4, d.y+1)
		pdf.CellFormat(d.w-8, 4.5, "• "+n, "", 0, "L", false, 0, "")
		d.y += 6.5
	}
	d.y += 2

	// ---- Footer ----
	d.need(10)
	pdf.SetFont(d.font, "", 7)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(d.x, d.y)
	pdf.CellFormat(d.w, 4, data.GeneratedBy, "", 0, "C", false, 0, "")
	d.y += 5
	pdf.SetXY(d.x, d.y)
	pdf.CellFormat(d.w, 4, labels.DocumentNo+": "+data.DocumentNo+"    |    "+labels.PrintDate+": "+data.PrintDate, "", 0, "C", false, 0, "")
}

func checkBoxText(label string, checked bool) string {
	mark := "[ ]"
	if checked {
		mark = "[x]"
	}
	return mark + " " + label
}

func renderLeaveFormPDFPage(c *gofpdf.Fpdf, font string, lang string, data leaveFormData, labels leaveFormLabels) {
	d := &leaveFormDraw{
		pdf:    c,
		font:   font,
		lang:   lang,
		labels: labels,
		x:      leaveFormLeft,
		w:      leaveFormW,
		y:      leaveFormTop,
	}
	renderLeaveFormPDF(d, data, labels)
}
