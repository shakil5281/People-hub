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

	// Outer border for payslip card
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.35)
	pdf.Rect(x, y, w, h, "D")

	// ---- 1. Header band: Company Name ----
	pdf.SetFillColor(15, 23, 42)
	pdf.Rect(x, y, w, 6.0*s, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 8.5*s)
	pdf.SetXY(x+1.5*s, y+1.0*s)
	pdf.CellFormat(w-3*s, 4.0*s, card.CompanyName, "", 0, "C", false, 0, "")
	curY := y + 6.0*s

	// ---- 2. Copy Label & Payslip Title ----
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.Rect(x, curY, w, 4.5*s, "DF")
	pdf.SetFont(font, "B", 7.5*s)
	pdf.SetTextColor(217, 119, 6) // Amber #D97706
	pdf.SetXY(x+1.5*s, curY+0.8*s)
	pdf.CellFormat(w-3*s, 3.0*s, fmt.Sprintf("%s   -   %s", card.CopyLabel, card.PayslipWord), "", 0, "C", false, 0, "")
	curY += 4.5 * s

	// ---- 3. Month & Payroll No ----
	pdf.SetFont(font, "", 5.5*s)
	pdf.SetTextColor(15, 23, 42)
	pdf.Rect(x, curY, w, 3.8*s, "D")
	pdf.SetXY(x+1.5*s, curY+0.6*s)
	metaStr := fmt.Sprintf("%s: %s   |   %s: %s   |   %s: %s",
		labels.PayrollMonth, card.PayrollMonth, labels.PayrollNo, card.PayrollNo, labels.PrintDateLabel, card.PrintDate)
	pdf.CellFormat(w-3*s, 2.6*s, metaStr, "", 0, "C", false, 0, "")
	curY += 3.8 * s

	// ---- 4. Employee Information ----
	curY += 1.2 * s
	curY = drawSectionTitlePDF(pdf, font, s, x, curY, w, labels.EmployeeInfo)
	curY = drawFieldGridPDF(pdf, font, s, x, curY, w, card.EmployeeInfo, 2)
	curY += 1.5 * s

	// ---- 5. Attendance Summary ----
	curY = drawSectionTitlePDF(pdf, font, s, x, curY, w, labels.Attendance)
	curY = drawFieldGridPDF(pdf, font, s, x, curY, w, card.Attendance, 4)
	curY += 1.5 * s

	// ---- 6. Earnings (left) + Deductions (right) side-by-side ----
	startY := curY
	halfW := (w - 2.0*s) / 2
	endY1 := drawMoneyTablePDF(pdf, font, s, x, startY, halfW, labels.Earnings, card.Earnings, card.EarningsTotal)
	endY2 := drawMoneyTablePDF(pdf, font, s, x+halfW+2.0*s, startY, halfW, labels.Deductions, card.Deductions, card.DeductionsTotal)
	if endY1 > endY2 {
		curY = endY1
	} else {
		curY = endY2
	}

	// ---- 7. Summary ----
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFont(font, "B", 5.2*s)
	pdf.SetTextColor(30, 58, 138)
	pdf.Rect(x, curY, halfW, 3.8*s, "DF")
	pdf.SetXY(x+1.5*s, curY+0.6*s)
	pdf.CellFormat(halfW*0.5, 2.6*s, labels.GrossSalary, "", 0, "L", false, 0, "")

	pdf.SetFont(font, "B", 5.4*s)
	pdf.SetTextColor(15, 23, 42)
	pdf.Rect(x+halfW+2.0*s, curY, halfW, 3.8*s, "DF")
	pdf.SetXY(x+halfW+3.5*s, curY+0.6*s)
	pdf.CellFormat(halfW-3.0*s, 2.6*s, card.GrossSalary, "", 0, "R", false, 0, "")
	curY += 3.8 * s

	// Net salary green box
	pdf.SetFillColor(21, 128, 61)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 8*s)
	pdf.Rect(x, curY, w, 5.0*s, "F")
	pdf.SetXY(x+2*s, curY+0.8*s)
	pdf.CellFormat(w-4*s, 3.4*s, fmt.Sprintf("%s : BDT %s", labels.NetSalary, card.NetSalary), "", 0, "C", false, 0, "")
	curY += 5.0*s + 2.0*s // Empty gap after Net Salary (Height: 2mm)

	// ---- 8. Footer signatures ----
	sigCells := []payslipField{
		{"", card.PreparedBy},
		{"", card.CheckedBy},
		{"", card.ApprovedBy},
		{"", card.EmployeeSig},
	}
	cellW := w / 4
	for i, f := range sigCells {
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetLineWidth(0.2)
		pdf.SetFont(font, "", 5.2*s)
		pdf.SetTextColor(15, 23, 42)
		pdf.Rect(x+float64(i)*cellW, curY, cellW, 8.5*s, "D")
		pdf.SetXY(x+float64(i)*cellW+0.5*s, curY+6.0*s)
		pdf.CellFormat(cellW-1*s, 2.2*s, f.Value, "", 0, "C", false, 0, "")
	}
	curY += 8.5*s + 1.0*s

	// ---- 9. Generated / confidential line ----
	pdf.SetFont(font, "", 4.5*s)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(x+1.5*s, curY)
	pdf.CellFormat(w-3*s, 2.2*s, card.GeneratedBy+"   •   "+card.Confidential, "", 0, "C", false, 0, "")
}

func drawSectionTitlePDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title string) float64 {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFont(font, "B", 5.8*s)
	pdf.SetTextColor(30, 58, 138)
	pdf.Rect(x, y, w, 3.8*s, "DF")
	pdf.SetXY(x+1.5*s, y+0.6*s)
	pdf.CellFormat(w-3*s, 2.6*s, title, "", 0, "L", false, 0, "")
	return y + 3.8*s
}

func drawFieldGridPDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, fields []payslipField, cols int) float64 {
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

		labelFontSize := 4.8 * s
		if cols == 4 && len(labelStr) > 10 {
			labelFontSize = 4.0 * s
		}
		pdf.SetFont(font, "B", labelFontSize)
		pdf.SetTextColor(100, 116, 139)

		labelWidthRatio := 0.44
		if cols == 4 {
			labelWidthRatio = 0.68
		}
		lblW := colW * labelWidthRatio
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
		pdf.CellFormat(valW-0.4*s, 2.2*s, valStr, "", 0, "R", false, 0, "")
	}
	rows := (len(fields) + cols - 1) / cols
	return y + float64(rows)*rowH
}

func drawMoneyTablePDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title string, rows []payslipRow, total string) float64 {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFont(font, "B", 5.2*s)
	pdf.SetTextColor(30, 58, 138)
	pdf.Rect(x, y, w, 3.8*s, "DF")
	pdf.SetXY(x+1.5*s, y+0.6*s)
	pdf.CellFormat(w*0.64, 2.6*s, title, "", 0, "L", false, 0, "")
	pdf.SetXY(x+1.5*s, y+0.6*s)
	pdf.CellFormat(w-3*s, 2.6*s, "BDT", "", 0, "R", false, 0, "")

	y2 := y + 3.8*s
	descW := w * 0.64
	amtW := w * 0.36
	for _, r := range rows {
		pdf.SetDrawColor(203, 213, 225)
		pdf.Rect(x, y2, w, 3.8*s, "D")
		labelStr := r.Label
		if len(labelStr) > 20 {
			pdf.SetFont(font, "", 4.0*s)
		} else {
			pdf.SetFont(font, "", 4.6*s)
		}
		pdf.SetXY(x+1.0*s, y2+0.8*s)
		pdf.CellFormat(descW-1.2*s, 2.2*s, labelStr, "", 0, "L", false, 0, "")

		pdf.SetFont(font, "", 4.6*s)
		pdf.SetXY(x+descW, y2+0.8*s)
		pdf.CellFormat(amtW-1.2*s, 2.2*s, r.Amount, "", 0, "R", false, 0, "")
		y2 += 3.8 * s
	}

	pdf.SetFillColor(241, 245, 249)
	pdf.SetFont(font, "B", 5.0*s)
	pdf.SetTextColor(15, 23, 42)
	pdf.Rect(x, y2, w, 3.8*s, "DF")
	pdf.SetXY(x+1.5*s, y2+0.8*s)
	pdf.CellFormat(w-3*s, 2.2*s, total, "", 0, "R", false, 0, "")
	return y2 + 3.8*s
}

func drawPayslipCutLines(pdf *gofpdf.Fpdf) {
	pdf.SetDrawColor(148, 163, 184)
	pdf.SetLineWidth(0.25)
	pdf.SetDashPattern([]float64{1.5, 1.2}, 0)

	// Center vertical cut line (X = 105.0mm)
	pdf.Line(105.0, 0, 105.0, 297.0)
	// Center horizontal cut line (Y = 148.5mm)
	pdf.Line(0, 148.5, 210.0, 148.5)

	pdf.SetDashPattern([]float64{0, 0}, 0)
	pdf.SetLineWidth(0.3)
	pdf.SetDrawColor(30, 58, 138)
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
