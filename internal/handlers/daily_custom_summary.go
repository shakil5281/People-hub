package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/xuri/excelize/v2"
)

type CustomDailySummaryRow struct {
	ParentSection string `json:"parent_section"`
	Section       string `json:"section"`
	Present       int64  `json:"present"`
	Absent        int64  `json:"absent"`
	Leave         int64  `json:"leave"`
	Others        int64  `json:"others"`
	Total         int64  `json:"total"`
	Remarks       string `json:"remarks"`
	IsSubtotal    bool   `json:"is_subtotal"`
	IsGrandTotal  bool   `json:"is_grand_total"`
	StyleType     string `json:"style_type"` // "normal", "subtotal", "worker_total", "staff_total", "grand_total"
}

type CustomDailySummaryData struct {
	CompanyName    string                  `json:"company_name"`
	CompanyAddress string                  `json:"company_address"`
	ReportTitle    string                  `json:"report_title"`
	Date           string                  `json:"date"`
	FormattedDate  string                  `json:"formatted_date"`
	Rows           []CustomDailySummaryRow `json:"rows"`
	GrandTotal     CustomDailySummaryRow   `json:"grand_total"`
}

func fetchCustomDailySummaryData(companyID, dateStr, lang string) (CustomDailySummaryData, error) {
	db := database.DB

	var comp models.Company
	if companyID != "" {
		db.Where("id = ?", companyID).First(&comp)
	} else {
		db.First(&comp)
	}

	compName := comp.CompanyNameEn
	if compName == "" {
		compName = "EKUSHE FASHIONS LTD"
	}
	compAddress := comp.AddressEn
	if compAddress == "" {
		compAddress = "Masterbari, Gazipur city, Gazipur."
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t = time.Now()
		dateStr = t.Format("2006-01-02")
	}
	formattedDate := t.Format("02-01-2006")

	// Query attendance records for the date
	var attendances []models.Attendance
	q := db.Preload("Employee.Department").Preload("Employee.SectionRef").Preload("Employee.DesignationRef").Preload("Employee.LineRef").Preload("Employee.GroupRef")
	if comp.ID != "" {
		q = q.Where("company_id = ?", comp.ID)
	}
	q.Where("date = ? AND deleted_at IS NULL", dateStr).Find(&attendances)

	// Bucket maps for counting
	// key: section/category label
	counts := make(map[string]map[string]int64)

	addCount := func(label, status string) {
		if _, ok := counts[label]; !ok {
			counts[label] = make(map[string]int64)
		}
		st := strings.ToLower(status)
		switch st {
		case "present", "late":
			counts[label]["present"]++
		case "absent":
			counts[label]["absent"]++
		case "on_leave", "leave":
			counts[label]["leave"]++
		default:
			counts[label]["others"]++
		}
		counts[label]["total"]++
	}

	for _, a := range attendances {
		emp := a.Employee
		dept := ""
		if emp.Department != nil {
			dept = emp.Department.Name
		}
		sec := ""
		if emp.SectionRef != nil {
			sec = emp.SectionRef.Name
		}
		desig := ""
		if emp.DesignationRef != nil {
			desig = emp.DesignationRef.Name
		}
		line := ""
		if emp.LineRef != nil {
			line = emp.LineRef.Name
		}

		grp := ""
		if emp.GroupRef != nil {
			grp = emp.GroupRef.Name
		}

		secLower := strings.ToLower(sec)
		deptLower := strings.ToLower(dept)
		desigLower := strings.ToLower(desig)
		grpLower := strings.ToLower(grp)

		isOfficeStaff := strings.Contains(deptLower, "admin") && strings.Contains(grpLower, "staff")
		isProductionStaff := strings.Contains(deptLower, "production") && strings.Contains(grpLower, "staff")
		isMechanicalStaff := strings.Contains(deptLower, "maintenance") ||
			strings.Contains(deptLower, "mechanical") ||
			strings.Contains(deptLower, "macanical") ||
			strings.Contains(secLower, "maintenance") ||
			strings.Contains(secLower, "mechanical") ||
			strings.Contains(secLower, "macanical") ||
			strings.Contains(desigLower, "mechanical") ||
			strings.Contains(desigLower, "macanical") ||
			strings.Contains(desigLower, "maintenance")

		isCleanerLoader := strings.Contains(desigLower, "cleaner") || strings.Contains(desigLower, "loader") || strings.Contains(desigLower, "sweeper")

		lineLower := strings.ToLower(line)
		isNonSewingLine := false
		for _, kw := range []string{"admin", "cutting", "electrical", "finishing", "mechanical", "production", "quality"} {
			if strings.Contains(lineLower, kw) {
				isNonSewingLine = true
				break
			}
		}

		if isOfficeStaff {
			addCount("Office - Staff", a.Status)
		} else if isMechanicalStaff {
			addCount("Macanical - Staff", a.Status)
		} else if isProductionStaff {
			addCount("Production Staff", a.Status)
		} else if isCleanerLoader {
			addCount("Loder/Cleaner", a.Status)
		} else if strings.Contains(secLower, "cutting") || strings.Contains(deptLower, "cutting") || strings.Contains(lineLower, "cutting") {
			addCount("Cutting", a.Status)
		} else if strings.Contains(secLower, "finishing") || strings.Contains(deptLower, "finishing") || strings.Contains(lineLower, "finishing") {
			addCount("Finishing", a.Status)
		} else if strings.Contains(secLower, "quality") || strings.Contains(deptLower, "quality") || strings.Contains(lineLower, "quality") {
			addCount("Quality", a.Status)
		} else if strings.Contains(secLower, "sewing") || strings.Contains(deptLower, "sewing") || (line != "" && !isNonSewingLine) {
			// Sewing breakdown
			lineNum := extractLineName(line)
			if strings.Contains(desigLower, "helper") {
				label := fmt.Sprintf("%s(Helper)", lineNum)
				addCount(label, a.Status)
			} else if strings.Contains(desigLower, "input") || strings.Contains(desigLower, "inputman") {
				addCount("Inputman", a.Status)
			} else if strings.Contains(desigLower, "iron") || strings.Contains(desigLower, "ironman") {
				addCount("Ironman", a.Status)
			} else if strings.Contains(desigLower, "operator") {
				label := fmt.Sprintf("%s(Operator)", lineNum)
				addCount(label, a.Status)
			} else {
				// Default operator line if line is set
				label := fmt.Sprintf("%s(Operator)", lineNum)
				addCount(label, a.Status)
			}
		} else {
			addCount("Line-01(Operator)", a.Status)
		}
	}

	getRow := func(parent, sec, style string) CustomDailySummaryRow {
		m := counts[sec]
		p := m["present"]
		a := m["absent"]
		l := m["leave"]
		o := m["others"]
		tot := m["total"]
		if tot == 0 && (p+a+l+o) > 0 {
			tot = p + a + l + o
		}
		isSub := style == "subtotal" || style == "worker_total" || style == "staff_total"
		isGrand := style == "grand_total"
		return CustomDailySummaryRow{
			ParentSection: parent,
			Section:       sec,
			Present:       p,
			Absent:        a,
			Leave:         l,
			Others:        o,
			Total:         tot,
			IsSubtotal:    isSub,
			IsGrandTotal:  isGrand,
			StyleType:     style,
		}
	}

	sumRows := func(parent, title, style string, items ...CustomDailySummaryRow) CustomDailySummaryRow {
		var p, a, l, o, tot int64
		for _, item := range items {
			p += item.Present
			a += item.Absent
			l += item.Leave
			o += item.Others
			tot += item.Total
		}
		return CustomDailySummaryRow{
			ParentSection: parent,
			Section:       title,
			Present:       p,
			Absent:        a,
			Leave:         l,
			Others:        o,
			Total:         tot,
			IsSubtotal:    true,
			StyleType:     style,
		}
	}

	var rows []CustomDailySummaryRow

	// 1. Non-sewing sections
	cCut := getRow("", "Cutting", "normal")
	cFin := getRow("", "Finishing", "normal")
	cQual := getRow("", "Quality", "normal")
	subNonSewing := sumRows("", "Total", "subtotal", cCut, cFin, cQual)

	rows = append(rows, cCut, cFin, cQual, subNonSewing)

	// 2. Sewing section: Fetch lines from DB to support dynamic line names (filtering out non-sewing lines)
	var dbLines []models.Line
	db.Where("deleted_at IS NULL").Order("name ASC").Find(&dbLines)

	lineLabels := []string{"Line-01", "Line-02", "Line-03", "Line-04", "Line-05", "Line-06", "Line-07"}
	if len(dbLines) > 0 {
		seen := make(map[string]bool)
		for _, l := range lineLabels {
			seen[l] = true
		}
		for _, dl := range dbLines {
			dlLower := strings.ToLower(dl.Name)
			isNonSewing := false
			for _, kw := range []string{"admin", "cutting", "electrical", "finishing", "mechanical", "production", "quality"} {
				if strings.Contains(dlLower, kw) {
					isNonSewing = true
					break
				}
			}
			if !isNonSewing {
				norm := extractLineName(dl.Name)
				if !seen[norm] {
					lineLabels = append(lineLabels, norm)
					seen[norm] = true
				}
			}
		}
	}

	var helpers []CustomDailySummaryRow
	for _, ll := range lineLabels {
		label := fmt.Sprintf("%s(Helper)", ll)
		helpers = append(helpers, getRow("Sewing", label, "normal"))
	}
	subHelper := sumRows("Sewing", "Total (Helper)", "subtotal", helpers...)

	ironInput := []CustomDailySummaryRow{
		getRow("Sewing", "Inputman", "normal"),
		getRow("Sewing", "Ironman", "normal"),
	}
	subIronInput := sumRows("Sewing", "Total (Iron/Input)", "subtotal", ironInput...)

	var operators []CustomDailySummaryRow
	for _, ll := range lineLabels {
		label := fmt.Sprintf("%s(Operator)", ll)
		operators = append(operators, getRow("Sewing", label, "normal"))
	}
	subOperator := sumRows("Sewing", "Total (Operator)", "subtotal", operators...)

	cleaner := getRow("Sewing", "Loder/Cleaner", "normal")

	totWorker := sumRows("Sewing", "Total -Worker", "worker_total", cCut, cFin, cQual, subHelper, subIronInput, subOperator, cleaner)

	for _, h := range helpers {
		rows = append(rows, h)
	}
	rows = append(rows, subHelper)
	for _, ii := range ironInput {
		rows = append(rows, ii)
	}
	rows = append(rows, subIronInput)
	for _, op := range operators {
		rows = append(rows, op)
	}
	rows = append(rows, subOperator, cleaner, totWorker)

	// 3. Staff section
	offStaff := getRow("", "Office - Staff", "normal")
	mechStaff := getRow("", "Macanical - Staff", "normal")
	prodStaff := getRow("", "Production Staff", "normal")
	totStaff := sumRows("", "Total-Staff", "staff_total", offStaff, mechStaff, prodStaff)

	rows = append(rows, offStaff, mechStaff, prodStaff, totStaff)

	// 4. Grand Total
	grandTotal := sumRows("", "Total", "grand_total", totWorker, totStaff)
	grandTotal.IsGrandTotal = true

	data := CustomDailySummaryData{
		CompanyName:    compName,
		CompanyAddress: compAddress,
		ReportTitle:    "Daily Attendance Report",
		Date:           dateStr,
		FormattedDate:  formattedDate,
		Rows:           rows,
		GrandTotal:     grandTotal,
	}

	return data, nil
}

func extractLineName(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return "Line-01"
	}
	var digits strings.Builder
	for _, r := range line {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	if digits.Len() > 0 {
		var n int
		fmt.Sscanf(digits.String(), "%d", &n)
		if n > 0 {
			return fmt.Sprintf("Line-%02d", n)
		}
	}
	if strings.HasPrefix(strings.ToLower(line), "line") {
		return line
	}
	return "Line-" + line
}

// GetCustomDailySummary godoc
//
//	@Summary      Get custom daily attendance summary report data
//	@Tags         Attendance
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Param        date query string false "Date (YYYY-MM-DD)"
//	@Success      200 {object} CustomDailySummaryData
//	@Router       /attendance/custom-daily-summary [get]
func (h *AttendanceHandler) GetCustomDailySummary(c *gin.Context) {
	companyID := c.Query("company_id")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	lang := c.DefaultQuery("lang", "en")

	data, err := fetchCustomDailySummaryData(companyID, dateStr, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetHelperCustomSummary godoc
//
//	@Summary      Get Helper custom summary data
//	@Tags         Attendance
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Param        date query string false "Date (YYYY-MM-DD)"
//	@Router       /attendance/custom-summary/helpers [get]
func (h *AttendanceHandler) GetHelperCustomSummary(c *gin.Context) {
	companyID := c.Query("company_id")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	lang := c.DefaultQuery("lang", "en")

	data, err := fetchCustomDailySummaryData(companyID, dateStr, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var helperRows []CustomDailySummaryRow
	for _, r := range data.Rows {
		if strings.Contains(r.Section, "Helper") {
			helperRows = append(helperRows, r)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":        data.Date,
		"helper_rows": helperRows,
	})
}

// GetOperatorCustomSummary godoc
//
//	@Summary      Get Operator custom summary data
//	@Tags         Attendance
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Param        date query string false "Date (YYYY-MM-DD)"
//	@Router       /attendance/custom-summary/operators [get]
func (h *AttendanceHandler) GetOperatorCustomSummary(c *gin.Context) {
	companyID := c.Query("company_id")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	lang := c.DefaultQuery("lang", "en")

	data, err := fetchCustomDailySummaryData(companyID, dateStr, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var operatorRows []CustomDailySummaryRow
	for _, r := range data.Rows {
		if strings.Contains(r.Section, "Operator") {
			operatorRows = append(operatorRows, r)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":          data.Date,
		"operator_rows": operatorRows,
	})
}

// ExportCustomDailySummaryExcel godoc
//
//	@Summary      Export custom daily attendance summary report to Excel
//	@Tags         Attendance
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        company_id query string false "Company ID"
//	@Param        date query string false "Date (YYYY-MM-DD)"
//	@Router       /attendance/custom-daily-summary/export/excel [get]
func (h *AttendanceHandler) ExportCustomDailySummaryExcel(c *gin.Context) {
	companyID := c.Query("company_id")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	lang := c.DefaultQuery("lang", "en")

	data, err := fetchCustomDailySummaryData(companyID, dateStr, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Daily Summary"
	f.SetSheetName("Sheet1", sheet)

	// Set column widths
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "G", 10)
	f.SetColWidth(sheet, "H", "H", 12)

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "6B21A8", Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Size: 10, Color: "334155", Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	dateStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "0F172A", Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	thinBorder := []excelize.Border{
		{Type: "left", Color: "CBD5E1", Style: 1},
		{Type: "top", Color: "CBD5E1", Style: 1},
		{Type: "right", Color: "CBD5E1", Style: 1},
		{Type: "bottom", Color: "CBD5E1", Style: 1},
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "1E3A8A", Family: "Calibri"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8FAFC"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})

	cellNormal, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "0F172A", Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})
	cellNormalLeft, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "0F172A", Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border:    thinBorder,
	})

	subtotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "0F172A", Family: "Calibri"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E0F2FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})

	workerTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "0F172A", Family: "Calibri"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"BAE6FD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})

	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF", Family: "Calibri"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"00A0E9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    thinBorder,
	})

	// Header rows
	f.MergeCell(sheet, "A1", "H1")
	f.SetCellValue(sheet, "A1", data.CompanyName)
	f.SetCellStyle(sheet, "A1", "H1", titleStyle)
	f.SetRowHeight(sheet, 1, 24)

	f.MergeCell(sheet, "A2", "H2")
	f.SetCellValue(sheet, "A2", data.CompanyAddress)
	f.SetCellStyle(sheet, "A2", "H2", subtitleStyle)
	f.SetRowHeight(sheet, 2, 16)

	f.MergeCell(sheet, "A3", "H3")
	f.SetCellValue(sheet, "A3", data.ReportTitle)
	f.SetCellStyle(sheet, "A3", "H3", subtitleStyle)
	f.SetRowHeight(sheet, 3, 16)

	f.SetCellValue(sheet, "H4", fmt.Sprintf("Date:- %s", data.FormattedDate))
	f.SetCellStyle(sheet, "H4", "H4", dateStyle)
	f.SetRowHeight(sheet, 4, 18)

	// Table Header
	headers := []string{"Section", "Section", "Present", "Abesnt", "Leave", "Others", "Total", "Remarks"}
	f.SetRowHeight(sheet, 5, 22)
	f.MergeCell(sheet, "A5", "B5")
	f.SetCellValue(sheet, "A5", "Section")
	f.SetCellStyle(sheet, "A5", "B5", headerStyle)

	cols := []string{"C", "D", "E", "F", "G", "H"}
	for i, col := range cols {
		cell := fmt.Sprintf("%s5", col)
		f.SetCellValue(sheet, cell, headers[i+2])
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Hide Gridlines
	f.SetSheetView(sheet, -1, &excelize.ViewOptions{
		ShowGridLines: func(b bool) *bool { return &b }(false),
	})

	rowIdx := 6
	sewingStartRow := 0
	sewingEndRow := 0

	for _, r := range data.Rows {
		f.SetRowHeight(sheet, rowIdx, 20)

		st := cellNormal
		stLeft := cellNormalLeft
		if r.StyleType == "subtotal" || r.StyleType == "staff_total" {
			st = subtotalStyle
			stLeft = subtotalStyle
		} else if r.StyleType == "worker_total" {
			st = workerTotalStyle
			stLeft = workerTotalStyle
		}

		if r.ParentSection == "Sewing" {
			if sewingStartRow == 0 {
				sewingStartRow = rowIdx
			}
			sewingEndRow = rowIdx
			f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIdx), r.Section)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIdx), fmt.Sprintf("B%d", rowIdx), stLeft)
		} else {
			f.MergeCell(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("B%d", rowIdx))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), r.Section)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("B%d", rowIdx), stLeft)
		}

		valStr := func(n int64) interface{} {
			return n
		}

		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIdx), valStr(r.Present))
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", rowIdx), fmt.Sprintf("C%d", rowIdx), st)

		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowIdx), valStr(r.Absent))
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", rowIdx), fmt.Sprintf("D%d", rowIdx), st)

		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowIdx), valStr(r.Leave))
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", rowIdx), fmt.Sprintf("E%d", rowIdx), st)

		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), valStr(r.Others))
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", rowIdx), fmt.Sprintf("F%d", rowIdx), st)

		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), valStr(r.Total))
		f.SetCellStyle(sheet, fmt.Sprintf("G%d", rowIdx), fmt.Sprintf("G%d", rowIdx), st)

		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowIdx), r.Remarks)
		f.SetCellStyle(sheet, fmt.Sprintf("H%d", rowIdx), fmt.Sprintf("H%d", rowIdx), st)

		rowIdx++
	}

	// Merge Sewing parent column vertically
	if sewingStartRow > 0 && sewingEndRow >= sewingStartRow {
		sewingCellStart := fmt.Sprintf("A%d", sewingStartRow)
		sewingCellEnd := fmt.Sprintf("A%d", sewingEndRow)
		f.MergeCell(sheet, sewingCellStart, sewingCellEnd)
		f.SetCellValue(sheet, sewingCellStart, "Sewing")
		sewingStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 11, Color: "0F172A", Family: "Calibri"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			Border:    thinBorder,
		})
		f.SetCellStyle(sheet, sewingCellStart, sewingCellEnd, sewingStyle)
	}

	// Grand Total Row
	f.SetRowHeight(sheet, rowIdx, 24)
	f.MergeCell(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("B%d", rowIdx))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), data.GrandTotal.Section)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("B%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIdx), data.GrandTotal.Present)
	f.SetCellStyle(sheet, fmt.Sprintf("C%d", rowIdx), fmt.Sprintf("C%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("D%d", rowIdx), data.GrandTotal.Absent)
	f.SetCellStyle(sheet, fmt.Sprintf("D%d", rowIdx), fmt.Sprintf("D%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("E%d", rowIdx), data.GrandTotal.Leave)
	f.SetCellStyle(sheet, fmt.Sprintf("E%d", rowIdx), fmt.Sprintf("E%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), data.GrandTotal.Others)
	f.SetCellStyle(sheet, fmt.Sprintf("F%d", rowIdx), fmt.Sprintf("F%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), data.GrandTotal.Total)
	f.SetCellStyle(sheet, fmt.Sprintf("G%d", rowIdx), fmt.Sprintf("G%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("H%d", rowIdx), "")
	f.SetCellStyle(sheet, fmt.Sprintf("H%d", rowIdx), fmt.Sprintf("H%d", rowIdx), grandTotalStyle)

	// Page setup: A4 size with custom margins
	marginLeftRight := 0.45
	marginTopBottom := 0.75
	f.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{
		Left:   &marginLeftRight,
		Right:  &marginLeftRight,
		Top:    &marginTopBottom,
		Bottom: &marginTopBottom,
	})
	orientation := "portrait"
	size := 9 // A4 paper size
	f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
		Orientation: &orientation,
		Size:        &size,
	})

	filename := fmt.Sprintf("daily_summary_%s.xlsx", dateStr)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// ExportCustomDailySummaryPDF godoc
//
//	@Summary      Export custom daily attendance summary report to PDF
//	@Tags         Attendance
//	@Produce      application/pdf
//	@Param        company_id query string false "Company ID"
//	@Param        date query string false "Date (YYYY-MM-DD)"
//	@Router       /attendance/custom-daily-summary/export/pdf [get]
func (h *AttendanceHandler) ExportCustomDailySummaryPDF(c *gin.Context) {
	companyID := c.Query("company_id")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	lang := c.DefaultQuery("lang", "en")

	data, err := fetchCustomDailySummaryData(companyID, dateStr, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 10, 15)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	font := "Helvetica"

	pageW := 180.0
	x := 15.0
	curY := 12.0

	// 1. Header
	pdf.SetFont(font, "B", 18)
	pdf.SetTextColor(107, 33, 168) // Purple #6B21A8
	pdf.SetXY(x, curY)
	pdf.CellFormat(pageW, 7, data.CompanyName, "", 0, "C", false, 0, "")
	curY += 7.5

	pdf.SetFont(font, "I", 9.5)
	pdf.SetTextColor(51, 65, 85)
	pdf.SetXY(x, curY)
	pdf.CellFormat(pageW, 4.5, data.CompanyAddress, "", 0, "C", false, 0, "")
	curY += 5.0

	pdf.SetFont(font, "I", 9.5)
	pdf.SetXY(x, curY)
	pdf.CellFormat(pageW, 4.5, data.ReportTitle, "", 0, "C", false, 0, "")
	curY += 5.5

	pdf.SetFont(font, "B", 9.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.SetXY(x, curY)
	pdf.CellFormat(pageW, 4.5, fmt.Sprintf("Date:- %s", data.FormattedDate), "", 0, "R", false, 0, "")
	curY += 6.5

	// Table setup
	colWParent := 20.0
	colWSec := 35.0
	colWVal := 20.0
	colWRem := 25.0
	rowH := 6.5

	// Draw Header Row
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(203, 213, 225)
	pdf.SetLineWidth(0.3)
	pdf.SetFont(font, "B", 8.5)
	pdf.SetTextColor(15, 23, 42)

	drawCell := func(cx, cy, cw, ch float64, txt, align string, bg bool, r, g, b int) {
		pdf.SetDrawColor(203, 213, 225)
		if bg {
			pdf.SetFillColor(r, g, b)
			pdf.Rect(cx, cy, cw, ch, "DF")
		} else {
			pdf.Rect(cx, cy, cw, ch, "D")
		}
		offsetY := (ch - 3.5) / 2.0
		pdf.SetXY(cx+0.5, cy+offsetY)
		pdf.CellFormat(cw-1.0, 3.5, txt, "", 0, align, false, 0, "")
	}

	pdf.Rect(x, curY, colWParent+colWSec, rowH, "DF")
	pdf.SetXY(x+0.5, curY+(rowH-3.5)/2.0)
	pdf.CellFormat(colWParent+colWSec-1.0, 3.5, "Section", "", 0, "C", false, 0, "")

	cx := x + colWParent + colWSec
	headers := []string{"Present", "Abesnt", "Leave", "Others", "Total", "Remarks"}
	widths := []float64{colWVal, colWVal, colWVal, colWVal, colWVal, colWRem}
	for i, h := range headers {
		drawCell(cx, curY, widths[i], rowH, h, "C", true, 248, 250, 252)
		cx += widths[i]
	}
	curY += rowH

	sewingStartY := 0.0
	sewingRows := 0

	strVal := func(n int64) string {
		return fmt.Sprint(n)
	}

	for _, r := range data.Rows {
		bg := false
		cr, cg, cb := 255, 255, 255
		pdf.SetFont(font, "", 8.0)
		pdf.SetTextColor(15, 23, 42)

		if r.StyleType == "subtotal" || r.StyleType == "staff_total" {
			bg = true
			cr, cg, cb = 224, 242, 254
			pdf.SetFont(font, "B", 8.0)
		} else if r.StyleType == "worker_total" {
			bg = true
			cr, cg, cb = 186, 230, 253
			pdf.SetFont(font, "B", 8.0)
		}

		if r.ParentSection == "Sewing" {
			if sewingStartY == 0 {
				sewingStartY = curY
			}
			sewingRows++
			drawCell(x+colWParent, curY, colWSec, rowH, r.Section, "L", bg, cr, cg, cb)
		} else {
			drawCell(x, curY, colWParent+colWSec, rowH, r.Section, "L", bg, cr, cg, cb)
		}

		cx = x + colWParent + colWSec
		drawCell(cx, curY, colWVal, rowH, strVal(r.Present), "C", bg, cr, cg, cb)
		cx += colWVal
		drawCell(cx, curY, colWVal, rowH, strVal(r.Absent), "C", bg, cr, cg, cb)
		cx += colWVal
		drawCell(cx, curY, colWVal, rowH, strVal(r.Leave), "C", bg, cr, cg, cb)
		cx += colWVal
		drawCell(cx, curY, colWVal, rowH, strVal(r.Others), "C", bg, cr, cg, cb)
		cx += colWVal
		drawCell(cx, curY, colWVal, rowH, strVal(r.Total), "C", bg, cr, cg, cb)
		cx += colWVal
		drawCell(cx, curY, colWRem, rowH, r.Remarks, "C", bg, cr, cg, cb)

		curY += rowH
	}

	// Draw Sewing Parent vertical box
	if sewingStartY > 0 && sewingRows > 0 {
		sewingH := float64(sewingRows) * rowH
		pdf.SetFillColor(255, 255, 255)
		pdf.Rect(x, sewingStartY, colWParent, sewingH, "D")
		pdf.SetFont(font, "B", 9.0)
		pdf.SetTextColor(15, 23, 42)
		pdf.SetXY(x, sewingStartY+sewingH/2.0-2.0)
		pdf.CellFormat(colWParent, 4.0, "Sewing", "", 0, "C", false, 0, "")
	}

	// Grand Total Row
	grandH := 7.5
	pdf.SetFont(font, "B", 8.5)
	pdf.SetTextColor(255, 255, 255)

	drawGrandCell := func(cx, cy, cw, ch float64, txt string) {
		pdf.SetDrawColor(203, 213, 225)
		pdf.SetFillColor(0, 160, 233)
		pdf.Rect(cx, cy, cw, ch, "DF")
		offsetY := (ch - 4.0) / 2.0
		pdf.SetXY(cx+0.5, cy+offsetY)
		pdf.CellFormat(cw-1.0, 4.0, txt, "", 0, "C", false, 0, "")
	}

	drawGrandCell(x, curY, colWParent+colWSec, grandH, data.GrandTotal.Section)
	cx = x + colWParent + colWSec
	drawGrandCell(cx, curY, colWVal, grandH, fmt.Sprint(data.GrandTotal.Present))
	cx += colWVal
	drawGrandCell(cx, curY, colWVal, grandH, fmt.Sprint(data.GrandTotal.Absent))
	cx += colWVal
	drawGrandCell(cx, curY, colWVal, grandH, fmt.Sprint(data.GrandTotal.Leave))
	cx += colWVal
	drawGrandCell(cx, curY, colWVal, grandH, fmt.Sprint(data.GrandTotal.Others))
	cx += colWVal
	drawGrandCell(cx, curY, colWVal, grandH, fmt.Sprint(data.GrandTotal.Total))
	cx += colWVal
	drawGrandCell(cx, curY, colWRem, grandH, "")

	filename := fmt.Sprintf("daily_summary_%s.pdf", dateStr)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}
