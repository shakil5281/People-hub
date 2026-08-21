package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/xuri/excelize/v2"
)

func colName(n int) string {
	name, _ := excelize.ColumnNumberToName(n)
	return name
}

// ExportEmployeesExcel godoc
//
//	@Summary      Export employees to Excel
//	@Description  Export filtered employee list to Excel with complete details
//	@Tags         Employees
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        company_id query string false "Filter by company"
//	@Param        department_id query string false "Filter by department"
//	@Param        section_id query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id query string false "Filter by line"
//	@Param        shift_id query string false "Filter by shift"
//	@Param        group_id query string false "Filter by group"
//	@Param        floor_id query string false "Filter by floor"
//	@Param        status query string false "Filter by status"
//	@Param        employee_id query string false "Search by employee ID"
//	@Param        gender query string false "Filter by gender"
//	@Param        blood_group query string false "Filter by blood group"
//	@Param        min_salary query string false "Minimum gross salary"
//	@Param        max_salary query string false "Maximum gross salary"
//	@Success      200  {file}  binary
//	@Router       /employees/export/excel [get]
func (h *EmployeeHandler) ExportExcel(c *gin.Context) {
	var employees []models.Employee
	query := database.DB.
		Preload("Company").
		Preload("Department").
		Preload("DesignationRef").
		Preload("SectionRef").
		Preload("LineRef").
		Preload("GroupRef").
		Preload("FloorRef").
		Preload("Shift").
		Preload("PresentDivision").
		Preload("PresentDistrict").
		Preload("PresentUpazila").
		Preload("PresentUnion").
		Preload("PermanentDivision").
		Preload("PermanentDistrict").
		Preload("PermanentUpazila").
		Preload("PermanentUnion")

	if v := c.Query("company_id"); v != "" {
		query = query.Where("company_id = ?", v)
	}
	if v := c.Query("department_id"); v != "" {
		query = query.Where("department_id = ?", v)
	}
	if v := c.Query("section_id"); v != "" {
		query = query.Where("section_id = ?", v)
	}
	if v := c.Query("designation_id"); v != "" {
		query = query.Where("designation_id = ?", v)
	}
	if v := c.Query("line_id"); v != "" {
		query = query.Where("line_id = ?", v)
	}
	if v := c.Query("shift_id"); v != "" {
		query = query.Where("shift_id = ?", v)
	}
	if v := c.Query("group_id"); v != "" {
		query = query.Where("group_id = ?", v)
	}
	if v := c.Query("floor_id"); v != "" {
		query = query.Where("floor_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("employee_id"); v != "" {
		query = query.Where("employee_id ILIKE ?", "%"+v+"%")
	}
	if v := c.Query("gender"); v != "" {
		query = query.Where("gender = ?", v)
	}
	if v := c.Query("blood_group"); v != "" {
		query = query.Where("blood_group = ?", v)
	}
	if v := c.Query("employee_type"); v != "" {
		query = query.Where("employee_type = ?", v)
	}
	if v := c.Query("min_salary"); v != "" {
		if minVal, err := strconv.ParseFloat(v, 64); err == nil {
			query = query.Where("gross_salary >= ?", minVal)
		}
	}
	if v := c.Query("max_salary"); v != "" {
		if maxVal, err := strconv.ParseFloat(v, 64); err == nil {
			query = query.Where("gross_salary <= ?", maxVal)
		}
	}

	if err := query.Order("LENGTH(employee_id) ASC, employee_id ASC").Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "Employees"
	f.SetSheetName("Sheet1", sheet)

	headerFont := &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11, Family: "Calibri"}
	headerFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"4472C4"}}
	headerAlign := &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}
	headerBorder := []excelize.Border{
		{Type: "left", Color: "FFFFFF", Style: 1},
		{Type: "top", Color: "FFFFFF", Style: 1},
		{Type: "bottom", Color: "FFFFFF", Style: 1},
		{Type: "right", Color: "FFFFFF", Style: 1},
	}

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font: headerFont, Fill: headerFill, Alignment: headerAlign, Border: headerBorder,
	})

	dataFont := &excelize.Font{Size: 10, Family: "Calibri"}
	dataBorder := []excelize.Border{
		{Type: "left", Color: "D9D9D9", Style: 1},
		{Type: "top", Color: "D9D9D9", Style: 1},
		{Type: "bottom", Color: "D9D9D9", Style: 1},
		{Type: "right", Color: "D9D9D9", Style: 1},
	}
	dataAlign := &excelize.Alignment{Vertical: "center", WrapText: true}
	centerAlign := &excelize.Alignment{Horizontal: "center", Vertical: "center"}

	styleData, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign})
	styleDataCenter, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: centerAlign})
	styleSalary, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, CustomNumFmt: &[]string{"#,##0.00"}[0]})
	styleDate, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, CustomNumFmt: &[]string{"yyyy-mm-dd"}[0]})

	altFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F7FC"}}
	styleDataAlt, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, Fill: altFill})
	styleDataCenterAlt, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: centerAlign, Fill: altFill})
	styleSalaryAlt, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, CustomNumFmt: &[]string{"#,##0.00"}[0], Fill: altFill})
	styleDateAlt, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, CustomNumFmt: &[]string{"yyyy-mm-dd"}[0], Fill: altFill})

	type col struct {
		header string
		width  float64
		center bool
		isDate bool
		isNum  bool
	}
	cols := []col{
		{"Employee ID", 14, true, false, false},
		{"Name (English)", 28, false, false, false},
		{"Name (Bangla)", 28, false, false, false},
		{"Designation", 22, false, false, false},
		{"Department", 22, false, false, false},
		{"Section", 18, false, false, false},
		{"Grade", 10, true, false, false},
		{"Line", 16, false, false, false},
		{"Group", 10, true, false, false},
		{"Floor", 10, true, false, false},
		{"Shift", 14, false, false, false},
		{"Punch No", 12, true, false, false},
		{"Phone", 14, false, false, false},
		{"Email", 28, false, false, false},
		{"NID", 18, false, false, false},
		{"Gender", 10, true, false, false},
		{"Blood Group", 12, true, false, false},
		{"Marital Status", 14, true, false, false},
		{"Religion", 12, true, false, false},
		{"Nationality", 14, true, false, false},
		{"Date of Birth", 14, true, true, false},
		{"Joining Date", 14, true, true, false},
		{"Employee Type", 16, false, false, false},
		{"Father's Name", 26, false, false, false},
		{"Mother's Name", 26, false, false, false},
		{"Spouse Name", 26, false, false, false},
		{"Emergency Contact", 22, false, false, false},
		{"Emergency Phone", 14, false, false, false},
		{"Dependents", 12, true, false, false},
		{"Present Address", 36, false, false, false},
		{"Present Division", 18, false, false, false},
		{"Present District", 18, false, false, false},
		{"Present Upazila", 18, false, false, false},
		{"Present Union", 18, false, false, false},
		{"Permanent Address", 36, false, false, false},
		{"Permanent Division", 18, false, false, false},
		{"Permanent District", 18, false, false, false},
		{"Permanent Upazila", 18, false, false, false},
		{"Permanent Union", 18, false, false, false},
		{"Gross Salary", 14, false, false, true},
		{"Basic Salary", 14, false, false, true},
		{"House Rent", 14, false, false, true},
		{"Medical Allowance", 14, false, false, true},
		{"Transport Allowance", 16, false, false, true},
		{"Food Allowance", 14, false, false, true},
		{"Other Allowance", 14, false, false, true},
		{"Account Type", 14, true, false, false},
		{"Account Number", 18, false, false, false},
		{"Overtime Status", 14, true, false, false},
		{"Status", 12, true, false, false},
	}

	for i, c := range cols {
		cell := colName(i+1) + "1"
		f.SetCellValue(sheet, cell, c.header)
		f.SetColWidth(sheet, colName(i+1), colName(i+1), c.width)
	}
	endCell := colName(len(cols)) + "1"
	f.SetCellStyle(sheet, "A1", endCell, styleHeader)
	f.SetRowHeight(sheet, 1, 30)

	timeNow := time.Now()
	for rowIdx, emp := range employees {
		row := rowIdx + 2
		isAlt := rowIdx%2 == 1

		var s, sc, sd, sn func(int) int
		if isAlt {
			s = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDataAlt)
				return 0
			}
			sc = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDataCenterAlt)
				return 0
			}
			sd = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDateAlt)
				return 0
			}
			sn = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleSalaryAlt)
				return 0
			}
		} else {
			s = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleData)
				return 0
			}
			sc = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDataCenter)
				return 0
			}
			sd = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDate)
				return 0
			}
			sn = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleSalary)
				return 0
			}
		}
		_ = s
		_ = sc
		_ = sd
		_ = sn

		sv := func(c int, v string) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sc(c) }
		svl := func(c int, v string) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); s(c) }
		svi := func(c int, v int) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sc(c) }
		svfl := func(c int, v float64) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sn(c) }
		svd := func(c int, v string) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sd(c) }

		sv(1, emp.EmployeeID)
		svl(2, emp.NameEn)
		svl(3, emp.NameBn)

		if emp.DesignationRef != nil {
			svl(4, emp.DesignationRef.Name)
		} else {
			svl(4, "")
		}
		if emp.Department != nil {
			svl(5, emp.Department.Name)
		} else {
			svl(5, "")
		}
		if emp.SectionRef != nil {
			svl(6, emp.SectionRef.Name)
		} else {
			svl(6, "")
		}
		sv(7, emp.Grade)
		if emp.LineRef != nil {
			svl(8, emp.LineRef.Name)
		} else {
			svl(8, "")
		}
		if emp.GroupRef != nil {
			sv(9, emp.GroupRef.Name)
		} else {
			sv(9, "")
		}
		if emp.FloorRef != nil {
			sv(10, emp.FloorRef.Name)
		} else {
			sv(10, "")
		}
		if emp.Shift != nil {
			svl(11, emp.Shift.Name)
		} else {
			svl(11, "")
		}
		sv(12, emp.PunchNumber)
		svl(13, emp.Phone)
		svl(14, emp.Email)
		svl(15, emp.NID)
		sv(16, emp.Gender)
		sv(17, emp.BloodGroup)
		sv(18, emp.MaritalStatus)
		sv(19, emp.Religion)
		sv(20, emp.Nationality)

		if emp.DateOfBirth != "" {
			svd(21, emp.DateOfBirth)
		} else {
			svd(21, "")
		}
		svd(22, emp.JoiningDate.Format("2006-01-02"))
		svl(23, emp.EmployeeType)

		svl(24, emp.FatherName)
		svl(25, emp.MotherName)
		svl(26, emp.SpouseName)
		svl(27, emp.EmergencyContact)
		svl(28, emp.EmergencyPhone)
		svi(29, emp.NumberOfDependents)

		svl(30, emp.PresentAddress)
		if emp.PresentDivision != nil {
			svl(31, emp.PresentDivision.Name)
		} else {
			svl(31, "")
		}
		if emp.PresentDistrict != nil {
			svl(32, emp.PresentDistrict.Name)
		} else {
			svl(32, "")
		}
		if emp.PresentUpazila != nil {
			svl(33, emp.PresentUpazila.Name)
		} else {
			svl(33, "")
		}
		if emp.PresentUnion != nil {
			svl(34, emp.PresentUnion.Name)
		} else {
			svl(34, "")
		}

		svl(35, emp.PermanentAddress)
		if emp.PermanentDivision != nil {
			svl(36, emp.PermanentDivision.Name)
		} else {
			svl(36, "")
		}
		if emp.PermanentDistrict != nil {
			svl(37, emp.PermanentDistrict.Name)
		} else {
			svl(37, "")
		}
		if emp.PermanentUpazila != nil {
			svl(38, emp.PermanentUpazila.Name)
		} else {
			svl(38, "")
		}
		if emp.PermanentUnion != nil {
			svl(39, emp.PermanentUnion.Name)
		} else {
			svl(39, "")
		}

		svfl(40, emp.GrossSalary)
		svfl(41, emp.BasicSalary)
		svfl(42, emp.HouseRent)
		svfl(43, emp.MedicalAllowance)
		svfl(44, emp.TransportAllowance)
		svfl(45, emp.FoodAllowance)
		svfl(46, emp.OtherAllowance)

		sv(47, emp.AccountType)
		svl(48, emp.AccountNumber)

		otStatus := "No"
		if emp.OverTimeStatus {
			otStatus = "Yes"
		}
		sv(49, otStatus)

		status := "Active"
		if emp.Status == "inactive" {
			status = "Inactive"
		}
		sv(50, status)

		svd(51, emp.CreatedAt.Format("2006-01-02"))
		_ = timeNow
	}

	lastCol := colName(len(cols))
	lastRow := len(employees) + 1
	f.SetRowHeight(sheet, 1, 30)
	for r := 2; r <= lastRow; r++ {
		f.SetRowHeight(sheet, r, 20)
	}
	f.SetSheetView(sheet, -1, &excelize.ViewOptions{
		ShowGridLines: func(b bool) *bool { return &b }(true),
	})
	f.AutoFilter(sheet, "A1:"+lastCol+strconv.Itoa(lastRow), []excelize.AutoFilterOptions{})

	freezeCell := "A2"
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: freezeCell,
		ActivePane:  "bottomLeft",
	})

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=employees_export.xlsx")
	f.Write(c.Writer)
}

// ExportEmployeesPDF godoc
//
// @Summary      Export employees to PDF
// @Description  Export filtered employee list to PDF format
// @Tags         Employees
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        company_id query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Param        status query string false "Filter by status"
// @Success      200  {file}  binary
// @Router       /employees/export/pdf [get]
func (h *EmployeeHandler) ExportPDF(c *gin.Context) {
	var employees []models.Employee
	query := database.DB.Preload("Company").Preload("Department").Preload("DesignationRef").Preload("SectionRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef")
	if v := c.Query("company_id"); v != "" {
		query = query.Where("company_id = ?", v)
	}
	if v := c.Query("department_id"); v != "" {
		query = query.Where("department_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if err := query.Order("LENGTH(employee_id) ASC, employee_id ASC").Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(5, 10, 5)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Employee List", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	headers := []string{"Emp. ID", "Name", "Designation", "Department", "Phone", "Status"}
	colWidths := []float64{20, 50, 40, 35, 35, 20}

	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(68, 114, 196)
	pdf.SetTextColor(255, 255, 255)
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetTextColor(0, 0, 0)
	for _, emp := range employees {
		pageHeight := 297.0
		if pdf.GetY() > pageHeight-20 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 8)
			pdf.SetFillColor(68, 114, 196)
			pdf.SetTextColor(255, 255, 255)
			for i, h := range headers {
				pdf.CellFormat(colWidths[i], 7, h, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetTextColor(0, 0, 0)
		}

		fill := false
		if pdf.GetY()/7/2 == 0 {
			fill = true
		}
		if fill {
			pdf.SetFillColor(240, 245, 255)
		}

		pdf.SetFont("Arial", "", 7)
		pdf.CellFormat(colWidths[0], 7, emp.EmployeeID, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[1], 7, truncate(emp.NameEn, 25), "1", 0, "L", fill, 0, "")
		desigName := ""
		if emp.DesignationRef != nil {
			desigName = emp.DesignationRef.Name
		}
		pdf.CellFormat(colWidths[2], 7, truncate(desigName, 20), "1", 0, "L", fill, 0, "")
		deptName := ""
		if emp.Department != nil {
			deptName = emp.Department.Name
		}
		pdf.CellFormat(colWidths[3], 7, truncate(deptName, 18), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[4], 7, emp.Phone, "1", 0, "L", fill, 0, "")
		status := "Active"
		if emp.Status == "inactive" {
			status = "Inactive"
		}
		pdf.CellFormat(colWidths[5], 7, status, "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
	}

	// Footer summary
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 8, fmt.Sprintf("Total Employees: %d", len(employees)), "", 1, "R", false, 0, "")

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=employees_export.pdf")
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}

func truncate(s string, maxLen int) string {
	if len([]rune(s)) > maxLen {
		return string([]rune(s)[:maxLen-3]) + "..."
	}
	return s
}

// ExportProfileExcel godoc
//
//	@Summary      Export employee profile to Excel
//	@Description  Export full employee profile with salary and attendance to Excel
//	@Tags         Employees
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        id     path  string  true  "Employee ID"
//	@Param        month  query int     false "Month"
//	@Param        year   query int     false "Year"
//	@Success      200    {file}  binary
//	@Router       /employees/{id}/profile/export/excel [get]
func (h *EmployeeHandler) ExportProfileExcel(c *gin.Context) {
	id := c.Param("id")

	var emp models.Employee
	if err := database.DB.Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	endDate := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	var attendanceCounts []struct {
		Status string
		Count  int64
	}
	database.DB.Model(&models.Attendance{}).
		Select("status, count(*) as count").
		Where("employee_id = ? AND date BETWEEN ? AND ? AND deleted_at IS NULL", emp.EmployeeID, startDate, endDate).
		Group("status").
		Find(&attendanceCounts)

	var salary models.Salary
	salaryErr := database.DB.Where("employee_id = ? AND month = ? AND year = ?", emp.EmployeeID, month, year).First(&salary).Error

	f := excelize.NewFile()
	sheet := "Profile"
	f.SetSheetName("Sheet1", sheet)

	sectionFont := &excelize.Font{Bold: true, Size: 12, Color: "1F4E79", Family: "Calibri"}
	sectionFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D6E4F0"}}
	labelFont := &excelize.Font{Bold: true, Size: 10, Family: "Calibri"}
	labelFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F2F2"}}
	dataFont := &excelize.Font{Size: 10, Family: "Calibri"}
	titleFont := &excelize.Font{Bold: true, Size: 16, Color: "1F4E79", Family: "Calibri"}

	styleSection, _ := f.NewStyle(&excelize.Style{
		Font: sectionFont, Fill: sectionFill,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	styleLabel, _ := f.NewStyle(&excelize.Style{
		Font: labelFont, Fill: labelFill,
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})
	styleData, _ := f.NewStyle(&excelize.Style{
		Font: dataFont,
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})
	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font: titleFont,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 50)

	row := 1

	// Title row
	if emp.Company.ID != "" {
		f.MergeCell(sheet, "A1", "B1")
		f.SetCellStyle(sheet, "A1", "B1", styleTitle)
		f.SetCellValue(sheet, "A1", emp.Company.CompanyNameEn)
		f.SetRowHeight(sheet, 1, 30)
		row = 2
	}

	f.MergeCell(sheet, colName(1)+strconv.Itoa(row), colName(2)+strconv.Itoa(row))
	f.SetCellStyle(sheet, colName(1)+strconv.Itoa(row), colName(2)+strconv.Itoa(row), styleTitle)
	f.SetCellValue(sheet, colName(1)+strconv.Itoa(row), emp.NameEn+" ("+emp.EmployeeID+")")
	f.SetRowHeight(sheet, row, 35)
	row++

	// Blank row
	row++

	sv := func(r int, v string) {
		f.SetCellValue(sheet, colName(2)+strconv.Itoa(r), v)
		f.SetCellStyle(sheet, colName(2)+strconv.Itoa(r), colName(2)+strconv.Itoa(r), styleData)
	}
	setLabel := func(r int, label string) {
		f.SetCellValue(sheet, colName(1)+strconv.Itoa(r), label)
		f.SetCellStyle(sheet, colName(1)+strconv.Itoa(r), colName(1)+strconv.Itoa(r), styleLabel)
	}
	setSection := func(r int, title string) {
		f.MergeCell(sheet, colName(1)+strconv.Itoa(r), colName(2)+strconv.Itoa(r))
		f.SetCellValue(sheet, colName(1)+strconv.Itoa(r), title)
		f.SetCellStyle(sheet, colName(1)+strconv.Itoa(r), colName(2)+strconv.Itoa(r), styleSection)
		f.SetRowHeight(sheet, r, 25)
	}

	// Personal Information
	setSection(row, "Personal Information")
	row++
	setLabel(row, "Full Name (English)")
	sv(row, emp.NameEn)
	row++
	setLabel(row, "Full Name (Bangla)")
	sv(row, emp.NameBn)
	row++
	setLabel(row, "Father's Name")
	sv(row, emp.FatherName)
	row++
	setLabel(row, "Mother's Name")
	sv(row, emp.MotherName)
	row++
	setLabel(row, "Date of Birth")
	sv(row, emp.DateOfBirth)
	row++
	setLabel(row, "Gender")
	sv(row, emp.Gender)
	row++
	setLabel(row, "Blood Group")
	sv(row, emp.BloodGroup)
	row++
	setLabel(row, "Marital Status")
	sv(row, emp.MaritalStatus)
	row++
	setLabel(row, "Religion")
	sv(row, emp.Religion)
	row++
	setLabel(row, "Nationality")
	sv(row, emp.Nationality)
	row++
	setLabel(row, "NID Number")
	sv(row, emp.NID)
	row++

	// Contact & Address
	setSection(row, "Contact & Address")
	row++
	setLabel(row, "Phone")
	sv(row, emp.Phone)
	row++
	setLabel(row, "Email")
	sv(row, emp.Email)
	row++
	setLabel(row, "Present Address")
	sv(row, emp.PresentAddress)
	row++
	setLabel(row, "Permanent Address")
	sv(row, emp.PermanentAddress)
	row++

	// Office Information
	setSection(row, "Office Information")
	row++
	setLabel(row, "Employee ID")
	sv(row, emp.EmployeeID)
	row++
	setLabel(row, "Punch Number")
	sv(row, emp.PunchNumber)
	row++
	setLabel(row, "Employee Type")
	sv(row, emp.EmployeeType)
	row++
	setLabel(row, "Grade")
	sv(row, emp.Grade)
	row++
	setLabel(row, "Joining Date")
	sv(row, emp.JoiningDate.Format("2006-01-02"))
	row++
	if emp.Department != nil {
		setLabel(row, "Department")
		sv(row, emp.Department.Name)
		row++
	}
	if emp.SectionRef != nil {
		setLabel(row, "Section")
		sv(row, emp.SectionRef.Name)
		row++
	}
	if emp.DesignationRef != nil {
		setLabel(row, "Designation")
		sv(row, emp.DesignationRef.Name)
		row++
	}
	if emp.LineRef != nil {
		setLabel(row, "Line")
		sv(row, emp.LineRef.Name)
		row++
	}
	if emp.GroupRef != nil {
		setLabel(row, "Group")
		sv(row, emp.GroupRef.Name)
		row++
	}
	if emp.FloorRef != nil {
		setLabel(row, "Floor")
		sv(row, emp.FloorRef.Name)
		row++
	}
	if emp.Shift != nil {
		setLabel(row, "Shift")
		sv(row, emp.Shift.Name)
		row++
	}
	otStatus := "No"
	if emp.OverTimeStatus {
		otStatus = "Yes"
	}
	setLabel(row, "Over Time")
	sv(row, otStatus)
	row++
	if emp.Company.ID != "" {
		setLabel(row, "Company")
		sv(row, emp.Company.CompanyNameEn)
		row++
	}

	// Family & Emergency
	setSection(row, "Family & Emergency")
	row++
	setLabel(row, "Spouse Name")
	sv(row, emp.SpouseName)
	row++
	setLabel(row, "Emergency Contact")
	sv(row, emp.EmergencyContact)
	row++
	setLabel(row, "Emergency Phone")
	sv(row, emp.EmergencyPhone)
	row++
	setLabel(row, "Dependents")
	sv(row, strconv.Itoa(emp.NumberOfDependents))
	row++

	// Bank Account
	setSection(row, "Bank Account")
	row++
	setLabel(row, "Account Type")
	sv(row, emp.AccountType)
	row++
	setLabel(row, "Account Number")
	sv(row, emp.AccountNumber)
	row++

	// Salary
	if salaryErr == nil {
		setSection(row, fmt.Sprintf("Salary — %d/%d", salary.Month, salary.Year))
		row++
		setLabel(row, "Gross Salary")
		sv(row, fmt.Sprintf("৳%.2f", salary.GrossSalary))
		row++
		setLabel(row, "Basic Salary")
		sv(row, fmt.Sprintf("৳%.2f", salary.BasicSalary))
		row++
		setLabel(row, "House Rent")
		sv(row, fmt.Sprintf("৳%.2f", salary.HouseRent))
		row++
		setLabel(row, "Medical Allowance")
		sv(row, fmt.Sprintf("৳%.2f", salary.MedicalAllowance))
		row++
		setLabel(row, "Transport Allowance")
		sv(row, fmt.Sprintf("৳%.2f", salary.TransportAllowance))
		row++
		setLabel(row, "Food Allowance")
		sv(row, fmt.Sprintf("৳%.2f", salary.FoodAllowance))
		row++
		setLabel(row, "Other Allowance")
		sv(row, fmt.Sprintf("৳%.2f", salary.OtherAllowance))
		row++
		setLabel(row, "Overtime Amount")
		sv(row, fmt.Sprintf("৳%.2f", salary.OvertimeAmount))
		row++
		setLabel(row, "Attendance Bonus")
		sv(row, fmt.Sprintf("৳%.2f", salary.AttendanceBonus))
		row++
		setLabel(row, "Absent Deduction")
		sv(row, fmt.Sprintf("৳%.2f", salary.AbsentDeduction))
		row++
		setLabel(row, "Other Deduction")
		sv(row, fmt.Sprintf("৳%.2f", salary.OtherDeduction))
		row++
		setLabel(row, "Total Deductions")
		sv(row, fmt.Sprintf("৳%.2f", salary.TotalDeductions))
		row++
		setLabel(row, "Net Salary")
		sv(row, fmt.Sprintf("৳%.2f", salary.NetSalary))
		row++
	}

	// Attendance Summary
	if len(attendanceCounts) > 0 {
		setSection(row, fmt.Sprintf("Attendance Summary — %d/%d", month, year))
		row++
		for _, a := range attendanceCounts {
			setLabel(row, a.Status)
			sv(row, strconv.FormatInt(a.Count, 10))
			row++
		}
	}

	f.SetSheetView(sheet, -1, &excelize.ViewOptions{
		ShowGridLines: func(b bool) *bool { return &b }(false),
	})

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=profile_%s.xlsx", emp.EmployeeID))
	f.Write(c.Writer)
}

// ExportProfilePDF godoc
//
//	@Summary      Export employee profile to PDF
//	@Description  Export full employee profile with salary and attendance to PDF
//	@Tags         Employees
//	@Security     BearerAuth
//	@Produce      application/pdf
//	@Param        id     path  string  true  "Employee ID"
//	@Param        month  query int     false "Month"
//	@Param        year   query int     false "Year"
//	@Success      200    {file}  binary
//	@Router       /employees/{id}/profile/export/pdf [get]
func (h *EmployeeHandler) ExportProfilePDF(c *gin.Context) {
	id := c.Param("id")

	var emp models.Employee
	if err := database.DB.Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	endDate := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	var attendanceCounts []struct {
		Status string
		Count  int64
	}
	database.DB.Model(&models.Attendance{}).
		Select("status, count(*) as count").
		Where("employee_id = ? AND date BETWEEN ? AND ? AND deleted_at IS NULL", emp.EmployeeID, startDate, endDate).
		Group("status").
		Find(&attendanceCounts)

	var salary models.Salary
	salaryErr := database.DB.Where("employee_id = ? AND month = ? AND year = ?", emp.EmployeeID, month, year).First(&salary).Error

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header bar
	pdf.SetFillColor(68, 114, 196)
	pdf.Rect(0, 0, 210, 40, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetY(8)
	companyName := ""
	if emp.Company.ID != "" {
		companyName = emp.Company.CompanyNameEn + " — "
	}
	pdf.CellFormat(0, 10, companyName+"Employee Profile", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 8, emp.NameEn+" ("+emp.EmployeeID+")", "", 1, "C", false, 0, "")

	// Status badge area
	pdf.SetY(44)
	pdf.SetTextColor(0, 0, 0)

	// Helper function
	fieldRow := func(label, value string) {
		pdf.SetX(15)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(240, 245, 250)
		pdf.CellFormat(55, 7, label, "1", 0, "L", true, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(0, 7, " "+value, "1", 1, "L", false, 0, "")
	}

	sectionHeader := func(title string) {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}
		pdf.SetX(15)
		pdf.SetFont("Arial", "B", 11)
		pdf.SetFillColor(68, 114, 196)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(0, 8, "  "+title, "1", 1, "L", true, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	sectionHeader("Personal Information")
	fieldRow("Full Name (English)", emp.NameEn)
	fieldRow("Full Name (Bangla)", emp.NameBn)
	fieldRow("Father's Name", emp.FatherName)
	fieldRow("Mother's Name", emp.MotherName)
	fieldRow("Date of Birth", emp.DateOfBirth)
	fieldRow("Gender", emp.Gender)
	fieldRow("Blood Group", emp.BloodGroup)
	fieldRow("Marital Status", emp.MaritalStatus)
	fieldRow("Religion", emp.Religion)
	fieldRow("Nationality", emp.Nationality)
	fieldRow("NID Number", emp.NID)

	sectionHeader("Contact & Address")
	fieldRow("Phone", emp.Phone)
	fieldRow("Email", emp.Email)
	fieldRow("Present Address", emp.PresentAddress)
	fieldRow("Permanent Address", emp.PermanentAddress)

	sectionHeader("Office Information")
	fieldRow("Employee ID", emp.EmployeeID)
	fieldRow("Punch Number", emp.PunchNumber)
	fieldRow("Employee Type", emp.EmployeeType)
	fieldRow("Grade", emp.Grade)
	fieldRow("Joining Date", emp.JoiningDate.Format("2006-01-02"))
	if emp.Department != nil {
		fieldRow("Department", emp.Department.Name)
	}
	if emp.SectionRef != nil {
		fieldRow("Section", emp.SectionRef.Name)
	}
	if emp.DesignationRef != nil {
		fieldRow("Designation", emp.DesignationRef.Name)
	}
	if emp.LineRef != nil {
		fieldRow("Line", emp.LineRef.Name)
	}
	if emp.GroupRef != nil {
		fieldRow("Group", emp.GroupRef.Name)
	}
	if emp.FloorRef != nil {
		fieldRow("Floor", emp.FloorRef.Name)
	}
	if emp.Shift != nil {
		fieldRow("Shift", emp.Shift.Name)
	}
	otStatus := "No"
	if emp.OverTimeStatus {
		otStatus = "Yes"
	}
	fieldRow("Over Time", otStatus)
	if emp.Company.ID != "" {
		fieldRow("Company", emp.Company.CompanyNameEn)
	}

	sectionHeader("Family & Emergency")
	fieldRow("Spouse Name", emp.SpouseName)
	fieldRow("Emergency Contact", emp.EmergencyContact)
	fieldRow("Emergency Phone", emp.EmergencyPhone)
	fieldRow("Dependents", strconv.Itoa(emp.NumberOfDependents))

	sectionHeader("Bank Account")
	fieldRow("Account Type", emp.AccountType)
	fieldRow("Account Number", emp.AccountNumber)

	// Salary
	if salaryErr == nil {
		sectionHeader(fmt.Sprintf("Salary — %d/%d", salary.Month, salary.Year))
		fieldRow("Gross Salary", fmt.Sprintf("৳%.2f", salary.GrossSalary))
		fieldRow("Basic Salary", fmt.Sprintf("৳%.2f", salary.BasicSalary))
		fieldRow("House Rent", fmt.Sprintf("৳%.2f", salary.HouseRent))
		fieldRow("Medical Allowance", fmt.Sprintf("৳%.2f", salary.MedicalAllowance))
		fieldRow("Transport Allowance", fmt.Sprintf("৳%.2f", salary.TransportAllowance))
		fieldRow("Food Allowance", fmt.Sprintf("৳%.2f", salary.FoodAllowance))
		fieldRow("Other Allowance", fmt.Sprintf("৳%.2f", salary.OtherAllowance))
		fieldRow("Overtime Amount", fmt.Sprintf("৳%.2f", salary.OvertimeAmount))
		fieldRow("Attendance Bonus", fmt.Sprintf("৳%.2f", salary.AttendanceBonus))
		fieldRow("Absent Deduction", fmt.Sprintf("৳%.2f", salary.AbsentDeduction))
		fieldRow("Other Deduction", fmt.Sprintf("৳%.2f", salary.OtherDeduction))
		fieldRow("Total Deductions", fmt.Sprintf("৳%.2f", salary.TotalDeductions))

		// Net salary highlight
		pdf.SetX(15)
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(68, 114, 196)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(55, 8, "  Net Salary", "1", 0, "L", true, 0, "")
		pdf.CellFormat(0, 8, "  ৳"+fmt.Sprintf("%.2f", salary.NetSalary), "1", 1, "L", true, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	// Attendance
	if len(attendanceCounts) > 0 {
		sectionHeader(fmt.Sprintf("Attendance Summary — %d/%d", month, year))
		for _, a := range attendanceCounts {
			fieldRow(a.Status, strconv.FormatInt(a.Count, 10))
		}
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=profile_%s.pdf", emp.EmployeeID))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}
