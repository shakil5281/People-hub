package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/xuri/excelize/v2"
)

const (
	payslipExcelCardCols = 4
	payslipExcelGapCols  = 1
)

var payslipExcelThinBorder = []excelize.Border{
	{Type: "left", Color: "CBD5E1", Style: 1},
	{Type: "right", Color: "CBD5E1", Style: 1},
	{Type: "top", Color: "CBD5E1", Style: 1},
	{Type: "bottom", Color: "CBD5E1", Style: 1},
}

type payslipExcelStyles struct {
	title, payslip, section, label, value, amount, total, net, sig, gen int
}

func newPayslipExcelStyles(f *excelize.File, lang string) payslipExcelStyles {
	fontName := "Calibri"
	if lang == "bn" {
		fontName = "SutonnyMJ"
	}
	title, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Family: fontName, Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#0F172A"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	payslip, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 12, Color: "#1E3A8A"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	section, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 9, Color: "#1E3A8A"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F8FAFC"}},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	label, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 8},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	value, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 8},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	amount, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 8},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	total, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 8, Color: "#1E3A8A"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F1F5F9"}},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	net, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#15803D"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	sig, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 8, Color: "#475569"},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	gen, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 7, Color: "#64748B"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	return payslipExcelStyles{title: title, payslip: payslip, section: section, label: label, value: value, amount: amount, total: total, net: net, sig: sig, gen: gen}
}

func payslipExcelCell(col, row int) string {
	name, _ := excelize.ColumnNumberToName(col)
	return fmt.Sprintf("%s%d", name, row)
}

func drawPayslipCardExcel(f *excelize.File, sheet string, col, row int, card *payslipCard, labels payslipLabels, st payslipExcelStyles) int {
	c0 := col
	c1 := col + 1
	c2 := col + 2
	c3 := col + 3

	merge := func(r int, from, to int) string {
		a := payslipExcelCell(from, r)
		b := payslipExcelCell(to, r)
		f.MergeCell(sheet, a, b)
		return a
	}
	set := func(r int, from, to int, v string, style int) {
		a := payslipExcelCell(from, r)
		b := payslipExcelCell(to, r)
		f.SetCellValue(sheet, a, v)
		f.SetCellStyle(sheet, a, b, style)
	}

	// ---- header ----
	f.SetRowHeight(sheet, row, 28)
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), card.BrandName+"  |  "+card.CompanyName)
	f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c3, row), st.title)
	row++

	f.SetRowHeight(sheet, row, 16)
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), fmt.Sprintf("%s   %s", card.CopyLabel, card.PayslipWord))
	f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c3, row), st.payslip)
	row++

	f.SetRowHeight(sheet, row, 14)
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), fmt.Sprintf("%s: %s    %s: %s    %s: %s",
		labels.PayrollMonth, card.PayrollMonth, labels.PayrollNo, card.PayrollNo, labels.PrintDateLabel, card.PrintDate))
	f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c3, row), st.value)
	row++

	// ---- employee info ----
	row = drawSectionRowExcel(f, sheet, row, c0, c3, labels.EmployeeInfo, st.section)
	row = drawPairGridExcel(f, sheet, row, c0, c1, c2, c3, card.EmployeeInfo, st.label, st.value)
	row++

	// ---- attendance ----
	row = drawSectionRowExcel(f, sheet, row, c0, c3, labels.Attendance, st.section)
	row = drawPairGridExcel(f, sheet, row, c0, c1, c2, c3, card.Attendance, st.label, st.value)
	row++

	// ---- earnings (left) + deductions (right) ----
	set(row, c0, c1, labels.Earnings, st.section)
	set(row, c2, c3, labels.Deductions, st.section)
	row++
	set(row, c0, c0, labels.Description, st.label)
	set(row, c1, c1, labels.Amount, st.amount)
	set(row, c2, c2, labels.Description, st.label)
	set(row, c3, c3, labels.Amount, st.amount)
	row++

	maxRows := len(card.Earnings)
	if len(card.Deductions) > maxRows {
		maxRows = len(card.Deductions)
	}
	for i := 0; i < maxRows; i++ {
		if i < len(card.Earnings) {
			set(row, c0, c0, card.Earnings[i].Label, st.value)
			set(row, c1, c1, card.Earnings[i].Amount, st.amount)
		}
		if i < len(card.Deductions) {
			set(row, c2, c2, card.Deductions[i].Label, st.value)
			set(row, c3, c3, card.Deductions[i].Amount, st.amount)
		}
		row++
	}
	set(row, c0, c0, labels.TotalEarnings, st.total)
	set(row, c1, c1, card.EarningsTotal, st.total)
	set(row, c2, c2, labels.TotalDeduction, st.total)
	set(row, c3, c3, card.DeductionsTotal, st.total)
	row++

	// ---- summary ----
	set(row, c0, c0, labels.GrossSalary, st.label)
	set(row, c1, c1, card.GrossSalary, st.amount)
	row++
	set(row, c0, c1, fmt.Sprintf("%s : BDT %s", labels.NetSalary, card.NetSalary), st.net)
	row++

	// ---- payment ----
	row = drawSectionRowExcel(f, sheet, row, c0, c3, labels.Payment, st.section)
	if len(card.Payment) > 0 {
		pairW := 2
		for i, pf := range card.Payment {
			base := c0 + (i%2)*pairW
			if i >= 2 {
				break
			}
			set(row, base, base, pf.Label, st.label)
			set(row, base+1, base+1, pf.Value, st.value)
		}
		row++
	}

	// ---- footer signatures ----
	row = drawSignatureRowExcel(f, sheet, row, c0, c3, card.PreparedBy, card.CheckedBy, st.sig)
	row = drawSignatureRowExcel(f, sheet, row, c0, c3, card.ApprovedBy, card.EmployeeSig, st.sig)
	row++

	// ---- generated / confidential ----
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), card.GeneratedBy+"    •    "+card.Confidential)
	f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c3, row), st.gen)
	row++

	return row
}

func drawSectionRowExcel(f *excelize.File, sheet string, row, from, to int, title string, style int) int {
	a := payslipExcelCell(from, row)
	b := payslipExcelCell(to, row)
	f.MergeCell(sheet, a, b)
	f.SetCellValue(sheet, a, title)
	f.SetCellStyle(sheet, a, b, style)
	f.SetRowHeight(sheet, row, 13)
	return row + 1
}

func drawPairGridExcel(f *excelize.File, sheet string, row, c0, c1, c2, c3 int, fields []payslipField, labelSt, valueSt int) int {
	for i := 0; i < len(fields); i += 2 {
		f.SetCellValue(sheet, payslipExcelCell(c0, row), fields[i].Label)
		f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c0, row), labelSt)
		f.SetCellValue(sheet, payslipExcelCell(c1, row), fields[i].Value)
		f.SetCellStyle(sheet, payslipExcelCell(c1, row), payslipExcelCell(c1, row), valueSt)
		if i+1 < len(fields) {
			f.SetCellValue(sheet, payslipExcelCell(c2, row), fields[i+1].Label)
			f.SetCellStyle(sheet, payslipExcelCell(c2, row), payslipExcelCell(c2, row), labelSt)
			f.SetCellValue(sheet, payslipExcelCell(c3, row), fields[i+1].Value)
			f.SetCellStyle(sheet, payslipExcelCell(c3, row), payslipExcelCell(c3, row), valueSt)
		}
		row++
	}
	return row
}

func drawSignatureRowExcel(f *excelize.File, sheet string, row, from, to int, left, right string, style int) int {
	mid := from + (to-from)/2 + 1
	a := payslipExcelCell(from, row)
	b := payslipExcelCell(mid-1, row)
	c := payslipExcelCell(mid, row)
	d := payslipExcelCell(to, row)
	f.MergeCell(sheet, a, b)
	f.MergeCell(sheet, c, d)
	f.SetCellValue(sheet, a, left)
	f.SetCellStyle(sheet, a, b, style)
	f.SetCellValue(sheet, c, right)
	f.SetCellStyle(sheet, c, d, style)
	f.SetRowHeight(sheet, row, 20)
	return row + 1
}

func exportSinglePayslipExcel(c *gin.Context, salary *models.Salary, month, year int, lang string) {
	labels := buildPayslipLabels(lang, salary.Company)
	card := buildPayslipCard(salary, month, year, lang, labels)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Payslip"
	f.SetSheetName("Sheet1", sheet)
	st := newPayslipExcelStyles(f, lang)

	setPayslipExcelColumnWidths(f, sheet)
	row := drawPayslipCardExcel(f, sheet, 1, 1, card, labels, st)

	f.SetActiveSheet(0)
	setPayslipExcelPrintLayout(f, sheet, row)

	langSuffix := lang
	if langSuffix == "" {
		langSuffix = "en"
	}
	filename := fmt.Sprintf("payslip_%s_%s_%d_%s.xlsx", salary.Employee.EmployeeID, monthName(month, "en"), year, langSuffix)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

func exportBulkPayslipExcel(c *gin.Context, salaries []models.Salary, month, year int, lang string) {
	labels := buildPayslipLabels(lang, salaries[0].Company)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Payslips"
	f.SetSheetName("Sheet1", sheet)
	st := newPayslipExcelStyles(f, lang)

	setPayslipExcelColumnWidths(f, sheet)

	// 2x2 grid: cards per page = 4
	blockRows := 40
	pageStep := blockRows * 2
	colSlot := payslipExcelCardCols + payslipExcelGapCols

	var maxRow int
	for i := range salaries {
		pos := i % 4
		page := i / 4
		col := 1 + (pos%2)*colSlot
		row := 1 + page*pageStep + (pos/2)*blockRows
		card := buildPayslipCard(&salaries[i], month, year, lang, labels)
		end := drawPayslipCardExcel(f, sheet, col, row, card, labels, st)
		if end > maxRow {
			maxRow = end
		}
	}

	// page breaks after each page (2 card rows)
	f.InsertPageBreak(sheet, payslipExcelCell(1, pageStep+1))
	totalPages := (len(salaries) + 3) / 4
	for p := 1; p < totalPages; p++ {
		f.InsertPageBreak(sheet, payslipExcelCell(1, p*pageStep+1))
	}

	setPayslipExcelPrintLayout(f, sheet, maxRow)

	langSuffix := lang
	if langSuffix == "" {
		langSuffix = "en"
	}
	filename := fmt.Sprintf("payslips_%s_%d_%s.xlsx", monthName(month, "en"), year, langSuffix)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

func setPayslipExcelColumnWidths(f *excelize.File, sheet string) {
	widths := map[string]float64{
		"A": 16, "B": 12, "C": 16, "D": 12, "E": 3,
		"F": 16, "G": 12, "H": 16, "I": 12, "J": 3,
		"K": 16, "L": 12, "M": 16, "N": 12, "O": 3,
		"P": 16, "Q": 12, "R": 16, "S": 12,
	}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}
}

func setPayslipExcelPrintLayout(f *excelize.File, sheet string, lastRow int) {
	orientation := "portrait"
	size := 9 // A4
	fitWidth := 1
	fitHeight := 0
	f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
		Size:       &size,
		Orientation: &orientation,
		FitToWidth: &fitWidth,
		FitToHeight: &fitHeight,
	})
	margin := 0.25
	f.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{
		Top:    &margin,
		Bottom: &margin,
		Left:   &margin,
		Right:  &margin,
	})
	f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Area",
		RefersTo: fmt.Sprintf("'%s'!$A$1:$S$%d", sheet, lastRow),
	})
}
