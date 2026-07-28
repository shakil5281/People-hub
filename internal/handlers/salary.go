package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

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
func (h *SalaryHandler) SheetExport(c *gin.Context) {
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

	f := excelize.NewFile()
	defer f.Close()

	monthName := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("January")

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:  &excelize.Font{Size: 10},
		NumFmt: 4,
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	headers := []string{
		"Sl", "Employee ID", "Name", "Designation", "Department", "Working Days",
		"Present", "Absent", "Late", "Leave", "Holiday", "Weekend",
		"Transport", "Food", "Medical", "Other", "Gross",
		"Absent Deduction", "PF", "Tax", "Total Deductions",
		"OT Hours", "OT Rate", "OT Amount", "Att. Bonus", "Net Salary", "Status",
	}
	widths := []float64{4, 14, 24, 18, 18, 10, 8, 8, 6, 6, 8, 8, 10, 10, 10, 10, 12, 12, 10, 8, 12, 8, 8, 10, 10, 12, 10}

	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue("Sheet1", fmt.Sprintf("%s1", col), h)
		f.SetColWidth("Sheet1", col, col, widths[i])
		f.SetCellStyle("Sheet1", fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), headerStyle)
	}
	f.SetRowHeight("Sheet1", 1, 30)

	for i, s := range salaries {
		row := i + 2
		empID := s.Employee.EmployeeID
		name := s.Employee.NameEn
		desig := ""
		dept := ""
		if s.Employee.DesignationRef != nil {
			desig = s.Employee.DesignationRef.Name
		}
		if s.Employee.Department != nil {
			dept = s.Employee.Department.Name
		}

		vals := []interface{}{
			i + 1, empID, name, desig, dept,
			s.TotalDays, s.PresentDays, s.AbsentDays, s.LateDays, s.LeaveDays, s.HolidayDays, s.WeekendDays,
			s.TransportAllowance, s.FoodAllowance, s.MedicalAllowance, s.OtherAllowance, s.GrossSalary,
			s.AbsentDeduction, s.ProvidentFund, s.Tax, s.TotalDeductions,
			s.OvertimeHours, s.OvertimeRate, s.OvertimeAmount, s.AttendanceBonus, s.NetSalary, s.Status,
		}

		for j, v := range vals {
			col := string(rune('A' + j))
			cell := fmt.Sprintf("%s%d", col, row)
			f.SetCellValue("Sheet1", cell, v)
			if j >= 12 && j <= 25 {
				f.SetCellStyle("Sheet1", cell, cell, moneyStyle)
			} else {
				f.SetCellStyle("Sheet1", cell, cell, dataStyle)
			}
		}
	}

	totalRow := len(salaries) + 2
	if len(salaries) > 0 {
		totalStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true, Size: 10, Color: "#006100"},
			Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
		})
		for j := 0; j < len(headers); j++ {
			col := string(rune('A' + j))
			cell := fmt.Sprintf("%s%d", col, totalRow)
			if j == 0 {
				f.SetCellValue("Sheet1", cell, "")
			} else if j == 1 {
				f.SetCellValue("Sheet1", cell, "Total")
			}
			f.SetCellStyle("Sheet1", cell, cell, totalStyle)
		}
		netTotal := 0.0
		grossTotal := 0.0
		for _, s := range salaries {
			netTotal += s.NetSalary
			grossTotal += s.GrossSalary
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("Q%d", totalRow), grossTotal)
		f.SetCellValue("Sheet1", fmt.Sprintf("Z%d", totalRow), netTotal)
		f.SetCellStyle("Sheet1", fmt.Sprintf("Q%d", totalRow), fmt.Sprintf("Q%d", totalRow), moneyStyle)
		f.SetCellStyle("Sheet1", fmt.Sprintf("Z%d", totalRow), fmt.Sprintf("Z%d", totalRow), moneyStyle)
	}

	filename := fmt.Sprintf("salary_sheet_%s_%d.xlsx", monthName, year)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// GetPayslip godoc
//
// @Summary      Get employee payslip
// @Description  Get salary record for a specific employee for a given month/year
// @Tags         Salary
// @Security     BearerAuth
// @Produce      json
// @Param        employee_id query string true "Employee ID"
// @Param        month       query int    true "Month (1-12)"
// @Param        year        query int    true "Year"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/payslip [get]
func (h *SalaryHandler) Payslip(c *gin.Context) {
	employeeID := c.Query("employee_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if employeeID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id, month, and year are required"})
		return
	}

	salary, err := h.salaryRepo.FindByEmployeeMonth(employeeID, month, year)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Salary not found for this employee/month"})
		return
	}

	c.JSON(http.StatusOK, salary)
}

// PayslipExport godoc
//
// @Summary      Export payslip to Excel
// @Description  Download a single employee payslip as Excel file
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        employee_id query string true "Employee ID"
// @Param        month       query int    true "Month (1-12)"
// @Param        year        query int    true "Year"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/payslip/export [get]
func (h *SalaryHandler) PayslipExport(c *gin.Context) {
	employeeID := c.Query("employee_id")
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if employeeID == "" || month == 0 || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id, month, and year are required"})
		return
	}

	salary, err := h.salaryRepo.FindByEmployeeMonth(employeeID, month, year)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Salary not found for this employee/month"})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	monthName := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("January")

	empName := ""
	if salary.Employee.NameEn != "" {
		empName = salary.Employee.NameEn
	}
	desig := ""
	if salary.Employee.DesignationRef != nil {
		desig = salary.Employee.DesignationRef.Name
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "#1F4E79"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	labelStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D6E4F0"}},
	})
	valueStyle, _ := f.NewStyle(&excelize.Style{
		Font:  &excelize.Font{Size: 11},
		NumFmt: 4,
	})
	netStyle, _ := f.NewStyle(&excelize.Style{
		Font:  &excelize.Font{Bold: true, Size: 14, Color: "#006100"},
		NumFmt: 4,
		Fill:  excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
	})
	sectionStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "#1F4E79"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9E2F3"}},
	})

	f.SetCellValue("Sheet1", "A1", "PAYSLIP")
	f.MergeCell("Sheet1", "A1", "B1")
	f.SetCellStyle("Sheet1", "A1", "A1", titleStyle)
	f.SetRowHeight("Sheet1", 1, 30)

	f.SetCellValue("Sheet1", "A2", "")
	f.SetCellValue("Sheet1", "B2", "")
	f.SetRowHeight("Sheet1", 2, 6)

	f.SetCellValue("Sheet1", "A3", fmt.Sprintf("Employee: %s", empName))
	f.SetCellValue("Sheet1", "B3", fmt.Sprintf("Code: %s", salary.Employee.EmployeeID))
	f.SetCellValue("Sheet1", "A4", fmt.Sprintf("Designation: %s", desig))
	f.SetCellValue("Sheet1", "B4", fmt.Sprintf("Period: %s %d", monthName, year))

	f.SetCellValue("Sheet1", "A6", "Earnings")
	f.SetCellStyle("Sheet1", "A6", "B6", sectionStyle)
	f.SetCellValue("Sheet1", "A7", "Component")
	f.SetCellValue("Sheet1", "B7", "Amount")
	f.SetCellStyle("Sheet1", "A7", "B7", headerStyle)

	earnings := [][2]string{
		{"Transport Allowance", fmt.Sprintf("%.2f", salary.TransportAllowance)},
		{"Food Allowance", fmt.Sprintf("%.2f", salary.FoodAllowance)},
		{"Medical Allowance", fmt.Sprintf("%.2f", salary.MedicalAllowance)},
		{"Other Allowance", fmt.Sprintf("%.2f", salary.OtherAllowance)},
		{"Gross Salary", fmt.Sprintf("%.2f", salary.GrossSalary)},
	}
	for i, e := range earnings {
		row := 8 + i
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), e[0])
		f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), labelStyle)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), e[1])
		if e[0] == "Gross Salary" {
			f.SetCellStyle("Sheet1", fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), netStyle)
		} else {
			f.SetCellStyle("Sheet1", fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), valueStyle)
		}
	}

	otRow := 8 + len(earnings) + 1
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", otRow), "Overtime & Bonus")
	f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", otRow), fmt.Sprintf("B%d", otRow), sectionStyle)

	headerRow := otRow + 1
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", headerRow), "Component")
	f.SetCellValue("Sheet1", fmt.Sprintf("B%d", headerRow), "Amount")
	f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", headerRow), fmt.Sprintf("B%d", headerRow), headerStyle)

	otItems := [][2]string{
		{"OT Hours", fmt.Sprintf("%.2f", salary.OvertimeHours)},
		{"OT Rate", fmt.Sprintf("%.2f", salary.OvertimeRate)},
		{"OT Amount", fmt.Sprintf("%.2f", salary.OvertimeAmount)},
		{"Attendance Bonus", fmt.Sprintf("%.2f", salary.AttendanceBonus)},
	}
	for i, item := range otItems {
		row := headerRow + 1 + i
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), item[0])
		f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), labelStyle)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), item[1])
		f.SetCellStyle("Sheet1", fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), valueStyle)
	}

	dedRow := headerRow + 1 + len(otItems) + 1
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", dedRow), "Deductions")
	f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", dedRow), fmt.Sprintf("B%d", dedRow), sectionStyle)

	dedHeaderRow := dedRow + 1
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", dedHeaderRow), "Component")
	f.SetCellValue("Sheet1", fmt.Sprintf("B%d", dedHeaderRow), "Amount")
	f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", dedHeaderRow), fmt.Sprintf("B%d", dedHeaderRow), headerStyle)

	dedItems := [][2]string{
		{"Provident Fund", fmt.Sprintf("%.2f", salary.ProvidentFund)},
		{"Tax", fmt.Sprintf("%.2f", salary.Tax)},
		{"Absent Deduction", fmt.Sprintf("%.2f", salary.AbsentDeduction)},
		{"Total Deductions", fmt.Sprintf("%.2f", salary.TotalDeductions)},
	}
	for i, item := range dedItems {
		row := dedHeaderRow + 1 + i
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), item[0])
		f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), labelStyle)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), item[1])
		if item[0] == "Total Deductions" {
			f.SetCellStyle("Sheet1", fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), netStyle)
		} else {
			f.SetCellStyle("Sheet1", fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), valueStyle)
		}
	}

	netRow := dedHeaderRow + 1 + len(dedItems) + 1
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", netRow), "Net Salary")
	f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", netRow), fmt.Sprintf("A%d", netRow), netStyle)
	f.SetCellValue("Sheet1", fmt.Sprintf("B%d", netRow), fmt.Sprintf("%.2f", salary.NetSalary))
	f.SetCellStyle("Sheet1", fmt.Sprintf("B%d", netRow), fmt.Sprintf("B%d", netRow), netStyle)

	attRow := netRow + 2
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", attRow), "Attendance Summary")
	f.SetCellStyle("Sheet1", fmt.Sprintf("A%d", attRow), fmt.Sprintf("B%d", attRow), sectionStyle)

	attHeaderRow := attRow + 1
	attHeaders := []string{"Present", "Absent", "Late", "Leave", "Holiday", "Weekend", "Total Days"}
	for i, h := range attHeaders {
		col := string(rune('A' + i))
		f.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, attHeaderRow), h)
		if col <= "B" {
			f.SetCellStyle("Sheet1", fmt.Sprintf("%s%d", col, attHeaderRow), fmt.Sprintf("%s%d", col, attHeaderRow), headerStyle)
		}
	}

	attDataRow := attHeaderRow + 1
	f.SetCellValue("Sheet1", fmt.Sprintf("A%d", attDataRow), salary.PresentDays)
	f.SetCellValue("Sheet1", fmt.Sprintf("B%d", attDataRow), salary.AbsentDays)
	f.SetCellValue("Sheet1", fmt.Sprintf("C%d", attDataRow), salary.LateDays)
	f.SetCellValue("Sheet1", fmt.Sprintf("D%d", attDataRow), salary.LeaveDays)
	f.SetCellValue("Sheet1", fmt.Sprintf("E%d", attDataRow), salary.HolidayDays)
	f.SetCellValue("Sheet1", fmt.Sprintf("F%d", attDataRow), salary.WeekendDays)
	f.SetCellValue("Sheet1", fmt.Sprintf("G%d", attDataRow), salary.TotalDays)

	f.SetColWidth("Sheet1", "A", "A", 30)
	f.SetColWidth("Sheet1", "B", "B", 18)

	filename := fmt.Sprintf("payslip_%s_%s_%d.xlsx", salary.Employee.EmployeeID, monthName, year)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
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
				key = groupKey{Name: s.Employee.LineRef.Name, ID: s.Employee.LineRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		default:
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

	f := excelize.NewFile()
	defer f.Close()

	monthName := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("January")

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	moneyStyle, _ := f.NewStyle(&excelize.Style{
		Font:  &excelize.Font{Size: 10},
		NumFmt: 4,
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	labelMap := map[string]string{
		"department": "Department",
		"section":    "Section",
		"designation": "Designation",
		"line":       "Line",
	}
	groupLabel := labelMap[groupBy]
	if groupLabel == "" {
		groupLabel = "Group"
	}

	headers := []string{groupLabel, "Employees", "Gross Total", "Basic Total", "House Rent", "Medical", "Transport", "Deductions", "Net Total"}
	widths := []float64{28, 10, 14, 14, 14, 14, 14, 14, 14}

	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue("Sheet1", fmt.Sprintf("%s1", col), h)
		f.SetColWidth("Sheet1", col, col, widths[i])
		f.SetCellStyle("Sheet1", fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), headerStyle)
	}
	f.SetRowHeight("Sheet1", 1, 30)

	row := 2
	var grandTotal groupData
	for key, d := range groupMap {
		vals := []interface{}{
			key.Name, d.Employees,
			d.GrossSalary, d.BasicSalary, d.HouseRent, d.Medical, d.Transport,
			d.Deductions, d.NetSalary,
		}
		for j, v := range vals {
			col := string(rune('A' + j))
			cell := fmt.Sprintf("%s%d", col, row)
			f.SetCellValue("Sheet1", cell, v)
			if j >= 2 {
				f.SetCellStyle("Sheet1", cell, cell, moneyStyle)
			} else {
				f.SetCellStyle("Sheet1", cell, cell, dataStyle)
			}
		}
		grandTotal.Employees += d.Employees
		grandTotal.BasicSalary += d.BasicSalary
		grandTotal.HouseRent += d.HouseRent
		grandTotal.Medical += d.Medical
		grandTotal.Transport += d.Transport
		grandTotal.GrossSalary += d.GrossSalary
		grandTotal.Deductions += d.Deductions
		grandTotal.NetSalary += d.NetSalary
		row++
	}

	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "#006100"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2EFDA"}},
	})
	grandVals := []interface{}{
		"Grand Total", grandTotal.Employees,
		grandTotal.GrossSalary, grandTotal.BasicSalary, grandTotal.HouseRent, grandTotal.Medical, grandTotal.Transport,
		grandTotal.Deductions, grandTotal.NetSalary,
	}
	for j, v := range grandVals {
		col := string(rune('A' + j))
		cell := fmt.Sprintf("%s%d", col, row)
		f.SetCellValue("Sheet1", cell, v)
		f.SetCellStyle("Sheet1", cell, cell, totalStyle)
		if j >= 2 {
			f.SetCellStyle("Sheet1", cell, cell, moneyStyle)
		}
	}

	filename := fmt.Sprintf("salary_summary_%s_%d.xlsx", monthName, year)
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
				key = groupKey{Name: s.Employee.LineRef.Name, ID: s.Employee.LineRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
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

	var summaries []map[string]interface{}

	for key, d := range groupMap {
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
