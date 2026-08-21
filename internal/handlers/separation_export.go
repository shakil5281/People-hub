package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"github.com/xuri/excelize/v2"
)

// ExportListExcel godoc
//
//	@Summary      Export separations list to Excel
//	@Description  Export filtered list of employee separations to Excel file
//	@Tags         Separations
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        employee       query string false "Filter by employee name"
//	@Param        employee_id    query string false "Filter by employee ID"
//	@Param        department_id  query string false "Filter by department"
//	@Param        type           query string false "Filter by type (Resign, Lefty, Close)"
//	@Param        status         query string false "Filter by status (Pending, Approved, Processed, Cancelled)"
//	@Param        company_id     query string false "Filter by company"
//	@Param        section_id     query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id        query string false "Filter by line"
//	@Param        group_id       query string false "Filter by group"
//	@Param        date_from      query string false "Filter by start date (YYYY-MM-DD)"
//	@Param        date_to        query string false "Filter by end date (YYYY-MM-DD)"
//	@Success      200            {file} binary
//	@Failure      500            {object} map[string]string
//	@Router       /separations/export/excel [get]
func (h *SeparationHandler) ExportListExcel(c *gin.Context) {
	employee := c.Query("employee")
	employeeID := c.Query("employee_id")
	departmentID := c.Query("department_id")
	sepType := c.Query("type")
	status := c.Query("status")
	companyID := c.Query("company_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	seps, err := h.repo.ListAllFiltered(employee, employeeID, departmentID, sepType, status, companyID, sectionID, designationID, lineID, groupID, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if companyID != "" {
		_ = database.DB.Where("id = ? AND deleted_at IS NULL", companyID).First(&company)
	}
	if company.ID == "" {
		_ = database.DB.Where("deleted_at IS NULL").First(&company)
	}

	periodStr := "All Time"
	if dateFrom != "" && dateTo != "" {
		periodStr = fmt.Sprintf("Period: %s to %s", dateFrom, dateTo)
	} else if dateFrom != "" {
		periodStr = fmt.Sprintf("From: %s", dateFrom)
	} else if dateTo != "" {
		periodStr = fmt.Sprintf("Up to: %s", dateTo)
	}

	empDesigMap := make(map[string]string)
	var empDesigs []struct {
		EmployeeID      string
		DesignationName string
	}
	database.DB.Table("employees").
		Select("employees.employee_id, designations.name as designation_name").
		Joins("LEFT JOIN designations ON designations.id = employees.designation_id").
		Where("employees.deleted_at IS NULL").
		Scan(&empDesigs)
	for _, ed := range empDesigs {
		empDesigMap[ed.EmployeeID] = ed.DesignationName
	}

	renderSeparationExcel(c, company, periodStr, seps, empDesigMap)
}

// ExportListPDF godoc
//
//	@Summary      Export separations list report to PDF
//	@Description  Export filtered list of employee separations to PDF report
//	@Tags         Separations
//	@Security     BearerAuth
//	@Produce      application/pdf
//	@Param        employee       query string false "Filter by employee name"
//	@Param        employee_id    query string false "Filter by employee ID"
//	@Param        department_id  query string false "Filter by department"
//	@Param        type           query string false "Filter by type (Resign, Lefty, Close)"
//	@Param        status         query string false "Filter by status (Pending, Approved, Processed, Cancelled)"
//	@Param        company_id     query string false "Filter by company"
//	@Param        section_id     query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id        query string false "Filter by line"
//	@Param        group_id       query string false "Filter by group"
//	@Param        date_from      query string false "Filter by start date (YYYY-MM-DD)"
//	@Param        date_to        query string false "Filter by end date (YYYY-MM-DD)"
//	@Param        lang           query string false "Language: en (default) or bn"
//	@Success      200            {file} binary
//	@Failure      500            {object} map[string]string
//	@Router       /separations/export/pdf [get]
func (h *SeparationHandler) ExportListPDF(c *gin.Context) {
	employee := c.Query("employee")
	employeeID := c.Query("employee_id")
	departmentID := c.Query("department_id")
	sepType := c.Query("type")
	status := c.Query("status")
	companyID := c.Query("company_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	lang := c.DefaultQuery("lang", "en")

	isBn := strings.ToLower(lang) == "bn" || strings.ToLower(lang) == "bangla"

	seps, err := h.repo.ListAllFiltered(employee, employeeID, departmentID, sepType, status, companyID, sectionID, designationID, lineID, groupID, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if companyID != "" {
		_ = database.DB.Where("id = ? AND deleted_at IS NULL", companyID).First(&company)
	}
	if company.ID == "" {
		_ = database.DB.Where("deleted_at IS NULL").First(&company)
	}

	periodStr := "All Time"
	if dateFrom != "" && dateTo != "" {
		periodStr = fmt.Sprintf("Period: %s to %s", dateFrom, dateTo)
	} else if dateFrom != "" {
		periodStr = fmt.Sprintf("From: %s", dateFrom)
	} else if dateTo != "" {
		periodStr = fmt.Sprintf("Up to: %s", dateTo)
	}

	empDesigMap := make(map[string]string)
	var empDesigs []struct {
		EmployeeID      string
		DesignationName string
	}
	database.DB.Table("employees").
		Select("employees.employee_id, designations.name as designation_name").
		Joins("LEFT JOIN designations ON designations.id = employees.designation_id").
		Where("employees.deleted_at IS NULL").
		Scan(&empDesigs)
	for _, ed := range empDesigs {
		empDesigMap[ed.EmployeeID] = ed.DesignationName
	}

	renderSeparationPDF(c, company, periodStr, seps, empDesigMap, isBn)
}

func renderSeparationExcel(c *gin.Context, company models.Company, periodStr string, seps []models.Separation, empDesigMap map[string]string) {
	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "EKUSHE FASHIONS LTD."
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Masterbari, Gazipur City, Gazipur."
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Separation Report"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"SL", "Employee ID", "Employee Name", "Department", "Designation", "Separation Type", "Effective Date", "Status", "Reason"}
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	lastCol := "I"

	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 25)
	f.SetColWidth(sheet, "D", "D", 20)
	f.SetColWidth(sheet, "E", "E", 22)
	f.SetColWidth(sheet, "F", "F", 16)
	f.SetColWidth(sheet, "G", "G", 15)
	f.SetColWidth(sheet, "H", "H", 14)
	f.SetColWidth(sheet, "I", "I", 30)

	borderColor := "262626"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9E2F3"}, Pattern: 1},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	cellNormal, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	cellLeft, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	headerRuns := []excelize.RichTextRun{
		{
			Text: companyName,
			Font: &excelize.Font{Bold: true, Size: 20},
		},
		{
			Text: "\n" + companyAddress,
			Font: &excelize.Font{Size: 11},
		},
		{
			Text: "\nEMPLOYEE SEPARATION REPORT",
			Font: &excelize.Font{Bold: true, Size: 11, Color: "DC2626"},
		},
		{
			Text: "\n" + periodStr,
			Font: &excelize.Font{Size: 11},
		},
	}

	_ = f.SetCellRichText(sheet, "A1", headerRuns)
	_ = f.MergeCell(sheet, "A1", lastCol+"1")

	headerStyleA1, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	})
	_ = f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyleA1)
	_ = f.SetRowHeight(sheet, 1, 75)

	headerRow := 2
	f.SetRowHeight(sheet, headerRow, 30)

	for i, h := range headers {
		cell := cols[i] + strconv.Itoa(headerRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	currentRow := headerRow + 1
	for idx, s := range seps {
		rStr := strconv.Itoa(currentRow)
		f.SetRowHeight(sheet, currentRow, 22)

		deptName := s.Department.Name
		desigName := empDesigMap[s.EmployeeID]

		f.SetCellValue(sheet, "A"+rStr, idx+1)
		f.SetCellValue(sheet, "B"+rStr, s.EmployeeID)
		f.SetCellValue(sheet, "C"+rStr, s.Employee)
		f.SetCellValue(sheet, "D"+rStr, deptName)
		f.SetCellValue(sheet, "E"+rStr, desigName)
		f.SetCellValue(sheet, "F"+rStr, s.Type)
		f.SetCellValue(sheet, "G"+rStr, s.Date)
		f.SetCellValue(sheet, "H"+rStr, s.Status)
		f.SetCellValue(sheet, "I"+rStr, s.Reason)

		f.SetCellStyle(sheet, "A"+rStr, "A"+rStr, cellNormal)
		f.SetCellStyle(sheet, "B"+rStr, "B"+rStr, cellNormal)
		f.SetCellStyle(sheet, "C"+rStr, "C"+rStr, cellLeft)
		f.SetCellStyle(sheet, "D"+rStr, "D"+rStr, cellLeft)
		f.SetCellStyle(sheet, "E"+rStr, "E"+rStr, cellLeft)
		f.SetCellStyle(sheet, "F"+rStr, "F"+rStr, cellNormal)
		f.SetCellStyle(sheet, "G"+rStr, "G"+rStr, cellNormal)
		f.SetCellStyle(sheet, "H"+rStr, "H"+rStr, cellNormal)
		f.SetCellStyle(sheet, "I"+rStr, "I"+rStr, cellLeft)

		currentRow++
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=separation_report_%s.xlsx", time.Now().Format("20060102_150405")))

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func renderSeparationPDF(c *gin.Context, company models.Company, periodStr string, seps []models.Separation, empDesigMap map[string]string, isBn bool) {
	companyName := companyDisplayName(company, isBn)
	companyAddress := company.AddressEn
	if isBn && company.AddressBn != "" {
		companyAddress = utils.UnicodeToBijoy(company.AddressBn)
	}
	if companyAddress == "" {
		companyAddress = "Masterbari, Gazipur City, Gazipur."
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(true, 10)

	font := "Arial"
	if isBn {
		font = loadBanglaFont(pdf)
	}

	pdf.AddPage()

	pdf.SetFont(font, "B", 14)
	pdf.SetTextColor(18, 58, 99)
	pdf.CellFormat(281, 7, companyName, "", 1, "C", false, 0, "")

	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(281, 5, companyAddress, "", 1, "C", false, 0, "")

	pdf.SetFont(font, "B", 11)
	pdf.SetTextColor(220, 38, 38)
	reportTitle := "EMPLOYEE SEPARATION REPORT"
	if isBn {
		reportTitle = utils.UnicodeToBijoy("কর্মচারী পদত্যাগ / অব্যাহতির রিপোর্ট")
	}
	pdf.CellFormat(281, 6, reportTitle, "", 1, "C", false, 0, "")

	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(281, 5, periodStr, "", 1, "C", false, 0, "")
	pdf.Ln(3)

	wSL := 10.0
	wEmpID := 24.0
	wEmpName := 48.0
	wDept := 36.0
	wDesig := 36.0
	wType := 22.0
	wDate := 22.0
	wStatus := 22.0
	wReason := 61.0

	pdf.SetFillColor(217, 226, 243)
	pdf.SetDrawColor(38, 38, 38)
	pdf.SetLineWidth(0.3)
	pdf.SetFont(font, "B", 9)
	pdf.SetTextColor(0, 0, 0)

	pdf.CellFormat(wSL, 7, "SL", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wEmpID, 7, "Emp ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wEmpName, 7, "Employee Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wDept, 7, "Department", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wDesig, 7, "Designation", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wType, 7, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wDate, 7, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wStatus, 7, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wReason, 7, "Reason", "1", 1, "C", true, 0, "")

	pdf.SetFont(font, "", 8.5)
	pdf.SetTextColor(20, 20, 20)

	for idx, s := range seps {
		deptName := s.Department.Name
		desigName := empDesigMap[s.EmployeeID]

		empName := s.Employee
		if isBn {
			deptName = utils.UnicodeToBijoy(deptName)
			desigName = utils.UnicodeToBijoy(desigName)
			empName = utils.UnicodeToBijoy(empName)
		}

		pdf.CellFormat(wSL, 6.5, strconv.Itoa(idx+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wEmpID, 6.5, s.EmployeeID, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wEmpName, 6.5, truncateString(empName, 26), "1", 0, "L", false, 0, "")
		pdf.CellFormat(wDept, 6.5, truncateString(deptName, 20), "1", 0, "L", false, 0, "")
		pdf.CellFormat(wDesig, 6.5, truncateString(desigName, 20), "1", 0, "L", false, 0, "")
		pdf.CellFormat(wType, 6.5, s.Type, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wDate, 6.5, s.Date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wStatus, 6.5, s.Status, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wReason, 6.5, truncateString(s.Reason, 35), "1", 1, "L", false, 0, "")
	}

	pdf.AliasNbPages("{nb}")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-10)
		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(281, 5, fmt.Sprintf("Page %d of {nb}   |   Print Date: %s", pdf.PageNo(), time.Now().Format("02-01-2006 15:04")), "", 0, "C", false, 0, "")
	})

	if pdf.Error() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": pdf.Error().Error()})
		return
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=separation_report_%s.pdf", time.Now().Format("20060102_150405")))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
