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
	payslipTop   = 10.0
	payslipLeft  = 8.0
)

const (
	payslipCardW = 97.0
	payslipCardH = 138.5
)

func exportSinglePayslipPDF(c *gin.Context, salary *models.Salary, month, year int, lang string) {
	labels := buildPayslipLabels(lang, salary.Company)
	card := buildPayslipCard(salary, month, year, lang, labels)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	font := payslipPDFFont(pdf, lang)

	// Single payslip: centered card filling the printable area.
	scale := 2.0
	cx := (payslipPageW - payslipCardW*scale) / 2
	cy := (payslipPageH - payslipCardH*scale) / 2
	drawPayslipCardPDF(pdf, font, cx, cy, payslipCardW*scale, payslipCardH*scale, card, labels)

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

	cols := []float64{payslipLeft, payslipLeft + payslipCardW}
	rows := []float64{payslipTop, payslipTop + payslipCardH}

	for i := range salaries {
		pos := i % 4
		if i > 0 && pos == 0 {
			pdf.AddPage()
		}
		card := buildPayslipCard(&salaries[i], month, year, lang, labels)
		drawPayslipCardPDF(pdf, font, cols[pos%2], rows[pos/2], payslipCardW, payslipCardH, card, labels)
		if pos == 3 {
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

	// ---- Header band ----
	pdf.SetFillColor(15, 23, 42)
	pdf.Rect(x, y, w, 14*s, "F")

	// Brand + company (left)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 8*s)
	pdf.SetXY(x+2*s, y+0.8*s)
	pdf.CellFormat(58*s, 3.4*s, card.BrandName, "", 0, "L", false, 0, "")

	pdf.SetFont(font, "B", 7*s)
	pdf.SetXY(x+2*s, y+3.8*s)
	pdf.CellFormat(58*s, 3.2*s, card.CompanyName, "", 0, "L", false, 0, "")

	pdf.SetFont(font, "", 5*s)
	pdf.SetTextColor(203, 213, 225)
	pdf.SetXY(x+2*s, y+6.8*s)
	pdf.CellFormat(58*s, 2.4*s, card.Address, "", 0, "L", false, 0, "")
	line := y + 9.0*s
	if card.Phone != "" || card.Email != "" {
		pdf.SetXY(x+2*s, line)
		pdf.CellFormat(58*s, 2.4*s, card.Phone+"   "+card.Email, "", 0, "L", false, 0, "")
		line += 2.6 * s
	}

	// PAYSLIP + copy (right)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 10*s)
	pdf.SetXY(x+62*s, y+0.8*s)
	pdf.CellFormat(30*s, 4*s, card.PayslipWord, "", 0, "R", false, 0, "")

	pdf.SetFont(font, "B", 5.2*s)
	pdf.SetTextColor(148, 163, 184)
	pdf.SetXY(x+62*s, y+4.8*s)
	pdf.CellFormat(30*s, 2.6*s, card.CopyLabel, "", 0, "R", false, 0, "")

	pdf.SetFont(font, "", 5.2*s)
	pdf.SetTextColor(203, 213, 225)
	pdf.SetXY(x+62*s, y+7.4*s)
	pdf.CellFormat(30*s, 2.6*s, labels.PayrollMonth+": "+card.PayrollMonth, "", 0, "R", false, 0, "")
	pdf.SetXY(x+62*s, y+9.8*s)
	pdf.CellFormat(30*s, 2.6*s, labels.PayrollNo+": "+card.PayrollNo, "", 0, "R", false, 0, "")
	pdf.SetXY(x+62*s, y+12.0*s)
	pdf.CellFormat(30*s, 2.2*s, labels.PrintDateLabel+": "+card.PrintDate, "", 0, "R", false, 0, "")

	curY := y + 14*s + 1.5*s

	// ---- Employee information ----
	curY = drawSectionTitlePDF(pdf, font, s, x, curY, w, labels.EmployeeInfo)
	curY = drawFieldGridPDF(pdf, font, s, x, curY, w, card.EmployeeInfo, 2)
	curY += 1.2 * s

	// ---- Attendance ----
	curY = drawSectionTitlePDF(pdf, font, s, x, curY, w, labels.Attendance)
	curY = drawFieldGridPDF(pdf, font, s, x, curY, w, card.Attendance, 4)
	curY += 1.2 * s

	// ---- Earnings (left) + Deductions (right) ----
	halfW := (w - 1.5*s) / 2
	curY = drawMoneyTablePDF(pdf, font, s, x, curY, halfW, labels.Earnings, card.Earnings, card.EarningsTotal)
	drawMoneyTablePDF(pdf, font, s, x+halfW+1.5*s, curY, halfW, labels.Deductions, card.Deductions, card.DeductionsTotal)

	// table height = title + header + max rows + total
	tableRows := len(card.Earnings)
	if len(card.Deductions) > tableRows {
		tableRows = len(card.Deductions)
	}
	curY += (4.6*s) + (2.8*s) + float64(tableRows)*(2.7*s) + (3.2*s)
	curY += 1.2 * s

	// ---- Summary ----
	summary := []payslipField{
		{labels.GrossSalary, card.GrossSalary},
		{labels.TotalEarnings, card.EarningsTotal},
		{labels.TotalDeductionLine, card.DeductionsTotal},
	}
	for _, fld := range summary {
		pdf.SetFillColor(248, 250, 252)
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetFont(font, "", 5.6*s)
		pdf.SetTextColor(30, 58, 138)
		pdf.Rect(x, curY, w, 3.2*s, "DF")
		pdf.SetXY(x+2*s, curY+0.5*s)
		pdf.CellFormat(w/2-2*s, 2.2*s, fld.Label, "", 0, "L", false, 0, "")
		pdf.SetFont(font, "B", 5.8*s)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(x+2*s, curY+0.5*s)
		pdf.CellFormat(w-4*s, 2.2*s, fld.Value, "", 0, "R", false, 0, "")
		curY += 3.2 * s
	}

	// Net salary green box
	pdf.SetFillColor(21, 128, 61)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 8*s)
	pdf.Rect(x, curY, w, 5.4*s, "F")
	pdf.SetXY(x+2*s, curY+0.9*s)
	pdf.CellFormat(w-4*s, 3.6*s, fmt.Sprintf("%s : BDT %s", labels.NetSalary, card.NetSalary), "", 0, "C", false, 0, "")
	curY += 5.4*s + 1.2*s

	// ---- Payment ----
	if len(card.Payment) > 0 {
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetFont(font, "B", 5.4*s)
		pdf.SetTextColor(15, 23, 42)
		pdf.Rect(x, curY, w, 4.4*s, "D")
		cellW := w / float64(len(card.Payment))
		for i, fld := range card.Payment {
			pdf.SetXY(x+float64(i)*cellW+1.5*s, curY+0.4*s)
			pdf.CellFormat(cellW-2*s, 1.8*s, fld.Label, "", 0, "L", false, 0, "")
			pdf.SetFont(font, "", 5.4*s)
			pdf.SetXY(x+float64(i)*cellW+1.5*s, curY+2.1*s)
			pdf.CellFormat(cellW-2*s, 2*s, fld.Value, "", 0, "L", false, 0, "")
			pdf.SetFont(font, "B", 5.4*s)
		}
		curY += 4.4*s + 1.2*s
	}

	// ---- Footer signatures ----
	sigCells := []payslipField{
		{"", card.PreparedBy},
		{"", card.CheckedBy},
		{"", card.ApprovedBy},
		{"", card.EmployeeSig},
	}
	cellW := w / 4
	for i, f := range sigCells {
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetFont(font, "", 5.4*s)
		pdf.SetTextColor(15, 23, 42)
		pdf.Rect(x+float64(i)*cellW, curY, cellW, 8*s, "D")
		pdf.SetXY(x+float64(i)*cellW+1*s, curY+5.6*s)
		pdf.CellFormat(cellW-2*s, 2.2*s, f.Value, "", 0, "C", false, 0, "")
	}
	curY += 8*s + 1.2*s

	// ---- Generated / confidential line ----
	pdf.SetFont(font, "", 4.6*s)
	pdf.SetTextColor(100, 116, 139)
	pdf.SetXY(x+2*s, curY)
	pdf.CellFormat(w-4*s, 2.2*s, card.GeneratedBy, "", 0, "L", false, 0, "")
	pdf.SetXY(x+2*s, curY)
	pdf.CellFormat(w-4*s, 2.2*s, card.Confidential, "", 0, "R", false, 0, "")
}

func drawSectionTitlePDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title string) float64 {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFont(font, "B", 6*s)
	pdf.SetTextColor(30, 58, 138)
	pdf.Rect(x, y, w, 4.6*s, "DF")
	pdf.SetXY(x+1.5*s, y+0.9*s)
	pdf.CellFormat(w-3*s, 2.8*s, title, "", 0, "L", false, 0, "")
	return y + 4.6*s
}

func drawFieldGridPDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, fields []payslipField, cols int) float64 {
	pdf.SetFont(font, "", 5.2*s)
	pdf.SetTextColor(15, 23, 42)
	rowH := 3.0 * s
	colW := w / float64(cols)
	for i, fld := range fields {
		col := i % cols
		row := i / cols
		px := x + float64(col)*colW
		py := y + float64(row)*rowH
		pdf.SetDrawColor(226, 232, 240)
		pdf.Rect(px, py, colW, rowH, "D")
		pdf.SetFont(font, "B", 5.2*s)
		pdf.SetTextColor(100, 116, 139)
		pdf.SetXY(px+1*s, py+0.5*s)
		pdf.CellFormat(colW*0.42, 2*s, fld.Label, "", 0, "L", false, 0, "")
		pdf.SetFont(font, "", 5.2*s)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(px+1*s+colW*0.42, py+0.5*s)
		pdf.CellFormat(colW-colW*0.42-1*s, 2*s, fld.Value, "", 0, "L", false, 0, "")
	}
	rows := (len(fields) + cols - 1) / cols
	return y + float64(rows)*rowH
}

func drawMoneyTablePDF(pdf *gofpdf.Fpdf, font string, s float64, x, y, w float64, title string, rows []payslipRow, total string) float64 {
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetFont(font, "B", 5.6*s)
	pdf.SetTextColor(30, 58, 138)
	pdf.Rect(x, y, w, 4.6*s, "DF")
	pdf.SetXY(x+1.5*s, y+0.9*s)
	pdf.CellFormat(w*0.6, 2.8*s, title, "", 0, "L", false, 0, "")
	pdf.SetXY(x+1.5*s, y+0.9*s)
	pdf.CellFormat(w-3*s, 2.8*s, "BDT", "", 0, "R", false, 0, "")

	y2 := y + 4.6*s
	pdf.SetFont(font, "", 5.2*s)
	pdf.SetTextColor(15, 23, 42)
	for _, r := range rows {
		pdf.SetDrawColor(226, 232, 240)
		pdf.Rect(x, y2, w, 2.7*s, "D")
		pdf.SetXY(x+1.5*s, y2+0.4*s)
		pdf.CellFormat(w-3*s, 2*s, r.Label, "", 0, "L", false, 0, "")
		pdf.SetXY(x+1.5*s, y2+0.4*s)
		pdf.CellFormat(w-3*s, 2*s, r.Amount, "", 0, "R", false, 0, "")
		y2 += 2.7 * s
	}

	pdf.SetFillColor(241, 245, 249)
	pdf.SetFont(font, "B", 5.4*s)
	pdf.SetTextColor(15, 23, 42)
	pdf.Rect(x, y2, w, 3.2*s, "DF")
	pdf.SetXY(x+1.5*s, y2+0.5*s)
	pdf.CellFormat(w-3*s, 2.2*s, total, "", 0, "R", false, 0, "")
	return y2 + 3.2*s
}

func drawPayslipCutLines(pdf *gofpdf.Fpdf) {
	pdf.SetDrawColor(148, 163, 184)
	pdf.SetLineWidth(0.25)
	pdf.SetDashPattern([]float64{1.2, 1.0}, 0)

	// vertical cut between columns
	cx := payslipLeft + payslipCardW
	pdf.Line(cx, payslipTop, cx, payslipTop+payslipCardH*2)
	// horizontal cut between rows
	cy := payslipTop + payslipCardH
	pdf.Line(payslipLeft, cy, payslipLeft+payslipCardW*2, cy)

	pdf.SetDashPattern([]float64{0, 0}, 0)
	pdf.SetLineWidth(0.3)
	pdf.SetDrawColor(30, 58, 138)
	drawScissorsIconPDF(pdf, cx, cy, 4.0)
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
