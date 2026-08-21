package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

const (
	payslipPageW = 210.0
	payslipPageH = 297.0
	payslipLeft  = 5.0
	payslipRight = 108.0
	payslipRow1  = 6.0
	payslipRow2  = 165.0
)

const (
	payslipCardW = 97.0
	payslipCardH = 129.0
)

func exportSinglePayslipPDF(c *gin.Context, salary *models.Salary, month, year int, lang string) {
	labels := buildPayslipLabels(lang, salary.Company)

	cardOffice := buildPayslipCardWithCopy(salary, month, year, lang, labels, labels.CopyOffice)
	cardEmp := buildPayslipCardWithCopy(salary, month, year, lang, labels, labels.CopyEmployee)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	font := payslipPDFFont(pdf, lang)

	// Single payslip: Left side Office Copy, Right side Employee Copy
	drawPayslipCardPDF(pdf, font, payslipLeft, payslipRow1, payslipCardW, payslipCardH, cardOffice, labels)
	drawPayslipCardPDF(pdf, font, payslipRight, payslipRow1, payslipCardW, payslipCardH, cardEmp, labels)
	drawPayslipCutLines(pdf)

	langSuffix := lang
	if langSuffix == "" {
		langSuffix = "en"
	}
	filename := fmt.Sprintf("payslip_%s_%s_%d_%s.pdf", salary.Employee.EmployeeID, monthName(month, "en"), year, langSuffix)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}

func exportBulkPayslipPDF(c *gin.Context, salaries []models.Salary, month, year int, lang string) {
	labels := buildPayslipLabels(lang, salaries[0].Company)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	font := payslipPDFFont(pdf, lang)

	// 2 employees per A4 page (4 cards total: Office & Employee copy for each)
	for i := range salaries {
		empIdxOnPage := i % 2
		if i > 0 && empIdxOnPage == 0 {
			pdf.AddPage()
		}

		cardOffice := buildPayslipCardWithCopy(&salaries[i], month, year, lang, labels, labels.CopyOffice)
		cardEmp := buildPayslipCardWithCopy(&salaries[i], month, year, lang, labels, labels.CopyEmployee)

		y := payslipRow1
		if empIdxOnPage == 1 {
			y = payslipRow2
		}

		drawPayslipCardPDF(pdf, font, payslipLeft, y, payslipCardW, payslipCardH, cardOffice, labels)
		drawPayslipCardPDF(pdf, font, payslipRight, y, payslipCardW, payslipCardH, cardEmp, labels)

		if empIdxOnPage == 1 || i == len(salaries)-1 {
			drawPayslipCutLines(pdf)
		}
	}

	langSuffix := lang
	if langSuffix == "" {
		langSuffix = "en"
	}
	filename := fmt.Sprintf("payslips_%s_%d_%s.pdf", monthName(month, "en"), year, langSuffix)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}

func payslipPDFFont(pdf *gofpdf.Fpdf, lang string) string {
	if lang == "bn" {
		return loadBanglaFont(pdf)
	}
	return "Helvetica"
}

func drawPayslipCardPDF(pdf *gofpdf.Fpdf, font string, x, y, w, h float64, card *payslipCard, labels payslipLabels) {
	s := w / payslipCardW

	// ---- 1. Header band: Company Name (White background, Black text, Gray border) ----
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(162, 162, 162)
	pdf.SetLineWidth(0.20)
	pdf.Rect(x, y, w, 6.5*s, "DF")
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont(font, "B", 9.5*s)
	pdf.SetXY(x+1.5*s, y+1.2*s)
	pdf.CellFormat(w-3*s, 4.0*s, card.CompanyName, "", 0, "C", false, 0, "")
	curY := y + 6.5*s

	// ---- 2. Month, Copy Label & Print Date ----
	pdf.SetFont(font, "", 6.5*s)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(162, 162, 162)
	pdf.SetLineWidth(0.20)
	pdf.Rect(x, curY, w, 4.3*s, "D")
	pdf.SetXY(x+1.5*s, curY+0.8*s)
	metaStr := fmt.Sprintf("%s: %s   |   %s   |   %s: %s",
		labels.PayrollMonth, card.PayrollMonth, card.CopyLabel, labels.PrintDateLabel, card.PrintDate)
	pdf.CellFormat(w-3*s, 2.6*s, metaStr, "", 0, "C", false, 0, "")
	curY += 4.3 * s

	// ---- 3. Employee Information ----
	curY += 1.2 * s
	curY = drawSectionTitlePDF(pdf, font, s, x, curY, w, labels.EmployeeInfo)
	curY = drawFieldGridPDF(pdf, font, s, x, curY, w, card.EmployeeInfo, 2)
	curY += 1.5 * s

	// ---- 4. Attendance Summary (5 columns: 10 fields in 2 rows) ----
	curY = drawSectionTitlePDF(pdf, font, s, x, curY, w, labels.Attendance)
	curY = drawFieldGridPDF(pdf, font, s, x, curY, w, card.Attendance, 5)
	curY += 1.5 * s

	// ---- 5. Earnings (left) + Deductions (right) side-by-side ----
	startY := curY
	halfW := (w - 2.0*s) / 2
	endY1 := drawMoneyTablePDF(pdf, font, s, x, startY, halfW, labels.Earnings, card.Earnings, card.EarningsTotal)
	endY2 := drawMoneyTablePDF(pdf, font, s, x+halfW+2.0*s, startY, halfW, labels.Deductions, card.Deductions, card.DeductionsTotal)
	if endY1 > endY2 {
		curY = endY1
	} else {
		curY = endY2
	}
	curY += 1.0 * s

	// Net salary black box (6.0mm height)
	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 9.5*s)
	pdf.Rect(x, curY, w, 6.0*s, "F")
	pdf.SetXY(x+2*s, curY+1.3*s)
	pdf.CellFormat(w-4*s, 3.4*s, fmt.Sprintf("%s : BDT %s", labels.NetSalary, card.NetSalary), "", 0, "C", false, 0, "")
	curY += 6.0*s + 1.5*s

	// ---- 6. Signatures (Left: Employee Signature, Right: Approved By) ----
	pdf.SetDrawColor(162, 162, 162)
	pdf.SetLineWidth(0.20)
	pdf.SetFont(font, "", 6.2*s)
	pdf.SetTextColor(0, 0, 0)

	halfSigW := w / 2.0
	// Left: Employee Signature
	pdf.Rect(x, curY, halfSigW, 10.0*s, "D")
	pdf.SetXY(x+0.5*s, curY+7.2*s)
	pdf.CellFormat(halfSigW-1.0*s, 2.2*s, card.EmployeeSig, "", 0, "C", false, 0, "")

	// Right: Approved By
	pdf.Rect(x+halfSigW, curY, halfSigW, 10.0*s, "D")
	pdf.SetXY(x+halfSigW+0.5*s, curY+7.2*s)
	pdf.CellFormat(halfSigW-1.0*s, 2.2*s, card.ApprovedBy, "", 0, "C", false, 0, "")

	curY += 10.0 * s

	// Outer border for payslip card (Fits exact height with zero bottom gap)
	cardH := curY - y
	pdf.SetDrawColor(162, 162, 162)
	pdf.SetLineWidth(0.20)
	pdf.Rect(x, y, w, cardH, "D")
}

func drawSectionTitlePDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title string) float64 {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(162, 162, 162)
	pdf.SetLineWidth(0.20)
	pdf.SetFont(font, "B", 6.8*s)
	pdf.SetTextColor(0, 0, 0)
	pdf.Rect(x, y, w, 4.3*s, "DF")
	pdf.SetXY(x+1.5*s, y+0.8*s)
	pdf.CellFormat(w-3*s, 2.6*s, title, "", 0, "L", false, 0, "")
	return y + 4.3*s
}

func drawFieldGridPDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, fields []payslipField, cols int) float64 {
	rowH := 4.3 * s
	colW := w / float64(cols)
	for i, fld := range fields {
		col := i % cols
		row := i / cols
		px := x + float64(col)*colW
		py := y + float64(row)*rowH
		pdf.SetDrawColor(162, 162, 162)
		pdf.SetLineWidth(0.20)
		pdf.Rect(px, py, colW, rowH, "D")

		labelStr := fld.Label
		valStr := fld.Value

		labelFontSize := 5.8 * s
		if cols >= 4 && len(labelStr) > 10 {
			labelFontSize = 4.6 * s
		}
		pdf.SetFont(font, "B", labelFontSize)
		pdf.SetTextColor(0, 0, 0)

		labelWidthRatio := 0.44
		if cols == 4 || cols == 5 {
			labelWidthRatio = 0.62
		}
		lblW := colW * labelWidthRatio
		pdf.SetXY(px+0.6*s, py+1.0*s)
		pdf.CellFormat(lblW-0.6*s, 2.2*s, labelStr, "", 0, "L", false, 0, "")

		valFontSize := 5.8 * s
		if cols >= 4 || len(valStr) > 20 {
			valFontSize = 4.8 * s
		}
		pdf.SetFont(font, "", valFontSize)
		pdf.SetTextColor(0, 0, 0)
		valW := colW - lblW
		pdf.SetXY(px+lblW, py+1.0*s)
		pdf.CellFormat(valW-0.4*s, 2.2*s, valStr, "", 0, "R", false, 0, "")
	}
	rows := (len(fields) + cols - 1) / cols
	return y + float64(rows)*rowH
}

func drawMoneyTablePDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title string, rows []payslipRow, total string) float64 {
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(162, 162, 162)
	pdf.SetLineWidth(0.20)
	pdf.SetFont(font, "B", 6.2*s)
	pdf.SetTextColor(0, 0, 0)
	pdf.Rect(x, y, w, 4.3*s, "DF")
	pdf.SetXY(x+1.5*s, y+0.8*s)
	pdf.CellFormat(w*0.64, 2.6*s, title, "", 0, "L", false, 0, "")
	pdf.SetXY(x+1.5*s, y+0.8*s)
	pdf.CellFormat(w-3*s, 2.6*s, "BDT", "", 0, "R", false, 0, "")

	y2 := y + 4.3*s
	descW := w * 0.64
	amtW := w * 0.36
	for _, r := range rows {
		pdf.SetDrawColor(162, 162, 162)
		pdf.SetLineWidth(0.20)
		pdf.Rect(x, y2, w, 4.3*s, "D")
		labelStr := r.Label
		if len(labelStr) > 20 {
			pdf.SetFont(font, "", 5.0*s)
		} else {
			pdf.SetFont(font, "", 5.6*s)
		}
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(x+1.0*s, y2+1.0*s)
		pdf.CellFormat(descW-1.2*s, 2.2*s, labelStr, "", 0, "L", false, 0, "")

		pdf.SetFont(font, "", 5.6*s)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(x+descW, y2+1.0*s)
		pdf.CellFormat(amtW-1.2*s, 2.2*s, r.Amount, "", 0, "R", false, 0, "")
		y2 += 4.3 * s
	}

	return y2
}

func drawPayslipCutLines(pdf *gofpdf.Fpdf) {
	pdf.SetDrawColor(162, 162, 162)
	pdf.SetLineWidth(0.20)
	pdf.SetDashPattern([]float64{1.5, 1.2}, 0)

	// Center vertical cut line (X = 105.0mm)
	pdf.Line(105.0, 0, 105.0, 297.0)
	// Center horizontal cut line (Y = 148.5mm)
	pdf.Line(0, 148.5, 210.0, 148.5)

	pdf.SetDashPattern([]float64{0, 0}, 0)
	pdf.SetLineWidth(0.20)
	pdf.SetDrawColor(162, 162, 162)
	drawScissorsIconPDF(pdf, 105.0, 148.5, 3.5)
}

func drawScissorsIconPDF(pdf *gofpdf.Fpdf, cx, cy, r float64) {
	// finger loops
	pdf.Circle(cx-r*1.5, cy-r*0.6, r*0.7, "D")
	pdf.Circle(cx+r*1.5, cy-r*0.6, r*0.7, "D")
	// blades crossing at center
	pdf.SetLineWidth(0.5)
	pdf.Line(cx-r*0.9, cy-r*0.3, cx, cy)
	pdf.Line(cx+r*0.9, cy-r*0.3, cx, cy)
	pdf.Line(cx-r*0.9, cy-r*0.3, cx+r*2.0, cy+r*1.2)
	pdf.Line(cx+r*0.9, cy-r*0.3, cx-r*2.0, cy+r*1.2)
}
