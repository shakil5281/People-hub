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

var payslipExcelCutBorder = []excelize.Border{
	{Type: "left", Color: "94A3B8", Style: 2},
	{Type: "right", Color: "94A3B8", Style: 2},
}

type payslipExcelStyles struct {
	title, payslip, section, label, value, amount, total, net, sig, gen, cut int
}

func newPayslipExcelStyles(f *excelize.File, lang string) payslipExcelStyles {
	fontName := "Calibri"
	if lang == "bn" {
		fontName = "SutonnyMJ"
	}
	title, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#0F172A"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	payslip, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 10, Color: "#D97706"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F8FAFC"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	section, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 8.5, Color: "#1E3A8A"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F8FAFC"}},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	label, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 8.0},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Vertical: "center", ShrinkToFit: true},
	})
	value, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 8.0},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Vertical: "center", ShrinkToFit: true},
	})
	amount, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 8.0},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", ShrinkToFit: true},
	})
	total, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 8.0, Color: "#1E3A8A"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F1F5F9"}},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center", ShrinkToFit: true},
	})
	net, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Bold: true, Size: 9.5, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#15803D"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	sig, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 7.5, Color: "#475569"},
		Border:    payslipExcelThinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	gen, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: fontName, Size: 6.5, Color: "#64748B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	cut, _ := f.NewStyle(&excelize.Style{
		Border: payslipExcelCutBorder,
	})
	return payslipExcelStyles{title: title, payslip: payslip, section: section, label: label, value: value, amount: amount, total: total, net: net, sig: sig, gen: gen, cut: cut}
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

	// ---- 1. Header: Company Name ----
	f.SetRowHeight(sheet, row, 22)
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), card.CompanyName)
	f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c3, row), st.title)
	row++

	// ---- 2. Copy Label & Payslip Title ----
	f.SetRowHeight(sheet, row, 16)
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), fmt.Sprintf("%s   -   %s", card.CopyLabel, card.PayslipWord))
	f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c3, row), st.payslip)
	row++

	// ---- 3. Month & Payroll No ----
	f.SetRowHeight(sheet, row, 17)
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), fmt.Sprintf("%s: %s   |   %s: %s   |   %s: %s",
		labels.PayrollMonth, card.PayrollMonth, labels.PayrollNo, card.PayrollNo, labels.PrintDateLabel, card.PrintDate))
	f.SetCellStyle(sheet, payslipExcelCell(c0, row), payslipExcelCell(c3, row), st.value)
	row++

	// ---- 4. Employee Information ----
	row = drawSectionRowExcel(f, sheet, row, c0, c3, labels.EmployeeInfo, st.section)
	row = drawPairGridExcel(f, sheet, row, c0, c1, c2, c3, card.EmployeeInfo, st.label, st.value)

	// ---- 5. Attendance Summary ----
	row = drawSectionRowExcel(f, sheet, row, c0, c3, labels.Attendance, st.section)
	row = drawPairGridExcel(f, sheet, row, c0, c1, c2, c3, card.Attendance, st.label, st.value)

	// ---- 6. Earnings & Deductions Side-by-Side ----
	f.SetRowHeight(sheet, row, 17)
	set(row, c0, c1, labels.Earnings, st.section)
	set(row, c2, c3, labels.Deductions, st.section)
	row++
	f.SetRowHeight(sheet, row, 17)
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
		f.SetRowHeight(sheet, row, 17)
		if i < len(card.Earnings) {
			set(row, c0, c0, card.Earnings[i].Label, st.value)
			set(row, c1, c1, card.Earnings[i].Amount, st.amount)
		} else {
			set(row, c0, c0, "", st.value)
			set(row, c1, c1, "", st.amount)
		}
		if i < len(card.Deductions) {
			set(row, c2, c2, card.Deductions[i].Label, st.value)
			set(row, c3, c3, card.Deductions[i].Amount, st.amount)
		} else {
			set(row, c2, c2, "", st.value)
			set(row, c3, c3, "", st.amount)
		}
		row++
	}
	f.SetRowHeight(sheet, row, 17)
	set(row, c0, c0, labels.TotalEarnings, st.total)
	set(row, c1, c1, card.EarningsTotal, st.total)
	set(row, c2, c2, labels.TotalDeduction, st.total)
	set(row, c3, c3, card.DeductionsTotal, st.total)
	row++

	// ---- 7. Summary & Net Salary ----
	f.SetRowHeight(sheet, row, 17)
	set(row, c0, c1, labels.GrossSalary, st.label)
	set(row, c2, c3, card.GrossSalary, st.amount)
	row++
	f.SetRowHeight(sheet, row, 20)
	merge(row, c0, c3)
	set(row, c0, c3, fmt.Sprintf("%s : BDT %s", labels.NetSalary, card.NetSalary), st.net)
	row++

	// Empty gap row after Net Salary (Height: 10)
	f.SetRowHeight(sheet, row, 10)
	row++

	// ---- 8. Footer Signatures ----
	row = drawSignatureRowExcel(f, sheet, row, c0, c3, card.PreparedBy, card.CheckedBy, st.sig)
	row = drawSignatureRowExcel(f, sheet, row, c0, c3, card.ApprovedBy, card.EmployeeSig, st.sig)

	// ---- 9. Confidential Line ----
	f.SetRowHeight(sheet, row, 14)
	merge(row, c0, c3)
	f.SetCellValue(sheet, payslipExcelCell(c0, row), card.GeneratedBy+"   •   "+card.Confidential)
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
	f.SetRowHeight(sheet, row, 17)
	return row + 1
}

func drawPairGridExcel(f *excelize.File, sheet string, row, c0, c1, c2, c3 int, fields []payslipField, labelSt, valueSt int) int {
	for i := 0; i < len(fields); i += 2 {
		f.SetRowHeight(sheet, row, 17)
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

	cardOffice := buildPayslipCardWithCopy(salary, month, year, lang, labels, labels.CopyOffice)
	cardEmp := buildPayslipCardWithCopy(salary, month, year, lang, labels, labels.CopyEmployee)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Payslip"
	f.SetSheetName("Sheet1", sheet)
	st := newPayslipExcelStyles(f, lang)

	setPayslipExcelColumnWidths(f, sheet)

	// Left side Office Copy (Cols A-D, Col 1), Right side Employee Copy (Cols F-I, Col 6)
	end1 := drawPayslipCardExcel(f, sheet, 1, 1, cardOffice, labels, st)
	end2 := drawPayslipCardExcel(f, sheet, 6, 1, cardEmp, labels, st)

	maxEnd := end1
	if end2 > maxEnd {
		maxEnd = end2
	}

	// Apply center cut border on Gap Column E
	for r := 1; r <= maxEnd; r++ {
		f.SetCellStyle(sheet, payslipExcelCell(5, r), payslipExcelCell(5, r), st.cut)
	}

	f.SetActiveSheet(0)
	setPayslipExcelPrintLayout(f, sheet, maxEnd)

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

	// 2 employees per A4 page (4 cards total per A4 page: 28 rows per card + 1 gap row = 29 rows per block)
	cardRows := 28
	gapRow := 1
	blockRows := cardRows + gapRow // 29 rows per employee block
	pageStep := blockRows * 2      // 58 rows per A4 page
	colSlot := payslipExcelCardCols + payslipExcelGapCols

	var maxRow int
	for i := range salaries {
		empIdxOnPage := i % 2
		page := i / 2
		row := 1 + page*pageStep + empIdxOnPage*blockRows

		cardOffice := buildPayslipCardWithCopy(&salaries[i], month, year, lang, labels, labels.CopyOffice)
		cardEmp := buildPayslipCardWithCopy(&salaries[i], month, year, lang, labels, labels.CopyEmployee)

		end1 := drawPayslipCardExcel(f, sheet, 1, row, cardOffice, labels, st)
		end2 := drawPayslipCardExcel(f, sheet, 1+colSlot, row, cardEmp, labels, st)

		end := end1
		if end2 > end {
			end = end2
		}
		if end > maxRow {
			maxRow = end
		}

		// Apply center cut border on Gap Column E for this block
		for r := row; r < row+cardRows; r++ {
			f.SetCellStyle(sheet, payslipExcelCell(5, r), payslipExcelCell(5, r), st.cut)
		}

		// Set height of center height gap row = 30
		if empIdxOnPage == 0 {
			gapRowIdx := row + cardRows
			f.SetRowHeight(sheet, gapRowIdx, 30)
		}
	}

	// Insert page breaks after every 2 employees (4 cards total = 58 rows)
	totalPages := (len(salaries) + 1) / 2
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
		"A": 14, "B": 17, "C": 14, "D": 17,
		"E": 6, // Center width gap = 6
		"F": 14, "G": 17, "H": 14, "I": 17,
	}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}
}

func setPayslipExcelPrintLayout(f *excelize.File, sheet string, lastRow int) {
	fitToPage := true
	f.SetSheetProps(sheet, &excelize.SheetPropsOptions{
		FitToPage: &fitToPage,
	})
	orientation := "portrait"
	size := 9 // A4 size
	fitWidth := 1
	fitHeight := 0 // Automatic per page break
	f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
		Size:        &size,
		Orientation: &orientation,
		FitToWidth:  &fitWidth,
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
		RefersTo: fmt.Sprintf("'%s'!$A$1:$I$%d", sheet, lastRow),
	})
}
