package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

const (
	sheetPDFPageW = 355.6
	sheetPDFPageH = 215.9
	sheetPDFLeft  = 5.0
	sheetPDFRight = 5.0
	sheetPDFTop   = 8.0
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
	company := salaries[0].Company
	compName := companyNameFor(lang, company)
	compAddr := companyAddress(company, lang)

	pdf := gofpdf.New("L", "mm", "Legal", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(sheetPDFLeft, sheetPDFTop, sheetPDFRight)

	font := "Helvetica"
	if lang == "bn" {
		font = loadBanglaFont(pdf)
	}

	// 25 Columns Definition matching Excel (345.6mm Total Printable Width)
	type colDef struct {
		header string
		width  float64
		align  string
	}
	cols := []colDef{
		{labels.Sl, 8, "C"},
		{labels.EmployeeID, 15, "C"},
		{labels.Name, 32, "L"},
		{labels.Designation, 26, "L"},
		{labels.WorkingDays, 10, "C"},
		{labels.Present, 10, "C"},
		{labels.Absent, 10, "C"},
		{labels.Late, 10, "C"},
		{labels.Leave, 10, "C"},
		{labels.Holiday, 10, "C"},
		{labels.Weekend, 10, "C"},
		{labels.BasicSalary, 15, "C"},
		{labels.HouseRent, 15, "C"},
		{labels.Medical, 14, "C"},
		{labels.Transport, 14, "C"},
		{labels.Food, 14, "C"},
		{labels.Gross, 16, "C"},
		{labels.AbsentDed, 15, "C"},
		{labels.TotalDeduction, 15, "C"},
		{labels.OTHours, 10, "C"},
		{labels.OTRate, 11, "C"},
		{labels.OTAmount, 13, "C"},
		{labels.AttBonus, 12, "C"},
		{labels.NetSalary, 18, "C"},
		{labels.Signature, 20.6, "C"},
	}

	// Group records into Office Staff, Production Staff, and Line Workers
	type pdfGroup struct {
		displayName string
		list        []models.Salary
	}
	var officeStaffList []models.Salary
	var productionStaffList []models.Salary
	lineWorkerMap := make(map[string][]models.Salary)
	var lineNames []string

	for _, s := range salaries {
		grpName := ""
		if s.Employee.GroupRef != nil {
			grpName = strings.ToLower(s.Employee.GroupRef.Name)
		}
		empType := strings.ToLower(s.Employee.EmployeeType)
		deptName := ""
		if s.Employee.Department != nil {
			deptName = strings.ToLower(s.Employee.Department.Name)
		}

		isStaff := strings.Contains(grpName, "staff") || strings.Contains(empType, "staff")

		if isStaff {
			if strings.Contains(deptName, "production") || strings.Contains(deptName, "maintenance") {
				productionStaffList = append(productionStaffList, s)
			} else {
				officeStaffList = append(officeStaffList, s)
			}
		} else {
			lineName := "No Line"
			if s.Employee.LineRef != nil && strings.TrimSpace(s.Employee.LineRef.Name) != "" {
				lineName = strings.TrimSpace(s.Employee.LineRef.Name)
			}
			if strings.EqualFold(lineName, "admin") {
				lineName = "Loader & Cleaner"
			}
			if _, exists := lineWorkerMap[lineName]; !exists {
				lineNames = append(lineNames, lineName)
			}
			lineWorkerMap[lineName] = append(lineWorkerMap[lineName], s)
		}
	}

	sort.Slice(lineNames, func(i, j int) bool {
		rankI := customOrderRank(lineNames[i])
		rankJ := customOrderRank(lineNames[j])
		if rankI != rankJ {
			return rankI < rankJ
		}
		return lineNames[i] < lineNames[j]
	})

	var groups []pdfGroup
	if len(officeStaffList) > 0 {
		groups = append(groups, pdfGroup{"Office Staff", officeStaffList})
	}
	if len(productionStaffList) > 0 {
		groups = append(groups, pdfGroup{"Production Staff", productionStaffList})
	}
	for _, lName := range lineNames {
		groups = append(groups, pdfGroup{lName, lineWorkerMap[lName]})
	}
	if len(groups) == 0 {
		groups = append(groups, pdfGroup{"All", salaries})
	}

	renderHeader := func(displayLine string) {
		pdf.SetFont(font, "B", 14)
		pdf.SetTextColor(15, 23, 42)
		pdf.Cell(0, 6, compName)
		pdf.Ln(6)

		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(100, 100, 100)
		if compAddr != "" {
			pdf.Cell(0, 4, compAddr)
			pdf.Ln(4)
		}

		pdf.SetFont(font, "B", 10)
		pdf.SetTextColor(15, 23, 42)
		title := fmt.Sprintf("%s - %s %d", labels.Title, monthLabel, year)
		pdf.Cell(0, 5, title)
		pdf.Ln(5)

		pdf.SetFont(font, "B", 9)
		pdf.SetTextColor(31, 41, 55)
		lineText := fmt.Sprintf("Line: %s", displayLine)
		if lang == "bn" {
			lineText = utils.UnicodeToBijoy(fmt.Sprintf("লাইন: %s", displayLine))
		}
		pdf.Cell(0, 5, lineText)
		pdf.Ln(6)

		// Table Header
		headerH := 7.0
		pdf.SetDrawColor(217, 217, 217)
		pdf.SetFillColor(68, 114, 196)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont(font, "B", 6)
		for _, c := range cols {
			pdf.CellFormat(c.width, headerH, c.header, "1", 0, c.align, true, 0, "")
		}
		pdf.Ln(headerH)
	}

	renderFooter := func() {
		pdf.SetY(sheetPDFPageH - 12)
		pdf.SetFont(font, "", 7)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(80, 4, "Prepared By", "", 0, "L", false, 0, "")
		pdf.CellFormat(185.6, 4, "Admin (A.G.M)                     Asst. General Manager", "", 0, "C", false, 0, "")
		pdf.CellFormat(80, 4, "Approved By", "", 0, "R", false, 0, "")
	}

	for groupIdx, grp := range groups {
		if groupIdx > 0 {
			renderFooter()
		}
		pdf.AddPage()
		renderHeader(grp.displayName)

		rowH := 6.0
		pdf.SetFont(font, "", 6)

		var totalBasic, totalHouse, totalMed, totalTrans, totalFood, totalGross float64
		var totalAbsent, totalDeductions, totalOTHours, totalOTAmount, totalBonus, totalNet float64

		for i, s := range grp.list {
			pdf.SetDrawColor(217, 217, 217)
			pdf.SetFillColor(255, 255, 255)
			pdf.SetTextColor(30, 30, 30)

			empName := employeeNameFor(lang, &s.Employee)
			desig := designationName(s.Employee.DesignationRef, lang)
			if lang == "bn" {
				empName = utils.UnicodeToBijoy(empName)
				desig = utils.UnicodeToBijoy(desig)
			}

			data := []struct {
				val   string
				align string
			}{
				{strconv.Itoa(i + 1), "C"},
				{s.Employee.EmployeeID, "C"},
				{empName, "L"},
				{desig, "L"},
				{strconv.Itoa(s.TotalDays), "C"},
				{strconv.Itoa(s.PresentDays), "C"},
				{strconv.Itoa(s.AbsentDays), "C"},
				{strconv.Itoa(s.LateDays), "C"},
				{strconv.Itoa(s.LeaveDays), "C"},
				{strconv.Itoa(s.HolidayDays), "C"},
				{strconv.Itoa(s.WeekendDays), "C"},
				{fmt.Sprintf("%.0f", s.BasicSalary), "C"},
				{fmt.Sprintf("%.0f", s.HouseRent), "C"},
				{fmt.Sprintf("%.0f", s.MedicalAllowance), "C"},
				{fmt.Sprintf("%.0f", s.TransportAllowance), "C"},
				{fmt.Sprintf("%.0f", s.FoodAllowance), "C"},
				{fmt.Sprintf("%.0f", s.GrossSalary), "C"},
				{fmt.Sprintf("%.0f", s.AbsentDeduction), "C"},
				{fmt.Sprintf("%.0f", s.TotalDeductions), "C"},
				{fmt.Sprintf("%.0f", s.OvertimeHours), "C"},
				{fmt.Sprintf("%.0f", s.OvertimeRate), "C"},
				{fmt.Sprintf("%.0f", s.OvertimeAmount), "C"},
				{fmt.Sprintf("%.0f", s.AttendanceBonus), "C"},
				{fmt.Sprintf("%.0f", s.NetSalary), "C"},
				{"", "C"},
			}

			for j, d := range data {
				pdf.CellFormat(cols[j].width, rowH, d.val, "1", 0, d.align, false, 0, "")
			}
			pdf.Ln(rowH)

			totalBasic += s.BasicSalary
			totalHouse += s.HouseRent
			totalMed += s.MedicalAllowance
			totalTrans += s.TransportAllowance
			totalFood += s.FoodAllowance
			totalGross += s.GrossSalary
			totalAbsent += s.AbsentDeduction
			totalDeductions += s.TotalDeductions
			totalOTHours += s.OvertimeHours
			totalOTAmount += s.OvertimeAmount
			totalBonus += s.AttendanceBonus
			totalNet += s.NetSalary

			if pdf.GetY() > sheetPDFPageH-20 {
				renderFooter()
				pdf.AddPage()
				renderHeader(grp.displayName)
			}
		}

		// Total Row for this Group
		pdf.SetDrawColor(217, 217, 217)
		pdf.SetFillColor(226, 239, 218)
		pdf.SetTextColor(0, 97, 0)
		pdf.SetFont(font, "B", 6)
		totalRowH := 6.5
		totalData := []struct {
			val   string
			align string
		}{
			{"", "C"},
			{"", "C"},
			{labels.TotalLabel, "L"},
			{"", "L"},
			{"", "C"},
			{"", "C"},
			{"", "C"},
			{"", "C"},
			{"", "C"},
			{"", "C"},
			{"", "C"},
			{fmt.Sprintf("%.0f", totalBasic), "C"},
			{fmt.Sprintf("%.0f", totalHouse), "C"},
			{fmt.Sprintf("%.0f", totalMed), "C"},
			{fmt.Sprintf("%.0f", totalTrans), "C"},
			{fmt.Sprintf("%.0f", totalFood), "C"},
			{fmt.Sprintf("%.0f", totalGross), "C"},
			{fmt.Sprintf("%.0f", totalAbsent), "C"},
			{fmt.Sprintf("%.0f", totalDeductions), "C"},
			{fmt.Sprintf("%.0f", totalOTHours), "C"},
			{"", "C"},
			{fmt.Sprintf("%.0f", totalOTAmount), "C"},
			{fmt.Sprintf("%.0f", totalBonus), "C"},
			{fmt.Sprintf("%.0f", totalNet), "C"},
			{"", "C"},
		}
		for j, d := range totalData {
			pdf.CellFormat(cols[j].width, totalRowH, d.val, "1", 0, d.align, true, 0, "")
		}
		pdf.Ln(totalRowH)
	}

	renderFooter()

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
// @Description  Download monthly salary summary as 5 landscape Legal pages
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

	buildGroupData := func(groupMode string) ([]groupKey, map[groupKey]*groupData) {
		gMap := make(map[groupKey]*groupData)
		for _, s := range salaries {
			var key groupKey
			switch groupMode {
			case "section":
				if s.Employee.SectionRef != nil && strings.TrimSpace(s.Employee.SectionRef.Name) != "" {
					key = groupKey{Name: sectionName(s.Employee.SectionRef, lang), ID: s.Employee.SectionRef.ID}
				} else {
					key = groupKey{Name: "No Section", ID: ""}
				}
			case "designation":
				if s.Employee.DesignationRef != nil && strings.TrimSpace(s.Employee.DesignationRef.Name) != "" {
					key = groupKey{Name: designationName(s.Employee.DesignationRef, lang), ID: s.Employee.DesignationRef.ID}
				} else {
					key = groupKey{Name: "No Designation", ID: ""}
				}
			case "line":
				if s.Employee.LineRef != nil && strings.TrimSpace(s.Employee.LineRef.Name) != "" {
					lName := lineName(s.Employee.LineRef, lang)
					if strings.EqualFold(strings.TrimSpace(s.Employee.LineRef.Name), "admin") {
						lName = "Loader & Cleaner"
					}
					key = groupKey{Name: lName, ID: s.Employee.LineRef.ID}
				} else {
					key = groupKey{Name: "No Line", ID: ""}
				}
			default:
				if s.Employee.Department != nil && strings.TrimSpace(s.Employee.Department.Name) != "" {
					key = groupKey{Name: departmentName(s.Employee.Department, lang), ID: s.Employee.Department.ID}
				} else {
					key = groupKey{Name: "No Department", ID: ""}
				}
			}
			if gMap[key] == nil {
				gMap[key] = &groupData{}
			}
			d := gMap[key]
			d.Employees++
			d.BasicSalary += s.BasicSalary
			d.HouseRent += s.HouseRent
			d.Medical += s.MedicalAllowance
			d.Transport += s.TransportAllowance
			d.GrossSalary += s.GrossSalary
			d.Deductions += s.TotalDeductions
			d.NetSalary += s.NetSalary
		}
		var keys []groupKey
		for k := range gMap {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			rankI := customOrderRank(keys[i].Name)
			rankJ := customOrderRank(keys[j].Name)
			if rankI != rankJ {
				return rankI < rankJ
			}
			return keys[i].Name < keys[j].Name
		})
		return keys, gMap
	}

	buildCustomSummaryData := func() ([]groupKey, map[groupKey]*groupData) {
		gMap := make(map[groupKey]*groupData)
		addRecord := func(kName string, s models.Salary) {
			key := groupKey{Name: kName}
			if gMap[key] == nil {
				gMap[key] = &groupData{}
			}
			d := gMap[key]
			d.Employees++
			d.BasicSalary += s.BasicSalary
			d.HouseRent += s.HouseRent
			d.Medical += s.MedicalAllowance
			d.Transport += s.TransportAllowance
			d.GrossSalary += s.GrossSalary
			d.Deductions += s.TotalDeductions
			d.NetSalary += s.NetSalary
		}

		for _, s := range salaries {
			grpName := ""
			if s.Employee.GroupRef != nil {
				grpName = strings.ToLower(s.Employee.GroupRef.Name)
			}
			empType := strings.ToLower(s.Employee.EmployeeType)
			deptName := ""
			if s.Employee.Department != nil {
				deptName = strings.ToLower(s.Employee.Department.Name)
			}

			isStaff := strings.Contains(grpName, "staff") || strings.Contains(empType, "staff")

			if isStaff {
				if strings.Contains(deptName, "production") || strings.Contains(deptName, "maintenance") {
					addRecord("Production Staff", s)
				} else {
					addRecord("Office Staff", s)
				}
			} else {
				lineName := "No Line"
				if s.Employee.LineRef != nil && strings.TrimSpace(s.Employee.LineRef.Name) != "" {
					lineName = strings.TrimSpace(s.Employee.LineRef.Name)
				}
				if strings.EqualFold(lineName, "admin") {
					lineName = "Loader & Cleaner"
				}
				addRecord(lineName, s)
			}
		}

		var keys []groupKey
		for k := range gMap {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			rankI := customOrderRank(keys[i].Name)
			rankJ := customOrderRank(keys[j].Name)
			if rankI != rankJ {
				return rankI < rankJ
			}
			return keys[i].Name < keys[j].Name
		})

		return keys, gMap
	}

	var company models.Company
	if len(salaries) > 0 {
		company = salaries[0].Company
	}
	compName := companyNameFor(lang, company)
	compAddr := companyAddress(company, lang)

	// Create Landscape A4 PDF (297mm x 210mm)
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(sheetPDFLeft, sheetPDFTop, sheetPDFRight)

	font := "Helvetica"
	if lang == "bn" {
		font = loadBanglaFont(pdf)
	}

	summaryTabs := []struct {
		groupMode    string
		groupLabel   string
		subtitleText string
	}{
		{"department", labels.Department, "Summary By Department"},
		{"section", labels.Section, "Summary By Section"},
		{"designation", labels.Designation, "Summary By Designation"},
		{"line", labels.GroupLabel, "Summary By Line"},
		{"custom", "Category / Line", "Custom Summary Report"},
	}

	type summaryCol struct {
		header string
		width  float64
		align  string
	}

	renderFooter := func() {
		pdf.SetY(210.0 - 12)
		pdf.SetFont(font, "", 7)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(70, 4, "Prepared By", "", 0, "L", false, 0, "")
		pdf.CellFormat(147, 4, "Admin (A.G.M)                     Asst. General Manager", "", 0, "C", false, 0, "")
		pdf.CellFormat(70, 4, "Approved By", "", 0, "R", false, 0, "")
	}

	for pageIdx, tab := range summaryTabs {
		if pageIdx > 0 {
			renderFooter()
		}
		pdf.AddPage()

		// Page Header
		pdf.SetFont(font, "B", 14)
		pdf.SetTextColor(15, 23, 42)
		pdf.Cell(0, 6, compName)
		pdf.Ln(6)

		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(100, 100, 100)
		if compAddr != "" {
			pdf.Cell(0, 4, compAddr)
			pdf.Ln(4)
		}

		pdf.SetFont(font, "B", 11)
		pdf.SetTextColor(15, 23, 42)
		reportTitle := fmt.Sprintf("Salary Summary Report - %s %d", monthLabel, year)
		if lang == "bn" {
			reportTitle = utils.UnicodeToBijoy(reportTitle)
		}
		pdf.Cell(0, 5, reportTitle)
		pdf.Ln(7)

		sCols := []summaryCol{
			{labels.Sl, 8, "C"},
			{tab.groupLabel, 50, "L"},
			{labels.Employees, 18, "C"},
			{labels.BasicTotal, 30, "C"},
			{labels.HouseRentH, 30, "C"},
			{labels.MedicalH, 25, "C"},
			{labels.TransportH, 25, "C"},
			{labels.GrossTotal, 34, "C"},
			{labels.Deductions, 30, "C"},
			{labels.NetTotal, 37, "C"},
		}

		headerH := 8.0
		pdf.SetDrawColor(217, 217, 217)
		pdf.SetFillColor(68, 114, 196)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont(font, "B", 7)
		for _, sc := range sCols {
			pdf.CellFormat(sc.width, headerH, sc.header, "1", 0, sc.align, true, 0, "")
		}
		pdf.Ln(headerH)

		var keys []groupKey
		var gMap map[groupKey]*groupData
		if tab.groupMode == "custom" {
			keys, gMap = buildCustomSummaryData()
		} else {
			keys, gMap = buildGroupData(tab.groupMode)
		}

		pdf.SetFont(font, "", 7)
		rowH := 6.5
		var grand groupData

		for i, key := range keys {
			d := gMap[key]
			pdf.SetDrawColor(217, 217, 217)
			pdf.SetFillColor(255, 255, 255)
			pdf.SetTextColor(30, 30, 30)

			gName := key.Name
			if lang == "bn" {
				gName = utils.UnicodeToBijoy(gName)
			}

			data := []struct {
				val   string
				align string
			}{
				{strconv.Itoa(i + 1), "C"},
				{gName, "L"},
				{strconv.Itoa(d.Employees), "C"},
				{fmt.Sprintf("%.0f", d.BasicSalary), "C"},
				{fmt.Sprintf("%.0f", d.HouseRent), "C"},
				{fmt.Sprintf("%.0f", d.Medical), "C"},
				{fmt.Sprintf("%.0f", d.Transport), "C"},
				{fmt.Sprintf("%.0f", d.GrossSalary), "C"},
				{fmt.Sprintf("%.0f", d.Deductions), "C"},
				{fmt.Sprintf("%.0f", d.NetSalary), "C"},
			}
			for j, dt := range data {
				pdf.CellFormat(sCols[j].width, rowH, dt.val, "1", 0, dt.align, false, 0, "")
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
		}

		// Grand Total Row
		pdf.SetDrawColor(217, 217, 217)
		pdf.SetFillColor(226, 239, 218)
		pdf.SetTextColor(0, 97, 0)
		pdf.SetFont(font, "B", 7)
		totalRowH := 7.0
		gData := []struct {
			val   string
			align string
		}{
			{"", "C"},
			{labels.GrandTotal, "L"},
			{strconv.Itoa(grand.Employees), "C"},
			{fmt.Sprintf("%.0f", grand.BasicSalary), "C"},
			{fmt.Sprintf("%.0f", grand.HouseRent), "C"},
			{fmt.Sprintf("%.0f", grand.Medical), "C"},
			{fmt.Sprintf("%.0f", grand.Transport), "C"},
			{fmt.Sprintf("%.0f", grand.GrossSalary), "C"},
			{fmt.Sprintf("%.0f", grand.Deductions), "C"},
			{fmt.Sprintf("%.0f", grand.NetSalary), "C"},
		}
		for j, dt := range gData {
			pdf.CellFormat(sCols[j].width, totalRowH, dt.val, "1", 0, dt.align, true, 0, "")
		}
		pdf.Ln(totalRowH)
	}

	renderFooter()

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

// DailySummaryExportPDF godoc
//
// @Summary      Export Daily Salary Summary to PDF
// @Description  Download daily salary summary as 5 landscape A4 pages
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        date           query string true  "Date (YYYY-MM-DD)"
// @Param        company_id     query string false "Company ID"
// @Param        lang           query string false "Language: en or bn (default en)"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/daily-summary/export/pdf [get]
func (h *SalaryHandler) DailySummaryExportPDF(c *gin.Context) {
	date := c.Query("date")
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	lang := c.DefaultQuery("lang", "en")

	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	records, err := h.salaryRepo.DailySheet(date, companyID,
		c.Query("department_id"), c.Query("section_id"),
		c.Query("designation_id"), c.Query("line_id"), c.Query("group_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if companyID != "" {
		_ = h.salaryRepo.DB().Where("id = ?", companyID).First(&company).Error
	}
	compName := companyNameFor(lang, company)
	compAddr := companyAddress(company, lang)
	labels := getSalarySheetLabels(lang)

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(sheetPDFLeft, sheetPDFTop, sheetPDFRight)

	font := "Helvetica"
	if lang == "bn" {
		font = loadBanglaFont(pdf)
	}

	summaryTabs := []struct {
		groupMode  string
		groupLabel string
	}{
		{"department", labels.Department},
		{"section", labels.Section},
		{"designation", labels.Designation},
		{"line", labels.GroupLabel},
		{"custom", "Category / Line"},
	}

	type summaryCol struct {
		header string
		width  float64
		align  string
	}

	renderFooter := func() {
		pdf.SetY(210.0 - 12)
		pdf.SetFont(font, "", 7)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(70, 4, "Prepared By", "", 0, "L", false, 0, "")
		pdf.CellFormat(147, 4, "Admin (A.G.M)                     Asst. General Manager", "", 0, "C", false, 0, "")
		pdf.CellFormat(70, 4, "Approved By", "", 0, "R", false, 0, "")
	}

	for pageIdx, tab := range summaryTabs {
		if pageIdx > 0 {
			renderFooter()
		}
		pdf.AddPage()

		pdf.SetFont(font, "B", 14)
		pdf.SetTextColor(15, 23, 42)
		pdf.Cell(0, 6, compName)
		pdf.Ln(6)

		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(100, 100, 100)
		if compAddr != "" {
			pdf.Cell(0, 4, compAddr)
			pdf.Ln(4)
		}

		pdf.SetFont(font, "B", 11)
		pdf.SetTextColor(15, 23, 42)
		reportTitle := fmt.Sprintf("Daily Salary Summary Report - %s", date)
		if lang == "bn" {
			reportTitle = utils.UnicodeToBijoy(reportTitle)
		}
		pdf.Cell(0, 5, reportTitle)
		pdf.Ln(7)

		sCols := []summaryCol{
			{labels.Sl, 10, "C"},
			{tab.groupLabel, 65, "L"},
			{labels.Employees, 22, "C"},
			{"Gross Salary", 40, "C"},
			{"Daily Rate", 38, "C"},
			{"OT Hours", 32, "C"},
			{"OT Amount", 38, "C"},
			{"Total Pay", 42, "C"},
		}

		headerH := 8.0
		pdf.SetDrawColor(217, 217, 217)
		pdf.SetFillColor(68, 114, 196)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont(font, "B", 7)
		for _, sc := range sCols {
			pdf.CellFormat(sc.width, headerH, sc.header, "1", 0, sc.align, true, 0, "")
		}
		pdf.Ln(headerH)

		type groupData struct {
			Name        string
			Employees   int
			GrossSalary float64
			DailyRate   float64
			OtHours     float64
			OtAmount    float64
			TotalPay    float64
		}
		gMap := make(map[string]*groupData)

		for _, r := range records {
			keyName := r.DepartmentName
			if tab.groupMode == "designation" {
				keyName = r.Designation
			} else if tab.groupMode == "section" || tab.groupMode == "line" || tab.groupMode == "custom" {
				keyName = r.DepartmentName
			}
			if strings.TrimSpace(keyName) == "" {
				keyName = "Other"
			}
			if gMap[keyName] == nil {
				gMap[keyName] = &groupData{Name: keyName}
			}
			d := gMap[keyName]
			d.Employees++
			d.GrossSalary += r.GrossSalary
			d.DailyRate += r.DailyRate
			d.OtHours += r.OtHours
			d.OtAmount += r.OtAmount
			d.TotalPay += r.TotalPay
		}

		var keys []string
		for k := range gMap {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			rankI := customOrderRank(keys[i])
			rankJ := customOrderRank(keys[j])
			if rankI != rankJ {
				return rankI < rankJ
			}
			return keys[i] < keys[j]
		})

		pdf.SetFont(font, "", 7)
		rowH := 6.5
		var grandEmp int
		var grandGross, grandDaily, grandOTHours, grandOTAmt, grandPay float64

		for i, k := range keys {
			d := gMap[k]
			pdf.SetDrawColor(217, 217, 217)
			pdf.SetFillColor(255, 255, 255)
			pdf.SetTextColor(30, 30, 30)

			gName := d.Name
			if lang == "bn" {
				gName = utils.UnicodeToBijoy(gName)
			}

			data := []struct {
				val   string
				align string
			}{
				{strconv.Itoa(i + 1), "C"},
				{gName, "L"},
				{strconv.Itoa(d.Employees), "C"},
				{fmt.Sprintf("%.0f", d.GrossSalary), "C"},
				{fmt.Sprintf("%.0f", d.DailyRate), "C"},
				{fmt.Sprintf("%.1f", d.OtHours), "C"},
				{fmt.Sprintf("%.0f", d.OtAmount), "C"},
				{fmt.Sprintf("%.0f", d.TotalPay), "C"},
			}
			for j, dt := range data {
				pdf.CellFormat(sCols[j].width, rowH, dt.val, "1", 0, dt.align, false, 0, "")
			}
			pdf.Ln(rowH)

			grandEmp += d.Employees
			grandGross += d.GrossSalary
			grandDaily += d.DailyRate
			grandOTHours += d.OtHours
			grandOTAmt += d.OtAmount
			grandPay += d.TotalPay
		}

		pdf.SetDrawColor(217, 217, 217)
		pdf.SetFillColor(226, 239, 218)
		pdf.SetTextColor(0, 97, 0)
		pdf.SetFont(font, "B", 7)
		totalRowH := 7.0
		gData := []struct {
			val   string
			align string
		}{
			{"", "C"},
			{labels.GrandTotal, "L"},
			{strconv.Itoa(grandEmp), "C"},
			{fmt.Sprintf("%.0f", grandGross), "C"},
			{fmt.Sprintf("%.0f", grandDaily), "C"},
			{fmt.Sprintf("%.1f", grandOTHours), "C"},
			{fmt.Sprintf("%.0f", grandOTAmt), "C"},
			{fmt.Sprintf("%.0f", grandPay), "C"},
		}
		for j, dt := range gData {
			pdf.CellFormat(sCols[j].width, totalRowH, dt.val, "1", 0, dt.align, true, 0, "")
		}
		pdf.Ln(totalRowH)
	}

	renderFooter()

	filename := fmt.Sprintf("daily_salary_summary_%s.pdf", date)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}
