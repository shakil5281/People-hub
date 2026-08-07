package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type separationFormLabels struct {
	SystemTitle string
	FormTitle   string

	Company   string
	Branch    string
	AppNo     string
	AppDate   string

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
	EmpType      string

	SepDetails  string
	SepType     string
	SepDate     string
	NoticePeriod string
	Reason      string

	ClearanceChecklist string
	IDCardReturned    string
	AssetsReturned    string
	AccountsCleared   string
	DuesPaid          string
	Yes               string
	No                string

	Approval       string
	EmployeeRole   string
	SupervisorRole string
	DeptHeadRole   string
	AccountsRole   string
	HRRole         string
	Signature      string
	Date           string
	Remarks        string

	GeneratedBy string
	DocumentNo  string
	PrintDate   string
}

var separationEnLabels = separationFormLabels{
	SystemTitle: "PEOPLEHUB HR & PAYROLL MANAGEMENT SYSTEM",
	FormTitle:   "EMPLOYEE RESIGNATION / SEPARATION FORM",

	Company: "Company",
	Branch:  "Branch",
	AppNo:   "Ref No",
	AppDate: "Date",

	EmployeeInfo: "EMPLOYEE INFORMATION",
	EmployeeID:   "Employee ID",
	CardNo:       "Card No",
	Name:         "Employee Name",
	Mobile:       "Mobile Number",
	Department:   "Department",
	Section:      "Section",
	Designation:  "Designation",
	Grade:        "Grade",
	Shift:        "Shift",
	JoiningDate:  "Joining Date",
	EmpType:      "Employment Type",

	SepDetails:   "SEPARATION DETAILS",
	SepType:      "Separation Type",
	SepDate:      "Effective Date",
	NoticePeriod: "Notice Period",
	Reason:       "Reason for Separation",

	ClearanceChecklist: "CLEARANCE CHECKLIST",
	IDCardReturned:     "ID Card Returned",
	AssetsReturned:     "Assets / Equipment Returned",
	AccountsCleared:    "Accounts / Loan Cleared",
	DuesPaid:           "Final Dues Settled",
	Yes:                "Yes",
	No:                 "No",

	Approval:       "APPROVAL & AUTHORIZATION WORKFLOW",
	EmployeeRole:   "Employee",
	SupervisorRole: "Supervisor",
	DeptHeadRole:   "Department Head",
	AccountsRole:   "Accounts / Audit",
	HRRole:         "HR Manager",
	Signature:      "Signature",
	Date:           "Date",
	Remarks:        "Remarks",

	GeneratedBy: "PeopleHub HR & Payroll System",
	DocumentNo:  "Document No",
	PrintDate:   "Print Date",
}

// SutonnyMJ Classic Font Labels for Resign Form
var separationBnLabels = separationFormLabels{
	SystemTitle: "wccevnve GBPAvi A¨vÛ ceIivj wm‡÷g",
	FormTitle:   "ceZ¨vMcÎ / B¯ÍdvcÎ (RESIGNATION FORM)",

	Company: "tKv¤úvbxi bvg",
	Branch:  "kvLv/KviLvbv",
	AppNo:   "m¥viK bs",
	AppDate: "ZvwiL",

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
	JoiningDate:  "PvKzix‡Z tvevMvtbi ZvwiL",
	EmpType:      "PvKzixi aiY",

	SepDetails:   "ceZ¨v‡Mi weeiY",
	SepType:      "ceZ¨v‡Mi aiY",
	SepDate:      "Kvh©Ki ZvwiL",
	NoticePeriod: "tbowUm wcwiqb",
	Reason:       "ceZ¨v‡Mi KviY",

	ClearanceChecklist: "QvojcÎ pK (CLEARANCE CHECKLIST)",
	IDCardReturned:     "AvBWw KviW jgv",
	AssetsReturned:     "ymivwx gvjvgvj jgv",
	AccountsCleared:    "wnmei/FY cmicvewb",
	DuesPaid:           "P~ovšÍ cvIbv cwi‡kva",
	Yes:                "n¨uv",
	No:                 "bv",

	Approval:       "Abzgveb cÖwµqv",
	EmployeeRole:   "Kg©Pvix",
	SupervisorRole: "mycvifvBRvi",
	DeptHeadRole:   "wefvMxq cÖavb",
	AccountsRole:   "wnmei/AwWU wefvM",
	HRRole:         "GBPAvi g¨vtbRvi",
	Signature:      "¯^v¶i",
	Date:           "ZvwiL",
	Remarks:        "gšÍee¨",

	GeneratedBy: "wccevnve GBPAvi A¨vÛ ceIivj wm‡÷g `viv ‰Zix",
	DocumentNo:  "bw_ b¤^vi",
	PrintDate:   "wcÖ‡›Ui ZvwiL",
}

// ExportPDF godoc
//
//	@Summary      Export resignation / separation form to PDF
//	@Description  Generate official resignation form PDF in English or SutonnyMJ Bangla
//	@Tags         Separations
//	@Security     BearerAuth
//	@Produce      application/pdf
//	@Param        id    path  string true  "Separation ID"
//	@Param        lang  query string false "Language: en (default) or bn"
//	@Success      200   {file} binary
//	@Failure      404   {object} map[string]string
//	@Failure      500   {object} map[string]string
//	@Router       /separations/{id}/export/pdf [get]
func (h *SeparationHandler) ExportPDF(c *gin.Context) {
	id := c.Param("id")
	lang := c.DefaultQuery("lang", "en")
	if lang != "bn" {
		lang = "en"
	}

	sep, err := h.repo.FindByID(id)
	if err != nil || sep == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "separation record not found"})
		return
	}

	labels := separationEnLabels
	if lang == "bn" {
		labels = separationBnLabels
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 10, 12)
	pdf.SetAutoPageBreak(true, 12)
	font := payslipPDFFont(pdf, lang)

	emp, _ := h.repo.FindEmployeeByCode(sep.EmployeeID)

	pdf.AddPage()
	drawSeparationFormPDF(pdf, font, lang, sep, emp, labels)

	filename := fmt.Sprintf("resign_form_%s_%s.pdf", sep.EmployeeID, sep.Date)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}

func drawSeparationFormPDF(pdf *gofpdf.Fpdf, font, lang string, sep *models.Separation, emp *models.Employee, labels separationFormLabels) {
	w := 186.0
	x := 12.0
	y := 10.0

	companyName := "PeopleHub Ltd."
	companyAddr := "Factory / Head Office, Bangladesh"
	companyPhone := ""

	if emp != nil && emp.Company.ID != "" {
		companyName = companyNameFor(lang, emp.Company)
		companyAddr = companyAddress(emp.Company, lang)
		companyPhone = emp.Company.Phone
	}
	if lang == "bn" {
		companyName = utils.UnicodeToBijoy(companyName)
		companyAddr = utils.UnicodeToBijoy(companyAddr)
	}

	// ---- Header band ----
	pdf.SetFillColor(15, 23, 42)
	pdf.Rect(x, y, w, 22, "F")
	pdf.SetFont(font, "B", 13)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x+2, y+2)
	pdf.CellFormat(w-4, 7, companyName, "", 0, "C", false, 0, "")

	pdf.SetFont(font, "", 7)
	pdf.SetTextColor(148, 163, 184)
	pdf.SetXY(x+2, y+9)
	pdf.CellFormat(w-4, 5, companyAddr+"  "+companyPhone, "", 0, "C", false, 0, "")

	pdf.SetFont(font, "B", 10)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x+2, y+14.5)
	pdf.CellFormat(w-4, 6, labels.FormTitle, "", 0, "C", false, 0, "")
	y += 25

	pdf.SetTextColor(0, 0, 0)

	// ---- Ref info ----
	empName := sep.Employee
	desigNameStr := ""
	deptNameStr := ""
	secNameStr := ""
	mobileStr := ""
	joinDateStr := "-"
	empTypeStr := ""
	gradeStr := ""

	if emp != nil {
		empName = employeeNameFor(lang, emp)
		desigNameStr = designationName(emp.DesignationRef, lang)
		deptNameStr = departmentName(emp.Department, lang)
		secNameStr = sectionName(emp.SectionRef, lang)
		mobileStr = emp.Phone
		empTypeStr = emp.EmployeeType
		gradeStr = emp.Grade
		if !emp.JoiningDate.IsZero() {
			joinDateStr = emp.JoiningDate.Format("02-01-2006")
		}
	} else if sep.Department.ID != "" {
		deptNameStr = departmentName(&sep.Department, lang)
	}
	if lang == "bn" {
		empName = utils.UnicodeToBijoy(empName)
		desigNameStr = utils.UnicodeToBijoy(desigNameStr)
		deptNameStr = utils.UnicodeToBijoy(deptNameStr)
		secNameStr = utils.UnicodeToBijoy(secNameStr)
		empTypeStr = utils.UnicodeToBijoy(empTypeStr)
	}

	drawPairRowPDF(pdf, font, x, y, w, labels.AppNo+": "+shortPayrollNo(sep.ID), labels.AppDate+": "+sep.Date)
	y += 7.5

	// ---- Employee Information ----
	y = drawFormSectionTitle(pdf, font, x, y, w, labels.EmployeeInfo)
	drawPairRowPDF(pdf, font, x, y, w, labels.EmployeeID+": "+sep.EmployeeID, labels.Name+": "+empName)
	y += 7
	drawPairRowPDF(pdf, font, x, y, w, labels.Department+": "+deptNameStr, labels.Section+": "+secNameStr)
	y += 7
	drawPairRowPDF(pdf, font, x, y, w, labels.Designation+": "+desigNameStr, labels.Grade+": "+gradeStr)
	y += 7
	drawPairRowPDF(pdf, font, x, y, w, labels.JoiningDate+": "+joinDateStr, labels.Mobile+": "+mobileStr)
	y += 7.5

	// ---- Separation Details ----
	y = drawFormSectionTitle(pdf, font, x, y, w, labels.SepDetails)
	sepTypeDisplay := sep.Type
	if lang == "bn" {
		switch strings.ToLower(sep.Type) {
		case "resign":
			sepTypeDisplay = "ceZ¨vMcÎ (Resign)"
		case "lefty":
			sepTypeDisplay = "tedUx (Lefty)"
		case "close":
			sepTypeDisplay = "vevm (Close)"
		}
	}
	drawPairRowPDF(pdf, font, x, y, w, labels.SepType+": "+sepTypeDisplay, labels.SepDate+": "+sep.Date)
	y += 7

	reasonStr := sep.Reason
	if reasonStr == "" {
		reasonStr = "-"
	}
	if lang == "bn" {
		reasonStr = utils.UnicodeToBijoy(reasonStr)
	}
	drawFullRowPDF(pdf, font, x, y, w, labels.Reason+": "+reasonStr)
	y += 8

	// ---- Clearance Checklist ----
	y = drawFormSectionTitle(pdf, font, x, y, w, labels.ClearanceChecklist)
	drawChecklistGridPDF(pdf, font, x, y, w, labels)
	y += 16

	// ---- Approval Workflow ----
	y = drawFormSectionTitle(pdf, font, x, y, w, labels.Approval)

	sigBoxW := w / 5
	sigBoxH := 24.0
	roles := []string{labels.EmployeeRole, labels.SupervisorRole, labels.DeptHeadRole, labels.AccountsRole, labels.HRRole}

	for i, r := range roles {
		cx := x + float64(i)*sigBoxW
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetLineWidth(0.3)
		pdf.Rect(cx, y, sigBoxW, sigBoxH, "D")

		pdf.SetFont(font, "B", 7)
		pdf.SetTextColor(30, 58, 138)
		pdf.SetXY(cx+1, y+2)
		pdf.CellFormat(sigBoxW-2, 4, r, "", 0, "C", false, 0, "")

		pdf.SetFont(font, "", 6.5)
		pdf.SetTextColor(100, 116, 139)
		pdf.SetXY(cx+1, y+sigBoxH-5)
		pdf.CellFormat(sigBoxW-2, 4, labels.Signature, "", 0, "C", false, 0, "")
	}
	y += sigBoxH + 4

	// ---- Footer ----
	pdf.SetFont(font, "", 7)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 4, labels.GeneratedBy, "", 0, "C", false, 0, "")
	y += 4.5
	pdf.SetXY(x, y)
	pdf.CellFormat(w, 4, fmt.Sprintf("%s: %s   |   %s: %s", labels.DocumentNo, shortPayrollNo(sep.ID), labels.PrintDate, time.Now().Format("02-01-2006")), "", 0, "C", false, 0, "")
}

func drawFormSectionTitle(pdf *gofpdf.Fpdf, font string, x, y, w float64, title string) float64 {
	pdf.SetFillColor(241, 245, 249)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.3)
	pdf.SetFont(font, "B", 8)
	pdf.SetTextColor(30, 58, 138)
	pdf.Rect(x, y, w, 6, "DF")
	pdf.SetXY(x+2, y+1)
	pdf.CellFormat(w-4, 4, title, "", 0, "L", false, 0, "")
	return y + 6
}

func drawPairRowPDF(pdf *gofpdf.Fpdf, font string, x, y, w float64, col1, col2 string) {
	half := w / 2
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.3)
	pdf.Rect(x, y, half, 7, "D")
	pdf.Rect(x+half, y, half, 7, "D")
	pdf.SetFont(font, "", 7.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetXY(x+2, y+1.2)
	pdf.CellFormat(half-4, 4.5, col1, "", 0, "L", false, 0, "")
	pdf.SetXY(x+half+2, y+1.2)
	pdf.CellFormat(half-4, 4.5, col2, "", 0, "L", false, 0, "")
}

func drawFullRowPDF(pdf *gofpdf.Fpdf, font string, x, y, w float64, text string) {
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.3)
	pdf.Rect(x, y, w, 7, "D")
	pdf.SetFont(font, "", 7.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetXY(x+2, y+1.2)
	pdf.CellFormat(w-4, 4.5, text, "", 0, "L", false, 0, "")
}

func drawChecklistGridPDF(pdf *gofpdf.Fpdf, font string, x, y, w float64, labels separationFormLabels) {
	half := w / 2
	items := [][2]string{
		{labels.IDCardReturned, labels.AssetsReturned},
		{labels.AccountsCleared, labels.DuesPaid},
	}
	pdf.SetFont(font, "", 7.5)
	pdf.SetTextColor(15, 23, 42)
	for i, row := range items {
		ry := y + float64(i)*7
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetLineWidth(0.3)
		pdf.Rect(x, ry, half, 7, "D")
		pdf.Rect(x+half, ry, half, 7, "D")

		pdf.SetXY(x+2, ry+1.2)
		pdf.CellFormat(half-20, 4.5, row[0], "", 0, "L", false, 0, "")
		pdf.SetXY(x+half-18, ry+1.2)
		pdf.CellFormat(16, 4.5, "[ ] "+labels.Yes+"  [ ] "+labels.No, "", 0, "R", false, 0, "")

		pdf.SetXY(x+half+2, ry+1.2)
		pdf.CellFormat(half-20, 4.5, row[1], "", 0, "L", false, 0, "")
		pdf.SetXY(x+w-18, ry+1.2)
		pdf.CellFormat(16, 4.5, "[ ] "+labels.Yes+"  [ ] "+labels.No, "", 0, "R", false, 0, "")
	}
}
