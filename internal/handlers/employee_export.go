package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	if v := c.Query("joining_date_from"); v != "" {
		query = query.Where("joining_date >= ?", v)
	} else if v := c.Query("joining_from"); v != "" {
		query = query.Where("joining_date >= ?", v)
	} else if v := c.Query("start_date"); v != "" {
		query = query.Where("joining_date >= ?", v)
	}
	if v := c.Query("joining_date_to"); v != "" {
		query = query.Where("joining_date <= ?", v)
	} else if v := c.Query("joining_to"); v != "" {
		query = query.Where("joining_date <= ?", v)
	} else if v := c.Query("end_date"); v != "" {
		query = query.Where("joining_date <= ?", v)
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
	sutonnyFont := &excelize.Font{Size: 11, Family: "SutonnyMJ"}
	dataBorder := []excelize.Border{
		{Type: "left", Color: "D9D9D9", Style: 1},
		{Type: "top", Color: "D9D9D9", Style: 1},
		{Type: "bottom", Color: "D9D9D9", Style: 1},
		{Type: "right", Color: "D9D9D9", Style: 1},
	}
	dataAlign := &excelize.Alignment{Vertical: "center", WrapText: true}
	centerAlign := &excelize.Alignment{Horizontal: "center", Vertical: "center"}

	styleData, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign})
	styleDataSutonny, _ := f.NewStyle(&excelize.Style{Font: sutonnyFont, Border: dataBorder, Alignment: dataAlign})
	styleDataCenter, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: centerAlign})
	styleSalary, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, CustomNumFmt: &[]string{"#,##0.00"}[0]})
	styleDate, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, CustomNumFmt: &[]string{"yyyy-mm-dd"}[0]})

	altFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F2F7FC"}}
	styleDataAlt, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: dataBorder, Alignment: dataAlign, Fill: altFill})
	styleDataSutonnyAlt, _ := f.NewStyle(&excelize.Style{Font: sutonnyFont, Border: dataBorder, Alignment: dataAlign, Fill: altFill})
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

		var s, ss, sc, sd, sn func(int) int
		if isAlt {
			s = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDataAlt)
				return 0
			}
			ss = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDataSutonnyAlt)
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
			ss = func(c int) int {
				f.SetCellStyle(sheet, colName(c)+strconv.Itoa(row), colName(c)+strconv.Itoa(row), styleDataSutonny)
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
		_ = ss
		_ = sc
		_ = sd
		_ = sn

		sv := func(c int, v string) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sc(c) }
		svl := func(c int, v string) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); s(c) }
		svs := func(c int, v string) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); ss(c) }
		svi := func(c int, v int) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sc(c) }
		svfl := func(c int, v float64) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sn(c) }
		svd := func(c int, v string) { f.SetCellValue(sheet, colName(c)+strconv.Itoa(row), v); sd(c) }

		sv(1, emp.EmployeeID)
		svl(2, emp.NameEn)
		svs(3, emp.NameBn)

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
	c.Header("Content-Disposition", "attachment; filename=\"ManPower list.xlsx\"")
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

func resolveEmployeeImagePath(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	u := strings.TrimSpace(rawURL)
	if strings.Contains(u, "/uploads/") {
		idx := strings.Index(u, "/uploads/")
		u = u[idx+1:]
	}
	u = strings.TrimPrefix(u, "/")
	if _, err := os.Stat(u); err == nil {
		return u
	}
	localPath := filepath.Join("uploads", strings.TrimPrefix(u, "uploads/"))
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	return ""
}

// ExportProfileExcel godoc
//
//	@Summary      Export employee profile to Excel
//	@Description  Export full employee profile with salary and attendance sheets to Excel
//	@Tags         Employees
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        id     path  string  true  "Employee ID"
//	@Success      200    {file}  binary
//	@Router       /employees/{id}/profile/export/excel [get]
func (h *EmployeeHandler) ExportProfileExcel(c *gin.Context) {
	id := c.Param("id")

	var emp models.Employee
	if err := database.DB.Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").
		Preload("PresentDivision").Preload("PresentDistrict").Preload("PresentUpazila").Preload("PresentUnion").
		Preload("PermanentDivision").Preload("PermanentDistrict").Preload("PermanentUpazila").Preload("PermanentUnion").
		First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	now := time.Now()
	var startYear, startMonth int
	if emp.JoiningDate.IsZero() {
		startYear = now.Year()
		startMonth = int(now.Month())
	} else {
		startYear = emp.JoiningDate.Year()
		startMonth = int(emp.JoiningDate.Month())
	}

	type MonthItem struct {
		Year      int
		Month     int
		MonthName string
	}
	var tenureMonths []MonthItem
	curYear := now.Year()
	curMonth := int(now.Month())

	for y := curYear; y >= startYear; y-- {
		mEnd := 12
		mStart := 1
		if y == curYear {
			mEnd = curMonth
		}
		if y == startYear {
			mStart = startMonth
		}
		for m := mEnd; m >= mStart; m-- {
			tenureMonths = append(tenureMonths, MonthItem{
				Year:      y,
				Month:     m,
				MonthName: time.Month(m).String(),
			})
		}
	}

	// Fetch all salaries
	var allSalaries []models.Salary
	database.DB.Where("employee_id = ?", emp.EmployeeID).
		Order("year desc, month desc").
		Find(&allSalaries)
	salaryMap := make(map[string]models.Salary)
	for _, s := range allSalaries {
		salaryMap[fmt.Sprintf("%d-%d", s.Year, s.Month)] = s
	}

	// Fetch all attendance
	type AttendanceRow struct {
		Year   int
		Month  int
		Status string
		Count  int64
	}
	var attRows []AttendanceRow
	database.DB.Model(&models.Attendance{}).
		Select("EXTRACT(YEAR FROM date) as year, EXTRACT(MONTH FROM date) as month, status, count(*) as count").
		Where("employee_id = ? AND deleted_at IS NULL", emp.EmployeeID).
		Group("EXTRACT(YEAR FROM date), EXTRACT(MONTH FROM date), status").
		Find(&attRows)

	type AttMonthStats struct {
		Present int64
		Absent  int64
		Late    int64
		Leave   int64
		Weekend int64
		Total   int64
	}
	attMap := make(map[string]*AttMonthStats)
	for _, r := range attRows {
		k := fmt.Sprintf("%d-%d", r.Year, r.Month)
		if attMap[k] == nil {
			attMap[k] = &AttMonthStats{}
		}
		st := attMap[k]
		st.Total += r.Count
		switch strings.ToLower(r.Status) {
		case "present":
			st.Present += r.Count
		case "absent":
			st.Absent += r.Count
		case "late":
			st.Late += r.Count
		case "on_leave", "leave":
			st.Leave += r.Count
		case "weekend", "holiday":
			st.Weekend += r.Count
		}
	}

	f := excelize.NewFile()

	// Styles
	sectionFont := &excelize.Font{Bold: true, Size: 11, Color: "1F4E79", Family: "Calibri"}
	sectionFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D6E4F0"}}
	labelFont := &excelize.Font{Bold: true, Size: 10, Family: "Calibri"}
	labelFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F7F9FB"}}
	dataFont := &excelize.Font{Size: 10, Family: "Calibri"}
	sutonnyFont := &excelize.Font{Size: 11, Family: "SutonnyMJ"}
	titleFont := &excelize.Font{Bold: true, Size: 15, Color: "1F4E79", Family: "Calibri"}
	thFont := &excelize.Font{Bold: true, Size: 10, Color: "FFFFFF", Family: "Calibri"}
	thFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1F4E79"}}

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
			{Type: "top", Color: "EFEFEF", Style: 1},
			{Type: "bottom", Color: "EFEFEF", Style: 1},
		},
	})
	styleData, _ := f.NewStyle(&excelize.Style{
		Font: dataFont,
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
			{Type: "top", Color: "EFEFEF", Style: 1},
			{Type: "bottom", Color: "EFEFEF", Style: 1},
		},
	})
	styleDataSutonny, _ := f.NewStyle(&excelize.Style{
		Font: sutonnyFont,
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
			{Type: "top", Color: "EFEFEF", Style: 1},
			{Type: "bottom", Color: "EFEFEF", Style: 1},
		},
	})
	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font: titleFont,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleTableHeader, _ := f.NewStyle(&excelize.Style{
		Font: thFont, Fill: thFill,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleDataCenter, _ := f.NewStyle(&excelize.Style{
		Font: dataFont,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
	})
	styleDataRight, _ := f.NewStyle(&excelize.Style{
		Font: dataFont,
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
	})

	// ==========================================
	// SHEET 1: Profile
	// ==========================================
	sheetProfile := "Profile"
	f.SetSheetName("Sheet1", sheetProfile)
	f.SetColWidth(sheetProfile, "A", "A", 28)
	f.SetColWidth(sheetProfile, "B", "B", 42)
	f.SetColWidth(sheetProfile, "C", "C", 24)
	f.SetColWidth(sheetProfile, "D", "D", 32)

	row := 1
	// Title
	compName := "PeopleHub HRM"
	if emp.Company.CompanyNameEn != "" {
		compName = emp.Company.CompanyNameEn
	}
	f.MergeCell(sheetProfile, "A1", "D1")
	f.SetCellStyle(sheetProfile, "A1", "D1", styleTitle)
	f.SetCellValue(sheetProfile, "A1", compName+" — Employee Profile")
	f.SetRowHeight(sheetProfile, 1, 32)
	row = 2

	// Employee Name & Subtitle
	desigName := ""
	if emp.DesignationRef != nil {
		desigName = emp.DesignationRef.Name
	}
	f.MergeCell(sheetProfile, "A2", "C2")
	f.SetCellValue(sheetProfile, "A2", emp.NameEn+" ("+emp.EmployeeID+") — "+desigName)
	f.SetCellStyle(sheetProfile, "A2", "C2", styleSection)
	f.SetRowHeight(sheetProfile, 2, 26)

	// Embed Photo
	imgPath := resolveEmployeeImagePath(emp.ImageURL)
	if imgPath != "" {
		if imgBytes, err := os.ReadFile(imgPath); err == nil {
			ext := strings.ToLower(filepath.Ext(imgPath))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				_ = f.AddPictureFromBytes(sheetProfile, "D2", &excelize.Picture{
					Extension: ext,
					File:      imgBytes,
					Format: &excelize.GraphicOptions{
						ScaleX:      0.28,
						ScaleY:      0.28,
						Positioning: "oneCell",
					},
				})
			}
		}
	}
	row = 4

	sv := func(r int, v string, isSutonny ...bool) {
		f.SetCellValue(sheetProfile, "B"+strconv.Itoa(r), v)
		if len(isSutonny) > 0 && isSutonny[0] {
			f.SetCellStyle(sheetProfile, "B"+strconv.Itoa(r), "B"+strconv.Itoa(r), styleDataSutonny)
		} else {
			f.SetCellStyle(sheetProfile, "B"+strconv.Itoa(r), "B"+strconv.Itoa(r), styleData)
		}
	}
	setLabel := func(r int, label string) {
		f.SetCellValue(sheetProfile, "A"+strconv.Itoa(r), label)
		f.SetCellStyle(sheetProfile, "A"+strconv.Itoa(r), "A"+strconv.Itoa(r), styleLabel)
	}
	setSection := func(r int, title string) {
		f.MergeCell(sheetProfile, "A"+strconv.Itoa(r), "D"+strconv.Itoa(r))
		f.SetCellValue(sheetProfile, "A"+strconv.Itoa(r), title)
		f.SetCellStyle(sheetProfile, "A"+strconv.Itoa(r), "D"+strconv.Itoa(r), styleSection)
		f.SetRowHeight(sheetProfile, r, 24)
	}

	// Personal Information
	setSection(row, "Personal Information / ব্যক্তিগত বিবরণ")
	row++
	setLabel(row, "Full Name (English)")
	sv(row, emp.NameEn)
	row++
	setLabel(row, "Full Name (Bangla)")
	sv(row, emp.NameBn, true)
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

	// Contact & Emergency
	setSection(row, "Contact & Emergency / যোগাযোগ ও জরুরী তথ্য")
	row++
	setLabel(row, "Phone Number")
	sv(row, emp.Phone)
	row++
	setLabel(row, "Email Address")
	sv(row, emp.Email)
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
	setLabel(row, "Number of Dependents")
	sv(row, strconv.Itoa(emp.NumberOfDependents))
	row++

	// Present Address
	setSection(row, "Present Address / বর্তমান ঠিকানা")
	row++
	setLabel(row, "Address Details (English)")
	sv(row, emp.PresentAddress)
	row++
	setLabel(row, "Address Details (Bangla)")
	sv(row, emp.PresentAddressBn, true)
	row++
	setLabel(row, "Post Office / Code")
	poCode := ""
	if emp.PresentPostOffice != nil && *emp.PresentPostOffice != "" {
		poCode = *emp.PresentPostOffice
	}
	if emp.PresentPostCode != nil && *emp.PresentPostCode != "" {
		if poCode != "" {
			poCode += " - " + *emp.PresentPostCode
		} else {
			poCode = *emp.PresentPostCode
		}
	}
	sv(row, poCode)
	row++
	if emp.PresentUnion != nil {
		setLabel(row, "Union")
		sv(row, emp.PresentUnion.Name)
		row++
	}
	if emp.PresentUpazila != nil {
		setLabel(row, "Upazila / Thana")
		sv(row, emp.PresentUpazila.Name)
		row++
	}
	if emp.PresentDistrict != nil {
		setLabel(row, "District")
		sv(row, emp.PresentDistrict.Name)
		row++
	}
	if emp.PresentDivision != nil {
		setLabel(row, "Division")
		sv(row, emp.PresentDivision.Name)
		row++
	}

	// Permanent Address
	setSection(row, "Permanent Address / স্থায়ী ঠিকানা")
	row++
	setLabel(row, "Address Details (English)")
	sv(row, emp.PermanentAddress)
	row++
	setLabel(row, "Address Details (Bangla)")
	sv(row, emp.PermanentAddressBn, true)
	row++
	setLabel(row, "Post Office / Code")
	permPoCode := ""
	if emp.PermanentPostOffice != nil && *emp.PermanentPostOffice != "" {
		permPoCode = *emp.PermanentPostOffice
	}
	if emp.PermanentPostCode != nil && *emp.PermanentPostCode != "" {
		if permPoCode != "" {
			permPoCode += " - " + *emp.PermanentPostCode
		} else {
			permPoCode = *emp.PermanentPostCode
		}
	}
	sv(row, permPoCode)
	row++
	if emp.PermanentUnion != nil {
		setLabel(row, "Union")
		sv(row, emp.PermanentUnion.Name)
		row++
	}
	if emp.PermanentUpazila != nil {
		setLabel(row, "Upazila / Thana")
		sv(row, emp.PermanentUpazila.Name)
		row++
	}
	if emp.PermanentDistrict != nil {
		setLabel(row, "District")
		sv(row, emp.PermanentDistrict.Name)
		row++
	}
	if emp.PermanentDivision != nil {
		setLabel(row, "Division")
		sv(row, emp.PermanentDivision.Name)
		row++
	}

	// Office Information
	setSection(row, "Office Information / প্রাতিষ্ঠানিক তথ্য")
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
	setLabel(row, "Over Time Allowed")
	sv(row, otStatus)
	row++

	// Bank Details
	setSection(row, "Bank Account Details / ব্যাংক তথ্য")
	row++
	setLabel(row, "Account Type")
	sv(row, emp.AccountType)
	row++
	setLabel(row, "Account Number")
	sv(row, emp.AccountNumber)
	row++

	f.SetSheetView(sheetProfile, -1, &excelize.ViewOptions{
		ShowGridLines: func(b bool) *bool { return &b }(true),
	})

	// ==========================================
	// SHEET 2: Salary History
	// ==========================================
	sheetSalary := "Salary History"
	f.NewSheet(sheetSalary)
	f.SetColWidth(sheetSalary, "A", "A", 18)
	f.SetColWidth(sheetSalary, "B", "B", 14)
	f.SetColWidth(sheetSalary, "C", "C", 14)
	f.SetColWidth(sheetSalary, "D", "D", 14)
	f.SetColWidth(sheetSalary, "E", "E", 14)
	f.SetColWidth(sheetSalary, "F", "F", 14)
	f.SetColWidth(sheetSalary, "G", "G", 14)
	f.SetColWidth(sheetSalary, "H", "H", 14)
	f.SetColWidth(sheetSalary, "I", "I", 14)
	f.SetColWidth(sheetSalary, "J", "J", 14)
	f.SetColWidth(sheetSalary, "K", "K", 14)
	f.SetColWidth(sheetSalary, "L", "L", 14)
	f.SetColWidth(sheetSalary, "M", "M", 16)
	f.SetColWidth(sheetSalary, "N", "N", 16)
	f.SetColWidth(sheetSalary, "O", "O", 15)

	f.MergeCell(sheetSalary, "A1", "O1")
	f.SetCellValue(sheetSalary, "A1", fmt.Sprintf("%s (%s) — Full Monthly Salary History", emp.NameEn, emp.EmployeeID))
	f.SetCellStyle(sheetSalary, "A1", "O1", styleTitle)
	f.SetRowHeight(sheetSalary, 1, 28)

	salaryHeaders := []string{
		"Month / Year", "Gross Salary", "Basic Salary", "House Rent", "Medical", "Transport",
		"Food", "Other", "OT Amount", "Bonus", "Absent Deduct", "Other Deduct",
		"Total Deduct", "Net Salary", "Status",
	}
	for colIdx, hText := range salaryHeaders {
		cell := colName(colIdx+1) + "2"
		f.SetCellValue(sheetSalary, cell, hText)
		f.SetCellStyle(sheetSalary, cell, cell, styleTableHeader)
	}
	f.SetRowHeight(sheetSalary, 2, 24)

	sRow := 3
	for _, m := range tenureMonths {
		sKey := fmt.Sprintf("%d-%d", m.Year, m.Month)
		s, exists := salaryMap[sKey]

		f.SetCellValue(sheetSalary, "A"+strconv.Itoa(sRow), fmt.Sprintf("%s %d", m.MonthName, m.Year))
		f.SetCellStyle(sheetSalary, "A"+strconv.Itoa(sRow), "A"+strconv.Itoa(sRow), styleDataCenter)

		if exists {
			f.SetCellValue(sheetSalary, "B"+strconv.Itoa(sRow), s.GrossSalary)
			f.SetCellValue(sheetSalary, "C"+strconv.Itoa(sRow), s.BasicSalary)
			f.SetCellValue(sheetSalary, "D"+strconv.Itoa(sRow), s.HouseRent)
			f.SetCellValue(sheetSalary, "E"+strconv.Itoa(sRow), s.MedicalAllowance)
			f.SetCellValue(sheetSalary, "F"+strconv.Itoa(sRow), s.TransportAllowance)
			f.SetCellValue(sheetSalary, "G"+strconv.Itoa(sRow), s.FoodAllowance)
			f.SetCellValue(sheetSalary, "H"+strconv.Itoa(sRow), s.OtherAllowance)
			f.SetCellValue(sheetSalary, "I"+strconv.Itoa(sRow), s.OvertimeAmount)
			f.SetCellValue(sheetSalary, "J"+strconv.Itoa(sRow), s.AttendanceBonus)
			f.SetCellValue(sheetSalary, "K"+strconv.Itoa(sRow), s.AbsentDeduction)
			f.SetCellValue(sheetSalary, "L"+strconv.Itoa(sRow), s.OtherDeduction)
			f.SetCellValue(sheetSalary, "M"+strconv.Itoa(sRow), s.TotalDeductions)
			f.SetCellValue(sheetSalary, "N"+strconv.Itoa(sRow), s.NetSalary)
			f.SetCellValue(sheetSalary, "O"+strconv.Itoa(sRow), s.Status)

			for cIdx := 2; cIdx <= 14; cIdx++ {
				cCell := colName(cIdx) + strconv.Itoa(sRow)
				f.SetCellStyle(sheetSalary, cCell, cCell, styleDataRight)
			}
			f.SetCellStyle(sheetSalary, "O"+strconv.Itoa(sRow), "O"+strconv.Itoa(sRow), styleDataCenter)
		} else {
			for cIdx := 2; cIdx <= 14; cIdx++ {
				cCell := colName(cIdx) + strconv.Itoa(sRow)
				f.SetCellValue(sheetSalary, cCell, "null")
				f.SetCellStyle(sheetSalary, cCell, cCell, styleDataCenter)
			}
			f.SetCellValue(sheetSalary, "O"+strconv.Itoa(sRow), "Not Generated")
			f.SetCellStyle(sheetSalary, "O"+strconv.Itoa(sRow), "O"+strconv.Itoa(sRow), styleDataCenter)
		}
		sRow++
	}

	// ==========================================
	// SHEET 3: Attendance History
	// ==========================================
	sheetAtt := "Attendance History"
	f.NewSheet(sheetAtt)
	f.SetColWidth(sheetAtt, "A", "A", 18)
	f.SetColWidth(sheetAtt, "B", "B", 14)
	f.SetColWidth(sheetAtt, "C", "C", 14)
	f.SetColWidth(sheetAtt, "D", "D", 14)
	f.SetColWidth(sheetAtt, "E", "E", 14)
	f.SetColWidth(sheetAtt, "F", "F", 14)
	f.SetColWidth(sheetAtt, "G", "G", 16)
	f.SetColWidth(sheetAtt, "H", "H", 16)

	f.MergeCell(sheetAtt, "A1", "H1")
	f.SetCellValue(sheetAtt, "A1", fmt.Sprintf("%s (%s) — Full Monthly Attendance History", emp.NameEn, emp.EmployeeID))
	f.SetCellStyle(sheetAtt, "A1", "H1", styleTitle)
	f.SetRowHeight(sheetAtt, 1, 28)

	attHeaders := []string{
		"Month / Year", "Total Days", "Present Days", "Absent Days",
		"Late Days", "Leave Days", "Weekend / Holiday", "Presence Rate %",
	}
	for colIdx, hText := range attHeaders {
		cell := colName(colIdx+1) + "2"
		f.SetCellValue(sheetAtt, cell, hText)
		f.SetCellStyle(sheetAtt, cell, cell, styleTableHeader)
	}
	f.SetRowHeight(sheetAtt, 2, 24)

	aRow := 3
	for _, m := range tenureMonths {
		aKey := fmt.Sprintf("%d-%d", m.Year, m.Month)
		st, exists := attMap[aKey]

		f.SetCellValue(sheetAtt, "A"+strconv.Itoa(aRow), fmt.Sprintf("%s %d", m.MonthName, m.Year))
		f.SetCellStyle(sheetAtt, "A"+strconv.Itoa(aRow), "A"+strconv.Itoa(aRow), styleDataCenter)

		if exists && st.Total > 0 {
			workingDays := st.Present + st.Absent + st.Late
			rate := 0
			if workingDays > 0 {
				rate = int(((st.Present + st.Late) * 100) / workingDays)
			}
			f.SetCellValue(sheetAtt, "B"+strconv.Itoa(aRow), st.Total)
			f.SetCellValue(sheetAtt, "C"+strconv.Itoa(aRow), st.Present)
			f.SetCellValue(sheetAtt, "D"+strconv.Itoa(aRow), st.Absent)
			f.SetCellValue(sheetAtt, "E"+strconv.Itoa(aRow), st.Late)
			f.SetCellValue(sheetAtt, "F"+strconv.Itoa(aRow), st.Leave)
			f.SetCellValue(sheetAtt, "G"+strconv.Itoa(aRow), st.Weekend)
			f.SetCellValue(sheetAtt, "H"+strconv.Itoa(aRow), fmt.Sprintf("%d%%", rate))
		} else {
			f.SetCellValue(sheetAtt, "B"+strconv.Itoa(aRow), 0)
			f.SetCellValue(sheetAtt, "C"+strconv.Itoa(aRow), 0)
			f.SetCellValue(sheetAtt, "D"+strconv.Itoa(aRow), 0)
			f.SetCellValue(sheetAtt, "E"+strconv.Itoa(aRow), 0)
			f.SetCellValue(sheetAtt, "F"+strconv.Itoa(aRow), 0)
			f.SetCellValue(sheetAtt, "G"+strconv.Itoa(aRow), 0)
			f.SetCellValue(sheetAtt, "H"+strconv.Itoa(aRow), "0%")
		}

		for cIdx := 2; cIdx <= 8; cIdx++ {
			cCell := colName(cIdx) + strconv.Itoa(aRow)
			f.SetCellStyle(sheetAtt, cCell, cCell, styleDataCenter)
		}
		aRow++
	}

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
//	@Success      200    {file}  binary
//	@Router       /employees/{id}/profile/export/pdf [get]
func (h *EmployeeHandler) ExportProfilePDF(c *gin.Context) {
	id := c.Param("id")

	var emp models.Employee
	if err := database.DB.Preload("Company").Preload("Department").Preload("Shift").Preload("SectionRef").Preload("DesignationRef").Preload("LineRef").Preload("GroupRef").Preload("FloorRef").
		Preload("PresentDivision").Preload("PresentDistrict").Preload("PresentUpazila").Preload("PresentUnion").
		Preload("PermanentDivision").Preload("PermanentDistrict").Preload("PermanentUpazila").Preload("PermanentUnion").
		First(&emp, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		return
	}

	now := time.Now()
	var startYear, startMonth int
	if emp.JoiningDate.IsZero() {
		startYear = now.Year()
		startMonth = int(now.Month())
	} else {
		startYear = emp.JoiningDate.Year()
		startMonth = int(emp.JoiningDate.Month())
	}

	type MonthItem struct {
		Year      int
		Month     int
		MonthName string
	}
	var tenureMonths []MonthItem
	curYear := now.Year()
	curMonth := int(now.Month())

	for y := curYear; y >= startYear; y-- {
		mEnd := 12
		mStart := 1
		if y == curYear {
			mEnd = curMonth
		}
		if y == startYear {
			mStart = startMonth
		}
		for m := mEnd; m >= mStart; m-- {
			tenureMonths = append(tenureMonths, MonthItem{
				Year:      y,
				Month:     m,
				MonthName: time.Month(m).String()[:3],
			})
		}
	}

	// Fetch all salaries
	var allSalaries []models.Salary
	database.DB.Where("employee_id = ?", emp.EmployeeID).
		Order("year desc, month desc").
		Find(&allSalaries)
	salaryMap := make(map[string]models.Salary)
	for _, s := range allSalaries {
		salaryMap[fmt.Sprintf("%d-%d", s.Year, s.Month)] = s
	}

	// Fetch all attendance
	type AttendanceRow struct {
		Year   int
		Month  int
		Status string
		Count  int64
	}
	var attRows []AttendanceRow
	database.DB.Model(&models.Attendance{}).
		Select("EXTRACT(YEAR FROM date) as year, EXTRACT(MONTH FROM date) as month, status, count(*) as count").
		Where("employee_id = ? AND deleted_at IS NULL", emp.EmployeeID).
		Group("EXTRACT(YEAR FROM date), EXTRACT(MONTH FROM date), status").
		Find(&attRows)

	type AttMonthStats struct {
		Present int64
		Absent  int64
		Late    int64
		Leave   int64
		Weekend int64
		Total   int64
	}
	attMap := make(map[string]*AttMonthStats)
	for _, r := range attRows {
		k := fmt.Sprintf("%d-%d", r.Year, r.Month)
		if attMap[k] == nil {
			attMap[k] = &AttMonthStats{}
		}
		st := attMap[k]
		st.Total += r.Count
		switch strings.ToLower(r.Status) {
		case "present":
			st.Present += r.Count
		case "absent":
			st.Absent += r.Count
		case "late":
			st.Late += r.Count
		case "on_leave", "leave":
			st.Leave += r.Count
		case "weekend", "holiday":
			st.Weekend += r.Count
		}
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.AddPage()

	// Header Bar
	pdf.SetFillColor(31, 78, 121)
	pdf.Rect(0, 0, 210, 36, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 17)
	pdf.SetY(6)
	compTitle := "PeopleHub HRM"
	if emp.Company.CompanyNameEn != "" {
		compTitle = emp.Company.CompanyNameEn
	}
	pdf.CellFormat(0, 8, compTitle+" — Employee Profile", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 9.5)
	pdfDesig := ""
	if emp.DesignationRef != nil {
		pdfDesig = emp.DesignationRef.Name
	}
	pdf.CellFormat(0, 6, emp.NameEn+" ("+emp.EmployeeID+") • "+pdfDesig, "", 1, "C", false, 0, "")

	// Embed Photo on Right if exists
	imgPath := resolveEmployeeImagePath(emp.ImageURL)
	if imgPath != "" {
		if _, err := os.Stat(imgPath); err == nil {
			pdf.ImageOptions(imgPath, 168, 5, 26, 26, false, gofpdf.ImageOptions{}, 0, "")
		}
	}

	pdf.SetY(40)
	pdf.SetTextColor(0, 0, 0)

	fieldRow := func(label, value string) {
		pdf.SetX(12)
		pdf.SetFont("Arial", "B", 8.5)
		pdf.SetFillColor(245, 248, 252)
		pdf.CellFormat(55, 6, label, "1", 0, "L", true, 0, "")
		pdf.SetFont("Arial", "", 8.5)
		pdf.CellFormat(0, 6, " "+value, "1", 1, "L", false, 0, "")
	}

	sectionHeader := func(title string) {
		if pdf.GetY() > 260 {
			pdf.AddPage()
		}
		pdf.Ln(2)
		pdf.SetX(12)
		pdf.SetFont("Arial", "B", 9.5)
		pdf.SetFillColor(31, 78, 121)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(0, 6.5, "  "+title, "1", 1, "L", true, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	// 1. Personal Info
	sectionHeader("Personal Information")
	fieldRow("Full Name (English)", emp.NameEn)
	fieldRow("Father's Name", emp.FatherName)
	fieldRow("Mother's Name", emp.MotherName)
	fieldRow("Date of Birth", emp.DateOfBirth)
	fieldRow("Gender / Blood Group", emp.Gender+" • Blood Group: "+emp.BloodGroup)
	fieldRow("Marital Status / Religion", emp.MaritalStatus+" • Religion: "+emp.Religion)
	fieldRow("Nationality / NID", emp.Nationality+" • NID: "+emp.NID)

	// 2. Contact & Emergency
	sectionHeader("Contact & Emergency")
	fieldRow("Phone / Email", emp.Phone+" • "+emp.Email)
	fieldRow("Spouse Name", emp.SpouseName)
	fieldRow("Emergency Contact / Phone", emp.EmergencyContact+" • "+emp.EmergencyPhone)
	fieldRow("Dependents", strconv.Itoa(emp.NumberOfDependents))

	// 3. Addresses
	sectionHeader("Present Address")
	fieldRow("Address Details", emp.PresentAddress)
	poPres := ""
	if emp.PresentPostOffice != nil && *emp.PresentPostOffice != "" {
		poPres = *emp.PresentPostOffice
	}
	if emp.PresentPostCode != nil && *emp.PresentPostCode != "" {
		poPres += " (Code: " + *emp.PresentPostCode + ")"
	}
	fieldRow("Post Office / Code", poPres)
	presAdmin := ""
	if emp.PresentUnion != nil {
		presAdmin += "Union: " + emp.PresentUnion.Name + " • "
	}
	if emp.PresentUpazila != nil {
		presAdmin += "Upazila: " + emp.PresentUpazila.Name + " • "
	}
	if emp.PresentDistrict != nil {
		presAdmin += "District: " + emp.PresentDistrict.Name + " • "
	}
	if emp.PresentDivision != nil {
		presAdmin += "Div: " + emp.PresentDivision.Name
	}
	fieldRow("Administrative Region", presAdmin)

	sectionHeader("Permanent Address")
	fieldRow("Address Details", emp.PermanentAddress)
	poPerm := ""
	if emp.PermanentPostOffice != nil && *emp.PermanentPostOffice != "" {
		poPerm = *emp.PermanentPostOffice
	}
	if emp.PermanentPostCode != nil && *emp.PermanentPostCode != "" {
		poPerm += " (Code: " + *emp.PermanentPostCode + ")"
	}
	fieldRow("Post Office / Code", poPerm)
	permAdmin := ""
	if emp.PermanentUnion != nil {
		permAdmin += "Union: " + emp.PermanentUnion.Name + " • "
	}
	if emp.PermanentUpazila != nil {
		permAdmin += "Upazila: " + emp.PermanentUpazila.Name + " • "
	}
	if emp.PermanentDistrict != nil {
		permAdmin += "District: " + emp.PermanentDistrict.Name + " • "
	}
	if emp.PermanentDivision != nil {
		permAdmin += "Div: " + emp.PermanentDivision.Name
	}
	fieldRow("Administrative Region", permAdmin)

	// 4. Employment & Bank
	sectionHeader("Office & Employment Information")
	fieldRow("Employee ID / Punch No", emp.EmployeeID+" • Punch: "+emp.PunchNumber)
	fieldRow("Employee Type / Grade", emp.EmployeeType+" • Grade: "+emp.Grade)
	fieldRow("Joining Date", emp.JoiningDate.Format("2006-01-02"))
	deptSec := ""
	if emp.Department != nil {
		deptSec += "Dept: " + emp.Department.Name + " • "
	}
	if emp.SectionRef != nil {
		deptSec += "Sec: " + emp.SectionRef.Name + " • "
	}
	if emp.DesignationRef != nil {
		deptSec += "Desig: " + emp.DesignationRef.Name
	}
	fieldRow("Department / Designation", deptSec)
	lineGroup := ""
	if emp.LineRef != nil {
		lineGroup += "Line: " + emp.LineRef.Name + " • "
	}
	if emp.GroupRef != nil {
		lineGroup += "Group: " + emp.GroupRef.Name + " • "
	}
	if emp.FloorRef != nil {
		lineGroup += "Floor: " + emp.FloorRef.Name
	}
	fieldRow("Line / Group / Floor", lineGroup)
	otStatusText := "No"
	if emp.OverTimeStatus {
		otStatusText = "Yes"
	}
	shiftText := ""
	if emp.Shift != nil {
		shiftText = emp.Shift.Name
	}
	fieldRow("Shift / Over Time", "Shift: "+shiftText+" • OT Allowed: "+otStatusText)
	fieldRow("Bank Account", "Type: "+emp.AccountType+" • A/C No: "+emp.AccountNumber)

	// ==========================================
	// Page 2: Salary & Attendance Timeline Tables
	// ==========================================
	pdf.AddPage()

	// Header for Page 2
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(31, 78, 121)
	pdf.CellFormat(0, 7, "Monthly Salary History Timeline", "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	// Salary Table Header
	pdf.SetFont("Arial", "B", 7.5)
	pdf.SetFillColor(31, 78, 121)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(24, 6, "Month", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 6, "Gross", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 6, "Basic", "1", 0, "R", true, 0, "")
	pdf.CellFormat(26, 6, "Allowances", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 6, "OT Amount", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 6, "Deductions", "1", 0, "R", true, 0, "")
	pdf.CellFormat(26, 6, "Net Salary", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 6, "Status", "1", 1, "C", true, 0, "")
	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("Arial", "", 7.5)
	for _, m := range tenureMonths {
		if pdf.GetY() > 270 {
			pdf.AddPage()
			// repeat header
			pdf.SetFont("Arial", "B", 7.5)
			pdf.SetFillColor(31, 78, 121)
			pdf.SetTextColor(255, 255, 255)
			pdf.CellFormat(24, 6, "Month", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 6, "Gross", "1", 0, "R", true, 0, "")
			pdf.CellFormat(22, 6, "Basic", "1", 0, "R", true, 0, "")
			pdf.CellFormat(26, 6, "Allowances", "1", 0, "R", true, 0, "")
			pdf.CellFormat(22, 6, "OT Amount", "1", 0, "R", true, 0, "")
			pdf.CellFormat(22, 6, "Deductions", "1", 0, "R", true, 0, "")
			pdf.CellFormat(26, 6, "Net Salary", "1", 0, "R", true, 0, "")
			pdf.CellFormat(22, 6, "Status", "1", 1, "C", true, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 7.5)
		}

		sKey := fmt.Sprintf("%d-%d", m.Year, m.Month)
		s, exists := salaryMap[sKey]
		mLabel := fmt.Sprintf("%s %d", m.MonthName, m.Year)

		if exists {
			totAllow := s.HouseRent + s.MedicalAllowance + s.TransportAllowance + s.FoodAllowance + s.OtherAllowance
			pdf.CellFormat(24, 5.5, mLabel, "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, fmt.Sprintf("%.0f", s.GrossSalary), "1", 0, "R", false, 0, "")
			pdf.CellFormat(22, 5.5, fmt.Sprintf("%.0f", s.BasicSalary), "1", 0, "R", false, 0, "")
			pdf.CellFormat(26, 5.5, fmt.Sprintf("%.0f", totAllow), "1", 0, "R", false, 0, "")
			pdf.CellFormat(22, 5.5, fmt.Sprintf("%.0f", s.OvertimeAmount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(22, 5.5, fmt.Sprintf("-%.0f", s.TotalDeductions), "1", 0, "R", false, 0, "")
			pdf.SetFont("Arial", "B", 7.5)
			pdf.CellFormat(26, 5.5, fmt.Sprintf("%.0f", s.NetSalary), "1", 0, "R", false, 0, "")
			pdf.SetFont("Arial", "", 7.5)
			pdf.CellFormat(22, 5.5, s.Status, "1", 1, "C", false, 0, "")
		} else {
			pdf.CellFormat(24, 5.5, mLabel, "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "null", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "null", "1", 0, "C", false, 0, "")
			pdf.CellFormat(26, 5.5, "null", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "null", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "null", "1", 0, "C", false, 0, "")
			pdf.CellFormat(26, 5.5, "null", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "Unprocessed", "1", 1, "C", false, 0, "")
		}
	}

	// Attendance Table Section
	if pdf.GetY() > 220 {
		pdf.AddPage()
	} else {
		pdf.Ln(4)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(31, 78, 121)
	pdf.CellFormat(0, 7, "Monthly Attendance Summary Timeline", "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("Arial", "B", 7.5)
	pdf.SetFillColor(31, 78, 121)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(26, 6, "Month", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 6, "Total Days", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 6, "Present", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 6, "Absent", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 6, "Late", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 6, "Leave", "1", 0, "C", true, 0, "")
	pdf.CellFormat(26, 6, "Weekend", "1", 0, "C", true, 0, "")
	pdf.CellFormat(24, 6, "Rate %", "1", 1, "C", true, 0, "")
	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("Arial", "", 7.5)
	for _, m := range tenureMonths {
		if pdf.GetY() > 270 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 7.5)
			pdf.SetFillColor(31, 78, 121)
			pdf.SetTextColor(255, 255, 255)
			pdf.CellFormat(26, 6, "Month", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 6, "Total Days", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 6, "Present", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 6, "Absent", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 6, "Late", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 6, "Leave", "1", 0, "C", true, 0, "")
			pdf.CellFormat(26, 6, "Weekend", "1", 0, "C", true, 0, "")
			pdf.CellFormat(24, 6, "Rate %", "1", 1, "C", true, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 7.5)
		}

		aKey := fmt.Sprintf("%d-%d", m.Year, m.Month)
		st, exists := attMap[aKey]
		mLabel := fmt.Sprintf("%s %d", m.MonthName, m.Year)

		if exists && st.Total > 0 {
			workingDays := st.Present + st.Absent + st.Late
			rate := 0
			if workingDays > 0 {
				rate = int(((st.Present + st.Late) * 100) / workingDays)
			}
			pdf.CellFormat(26, 5.5, mLabel, "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, strconv.FormatInt(st.Total, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, strconv.FormatInt(st.Present, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, strconv.FormatInt(st.Absent, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, strconv.FormatInt(st.Late, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, strconv.FormatInt(st.Leave, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(26, 5.5, strconv.FormatInt(st.Weekend, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(24, 5.5, fmt.Sprintf("%d%%", rate), "1", 1, "C", false, 0, "")
		} else {
			pdf.CellFormat(26, 5.5, mLabel, "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "0", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "0", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "0", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "0", "1", 0, "C", false, 0, "")
			pdf.CellFormat(22, 5.5, "0", "1", 0, "C", false, 0, "")
			pdf.CellFormat(26, 5.5, "0", "1", 0, "C", false, 0, "")
			pdf.CellFormat(24, 5.5, "0%", "1", 1, "C", false, 0, "")
		}
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=profile_%s.pdf", emp.EmployeeID))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}
