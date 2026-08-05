package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/xuri/excelize/v2"
)

// monthlyReportParams holds the validated filter inputs shared by the monthly
// report Excel and PDF exports.
type monthlyReportParams struct {
	year, month, companyID string
	departmentID           string
	sectionID              string
	designationID          string
	lineID                 string
	groupID                string
	shiftID                string
	employeeID             string
	startStr               string
	endStr                 string
	monthLabel             string
}

func parseMonthlyReportParams(c *gin.Context) (*monthlyReportParams, error) {
	year := c.Query("year")
	month := c.Query("month")
	companyID := c.Query("company_id")

	if year == "" || month == "" || companyID == "" {
		return nil, fmt.Errorf("year, month, and company_id are required")
	}

	y, err := time.Parse("2006", year)
	if err != nil {
		return nil, fmt.Errorf("invalid year")
	}
	m, err := time.Parse("1", month)
	if err != nil {
		return nil, fmt.Errorf("invalid month")
	}

	startDate := time.Date(y.Year(), m.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	return &monthlyReportParams{
		year:          year,
		month:         month,
		companyID:     companyID,
		departmentID:  c.Query("department_id"),
		sectionID:     c.Query("section_id"),
		designationID: c.Query("designation_id"),
		lineID:        c.Query("line_id"),
		groupID:       c.Query("group_id"),
		shiftID:       c.Query("shift_id"),
		employeeID:    c.Query("employee_id"),
		startStr:      startDate.Format("2006-01-02"),
		endStr:        endDate.Format("2006-01-02"),
		monthLabel:    m.Month().String() + " " + year,
	}, nil
}

func (p *monthlyReportParams) load(attRepo *repository.AttendanceRepository) ([]map[string]interface{}, error) {
	return attRepo.MonthlyReport(
		p.startStr, p.endStr, p.companyID,
		p.departmentID, p.sectionID, p.designationID, p.lineID, p.groupID, p.shiftID, p.employeeID,
	)
}

// monthReportCompany returns the company record used in the export header.
func monthReportCompany() models.Company {
	var company models.Company
	database.DB.First(&company)
	return company
}

func monthlyMapInt(m map[string]interface{}, key string) int64 {
	switch v := m[key].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	}
	return 0
}

// ExportMonthlyReportExcel godoc
//
//	@Summary      Export monthly attendance report to Excel
//	@Description  Download per-employee monthly attendance summary as an Excel file
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        year           query int    true  "Year"
//	@Param        month          query int    true  "Month (1-12)"
//	@Param        company_id     query string true  "Company ID"
//	@Param        department_id  query string false "Filter by department"
//	@Param        section_id     query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id        query string false "Filter by line"
//	@Param        group_id       query string false "Filter by group"
//	@Param        shift_id       query string false "Filter by shift"
//	@Param        employee_id    query string false "Search by employee ID (partial match)"
//	@Success      200            {file}  file
//	@Failure      400            {object}  map[string]string
//	@Failure      500            {object}  map[string]string
//	@Router       /attendance/monthly-report/export/excel [get]
func (h *AttendanceHandler) ExportMonthlyReportExcel(c *gin.Context) {
	p, err := parseMonthlyReportParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := p.load(h.attendanceRepo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	company := monthReportCompany()
	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	// Build the workbook in memory and stream it back to the client.
	excelBytes, err := buildMonthlyReportExcel(companyName, companyAddress, p.monthLabel, results)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=monthly_attendance_report_%s_%s.xlsx", p.month, p.year))
	c.Writer.Write(excelBytes)
}

func buildMonthlyReportExcel(companyName, companyAddress, monthLabel string, results []map[string]interface{}) ([]byte, error) {
	const nCols = 15

	cols := []struct {
		header string
		width  float64
	}{
		{"Sl", 5}, {"Emp. ID", 12}, {"Name", 26}, {"Designation", 24}, {"Department", 20},
		{"Shift", 14}, {"Present", 10}, {"Absent", 10}, {"Late", 9}, {"Leave", 9},
		{"Weekend", 10}, {"Half Day", 10}, {"Holiday", 10}, {"OverTime", 10}, {"Total", 9},
	}

	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", "Monthly Report")

	thinBorder := []excelize.Border{
		{Type: "left", Color: "333333", Style: 1}, {Type: "top", Color: "333333", Style: 1},
		{Type: "bottom", Color: "333333", Style: 1}, {Type: "right", Color: "333333", Style: 1},
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	reportTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 13, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4470AF"}, Pattern: 1},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	textLeftStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	numCenterStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	totalRowStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9E1F2"}, Pattern: 1},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	endCol := colNameAttendance(nCols)

	for i, col := range cols {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth("Monthly Report", colLetter, colLetter, col.width)
	}

	// Header block
	f.SetCellValue("Monthly Report", "A1", companyName)
	f.MergeCell("Monthly Report", "A1", endCol+"1")
	f.SetCellStyle("Monthly Report", "A1", endCol+"1", titleStyle)
	f.SetRowHeight("Monthly Report", 1, 30)

	f.SetCellValue("Monthly Report", "A2", companyAddress)
	f.MergeCell("Monthly Report", "A2", endCol+"2")
	f.SetCellStyle("Monthly Report", "A2", endCol+"2", subtitleStyle)
	f.SetRowHeight("Monthly Report", 2, 18)

	f.SetCellValue("Monthly Report", "A3", "MONTHLY ATTENDANCE REPORT - "+monthLabel)
	f.MergeCell("Monthly Report", "A3", endCol+"3")
	f.SetCellStyle("Monthly Report", "A3", endCol+"3", reportTitleStyle)
	f.SetRowHeight("Monthly Report", 3, 20)

	// Column headers
	headerRow := 4
	for i, col := range cols {
		axis, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		f.SetCellValue("Monthly Report", axis, col.header)
		f.SetCellStyle("Monthly Report", axis, axis, headerStyle)
	}
	f.SetRowHeight("Monthly Report", headerRow, 22)

	// Data rows
	totals := make([]int64, len(cols))
	row := headerRow + 1
	for idx, rec := range results {
		empID := fmt.Sprintf("%v", rec["emp_id"])
		if empID == "<nil>" {
			empID = ""
		}
		name := fmt.Sprintf("%v", rec["employee_name"])
		if name == "<nil>" {
			name = ""
		}
		desig := fmt.Sprintf("%v", rec["designation_name"])
		if desig == "<nil>" {
			desig = ""
		}
		dept := fmt.Sprintf("%v", rec["department_name"])
		if dept == "<nil>" {
			dept = ""
		}
		shift := fmt.Sprintf("%v", rec["shift_name"])
		if shift == "<nil>" {
			shift = ""
		}

		values := []struct {
			val    interface{}
			center bool
		}{
			{idx + 1, true},
			{empID, false},
			{name, false},
			{desig, false},
			{dept, false},
			{shift, false},
			{monthlyMapInt(rec, "present"), true},
			{monthlyMapInt(rec, "absent"), true},
			{monthlyMapInt(rec, "late"), true},
			{monthlyMapInt(rec, "leave"), true},
			{monthlyMapInt(rec, "weekend"), true},
			{monthlyMapInt(rec, "half_day"), true},
			{monthlyMapInt(rec, "holiday"), true},
			{monthlyMapInt(rec, "over_time"), true},
			{monthlyMapInt(rec, "total_days"), true},
		}

		for j, v := range values {
			axis, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue("Monthly Report", axis, v.val)
			if v.center {
				f.SetCellStyle("Monthly Report", axis, axis, numCenterStyle)
			} else {
				f.SetCellStyle("Monthly Report", axis, axis, textLeftStyle)
			}
			if num, ok := v.val.(int64); ok {
				if j >= 6 {
					totals[j] += num
				}
			}
		}
		f.SetRowHeight("Monthly Report", row, 18)
		row++
	}

	// Totals row
	f.SetCellValue("Monthly Report", "A"+strconv.Itoa(row), "Total")
	f.SetCellStyle("Monthly Report", "A"+strconv.Itoa(row), "A"+strconv.Itoa(row), totalRowStyle)
	for j := 1; j < nCols; j++ {
		axis, _ := excelize.CoordinatesToCellName(j+1, row)
		if j >= 6 {
			f.SetCellValue("Monthly Report", axis, totals[j])
		} else {
			f.SetCellValue("Monthly Report", axis, "")
		}
		f.SetCellStyle("Monthly Report", axis, axis, totalRowStyle)
	}
	f.SetRowHeight("Monthly Report", row, 18)

	// Freeze the header so columns stay visible while scrolling.
	f.SetPanes("Monthly Report", &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      1,
		YSplit:      headerRow,
		TopLeftCell: "B5",
		ActivePane:  "bottomRight",
	})

	// Print setup so exports fit on A4 landscape.
	orientation := "landscape"
	pageSize := 9
	fitWidth := 1
	fitHeight := 0
	f.SetPageLayout("Monthly Report", &excelize.PageLayoutOptions{Orientation: &orientation, Size: &pageSize, FitToWidth: &fitWidth, FitToHeight: &fitHeight})
	f.SetPageMargins("Monthly Report", &excelize.PageLayoutMarginsOptions{Left: ptr(0.3), Right: ptr(0.3), Top: ptr(0.4), Bottom: ptr(0.4)})
	f.SetSheetView("Monthly Report", -1, &excelize.ViewOptions{ShowGridLines: ptrBool(false)})

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportMonthlyReportPDF godoc
//
//	@Summary      Export monthly attendance report to PDF
//	@Description  Download per-employee monthly attendance summary as a landscape PDF
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/pdf
//	@Param        year           query int    true  "Year"
//	@Param        month          query int    true  "Month (1-12)"
//	@Param        company_id     query string true  "Company ID"
//	@Param        department_id  query string false "Filter by department"
//	@Param        section_id     query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id        query string false "Filter by line"
//	@Param        group_id       query string false "Filter by group"
//	@Param        shift_id       query string false "Filter by shift"
//	@Param        employee_id    query string false "Search by employee ID (partial match)"
//	@Success      200            {file}  file
//	@Failure      400            {object}  map[string]string
//	@Failure      500            {object}  map[string]string
//	@Router       /attendance/monthly-report/export/pdf [get]
func (h *AttendanceHandler) ExportMonthlyReportPDF(c *gin.Context) {
	p, err := parseMonthlyReportParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := p.load(h.attendanceRepo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	company := monthReportCompany()
	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	excelBytes, err := buildMonthlyReportPDF(companyName, companyAddress, p.monthLabel, results)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=monthly_attendance_report_%s_%s.pdf", p.month, p.year))
	c.Writer.Write(excelBytes)
}

func buildMonthlyReportPDF(companyName, companyAddress, monthLabel string, results []map[string]interface{}) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(6, 10, 6)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	font := "Helvetica"

	type colDef struct {
		header string
		width  float64
		align  string
	}
	cols := []colDef{
		{"Sl", 8, "C"},
		{"Emp. ID", 18, "C"},
		{"Name", 34, "L"},
		{"Designation", 30, "L"},
		{"Department", 28, "L"},
		{"Shift", 18, "C"},
		{"Present", 12, "C"},
		{"Absent", 12, "C"},
		{"Late", 11, "C"},
		{"Leave", 11, "C"},
		{"Weekend", 12, "C"},
		{"Half Day", 13, "C"},
		{"Holiday", 12, "C"},
		{"OverTime", 12, "C"},
		{"Total", 12, "C"},
	}

	headerH := 8.0
	rowH := 6.5

	// Header block
	pdf.SetFont(font, "B", 13)
	pdf.SetTextColor(15, 23, 42)
	pdf.Cell(0, 8, companyName)
	pdf.Ln(8)

	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, companyAddress)
	pdf.Ln(5)

	pdf.SetFont(font, "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.Cell(0, 7, "MONTHLY ATTENDANCE REPORT - "+monthLabel)
	pdf.Ln(10)

	// Totals computed across all data rows for the footer.
	var totalPresent, totalAbsent, totalLate, totalLeave, totalWeekend, totalHalfDay, totalHoliday, totalOT, totalDays int64

	// Column header
	pdf.SetFillColor(68, 114, 196)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(font, "B", 6)
	for _, c := range cols {
		pdf.CellFormat(c.width, headerH, c.header, "1", 0, c.align, true, 0, "")
	}
	pdf.Ln(headerH)

	// Data rows
	for i, rec := range results {
		bg := i%2 == 0
		if bg {
			pdf.SetFillColor(240, 244, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(30, 30, 30)

		empID := fmt.Sprintf("%v", rec["emp_id"])
		if empID == "<nil>" {
			empID = ""
		}
		name := fmt.Sprintf("%v", rec["employee_name"])
		if name == "<nil>" {
			name = ""
		}
		desig := fmt.Sprintf("%v", rec["designation_name"])
		if desig == "<nil>" {
			desig = ""
		}
		dept := fmt.Sprintf("%v", rec["department_name"])
		if dept == "<nil>" {
			dept = ""
		}
		shift := fmt.Sprintf("%v", rec["shift_name"])
		if shift == "<nil>" {
			shift = ""
		}

		values := []struct {
			val   string
			align string
		}{
			{strconv.Itoa(i + 1), "C"},
			{truncate(empID, 10), "C"},
			{truncate(name, 22), "L"},
			{truncate(desig, 20), "L"},
			{truncate(dept, 18), "L"},
			{truncate(shift, 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "present"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "absent"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "late"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "leave"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "weekend"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "half_day"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "holiday"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "over_time"), 10), "C"},
			{strconv.FormatInt(monthlyMapInt(rec, "total_days"), 10), "C"},
		}

		pdf.SetFont(font, "", 6)
		for j, d := range values {
			pdf.CellFormat(cols[j].width, rowH, d.val, "1", 0, d.align, bg, 0, "")
		}
		pdf.Ln(rowH)

		totalPresent += monthlyMapInt(rec, "present")
		totalAbsent += monthlyMapInt(rec, "absent")
		totalLate += monthlyMapInt(rec, "late")
		totalLeave += monthlyMapInt(rec, "leave")
		totalWeekend += monthlyMapInt(rec, "weekend")
		totalHalfDay += monthlyMapInt(rec, "half_day")
		totalHoliday += monthlyMapInt(rec, "holiday")
		totalOT += monthlyMapInt(rec, "over_time")
		totalDays += monthlyMapInt(rec, "total_days")

		if pdf.GetY() > 200-15 {
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

	// Totals row
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
		{"Total", "L"},
		{"", "C"},
		{strconv.FormatInt(totalPresent, 10), "C"},
		{strconv.FormatInt(totalAbsent, 10), "C"},
		{strconv.FormatInt(totalLate, 10), "C"},
		{strconv.FormatInt(totalLeave, 10), "C"},
		{strconv.FormatInt(totalWeekend, 10), "C"},
		{strconv.FormatInt(totalHalfDay, 10), "C"},
		{strconv.FormatInt(totalHoliday, 10), "C"},
		{strconv.FormatInt(totalOT, 10), "C"},
		{strconv.FormatInt(totalDays, 10), "C"},
	}
	for j, d := range totalData {
		pdf.CellFormat(cols[j].width, totalRowH, d.val, "1", 0, d.align, true, 0, "")
	}
	pdf.Ln(totalRowH)

	// Footer
	pdf.Ln(5)
	pdf.SetFont(font, "", 8)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total Employees: %d", len(results)), "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
