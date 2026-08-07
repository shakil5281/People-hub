package handlers

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/xuri/excelize/v2"
)

type SalaryHandler struct {
	salaryService *service.SalaryService
	salaryRepo    *repository.SalaryRepository
}

func NewSalaryHandler(
	salaryService *service.SalaryService,
	salaryRepo *repository.SalaryRepository,
) *SalaryHandler {
	return &SalaryHandler{
		salaryService: salaryService,
		salaryRepo:    salaryRepo,
	}
}

type ProcessSalaryRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
	Month     int    `json:"month" binding:"required"`
	Year      int    `json:"year" binding:"required"`
}

// ProcessSalary godoc
//
// @Summary      Process monthly salary
// @Description  Calculate and generate salary for all active employees for a given month/year
// @Tags         Salary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body ProcessSalaryRequest true "Salary process params"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/process [post]
func (h *SalaryHandler) Process(c *gin.Context) {
	var req ProcessSalaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	result, err := h.salaryService.ProcessMonth(req.CompanyID, req.Month, req.Year, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Salary processed for %d employees", result.Processed),
		"processed": result.Processed,
		"total":     result.Total,
		"month":     req.Month,
		"year":      req.Year,
	})
}

// ListSalarySheet godoc
//
// @Summary      List salary sheet
// @Description  Get salary records for all employees for a given month/year
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        company_id     query string true  "Company ID"
// @Param        month          query int    true  "Month (1-12)"
// @Param        year           query int    true  "Year"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        shift_id       query string false "Filter by shift"
// @Param        employee_id    query string false "Search by employee ID (partial match)"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/sheet [get]
func (h *SalaryHandler) Sheet(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

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

	totals := map[string]float64{
		"basic_salary": 0, "house_rent": 0, "medical_allowance": 0,
		"transport_allowance": 0, "food_allowance": 0, "other_allowance": 0,
		"gross_salary": 0, "provident_fund": 0, "tax": 0,
		"absent_deduction": 0, "total_deductions": 0, "net_salary": 0,
		"overtime_hours": 0, "overtime_amount": 0, "attendance_bonus": 0,
		"total_days": 0, "present_days": 0, "absent_days": 0,
		"late_days": 0, "leave_days": 0, "holiday_days": 0, "weekend_days": 0,
	}
	for _, s := range salaries {
		totals["basic_salary"] += s.BasicSalary
		totals["house_rent"] += s.HouseRent
		totals["medical_allowance"] += s.MedicalAllowance
		totals["transport_allowance"] += s.TransportAllowance
		totals["food_allowance"] += s.FoodAllowance
		totals["other_allowance"] += s.OtherAllowance
		totals["gross_salary"] += s.GrossSalary
		totals["provident_fund"] += s.ProvidentFund
		totals["tax"] += s.Tax
		totals["absent_deduction"] += s.AbsentDeduction
		totals["total_deductions"] += s.TotalDeductions
		totals["net_salary"] += s.NetSalary
		totals["overtime_hours"] += s.OvertimeHours
		totals["overtime_amount"] += s.OvertimeAmount
		totals["attendance_bonus"] += s.AttendanceBonus
		totals["total_days"] += float64(s.TotalDays)
		totals["present_days"] += float64(s.PresentDays)
		totals["absent_days"] += float64(s.AbsentDays)
		totals["late_days"] += float64(s.LateDays)
		totals["leave_days"] += float64(s.LeaveDays)
		totals["holiday_days"] += float64(s.HolidayDays)
		totals["weekend_days"] += float64(s.WeekendDays)
	}

	c.JSON(http.StatusOK, gin.H{
		"salaries": salaries,
		"total":    len(salaries),
		"totals":   totals,
		"month":    month,
		"year":     year,
	})
}

// SheetExport godoc
//
// @Summary      Export salary sheet to Excel
// @Description  Download monthly salary sheet as Excel file
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        company_id     query string true  "Company ID"
// @Param        month          query int    true  "Month (1-12)"
// @Param        year           query int    true  "Year"
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
// @Router       /salary/sheet/export [get]


func companyAddressFor(lang string, c models.Company) string {
	if lang == "bn" && c.AddressBn != "" {
		return c.AddressBn
	}
	if c.AddressEn != "" {
		return c.AddressEn
	}
	return ""
}

func customOrderRank(name string) int {
	clean := strings.ToLower(strings.TrimSpace(name))
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, "_", "")

	switch {
	case strings.Contains(clean, "officestaff"):
		return 1
	case strings.Contains(clean, "productionstaff"):
		return 2
	case strings.Contains(clean, "line1") || strings.Contains(clean, "line01"):
		return 3
	case strings.Contains(clean, "line2") || strings.Contains(clean, "line02"):
		return 4
	case strings.Contains(clean, "line3") || strings.Contains(clean, "line03"):
		return 5
	case strings.Contains(clean, "line4") || strings.Contains(clean, "line04"):
		return 6
	case strings.Contains(clean, "line5") || strings.Contains(clean, "line05"):
		return 7
	case strings.Contains(clean, "line6") || strings.Contains(clean, "line06"):
		return 8
	case strings.Contains(clean, "line7") || strings.Contains(clean, "line07"):
		return 9
	case strings.Contains(clean, "cutting"):
		return 10
	case strings.Contains(clean, "finishing"):
		return 11
	case strings.Contains(clean, "quality"):
		return 12
	case strings.Contains(clean, "loader") || strings.Contains(clean, "cleaner") || strings.Contains(clean, "admin"):
		return 13
	default:
		return 100
	}
}

func (h *SalaryHandler) SheetExport(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")
	lang := c.DefaultQuery("lang", "en")
	mode := c.DefaultQuery("mode", "salary") // "master" (Staff/Worker tabs) or "salary"/"line" (Line-wise tabs)

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

	var company models.Company
	if len(salaries) > 0 {
		company = salaries[0].Company
	}
	compName := companyNameFor(lang, company)
	compAddr := companyAddressFor(lang, company)
	if compName == "-" || compName == "" {
		compName = "Company Name"
	}

	labels := getSalarySheetLabels(lang)
	monthLabel := monthName(month, lang)

	fontFamily := ""
	if lang == "bn" {
		fontFamily = "SutonnyMJ"
		labels = bijoyLabels(labels)
		monthLabel = toBijoy(monthLabel)
		compName = toBijoy(compName)
		compAddr = toBijoy(compAddr)
	}

	f := excelize.NewFile()
	defer f.Close()

	compNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Family: fontFamily, Color: "#1F2937"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	compAddrStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 9, Family: fontFamily, Color: "#4B5563"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	reportTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Family: fontFamily, Color: "#111827"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	lineHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 9, Family: fontFamily, Color: "#1F2937"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	cellBorder := []excelize.Border{
		{Type: "left", Color: "D9D9D9", Style: 1},
		{Type: "right", Color: "D9D9D9", Style: 1},
		{Type: "top", Color: "D9D9D9", Style: 1},
		{Type: "bottom", Color: "D9D9D9", Style: 1},
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Color: "#FFFFFF", Family: fontFamily},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		NumFmt:    3,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	leftDataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Color: "#006100", Family: fontFamily},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})

	headers := []string{
		labels.Sl, labels.EmployeeID, labels.Name, labels.Designation, labels.WorkingDays,
		labels.Present, labels.Absent, labels.Late, labels.Leave, labels.Holiday, labels.Weekend,
		labels.BasicSalary, labels.HouseRent, labels.Medical, labels.Transport, labels.Food, labels.Gross,
		labels.AbsentDed, labels.TotalDeduction,
		labels.OTHours, labels.OTRate, labels.OTAmount, labels.AttBonus, labels.NetSalary, labels.Signature,
	}

	isMoneyCol := func(j int) bool {
		return (j >= 11 && j <= 18) || (j >= 20 && j <= 23)
	}

	buildSheet := func(sheetName string, displayLineName string, list []models.Salary) {
		orientation := "landscape"
		paperSize := 5 // Legal Landscape
		fitToWidth := 1
		fitToHeight := 0
		f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitToWidth,
			FitToHeight: &fitToHeight,
		})

		margin := 0.25
		topMargin := 0.75
		f.SetPageMargins(sheetName, &excelize.PageLayoutMarginsOptions{
			Left:   &margin,
			Right:  &margin,
			Top:    &topMargin,
			Bottom: &topMargin,
		})

		showGrid := false
		_ = f.SetSheetView(sheetName, -1, &excelize.ViewOptions{
			ShowGridLines: &showGrid,
		})

		// Page setup custom footer for high office signatures
		_ = f.SetHeaderFooter(sheetName, &excelize.HeaderFooterOptions{
			OddFooter: "&LPrepared By&CAdmin (A.G.M)                     Asst. General Manager&RApproved By",
		})

		// Row 1: Company Name
		_ = f.MergeCell(sheetName, "A1", "Y1")
		f.SetCellValue(sheetName, "A1", compName)
		f.SetCellStyle(sheetName, "A1", "Y1", compNameStyle)
		f.SetRowHeight(sheetName, 1, 26)

		// Row 2: Company Address
		_ = f.MergeCell(sheetName, "A2", "Y2")
		f.SetCellValue(sheetName, "A2", compAddr)
		f.SetCellStyle(sheetName, "A2", "Y2", compAddrStyle)
		f.SetRowHeight(sheetName, 2, 20)

		// Row 3: Report Name
		reportNameText := fmt.Sprintf("%s - %s %d", labels.Title, monthLabel, year)
		if mode == "master" {
			reportNameText = fmt.Sprintf("Master Sheet - %s %d", monthLabel, year)
			if lang == "bn" {
				reportNameText = toBijoy(reportNameText)
			}
		}
		_ = f.MergeCell(sheetName, "A3", "Y3")
		f.SetCellValue(sheetName, "A3", reportNameText)
		f.SetCellStyle(sheetName, "A3", "Y3", reportTitleStyle)
		f.SetRowHeight(sheetName, 3, 22)

		// Row 4: Line Name (Left side)
		lineText := fmt.Sprintf("Line: %s", displayLineName)
		if lang == "bn" {
			lineText = toBijoy(fmt.Sprintf("লাইন: %s", displayLineName))
		}
		_ = f.MergeCell(sheetName, "A4", "D4")
		f.SetCellValue(sheetName, "A4", lineText)
		f.SetCellStyle(sheetName, "A4", "D4", lineHeaderStyle)
		f.SetRowHeight(sheetName, 4, 20)

		// Row 5: Column Headers
		for i, h := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s5", col), h)
			f.SetCellStyle(sheetName, fmt.Sprintf("%s5", col), fmt.Sprintf("%s5", col), headerStyle)
		}
		f.SetRowHeight(sheetName, 5, 32)

		// Row 6 onwards: Data Rows
		for i, s := range list {
			row := i + 6
			empID := s.Employee.EmployeeID
			name := bijoyText(employeeNameFor(lang, &s.Employee), lang)
			desig := bijoyText(designationName(s.Employee.DesignationRef, lang), lang)

			vals := []interface{}{
				i + 1, empID, name, desig,
				s.TotalDays, s.PresentDays, s.AbsentDays, s.LateDays, s.LeaveDays, s.HolidayDays, s.WeekendDays,
				s.BasicSalary, s.HouseRent, s.MedicalAllowance, s.TransportAllowance, s.FoodAllowance, s.GrossSalary,
				s.AbsentDeduction, s.TotalDeductions,
				s.OvertimeHours, s.OvertimeRate, s.OvertimeAmount, s.AttendanceBonus, s.NetSalary, "",
			}

			for j, v := range vals {
				col, _ := excelize.ColumnNumberToName(j + 1)
				cell := fmt.Sprintf("%s%d", col, row)
				f.SetCellValue(sheetName, cell, v)
				if j == 2 || j == 3 {
					f.SetCellStyle(sheetName, cell, cell, leftDataStyle)
				} else if isMoneyCol(j) {
					f.SetCellStyle(sheetName, cell, cell, moneyStyle)
				} else {
					f.SetCellStyle(sheetName, cell, cell, centerStyle)
				}
			}
			f.SetRowHeight(sheetName, row, 40)
		}

		// Exact column widths specified by user:
		// A=5, B=8, C=20, D=16, E=7, F..K=7, L..P=8, Q..S=8, T..W=6, X=9, Y=15
		fixedWidths := []float64{
			5.0,  // A: Sl
			8.0,  // B: Employee ID
			20.0, // C: Name
			16.0, // D: Designation
			7.0,  // E: Working Days
			7.0,  // F: Present
			7.0,  // G: Absent
			7.0,  // H: Late
			7.0,  // I: Leave
			7.0,  // J: Holiday
			7.0,  // K: Weekend
			8.0,  // L: Basic Salary
			8.0,  // M: House Rent
			8.0,  // N: Medical
			8.0,  // O: Transport
			8.0,  // P: Food
			8.0,  // Q: Gross Salary
			8.0,  // R: Absent Deduction
			8.0,  // S: Total Deduction
			6.0,  // T: OT Hours
			6.0,  // U: OT Rate
			6.0,  // V: OT Amount
			6.0,  // W: Att Bonus
			9.0,  // X: Net Salary
			15.0, // Y: Signature
		}

		for j := 0; j < len(headers); j++ {
			col, _ := excelize.ColumnNumberToName(j + 1)
			w := fixedWidths[j]
			f.SetColWidth(sheetName, col, col, w)
		}

		totalRow := len(list) + 6
		if len(list) > 0 {
			for j := 0; j < len(headers); j++ {
				col, _ := excelize.ColumnNumberToName(j + 1)
				cell := fmt.Sprintf("%s%d", col, totalRow)
				if j <= 1 {
					if j == 0 {
						f.SetCellValue(sheetName, cell, "")
					} else {
						f.SetCellValue(sheetName, cell, labels.TotalLabel)
					}
					f.SetCellStyle(sheetName, cell, cell, totalStyle)
				} else if isMoneyCol(j) {
					firstDataRow := 6
					lastDataRow := totalRow - 1
					formula := fmt.Sprintf("=SUM(%s%d:%s%d)", col, firstDataRow, col, lastDataRow)
					f.SetCellFormula(sheetName, cell, formula)
					f.SetCellStyle(sheetName, cell, cell, moneyStyle)
				} else {
					f.SetCellValue(sheetName, cell, "")
					f.SetCellStyle(sheetName, cell, cell, totalStyle)
				}
			}
			f.SetRowHeight(sheetName, totalRow, 30)
		}
	}

	if mode == "master" {
		var staffList []models.Salary
		var workerList []models.Salary

		for _, s := range salaries {
			grpName := ""
			if s.Employee.GroupRef != nil {
				grpName = strings.ToLower(s.Employee.GroupRef.Name)
			}
			empType := strings.ToLower(s.Employee.EmployeeType)

			if strings.Contains(grpName, "worker") || strings.Contains(empType, "worker") {
				workerList = append(workerList, s)
			} else {
				staffList = append(staffList, s)
			}
		}

		createdCount := 0
		if len(staffList) > 0 {
			f.SetSheetName("Sheet1", "Staff")
			buildSheet("Staff", "Staff", staffList)
			createdCount++
		}
		if len(workerList) > 0 {
			if createdCount == 0 {
				f.SetSheetName("Sheet1", "Worker")
			} else {
				f.NewSheet("Worker")
			}
			buildSheet("Worker", "Worker", workerList)
			createdCount++
		}
		if createdCount == 0 {
			f.SetSheetName("Sheet1", "Master Sheet")
			buildSheet("Master Sheet", "All", salaries)
		}

	} else {
		// mode == "salary" or "line" -> Custom Office Staff, Production Staff & Line-wise Worker sheets
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
				// Line-wise Worker
				lineName := "No Line"
				if s.Employee.LineRef != nil && strings.TrimSpace(s.Employee.LineRef.Name) != "" {
					lineName = strings.TrimSpace(s.Employee.LineRef.Name)
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

		createdSheetCount := 0

		addSalarySheet := func(rawSheetName, displayLine string, list []models.Salary) {
			if len(list) == 0 {
				return
			}
			sheetName := sanitizeSheetName(rawSheetName)
			if createdSheetCount == 0 {
				f.SetSheetName("Sheet1", sheetName)
			} else {
				f.NewSheet(sheetName)
			}
			buildSheet(sheetName, displayLine, list)
			createdSheetCount++
		}

		// 1. Office Staff sheet (Group: Staff, Dept: Admin/Office/Management)
		addSalarySheet("Office Staff", "Office Staff", officeStaffList)

		// 2. Production Staff sheet (Group: Staff, Dept: Production + Maintenance)
		addSalarySheet("Production Staff", "Production Staff", productionStaffList)

		// 3. Line-wise Worker sheets
		for _, lineName := range lineNames {
			displayLineName := lineName
			sheetTitle := lineName
			if strings.EqualFold(strings.TrimSpace(lineName), "admin") {
				displayLineName = "Loader & Cleaner"
				sheetTitle = "Loader & Cleaner"
			}
			addSalarySheet(sheetTitle, displayLineName, lineWorkerMap[lineName])
		}

		// Fallback if no specific list was populated
		if createdSheetCount == 0 {
			f.SetSheetName("Sheet1", "Salary Sheet")
			buildSheet("Salary Sheet", "All Lines", salaries)
		}
	}

	showGrid := false
	for _, sheet := range f.GetSheetList() {
		_ = f.SetSheetView(sheet, -1, &excelize.ViewOptions{
			ShowGridLines: &showGrid,
		})
	}

	firstSheet := f.GetSheetList()[0]
	sheetIdx, _ := f.GetSheetIndex(firstSheet)
	f.SetActiveSheet(sheetIdx)

	prefix := "salary_sheet"
	if mode == "master" {
		prefix = "master_sheet"
	}

	langSuffix := ""
	if lang == "bn" {
		langSuffix = "_bn"
	}
	filename := fmt.Sprintf("%s_%s_%d%s.xlsx", prefix, monthLabel, year, langSuffix)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// GetPayslip godoc
//
// @Summary      Get employee payslip(s)
// @Description  Get salary record for a specific employee or paginated list with filters
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        employee_id    query string false "Employee ID (omit for paginated list)"
// @Param        company_id     query string false "Company ID (required for paginated list)"
// @Param        month          query int    true  "Month (1-12)"
// @Param        year           query int    true  "Year"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        shift_id       query string false "Filter by shift"
// @Param        page           query int    false "Page number (default 1)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/payslip [get]
func (h *SalaryHandler) Payslip(c *gin.Context) {
	employeeID := c.Query("employee_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month and year are required"})
		return
	}

	if employeeID != "" {
		salary, err := h.salaryRepo.FindByEmployeeMonth(employeeID, month, year)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Salary not found for this employee/month"})
			return
		}
		c.JSON(http.StatusOK, salary)
		return
	}

	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required for paginated payslip list"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 4

	salaries, total, err := h.salaryRepo.ListPayslips(repository.SalaryFilter{
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
	}, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"salaries":    salaries,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
		"month":       month,
		"year":        year,
	})
}

func mustColumnName(n int) string {
	s, err := excelize.ColumnNumberToName(n)
	if err != nil {
		return "Z"
	}
	return s
}

// SummaryExport godoc
//
// @Summary      Export salary summary to Excel
// @Description  Download monthly salary summary as Excel file grouped by department/section/designation/line
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        company_id     query string true  "Company ID"
// @Param        month          query int    true  "Month (1-12)"
// @Param        year           query int    true  "Year"
// @Param        group_by       query string false "Group by: department|section|designation|line"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/summary/export [get]
func (h *SalaryHandler) SummaryExport(c *gin.Context) {
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

	var company models.Company
	if len(salaries) > 0 {
		company = salaries[0].Company
	}
	compName := companyNameFor(lang, company)
	compAddr := companyAddressFor(lang, company)
	if compName == "-" || compName == "" {
		compName = "Company Name"
	}

	fontFamily := ""
	if lang == "bn" {
		fontFamily = "SutonnyMJ"
		labels = bijoyLabels(labels)
		monthLabel = toBijoy(monthLabel)
		compName = toBijoy(compName)
		compAddr = toBijoy(compAddr)
	}

	f := excelize.NewFile()
	defer f.Close()

	compNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Family: fontFamily, Color: "#1F2937"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	compAddrStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 9, Family: fontFamily, Color: "#4B5563"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	reportTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Family: fontFamily, Color: "#111827"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	cellBorder := []excelize.Border{
		{Type: "left", Color: "D9D9D9", Style: 1},
		{Type: "right", Color: "D9D9D9", Style: 1},
		{Type: "top", Color: "D9D9D9", Style: 1},
		{Type: "bottom", Color: "D9D9D9", Style: 1},
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Color: "#FFFFFF", Family: fontFamily},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		NumFmt:    3,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Color: "#006100", Family: fontFamily},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})
	totalMoneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Color: "#006100", Family: fontFamily},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
		NumFmt:    3,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    cellBorder,
	})

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
		var keys []groupKey

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
				keys = append(keys, key)
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
		var keys []groupKey

		addRecord := func(kName string, s models.Salary) {
			key := groupKey{Name: kName}
			if gMap[key] == nil {
				gMap[key] = &groupData{}
				keys = append(keys, key)
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

	summaryTabs := []struct {
		groupMode    string
		groupLabel   string
		sheetName    string
		subtitleText string
	}{
		{"department", labels.Department, "Department Wise", "Summary By Department"},
		{"section", labels.Section, "Section Wise", "Summary By Section"},
		{"designation", labels.Designation, "Designation Wise", "Summary By Designation"},
		{"line", labels.GroupLabel, "Line Wise", "Summary By Line"},
		{"custom", "Category / Line", "Custom Summary", "Custom Summary Report"},
	}

	for tabIdx, tab := range summaryTabs {
		sheetName := tab.sheetName
		if tabIdx == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}

		orientation := "landscape"
		paperSize := 9 // A4 Landscape
		fitToWidth := 1
		fitToHeight := 0
		f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitToWidth,
			FitToHeight: &fitToHeight,
		})

		margin := 0.25
		topMargin := 0.75
		f.SetPageMargins(sheetName, &excelize.PageLayoutMarginsOptions{
			Left:   &margin,
			Right:  &margin,
			Top:    &topMargin,
			Bottom: &topMargin,
		})

		showGrid := false
		_ = f.SetSheetView(sheetName, -1, &excelize.ViewOptions{
			ShowGridLines: &showGrid,
		})

		_ = f.SetHeaderFooter(sheetName, &excelize.HeaderFooterOptions{
			OddFooter: "&LPrepared By&CAdmin (A.G.M)                     Asst. General Manager&RApproved By",
		})

		// Row 1: Company Name
		_ = f.MergeCell(sheetName, "A1", "J1")
		f.SetCellValue(sheetName, "A1", compName)
		f.SetCellStyle(sheetName, "A1", "J1", compNameStyle)
		f.SetRowHeight(sheetName, 1, 26)

		// Row 2: Company Address
		_ = f.MergeCell(sheetName, "A2", "J2")
		f.SetCellValue(sheetName, "A2", compAddr)
		f.SetCellStyle(sheetName, "A2", "J2", compAddrStyle)
		f.SetRowHeight(sheetName, 2, 20)

		// Row 3: Report Name
		reportNameText := fmt.Sprintf("Salary Summary Report - %s %d", monthLabel, year)
		if lang == "bn" {
			reportNameText = toBijoy(reportNameText)
		}
		_ = f.MergeCell(sheetName, "A3", "J3")
		f.SetCellValue(sheetName, "A3", reportNameText)
		f.SetCellStyle(sheetName, "A3", "J3", reportTitleStyle)
		f.SetRowHeight(sheetName, 3, 22)

		// Row 4: Column Headers
		headers := []string{labels.Sl, tab.groupLabel, labels.Employees, labels.BasicTotal, labels.HouseRentH, labels.MedicalH, labels.TransportH, labels.GrossTotal, labels.Deductions, labels.NetTotal}
		for i, h := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s4", col), h)
			f.SetCellStyle(sheetName, fmt.Sprintf("%s4", col), fmt.Sprintf("%s4", col), headerStyle)
		}
		f.SetRowHeight(sheetName, 4, 32)

		var keys []groupKey
		var gMap map[groupKey]*groupData
		if tab.groupMode == "custom" {
			keys, gMap = buildCustomSummaryData()
		} else {
			keys, gMap = buildGroupData(tab.groupMode)
		}

		// Row 5 onwards: Data Rows
		var grandEmployees int
		var grandBasic, grandHouse, grandMedical, grandTransport, grandGross, grandDeductions, grandNet float64

		for i, key := range keys {
			row := i + 5
			d := gMap[key]

			vals := []interface{}{
				i + 1, bijoyText(key.Name, lang), d.Employees,
				d.BasicSalary, d.HouseRent, d.Medical, d.Transport, d.GrossSalary, d.Deductions, d.NetSalary,
			}

			for j, v := range vals {
				col, _ := excelize.ColumnNumberToName(j + 1)
				cell := fmt.Sprintf("%s%d", col, row)
				f.SetCellValue(sheetName, cell, v)
				if j == 0 || j == 2 {
					f.SetCellStyle(sheetName, cell, cell, centerStyle)
				} else if j == 1 {
					f.SetCellStyle(sheetName, cell, cell, dataStyle)
				} else {
					f.SetCellStyle(sheetName, cell, cell, moneyStyle)
				}
			}
			f.SetRowHeight(sheetName, row, 28)

			grandEmployees += d.Employees
			grandBasic += d.BasicSalary
			grandHouse += d.HouseRent
			grandMedical += d.Medical
			grandTransport += d.Transport
			grandGross += d.GrossSalary
			grandDeductions += d.Deductions
			grandNet += d.NetSalary
		}

		totalRow := len(keys) + 5
		if len(keys) > 0 {
			totalVals := []interface{}{
				"", labels.GrandTotal, grandEmployees,
				grandBasic, grandHouse, grandMedical, grandTransport, grandGross, grandDeductions, grandNet,
			}
			for j, v := range totalVals {
				col, _ := excelize.ColumnNumberToName(j + 1)
				cell := fmt.Sprintf("%s%d", col, totalRow)
				f.SetCellValue(sheetName, cell, v)
				if j >= 3 {
					f.SetCellStyle(sheetName, cell, cell, totalMoneyStyle)
				} else {
					f.SetCellStyle(sheetName, cell, cell, totalStyle)
				}
			}
			f.SetRowHeight(sheetName, totalRow, 30)
		}

		// Calculate 100% dynamic column widths
		for j := 0; j < len(headers); j++ {
			col, _ := excelize.ColumnNumberToName(j + 1)
			hText := headers[j]
			maxLen := utf8.RuneCountInString(hText)

			words := strings.Split(hText, " ")
			longestWord := 0
			for _, wStr := range words {
				wLen := utf8.RuneCountInString(wStr)
				if wLen > longestWord {
					longestWord = wLen
				}
			}

			for _, key := range keys {
				d := gMap[key]
				vals := []interface{}{
					0, bijoyText(key.Name, lang), d.Employees,
					d.BasicSalary, d.HouseRent, d.Medical, d.Transport, d.GrossSalary, d.Deductions, d.NetSalary,
				}
				sVal := fmt.Sprintf("%v", vals[j])
				if j >= 3 && vals[j] != "" {
					if fVal, ok := vals[j].(float64); ok {
						sVal = fmt.Sprintf("%.2f", fVal)
						absVal := math.Abs(fVal)
						if absVal >= 1000 {
							sVal += ","
						}
						if absVal >= 100000 {
							sVal += ","
						}
					}
				}
				vLen := utf8.RuneCountInString(sVal)
				if vLen > maxLen {
					maxLen = vLen
				}
			}

			w := float64(maxLen) + 2.5
			minHeaderW := float64(longestWord) + 2.5
			if w < minHeaderW {
				w = minHeaderW
			}

			if j >= 3 {
				if w < 11.0 {
					w = 11.0
				}
			} else if j == 1 {
				if w < 16.0 {
					w = 16.0
				}
			} else {
				if w < 6.0 {
					w = 6.0
				}
			}

			f.SetColWidth(sheetName, col, col, w)
		}
	}

	firstSheet := f.GetSheetList()[0]
	sheetIdx, _ := f.GetSheetIndex(firstSheet)
	f.SetActiveSheet(sheetIdx)

	langSuffix := ""
	if lang == "bn" {
		langSuffix = "_bn"
	}
	filename := fmt.Sprintf("salary_summary_%s_%d%s.xlsx", monthLabel, year, langSuffix)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// ListSalaries godoc
//
// @Summary      List salary records
// @Description  Get salary records summary (grouped by department)
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        company_id query string true  "Company ID"
// @Param        month      query int    true  "Month (1-12)"
// @Param        year       query int    true  "Year"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/list [get]
func (h *SalaryHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id, month, and year are required"})
		return
	}

	salaries, err := h.salaryRepo.ListAllByMonth(companyID, month, year, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	deptSummary := make(map[string]map[string]interface{})
	for _, s := range salaries {
		deptName := ""
		if s.Employee.Department != nil {
			deptName = s.Employee.Department.Name
		}
		if deptName == "" {
			deptName = "Unknown"
		}
		if _, ok := deptSummary[deptName]; !ok {
			deptSummary[deptName] = map[string]interface{}{
				"department":       deptName,
				"employees":        0,
				"basic_salary":     0.0,
				"house_rent":       0.0,
				"medical":          0.0,
				"transport":        0.0,
				"gross_salary":     0.0,
				"total_deductions": 0.0,
				"net_salary":       0.0,
			}
		}
		ds := deptSummary[deptName]
		ds["employees"] = ds["employees"].(int) + 1
		ds["basic_salary"] = ds["basic_salary"].(float64) + s.BasicSalary
		ds["house_rent"] = ds["house_rent"].(float64) + s.HouseRent
		ds["medical"] = ds["medical"].(float64) + s.MedicalAllowance
		ds["transport"] = ds["transport"].(float64) + s.TransportAllowance
		ds["gross_salary"] = ds["gross_salary"].(float64) + s.GrossSalary
		ds["total_deductions"] = ds["total_deductions"].(float64) + s.TotalDeductions
		ds["net_salary"] = ds["net_salary"].(float64) + s.NetSalary
	}

	var summaries []map[string]interface{}
	for _, v := range deptSummary {
		summaries = append(summaries, v)
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries": summaries,
		"total":     len(summaries),
	})
}

// SalarySummary godoc
//
// @Summary      Salary summary by department/section/designation/line
// @Description  Get salary summary grouped by specified level with grand totals
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        company_id     query string true  "Company ID"
// @Param        month          query int    true  "Month (1-12)"
// @Param        year           query int    true  "Year"
// @Param        group_by       query string false "Group by: department|section|designation|line (default: department)"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/summary [get]
func (h *SalaryHandler) Summary(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id, month, and year are required"})
		return
	}

	groupBy := c.DefaultQuery("group_by", "department")

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

	type groupKey struct {
		Name string
		ID   string
	}
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
	totalEmployees := 0
	var grandTotals = map[string]float64{
		"basic_salary": 0, "house_rent": 0, "medical": 0,
		"transport": 0, "gross_salary": 0, "deductions": 0, "net_salary": 0,
	}

	for _, s := range salaries {
		var key groupKey
		switch groupBy {
		case "section":
			if s.Employee.SectionRef != nil {
				key = groupKey{Name: s.Employee.SectionRef.Name, ID: s.Employee.SectionRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		case "designation":
			if s.Employee.DesignationRef != nil {
				key = groupKey{Name: s.Employee.DesignationRef.Name, ID: s.Employee.DesignationRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		case "line":
			if s.Employee.LineRef != nil {
				lName := s.Employee.LineRef.Name
				if strings.EqualFold(strings.TrimSpace(lName), "admin") {
					lName = "Loader & Cleaner"
				}
				key = groupKey{Name: lName, ID: s.Employee.LineRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		case "custom":
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
					key = groupKey{Name: "Production Staff", ID: "production_staff"}
				} else {
					key = groupKey{Name: "Office Staff", ID: "office_staff"}
				}
			} else {
				lName := "No Line"
				if s.Employee.LineRef != nil && strings.TrimSpace(s.Employee.LineRef.Name) != "" {
					lName = strings.TrimSpace(s.Employee.LineRef.Name)
				}
				if strings.EqualFold(lName, "admin") {
					lName = "Loader & Cleaner"
				}
				key = groupKey{Name: lName, ID: lName}
			}
		default: // department
			if s.Employee.Department != nil {
				key = groupKey{Name: s.Employee.Department.Name, ID: s.Employee.Department.ID}
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

	var keys []groupKey
	for k := range groupMap {
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

	var summaries []map[string]interface{}
	for _, key := range keys {
		d := groupMap[key]
		summaries = append(summaries, map[string]interface{}{
			"group_key":    key.Name,
			"group_id":     key.ID,
			"employees":    d.Employees,
			"basic_salary": d.BasicSalary,
			"house_rent":   d.HouseRent,
			"medical":      d.Medical,
			"transport":    d.Transport,
			"gross_salary": d.GrossSalary,
			"deductions":   d.Deductions,
			"net_salary":   d.NetSalary,
		})
		totalEmployees += d.Employees
		grandTotals["basic_salary"] += d.BasicSalary
		grandTotals["house_rent"] += d.HouseRent
		grandTotals["medical"] += d.Medical
		grandTotals["transport"] += d.Transport
		grandTotals["gross_salary"] += d.GrossSalary
		grandTotals["deductions"] += d.Deductions
		grandTotals["net_salary"] += d.NetSalary
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries":       summaries,
		"total":           len(summaries),
		"total_employees": totalEmployees,
		"grand_totals":    grandTotals,
	})
}

// DailySheet godoc
//
// @Summary      Daily salary sheet
// @Description  Get daily salary calculations for a specific date
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        date           query string true  "Date (YYYY-MM-DD)"
// @Param        company_id     query string false "Company ID (defaults to JWT company)"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/daily-sheet [get]
func (h *SalaryHandler) DailySheet(c *gin.Context) {
	date := c.Query("date")
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}

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

	var totals = map[string]float64{
		"employees":    0,
		"gross_salary": 0,
		"daily_rate":   0,
		"ot_hours":     0,
		"ot_amount":    0,
		"total_pay":    0,
	}
	for _, r := range records {
		totals["employees"]++
		totals["gross_salary"] += r.GrossSalary
		totals["daily_rate"] += r.DailyRate
		totals["ot_hours"] += r.OtHours
		totals["ot_amount"] += r.OtAmount
		totals["total_pay"] += r.TotalPay
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   len(records),
		"totals":  totals,
		"date":    date,
	})
}

// DailySheetExport godoc
//
// @Summary      Export Daily Salary Sheet to Excel
// @Description  Download daily salary sheet calculation as Excel file
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        date           query string true  "Date (YYYY-MM-DD)"
// @Param        company_id     query string false "Company ID"
// @Param        lang           query string false "Language: en or bn (default en)"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/daily-sheet/export [get]
func (h *SalaryHandler) DailySheetExport(c *gin.Context) {
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

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Daily Salary Sheet"
	f.SetSheetName("Sheet1", sheetName)

	orientation := "landscape"
	paperSize := 9 // A4 Landscape
	fitToWidth := 1
	fitToHeight := 0
	f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
		Orientation: &orientation,
		Size:        &paperSize,
		FitToWidth:  &fitToWidth,
		FitToHeight: &fitToHeight,
	})

	showGrid := false
	_ = f.SetSheetView(sheetName, -1, &excelize.ViewOptions{
		ShowGridLines: &showGrid,
	})

	_ = f.SetHeaderFooter(sheetName, &excelize.HeaderFooterOptions{
		OddFooter: "&LPrepared By&CAdmin (A.G.M)                     Asst. General Manager&RApproved By",
	})

	fontFamily := "Calibri"
	if lang == "bn" {
		fontFamily = "SutonnyMJ"
	}

	compNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Family: fontFamily, Color: "#0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	compAddrStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 9, Family: fontFamily, Color: "#64748B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	reportTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Family: fontFamily, Color: "#111827"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Family: fontFamily, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		NumFmt:    3,
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Family: fontFamily, Color: "#006100"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E2EFDA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	totalMoneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Family: fontFamily, Color: "#006100"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E2EFDA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		NumFmt:    3,
	})

	_ = f.MergeCell(sheetName, "A1", "N1")
	f.SetCellValue(sheetName, "A1", compName)
	f.SetCellStyle(sheetName, "A1", "N1", compNameStyle)
	f.SetRowHeight(sheetName, 1, 26)

	_ = f.MergeCell(sheetName, "A2", "N2")
	f.SetCellValue(sheetName, "A2", compAddr)
	f.SetCellStyle(sheetName, "A2", "N2", compAddrStyle)
	f.SetRowHeight(sheetName, 2, 20)

	_ = f.MergeCell(sheetName, "A3", "N3")
	title := fmt.Sprintf("Daily Salary Sheet - %s", date)
	f.SetCellValue(sheetName, "A3", title)
	f.SetCellStyle(sheetName, "A3", "N3", reportTitleStyle)
	f.SetRowHeight(sheetName, 3, 22)

	headers := []string{"SL", "Code", "Employee Name", "Designation", "Department", "Status", "In", "Out", "Hours", "OT Hours", "Gross Salary", "Daily Rate", "OT Amount", "Total Pay"}
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheetName, fmt.Sprintf("%s4", col), h)
		f.SetCellStyle(sheetName, fmt.Sprintf("%s4", col), fmt.Sprintf("%s4", col), headerStyle)
	}
	f.SetRowHeight(sheetName, 4, 32)

	var totGross, totDaily, totOTHours, totOTAmt, totPay float64

	for i, r := range records {
		row := i + 5
		vals := []interface{}{
			i + 1, r.EmployeeID, bijoyText(r.EmployeeName, lang), bijoyText(r.Designation, lang), bijoyText(r.DepartmentName, lang),
			r.Status, r.CheckIn, r.CheckOut, r.TotalHours, r.OtHours,
			r.GrossSalary, r.DailyRate, r.OtAmount, r.TotalPay,
		}

		for j, v := range vals {
			col, _ := excelize.ColumnNumberToName(j + 1)
			cell := fmt.Sprintf("%s%d", col, row)
			f.SetCellValue(sheetName, cell, v)
			if j == 2 || j == 3 || j == 4 {
				f.SetCellStyle(sheetName, cell, cell, dataStyle)
			} else if j >= 10 {
				f.SetCellStyle(sheetName, cell, cell, moneyStyle)
			} else {
				f.SetCellStyle(sheetName, cell, cell, centerStyle)
			}
		}
		f.SetRowHeight(sheetName, row, 26)

		totGross += r.GrossSalary
		totDaily += r.DailyRate
		totOTHours += r.OtHours
		totOTAmt += r.OtAmount
		totPay += r.TotalPay
	}

	totalRow := len(records) + 5
	if len(records) > 0 {
		totalVals := []interface{}{
			"", "", "Total", "", "", "", "", "", "", totOTHours,
			totGross, totDaily, totOTAmt, totPay,
		}
		for j, v := range totalVals {
			col, _ := excelize.ColumnNumberToName(j + 1)
			cell := fmt.Sprintf("%s%d", col, totalRow)
			f.SetCellValue(sheetName, cell, v)
			if j >= 10 {
				f.SetCellStyle(sheetName, cell, cell, totalMoneyStyle)
			} else {
				f.SetCellStyle(sheetName, cell, cell, totalStyle)
			}
		}
		f.SetRowHeight(sheetName, totalRow, 28)
	}

	filename := fmt.Sprintf("daily_salary_sheet_%s.xlsx", date)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// DailySummary godoc
//
// @Summary      Daily salary sheet summary
// @Description  Get daily salary calculations summary grouped by department/section/designation/line/custom
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        date           query string true  "Date (YYYY-MM-DD)"
// @Param        company_id     query string false "Company ID"
// @Param        group_by       query string false "Group by: department|section|designation|line|custom"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/daily-summary [get]
func (h *SalaryHandler) DailySummary(c *gin.Context) {
	date := c.Query("date")
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	groupBy := c.DefaultQuery("group_by", "department")

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

	type groupData struct {
		Name        string  `json:"group_key"`
		Employees   int     `json:"employees"`
		GrossSalary float64 `json:"gross_salary"`
		DailyRate   float64 `json:"daily_rate"`
		OtHours     float64 `json:"ot_hours"`
		OtAmount    float64 `json:"ot_amount"`
		TotalPay    float64 `json:"total_pay"`
	}

	gMap := make(map[string]*groupData)

	for _, r := range records {
		keyName := r.DepartmentName
		if groupBy == "designation" {
			keyName = r.Designation
		} else if groupBy == "section" || groupBy == "line" || groupBy == "custom" {
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

	var summaries []*groupData
	var totals = map[string]float64{
		"employees": 0, "gross_salary": 0, "daily_rate": 0,
		"ot_hours": 0, "ot_amount": 0, "total_pay": 0,
	}

	for _, k := range keys {
		d := gMap[k]
		summaries = append(summaries, d)
		totals["employees"] += float64(d.Employees)
		totals["gross_salary"] += d.GrossSalary
		totals["daily_rate"] += d.DailyRate
		totals["ot_hours"] += d.OtHours
		totals["ot_amount"] += d.OtAmount
		totals["total_pay"] += d.TotalPay
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries": summaries,
		"total":     len(summaries),
		"totals":    totals,
		"date":      date,
	})
}

// DailySummaryExport godoc
//
// @Summary      Export Daily Salary Summary to Excel
// @Description  Download daily salary summary as Excel file
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        date           query string true  "Date (YYYY-MM-DD)"
// @Param        company_id     query string false "Company ID"
// @Param        lang           query string false "Language: en or bn (default en)"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/daily-summary/export [get]
func (h *SalaryHandler) DailySummaryExport(c *gin.Context) {
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

	f := excelize.NewFile()
	defer f.Close()

	fontFamily := "Calibri"
	if lang == "bn" {
		fontFamily = "SutonnyMJ"
	}

	compNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Family: fontFamily, Color: "#0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	compAddrStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 9, Family: fontFamily, Color: "#64748B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	reportTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Family: fontFamily, Color: "#111827"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Family: fontFamily, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 8, Family: fontFamily},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		NumFmt:    3,
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Family: fontFamily, Color: "#006100"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E2EFDA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	totalMoneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 8, Family: fontFamily, Color: "#006100"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E2EFDA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		NumFmt:    3,
	})

	summaryTabs := []struct {
		groupMode  string
		groupLabel string
		sheetName  string
	}{
		{"department", labels.Department, "Department Wise"},
		{"section", labels.Section, "Section Wise"},
		{"designation", labels.Designation, "Designation Wise"},
		{"line", labels.GroupLabel, "Line Wise"},
		{"custom", "Category / Line", "Custom Summary"},
	}

	for tabIdx, tab := range summaryTabs {
		sheetName := tab.sheetName
		if tabIdx == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}

		orientation := "landscape"
		paperSize := 9 // A4 Landscape
		fitToWidth := 1
		fitToHeight := 0
		f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitToWidth,
			FitToHeight: &fitToHeight,
		})

		showGrid := false
		_ = f.SetSheetView(sheetName, -1, &excelize.ViewOptions{
			ShowGridLines: &showGrid,
		})

		_ = f.SetHeaderFooter(sheetName, &excelize.HeaderFooterOptions{
			OddFooter: "&LPrepared By&CAdmin (A.G.M)                     Asst. General Manager&RApproved By",
		})

		_ = f.MergeCell(sheetName, "A1", "H1")
		f.SetCellValue(sheetName, "A1", compName)
		f.SetCellStyle(sheetName, "A1", "H1", compNameStyle)
		f.SetRowHeight(sheetName, 1, 26)

		_ = f.MergeCell(sheetName, "A2", "H2")
		f.SetCellValue(sheetName, "A2", compAddr)
		f.SetCellStyle(sheetName, "A2", "H2", compAddrStyle)
		f.SetRowHeight(sheetName, 2, 20)

		_ = f.MergeCell(sheetName, "A3", "H3")
		reportTitleText := fmt.Sprintf("Daily Salary Summary Report - %s", date)
		if lang == "bn" {
			reportTitleText = toBijoy(reportTitleText)
		}
		f.SetCellValue(sheetName, "A3", reportTitleText)
		f.SetCellStyle(sheetName, "A3", "H3", reportTitleStyle)
		f.SetRowHeight(sheetName, 3, 22)

		headers := []string{labels.Sl, tab.groupLabel, labels.Employees, "Gross Salary", "Daily Rate", "OT Hours", "OT Amount", "Total Pay"}
		for i, h := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s4", col), h)
			f.SetCellStyle(sheetName, fmt.Sprintf("%s4", col), fmt.Sprintf("%s4", col), headerStyle)
		}
		f.SetRowHeight(sheetName, 4, 32)

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

		var grandEmp int
		var grandGross, grandDaily, grandOTHours, grandOTAmt, grandPay float64

		for i, k := range keys {
			row := i + 5
			d := gMap[k]
			vals := []interface{}{
				i + 1, bijoyText(d.Name, lang), d.Employees,
				d.GrossSalary, d.DailyRate, d.OtHours, d.OtAmount, d.TotalPay,
			}
			for j, v := range vals {
				col, _ := excelize.ColumnNumberToName(j + 1)
				cell := fmt.Sprintf("%s%d", col, row)
				f.SetCellValue(sheetName, cell, v)
				if j == 0 || j == 2 {
					f.SetCellStyle(sheetName, cell, cell, centerStyle)
				} else if j == 1 {
					f.SetCellStyle(sheetName, cell, cell, dataStyle)
				} else {
					f.SetCellStyle(sheetName, cell, cell, moneyStyle)
				}
			}
			f.SetRowHeight(sheetName, row, 26)

			grandEmp += d.Employees
			grandGross += d.GrossSalary
			grandDaily += d.DailyRate
			grandOTHours += d.OtHours
			grandOTAmt += d.OtAmount
			grandPay += d.TotalPay
		}

		totalRow := len(keys) + 5
		if len(keys) > 0 {
			totalVals := []interface{}{
				"", labels.GrandTotal, grandEmp,
				grandGross, grandDaily, grandOTHours, grandOTAmt, grandPay,
			}
			for j, v := range totalVals {
				col, _ := excelize.ColumnNumberToName(j + 1)
				cell := fmt.Sprintf("%s%d", col, totalRow)
				f.SetCellValue(sheetName, cell, v)
				if j >= 3 {
					f.SetCellStyle(sheetName, cell, cell, totalMoneyStyle)
				} else {
					f.SetCellStyle(sheetName, cell, cell, totalStyle)
				}
			}
			f.SetRowHeight(sheetName, totalRow, 28)
		}
	}

	firstSheet := f.GetSheetList()[0]
	sheetIdx, _ := f.GetSheetIndex(firstSheet)
	f.SetActiveSheet(sheetIdx)

	filename := fmt.Sprintf("daily_salary_summary_%s.xlsx", date)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// BankSheet godoc
//
// @Summary      Bank sheet (salary bank transfer)
// @Description  Get salary records filtered by account type and group for bank transfer
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        company_id    query string true  "Company ID"
// @Param        month         query int    true  "Month (1-12)"
// @Param        year          query int    true  "Year"
// @Param        group_id      query string false "Filter by group ID"
// @Param        account_type  query string false "Filter by account type (mCash/Card/Bank)"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /salary/bank-sheet [get]
func (h *SalaryHandler) BankSheet(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id, month, and year are required"})
		return
	}

	salaries, err := h.salaryRepo.ListAllByMonthFiltered(repository.SalaryFilter{
		CompanyID:   companyID,
		Month:       month,
		Year:        year,
		GroupID:     c.Query("group_id"),
		AccountType: c.Query("account_type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totals := map[string]float64{
		"gross_salary": 0, "net_salary": 0,
	}
	for _, s := range salaries {
		totals["gross_salary"] += s.GrossSalary
		totals["net_salary"] += s.NetSalary
	}

	c.JSON(http.StatusOK, gin.H{
		"salaries": salaries,
		"total":    len(salaries),
		"totals":   totals,
		"month":    month,
		"year":     year,
	})
}

type bankSheetItem struct {
	EmployeeID    string
	Name          string
	AccountNumber string
	NetSalary     float64
}

type bankSheetStyles struct {
	header    int
	data      int
	line      int
	subtotal  int
	money     int
	moneyBold int
}

func newBankSheetStyles(f *excelize.File) *bankSheetStyles {
	header, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#FFFFFF", Style: 1},
			{Type: "right", Color: "#FFFFFF", Style: 1},
			{Type: "top", Color: "#FFFFFF", Style: 1},
			{Type: "bottom", Color: "#FFFFFF", Style: 1},
		},
	})
	data, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	line, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D6E4F0"}},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	subtotal, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "#006100"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	money, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		NumFmt:    4,
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	moneyBold, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "#006100"},
		NumFmt:    4,
		Alignment: &excelize.Alignment{Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	return &bankSheetStyles{header, data, line, subtotal, money, moneyBold}
}

func writeSummarySheet(f *excelize.File, sheet string, salaries []models.Salary, styles *bankSheetStyles) {
	for i, h := range summaryHeaders {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, fmt.Sprintf("%s1", col), h)
		f.SetColWidth(sheet, col, col, summaryWidths[i])
		f.SetCellStyle(sheet, fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), styles.header)
	}
	f.SetRowHeight(sheet, 1, 30)

	lineMap := make(map[string][]bankSheetItem)
	for _, s := range salaries {
		lineName := "No Line"
		if s.Employee.LineRef != nil {
			lineName = s.Employee.LineRef.Name
		}
		lineMap[lineName] = append(lineMap[lineName], bankSheetItem{
			EmployeeID:    s.Employee.EmployeeID,
			Name:          s.Employee.NameEn,
			AccountNumber: s.Employee.AccountNumber,
			NetSalary:     s.NetSalary,
		})
	}

	sortedNames := make([]string, 0, len(lineMap))
	for k := range lineMap {
		sortedNames = append(sortedNames, k)
	}
	sort.Strings(sortedNames)

	row := 2
	sl := 0
	for _, lineName := range sortedNames {
		items := lineMap[lineName]
		var lineTotal float64
		for _, it := range items {
			lineTotal += it.NetSalary
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("%s  (%d employees)", lineName, len(items)))
		f.MergeCell(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("E%d", row))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), "")
		for i := 0; i < 6; i++ {
			col := string(rune('A' + i))
			f.SetCellStyle(sheet, fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), styles.line)
		}
		f.SetRowHeight(sheet, row, 22)
		row++

		for _, it := range items {
			sl++
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sl)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles.data)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), lineName)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles.data)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), it.EmployeeID)
			f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styles.data)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), it.Name)
			f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styles.data)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), it.AccountNumber)
			f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styles.data)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", row), it.NetSalary)
			f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styles.money)
			f.SetRowHeight(sheet, row, 20)
			row++
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "")
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "")
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "")
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "Line Total")
		for i := 0; i < 5; i++ {
			col := string(rune('A' + i))
			f.SetCellStyle(sheet, fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), styles.subtotal)
		}
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), lineTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styles.moneyBold)
		f.SetRowHeight(sheet, row, 22)
		row++
	}

	f.SetSheetView(sheet, -1, &excelize.ViewOptions{
		ShowGridLines: func(b bool) *bool { return &b }(false),
	})
}

var flatHeaders = []string{"Sl", "Employee ID", "Name", "Account Number", "Net Salary"}
var flatWidths = []float64{6, 16, 30, 22, 16}
var summaryHeaders = []string{"Sl", "Line", "Employee ID", "Name", "Account Number", "Net Salary"}
var summaryWidths = []float64{6, 16, 16, 30, 22, 16}

func writeFlatSheet(f *excelize.File, sheet string, salaries []models.Salary, styles *bankSheetStyles) {
	for i, h := range flatHeaders {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, fmt.Sprintf("%s1", col), h)
		f.SetColWidth(sheet, col, col, flatWidths[i])
		f.SetCellStyle(sheet, fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), styles.header)
	}
	f.SetRowHeight(sheet, 1, 30)

	var total float64
	for i, s := range salaries {
		r := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), i+1)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", r), fmt.Sprintf("A%d", r), styles.data)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), s.Employee.EmployeeID)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf("B%d", r), styles.data)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), s.Employee.NameEn)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", r), fmt.Sprintf("C%d", r), styles.data)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), s.Employee.AccountNumber)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", r), fmt.Sprintf("D%d", r), styles.data)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), s.NetSalary)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", r), fmt.Sprintf("E%d", r), styles.money)
		f.SetRowHeight(sheet, r, 20)
		total += s.NetSalary
	}

	lastRow := len(salaries) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), "")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", lastRow), "")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", lastRow), "")
	f.SetCellValue(sheet, fmt.Sprintf("D%d", lastRow), "Total")
	for i := 0; i < 4; i++ {
		col := string(rune('A' + i))
		f.SetCellStyle(sheet, fmt.Sprintf("%s%d", col, lastRow), fmt.Sprintf("%s%d", col, lastRow), styles.subtotal)
	}
	f.SetCellValue(sheet, fmt.Sprintf("E%d", lastRow), total)
	f.SetCellStyle(sheet, fmt.Sprintf("E%d", lastRow), fmt.Sprintf("E%d", lastRow), styles.moneyBold)
	f.SetRowHeight(sheet, lastRow, 22)

	f.SetSheetView(sheet, -1, &excelize.ViewOptions{
		ShowGridLines: func(b bool) *bool { return &b }(false),
	})
}

// BankSheetExportAll godoc
//
// @Summary      Export bank sheet (all tabs) to Excel
// @Description  Download salary bank transfer data as multi-sheet Excel with Summary, Staff-mCash, Staff-Card, Worker-mCash, Worker-Card tabs
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        company_id      query string true  "Company ID"
// @Param        month           query int    true  "Month (1-12)"
// @Param        year            query int    true  "Year"
// @Param        staff_group_id  query string false "Staff group ID"
// @Param        worker_group_id query string false "Worker group ID"
// @Success      200  {file}  file
// @Failure      500  {object}  map[string]string
// @Router       /salary/bank-sheet/export-all [get]
func (h *SalaryHandler) BankSheetExportAll(c *gin.Context) {
	companyID := c.Query("company_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")
	staffGroupID := c.Query("staff_group_id")
	workerGroupID := c.Query("worker_group_id")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id, month, and year are required"})
		return
	}

	// Fetch all 5 datasets
	type sheetDef struct {
		name        string
		groupID     string
		accountType string
	}
	defs := []sheetDef{
		{name: "Summary", groupID: "", accountType: ""},
		{name: "Staff-mCash", groupID: staffGroupID, accountType: "mCash"},
		{name: "Staff-Card", groupID: staffGroupID, accountType: "Card"},
		{name: "Worker-mCash", groupID: workerGroupID, accountType: "mCash"},
		{name: "Worker-Card", groupID: workerGroupID, accountType: "Card"},
	}

	type sheetData struct {
		def  sheetDef
		data []models.Salary
	}
	var results []sheetData

	for _, d := range defs {
		salaries, err := h.salaryRepo.ListAllByMonthFiltered(repository.SalaryFilter{
			CompanyID:   companyID,
			Month:       month,
			Year:        year,
			GroupID:     d.groupID,
			AccountType: d.accountType,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": d.name + ": " + err.Error()})
			return
		}
		results = append(results, sheetData{def: d, data: salaries})
	}

	f := excelize.NewFile()
	defer f.Close()

	styles := newBankSheetStyles(f)

	for idx, rd := range results {
		sheetName := rd.def.name
		if idx == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}
		if sheetName == "Summary" {
			writeSummarySheet(f, sheetName, rd.data, styles)
		} else {
			writeFlatSheet(f, sheetName, rd.data, styles)
		}
	}

	filename := fmt.Sprintf("bank_sheet_%d_%02d.xlsx", year, month)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}
