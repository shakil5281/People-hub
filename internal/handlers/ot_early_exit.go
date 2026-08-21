package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/xuri/excelize/v2"
)

// OtEarlyExitHandler exposes early-exit OT deduction processing and reporting.
type OtEarlyExitHandler struct {
	otRepo    *repository.OtEarlyExitRepository
	otService *service.OtEarlyExitService
}

func NewOtEarlyExitHandler(otRepo *repository.OtEarlyExitRepository, otService *service.OtEarlyExitService) *OtEarlyExitHandler {
	return &OtEarlyExitHandler{otRepo: otRepo, otService: otService}
}

type otEarlyExitProcessRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
	Month     int    `json:"month" binding:"required"`
	Year      int    `json:"year" binding:"required"`
}

// ComputeOtEarlyExit godoc
//
//	@Summary      Compute early-exit OT deductions
//	@Description  Recompute the monthly ledger of early-exit shortfall hours deducted from employee overtime
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body otEarlyExitProcessRequest true "Company, month, and year"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /attendance/ot-early-exit/process [post]
func (h *OtEarlyExitHandler) Compute(c *gin.Context) {
	var req otEarlyExitProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	result, err := h.otService.ComputeEarlyExitDeductions(req.CompanyID, req.Month, req.Year, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             fmt.Sprintf("Computed %d early-exit deductions for %d employees", result.TotalRecords, result.AffectedEmployees),
		"records":             result.TotalRecords,
		"affected_employees":  result.AffectedEmployees,
		"month":               result.Month,
		"year":                result.Year,
		"company_id":          result.CompanyID,
	})
}

// ListOtEarlyExit godoc
//
//	@Summary      List early-exit OT deductions
//	@Description  Get the paginated ledger of early-exit shortfall deductions
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id     query string true  "Company ID"
//	@Param        month          query int    true  "Month (1-12)"
//	@Param        year           query int    true  "Year"
//	@Param        department_id  query string false "Filter by department"
//	@Param        section_id     query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id        query string false "Filter by line"
//	@Param        group_id       query string false "Filter by group"
//	@Param        shift_id       query string false "Filter by shift"
//	@Param        employee_id    query string false "Search by employee ID (partial match)"
//	@Param        page           query int    false "Page number (default: 1)"
//	@Param        limit          query int    false "Page size (default: 20)"
//	@Success      200  {object}  utils.PaginatedResponse
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /attendance/ot-early-exit [get]
func (h *OtEarlyExitHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}
	if year == 0 {
		year = time.Now().Year()
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 20
	}

	records, total, err := h.otRepo.List(
		companyID, month, year,
		c.Query("department_id"), c.Query("section_id"), c.Query("designation_id"),
		c.Query("line_id"), c.Query("group_id"), c.Query("shift_id"), c.Query("employee_id"),
		page, limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Month totals for stat cards matching active filters
	totalShortfall, affected, _ := h.otRepo.ListStats(
		companyID, month, year,
		c.Query("department_id"), c.Query("section_id"), c.Query("designation_id"),
		c.Query("line_id"), c.Query("group_id"), c.Query("shift_id"), c.Query("employee_id"),
	)

	c.JSON(http.StatusOK, gin.H{
		"records":         records,
		"total":           total,
		"page":            page,
		"limit":           limit,
		"total_shortfall": round2(totalShortfall),
		"affected_employees": affected,
		"month":           month,
		"year":            year,
	})
}

// ExportOtEarlyExitExcel godoc
//
//	@Summary      Export early-exit overtime deductions to Excel
//	@Description  Generate and download Excel report of early-exit deductions
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        company_id     query string true  "Company ID"
//	@Param        month          query int    true  "Month (1-12)"
//	@Param        year           query int    true  "Year"
//	@Success      200  {file}  file
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /attendance/ot-early-exit/export/excel [get]
func (h *OtEarlyExitHandler) ExportExcel(c *gin.Context) {
	companyID := c.Query("company_id")
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}
	if year == 0 {
		year = time.Now().Year()
	}

	records, _, err := h.otRepo.List(
		companyID, month, year,
		c.Query("department_id"), c.Query("section_id"), c.Query("designation_id"),
		c.Query("line_id"), c.Query("group_id"), c.Query("shift_id"), c.Query("employee_id"),
		1, 100000,
	)
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

	monthLabel := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("January") + " " + strconv.Itoa(year)

	f, err := buildOtEarlyExitExcel(companyName, companyAddress, monthLabel, records)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=ot_early_exit_%s_%d_%d.xlsx", monthLabel, month, year))
	c.Writer.Write(buf.Bytes())
}

func buildOtEarlyExitExcel(companyName, companyAddress, monthLabel string, records []map[string]interface{}) (*excelize.File, error) {
	const nCols = 10

	cols := []struct {
		header string
		width  float64
	}{
		{"Sl", 5}, {"Emp. ID", 12}, {"Name", 26}, {"Designation", 24}, {"Department", 20},
		{"Date", 12}, {"Shift", 10}, {"Expected", 10}, {"Worked", 10}, {"Shortfall", 12},
	}

	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "OT Early Exit")

	thinBorder := []excelize.Border{
		{Type: "left", Color: "333333", Style: 1}, {Type: "top", Color: "333333", Style: 1},
		{Type: "bottom", Color: "333333", Style: 1}, {Type: "right", Color: "333333", Style: 1},
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	reportTitleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 13, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"4470AF"}, Pattern: 1}, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}})
	textLeftStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"}, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"}})
	numCenterStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"}, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	totalRowStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"D9E1F2"}, Pattern: 1}, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})

	endCol := colNameAttendance(nCols)

	for i, col := range cols {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth("OT Early Exit", colLetter, colLetter, col.width)
	}

	f.SetCellValue("OT Early Exit", "A1", companyName)
	f.MergeCell("OT Early Exit", "A1", endCol+"1")
	f.SetCellStyle("OT Early Exit", "A1", endCol+"1", titleStyle)
	f.SetRowHeight("OT Early Exit", 1, 30)

	f.SetCellValue("OT Early Exit", "A2", companyAddress)
	f.MergeCell("OT Early Exit", "A2", endCol+"2")
	f.SetCellStyle("OT Early Exit", "A2", endCol+"2", subtitleStyle)
	f.SetRowHeight("OT Early Exit", 2, 18)

	f.SetCellValue("OT Early Exit", "A3", "OT EARLY EXIT DEDUCTION - "+monthLabel)
	f.MergeCell("OT Early Exit", "A3", endCol+"3")
	f.SetCellStyle("OT Early Exit", "A3", endCol+"3", reportTitleStyle)
	f.SetRowHeight("OT Early Exit", 3, 20)

	headerRow := 4
	for i, col := range cols {
		axis, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		f.SetCellValue("OT Early Exit", axis, col.header)
		f.SetCellStyle("OT Early Exit", axis, axis, headerStyle)
	}
	f.SetRowHeight("OT Early Exit", headerRow, 22)

	var totalShortfall float64
	row := headerRow + 1
	for idx, rec := range records {
		empID := fmt.Sprintf("%v", rec["employee_id"])
		name := fmt.Sprintf("%v", rec["employee_name"])
		desig := fmt.Sprintf("%v", rec["designation"])
		dept := fmt.Sprintf("%v", rec["department"])
		date := fmt.Sprintf("%v", rec["date"])
		if len(date) > 10 {
			date = date[:10]
		}
		shift := fmt.Sprintf("%v", rec["shift_start"]) + " - " + fmt.Sprintf("%v", rec["shift_end"])
		expected := math.Round(numOrZero(rec["expected_hours"]))
		worked := math.Round(numOrZero(rec["worked_hours"]))
		shortfall := math.Round(numOrZero(rec["shortfall_hours"]))
		totalShortfall += shortfall

		values := []struct {
			val    interface{}
			center bool
		}{
			{idx + 1, true},
			{empID, false},
			{name, false},
			{desig, false},
			{dept, false},
			{date, true},
			{shift, true},
			{expected, true},
			{worked, true},
			{shortfall, true},
		}
		for j, v := range values {
			axis, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue("OT Early Exit", axis, v.val)
			if v.center {
				f.SetCellStyle("OT Early Exit", axis, axis, numCenterStyle)
			} else {
				f.SetCellStyle("OT Early Exit", axis, axis, textLeftStyle)
			}
		}
		f.SetRowHeight("OT Early Exit", row, 18)
		row++
	}

	f.SetCellValue("OT Early Exit", "A"+strconv.Itoa(row), "Total")
	f.SetCellStyle("OT Early Exit", "A"+strconv.Itoa(row), "A"+strconv.Itoa(row), totalRowStyle)
	f.SetCellValue("OT Early Exit", "J"+strconv.Itoa(row), totalShortfall)
	f.SetCellStyle("OT Early Exit", "J"+strconv.Itoa(row), "J"+strconv.Itoa(row), totalRowStyle)
	f.SetRowHeight("OT Early Exit", row, 18)

	f.SetPanes("OT Early Exit", &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      1,
		YSplit:      headerRow,
		TopLeftCell: "B5",
		ActivePane:  "bottomRight",
	})

	orientation := "landscape"
	pageSize := 9
	fitWidth := 1
	fitHeight := 0
	f.SetPageLayout("OT Early Exit", &excelize.PageLayoutOptions{Orientation: &orientation, Size: &pageSize, FitToWidth: &fitWidth, FitToHeight: &fitHeight})
	f.SetPageMargins("OT Early Exit", &excelize.PageLayoutMarginsOptions{Left: ptr(0.3), Right: ptr(0.3), Top: ptr(0.4), Bottom: ptr(0.4)})
	f.SetSheetView("OT Early Exit", -1, &excelize.ViewOptions{ShowGridLines: ptrBool(false)})

	return f, nil
}

// round2 rounds a float to two decimals for display.
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// numOrZero safely extracts a numeric value from an Excel cell source.
func numOrZero(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}
