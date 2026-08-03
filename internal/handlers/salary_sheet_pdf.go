package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

const (
	sheetPDFPageW = 355.6
	sheetPDFPageH = 215.9
	sheetPDFLeft  = 8.0
	sheetPDFRight = 8.0
	sheetPDFTop   = 10.0
)

// SheetExportPDF godoc
//
// @Summary      Export salary sheet to PDF
// @Description  Download monthly salary sheet as landscape PDF with English or Bengali headers
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        company_id     query string true  "Company ID"
// @Param        month          query int    true  "Month (1-12)"
// @Param        year           query int    true  "Year"
// @Param        lang           query string false "Language: en or bn (default en)"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        shift_id       query string false "Filter by shift"
// @Param        employee_id    query string false "Search by employee ID (partial match)"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/sheet/export/pdf [get]
func (h *SalaryHandler) SheetExportPDF(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")
	lang := c.DefaultQuery("lang", "en")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id, month, and year are required"})
		return
	}

	salaries, err := h.salaryRepo.ListAllByMonthFiltered(repository.SalaryFilter{
		CompanyID:     companyID,
		Month:         month,
		Year:          year,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		ShiftID:       c.Query("shift_id"),
		EmployeeID:    c.Query("employee_id"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(salaries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No salary records found"})
		return
	}

	labels := getSalarySheetLabels(lang)
	monthLabel := monthName(month, lang)

	// Fetch company from first salary record
	company := salaries[0].Company

	compName := companyNameFor(lang, company)

	pdf := gofpdf.New("L", "mm", "Legal", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(sheetPDFLeft, sheetPDFTop, sheetPDFRight)
	pdf.AddPage()

	font := "Helvetica"
	if lang == "bn" {
		font = loadBanglaFont(pdf)
	}

	// --- Header ---
	pdf.SetFont(font, "B", 14)
	pdf.SetTextColor(15, 23, 42)
	pdf.Cell(0, 7, compName)
	pdf.Ln(7)

	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(100, 100, 100)
	addr := companyAddress(company, lang)
	if addr != "" {
		pdf.Cell(0, 5, addr)
		pdf.Ln(5)
	}

	pdf.SetFont(font, "B", 11)
	pdf.SetTextColor(15, 23, 42)
	title := fmt.Sprintf("%s - %s %d", labels.Title, monthLabel, year)
	pdf.Cell(0, 7, title)
	pdf.Ln(10)

	// --- Table ---
	type colDef struct {
		header string
		width  float64
		align  string
	}
	cols := []colDef{
		{labels.Sl, 8, "C"},
		{labels.EmployeeID, 18, "C"},
		{labels.Name, 28, "L"},
		{labels.Designation, 22, "L"},
		{labels.Department, 22, "L"},
		{labels.WorkingDays, 12, "C"},
		{labels.Present, 10, "C"},
		{labels.Absent, 10, "C"},
		{labels.Late, 8, "C"},
		{labels.Leave, 8, "C"},
		{labels.Holiday, 10, "C"},
		{labels.Weekend, 10, "C"},
		{labels.BasicSalary, 16, "R"},
		{labels.HouseRent, 16, "R"},
		{labels.Medical, 14, "R"},
		{labels.Transport, 14, "R"},
		{labels.Food, 14, "R"},
		{labels.OtherAllowance, 12, "R"},
		{labels.Gross, 16, "R"},
		{labels.AbsentDed, 14, "R"},
		{labels.OTHours, 10, "C"},
		{labels.OTAmount, 14, "R"},
		{labels.AttBonus, 14, "R"},
		{labels.NetSalary, 16, "R"},
	}

	headerH := 8.0
	rowH := 6.5

	// Header row
	pdf.SetFillColor(68, 114, 196)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 6)
	for _, c := range cols {
		pdf.CellFormat(c.width, headerH, c.header, "1", 0, c.align, true, 0, "")
	}
	pdf.Ln(headerH)

	// Data rows
	pdf.SetFont(font, "", 6)
	totalBasic := 0.0
	totalHouse := 0.0
	totalMed := 0.0
	totalTrans := 0.0
	totalFood := 0.0
	totalGross := 0.0
	totalAbsent := 0.0
	totalOT := 0.0
	totalBonus := 0.0
	totalNet := 0.0

	for i, s := range salaries {
		bg := i%2 == 0
		if bg {
			pdf.SetFillColor(240, 244, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(30, 30, 30)

		empName := employeeNameFor(lang, &s.Employee)
		desig := designationName(s.Employee.DesignationRef, lang)
		dept := departmentName(s.Employee.Department, lang)

		data := []struct {
			val   string
			align string
		}{
			{strconv.Itoa(i + 1), "C"},
			{s.Employee.EmployeeID, "C"},
			{empName, "L"},
			{desig, "L"},
			{dept, "L"},
			{strconv.Itoa(s.TotalDays), "C"},
			{strconv.Itoa(s.PresentDays), "C"},
			{strconv.Itoa(s.AbsentDays), "C"},
			{strconv.Itoa(s.LateDays), "C"},
			{strconv.Itoa(s.LeaveDays), "C"},
			{strconv.Itoa(s.HolidayDays), "C"},
			{strconv.Itoa(s.WeekendDays), "C"},
			{fmt.Sprintf("%.0f", s.BasicSalary), "R"},
			{fmt.Sprintf("%.0f", s.HouseRent), "R"},
			{fmt.Sprintf("%.0f", s.MedicalAllowance), "R"},
			{fmt.Sprintf("%.0f", s.TransportAllowance), "R"},
			{fmt.Sprintf("%.0f", s.FoodAllowance), "R"},
			{fmt.Sprintf("%.0f", s.OtherAllowance), "R"},
			{fmt.Sprintf("%.0f", s.GrossSalary), "R"},
			{fmt.Sprintf("%.0f", s.AbsentDeduction), "R"},
			{fmt.Sprintf("%.1f", s.OvertimeHours), "C"},
			{fmt.Sprintf("%.0f", s.OvertimeAmount), "R"},
			{fmt.Sprintf("%.0f", s.AttendanceBonus), "R"},
			{fmt.Sprintf("%.0f", s.NetSalary), "R"},
		}

		for j, d := range data {
			pdf.CellFormat(cols[j].width, rowH, d.val, "1", 0, d.align, bg, 0, "")
		}
		pdf.Ln(rowH)

		totalBasic += s.BasicSalary
		totalHouse += s.HouseRent
		totalMed += s.MedicalAllowance
		totalTrans += s.TransportAllowance
		totalFood += s.FoodAllowance
		totalGross += s.GrossSalary
		totalAbsent += s.AbsentDeduction
		totalOT += s.OvertimeAmount
		totalBonus += s.AttendanceBonus
		totalNet += s.NetSalary

		// Page break check
		if pdf.GetY() > sheetPDFPageH-15 {
			pdf.AddPage()
			pdf.SetFillColor(68, 114, 196)
			pdf.SetTextColor(255, 255, 255)
			pdf.SetFont(font, "B", 6)
			for _, c := range cols {
				pdf.CellFormat(c.width, headerH, c.header, "1", 0, c.align, true, 0, "")
			}
			pdf.Ln(headerH)
			pdf.SetFont(font, "", 6)
		}
	}

	// Total row
	pdf.SetFillColor(226, 239, 218)
	pdf.SetTextColor(0, 97, 0)
	pdf.SetFont(font, "B", 6)
	totalRowH := 7.0
	totalData := []struct {
		val   string
		align string
	}{
		{"", "C"},
		{"", "C"},
		{"", "L"},
		{"", "L"},
		{labels.TotalLabel, "L"},
		{"", "C"},
		{"", "C"},
		{"", "C"},
		{"", "C"},
		{"", "C"},
		{"", "C"},
		{"", "C"},
		{fmt.Sprintf("%.0f", totalBasic), "R"},
		{fmt.Sprintf("%.0f", totalHouse), "R"},
		{fmt.Sprintf("%.0f", totalMed), "R"},
		{fmt.Sprintf("%.0f", totalTrans), "R"},
		{fmt.Sprintf("%.0f", totalFood), "R"},
		{"", "R"},
		{fmt.Sprintf("%.0f", totalGross), "R"},
		{fmt.Sprintf("%.0f", totalAbsent), "R"},
		{"", "C"},
		{fmt.Sprintf("%.0f", totalOT), "R"},
		{fmt.Sprintf("%.0f", totalBonus), "R"},
		{fmt.Sprintf("%.0f", totalNet), "R"},
	}
	for j, d := range totalData {
		pdf.CellFormat(cols[j].width, totalRowH, d.val, "1", 0, d.align, true, 0, "")
	}
	pdf.Ln(totalRowH)

	// Footer
	pdf.Ln(5)
	pdf.SetFont(font, "", 7)
	pdf.SetTextColor(130, 130, 130)
	footerText := fmt.Sprintf("%s | %s: %s", "Generated by PeopleHub", monthLabel, year)
	if lang == "bn" {
		footerText = fmt.Sprintf("%s | %s %d", "পিপলহাব দ্বারা উৎপন্ন", monthLabel, year)
	}
	pdf.Cell(0, 5, footerText)
	pdf.Ln(3)
	printDate := fmt.Sprintf("%s: %s", "Print Date", time.Now().Format("02-01-2006 15:04"))
	if lang == "bn" {
		printDate = fmt.Sprintf("%s: %s", "মুদ্রণের তারিখ", time.Now().Format("02-01-2006 15:04"))
	}
	pdf.Cell(0, 5, printDate)

	langSuffix := "en"
	if lang == "bn" {
		langSuffix = "bn"
	}
	filename := fmt.Sprintf("salary_sheet_%s_%d_%s.pdf", monthLabel, year, langSuffix)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}

// SummaryExportPDF godoc
//
// @Summary      Export salary summary to PDF
// @Description  Download monthly salary summary as portrait PDF
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        company_id     query string true  "Company ID"
// @Param        month          query int    true  "Month (1-12)"
// @Param        year           query int    true  "Year"
// @Param        lang           query string false "Language: en or bn (default en)"
// @Param        group_by       query string false "Group by: department|section|designation|line"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/summary/export/pdf [get]
func (h *SalaryHandler) SummaryExportPDF(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")
	lang := c.DefaultQuery("lang", "en")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id, month, and year are required"})
		return
	}

	groupBy := c.DefaultQuery("group_by", "department")
	labels := getSalarySheetLabels(lang)
	monthLabel := monthName(month, lang)

	salaries, err := h.salaryRepo.ListAllByMonthFiltered(repository.SalaryFilter{
		CompanyID:     companyID,
		Month:         month,
		Year:          year,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type groupKey struct{ Name, ID string }
	type groupData struct {
		Employees   int
		BasicSalary float64
		HouseRent   float64
		Medical     float64
		Transport   float64
		GrossSalary float64
		Deductions  float64
		NetSalary   float64
	}
	groupMap := make(map[groupKey]*groupData)

	for _, s := range salaries {
		var key groupKey
		switch groupBy {
		case "section":
			if s.Employee.SectionRef != nil {
				key = groupKey{Name: sectionName(s.Employee.SectionRef, lang), ID: s.Employee.SectionRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		case "designation":
			if s.Employee.DesignationRef != nil {
				key = groupKey{Name: designationName(s.Employee.DesignationRef, lang), ID: s.Employee.DesignationRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		case "line":
			if s.Employee.LineRef != nil {
				key = groupKey{Name: lineName(s.Employee.LineRef, lang), ID: s.Employee.LineRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		default:
			if s.Employee.Department != nil {
				key = groupKey{Name: departmentName(s.Employee.Department, lang), ID: s.Employee.Department.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		}
		if groupMap[key] == nil {
			groupMap[key] = &groupData{}
		}
		d := groupMap[key]
		d.Employees++
		d.BasicSalary += s.BasicSalary
		d.HouseRent += s.HouseRent
		d.Medical += s.MedicalAllowance
		d.Transport += s.TransportAllowance
		d.GrossSalary += s.GrossSalary
		d.Deductions += s.TotalDeductions
		d.NetSalary += s.NetSalary
	}

	var company models.Company
	if len(salaries) > 0 {
		company = salaries[0].Company
	}
	compName := companyNameFor(lang, company)

	pdf := gofpdf.New("P", "mm", "Legal", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	font := "Helvetica"
	if lang == "bn" {
		font = loadBanglaFont(pdf)
	}

	// Header
	pdf.SetFont(font, "B", 14)
	pdf.SetTextColor(15, 23, 42)
	pdf.Cell(0, 7, compName)
	pdf.Ln(7)

	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(100, 100, 100)
	addr := companyAddress(company, lang)
	if addr != "" {
		pdf.Cell(0, 5, addr)
		pdf.Ln(5)
	}

	groupLabelMap := map[string]string{
		"department":  labels.Department,
		"section":     labels.Section,
		"designation": labels.Designation,
		"line":        labels.GroupLabel,
	}
	groupLabel := groupLabelMap[groupBy]
	if groupLabel == "" {
		groupLabel = labels.GroupLabel
	}

	pdf.SetFont(font, "B", 12)
	pdf.SetTextColor(15, 23, 42)
	title := fmt.Sprintf("%s - %s %d", labels.Title, monthLabel, year)
	pdf.Cell(0, 8, title)
	pdf.Ln(12)

	// Table
	type summaryCol struct {
		header string
		width  float64
		align  string
	}
	sCols := []summaryCol{
		{labels.Sl, 10, "C"},
		{groupLabel, 50, "L"},
		{labels.Employees, 18, "C"},
		{labels.GrossTotal, 30, "R"},
		{labels.BasicTotal, 28, "R"},
		{labels.HouseRentH, 28, "R"},
		{labels.MedicalH, 25, "R"},
		{labels.TransportH, 25, "R"},
		{labels.Deductions, 28, "R"},
		{labels.NetTotal, 30, "R"},
	}

	headerH := 9.0
	rowH := 7.5

	pdf.SetFillColor(68, 114, 196)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 8)
	for _, sc := range sCols {
		pdf.CellFormat(sc.width, headerH, sc.header, "1", 0, sc.align, true, 0, "")
	}
	pdf.Ln(headerH)

	pdf.SetFont(font, "", 8)
	var grand groupData
	i := 0
	for key, d := range groupMap {
		_ = key
		bg := i%2 == 0
		if bg {
			pdf.SetFillColor(240, 244, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(30, 30, 30)

		data := []struct {
			val   string
			align string
		}{
			{strconv.Itoa(i + 1), "C"},
			{key.Name, "L"},
			{strconv.Itoa(d.Employees), "C"},
			{fmt.Sprintf("%.0f", d.GrossSalary), "R"},
			{fmt.Sprintf("%.0f", d.BasicSalary), "R"},
			{fmt.Sprintf("%.0f", d.HouseRent), "R"},
			{fmt.Sprintf("%.0f", d.Medical), "R"},
			{fmt.Sprintf("%.0f", d.Transport), "R"},
			{fmt.Sprintf("%.0f", d.Deductions), "R"},
			{fmt.Sprintf("%.0f", d.NetSalary), "R"},
		}
		for j, dt := range data {
			pdf.CellFormat(sCols[j].width, rowH, dt.val, "1", 0, dt.align, bg, 0, "")
		}
		pdf.Ln(rowH)

		grand.Employees += d.Employees
		grand.BasicSalary += d.BasicSalary
		grand.HouseRent += d.HouseRent
		grand.Medical += d.Medical
		grand.Transport += d.Transport
		grand.GrossSalary += d.GrossSalary
		grand.Deductions += d.Deductions
		grand.NetSalary += d.NetSalary
		i++
	}

	// Grand total
	pdf.SetFillColor(226, 239, 218)
	pdf.SetTextColor(0, 97, 0)
	pdf.SetFont(font, "B", 8)
	totalRowH := 8.0
	gData := []struct {
		val   string
		align string
	}{
		{"", "C"},
		{labels.GrandTotal, "L"},
		{strconv.Itoa(grand.Employees), "C"},
		{fmt.Sprintf("%.0f", grand.GrossSalary), "R"},
		{fmt.Sprintf("%.0f", grand.BasicSalary), "R"},
		{fmt.Sprintf("%.0f", grand.HouseRent), "R"},
		{fmt.Sprintf("%.0f", grand.Medical), "R"},
		{fmt.Sprintf("%.0f", grand.Transport), "R"},
		{fmt.Sprintf("%.0f", grand.Deductions), "R"},
		{fmt.Sprintf("%.0f", grand.NetSalary), "R"},
	}
	for j, dt := range gData {
		pdf.CellFormat(sCols[j].width, totalRowH, dt.val, "1", 0, dt.align, true, 0, "")
	}
	pdf.Ln(totalRowH)

	// Footer
	pdf.Ln(5)
	pdf.SetFont(font, "", 7)
	pdf.SetTextColor(130, 130, 130)
	genBy := "Generated by PeopleHub"
	if lang == "bn" {
		genBy = "পিপলহাব দ্বারা উৎপন্ন"
	}
	pdf.Cell(0, 5, fmt.Sprintf("%s | %s", genBy, time.Now().Format("02-01-2006 15:04")))

	langSuffix := "en"
	if lang == "bn" {
		langSuffix = "bn"
	}
	filename := fmt.Sprintf("salary_summary_%s_%d_%s.pdf", monthLabel, year, langSuffix)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}
