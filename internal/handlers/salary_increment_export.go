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
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"github.com/xuri/excelize/v2"
)

// ExportExcel godoc
// @Summary      Export increment report to Excel
// @Description  Download salary increment report as Excel file
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        company_id     query string true  "Company ID"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        month          query int    false "Month (1-12)"
// @Param        year           query int    false "Year"
// @Param        status         query string false "Filter by status"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments/export/excel [get]
func (h *SalaryIncrementHandler) ExportExcel(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	increments, err := h.incrementRepo.List(repository.IncrementFilter{
		CompanyID:     companyID,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		IncrementType: c.Query("increment_type"),
		Month:         month,
		Year:          year,
		Status:        c.Query("status"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	_ = database.DB.First(&company, "id = ?", companyID).Error

	f := excelize.NewFile()
	sheetName := "Increments"
	f.SetSheetName("Sheet1", sheetName)

	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	styleSubtitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 25)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 15)
	f.SetColWidth(sheetName, "F", "F", 18)
	f.SetColWidth(sheetName, "G", "G", 15)
	f.SetColWidth(sheetName, "H", "H", 15)
	f.SetColWidth(sheetName, "I", "I", 15)
	f.SetColWidth(sheetName, "J", "J", 15)

	f.MergeCell(sheetName, "A1", "J1")
	f.SetCellValue(sheetName, "A1", company.CompanyNameEn)
	f.SetCellStyle(sheetName, "A1", "A1", styleTitle)

	f.MergeCell(sheetName, "A2", "J2")
	f.SetCellValue(sheetName, "A2", company.AddressEn)
	f.SetCellStyle(sheetName, "A2", "A2", styleSubtitle)

	f.MergeCell(sheetName, "A3", "J3")
	f.SetCellValue(sheetName, "A3", "Salary Increments Report")
	f.SetCellStyle(sheetName, "A3", "A3", styleSubtitle)

	f.MergeCell(sheetName, "A4", "J4")
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Report Date: %s", time.Now().Format("02-01-2006")))
	f.SetCellStyle(sheetName, "A4", "A4", styleSubtitle)

	headers := []string{"Employee ID", "Name", "Designation", "Department", "Current Gross", "Increment Amount", "New Gross", "Increment Date", "Effective Date", "Status"}
	row := 6
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styleHeader)
	}

	for _, inc := range increments {
		row++
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), inc.EmployeeID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), inc.Employee.NameEn)
		if inc.Employee.DesignationRef != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), inc.Employee.DesignationRef.Name)
		}
		if inc.Employee.Department != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), inc.Employee.Department.Name)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), inc.PreviousGross)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), inc.IncrementAmount)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), inc.NewGross)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), utils.FormatBillDate(inc.IncrementDate))
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), utils.FormatBillDate(inc.EffectiveDate))
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), inc.Status)
	}

	row += 4
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row+1), "Prepared By")

	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row+1), "Checked By")

	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row+1), "Authorized By")

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=increments_%s.xlsx", time.Now().Format("20060102_150405")))
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
	}
}

// ExportPDF godoc
// @Summary      Export increment report to PDF
// @Description  Download salary increment report as PDF file
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        company_id     query string true  "Company ID"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        month          query int    false "Month (1-12)"
// @Param        year           query int    false "Year"
// @Param        status         query string false "Filter by status"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /salary/increments/export/pdf [get]
func (h *SalaryIncrementHandler) ExportPDF(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	increments, err := h.incrementRepo.List(repository.IncrementFilter{
		CompanyID:     companyID,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		IncrementType: c.Query("increment_type"),
		Month:         month,
		Year:          year,
		Status:        c.Query("status"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	_ = database.DB.First(&company, "id = ?", companyID).Error

	pdf := gofpdf.New("L", "mm", "A4", "")

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 14)
		pdf.CellFormat(280, 8, company.CompanyNameEn, "", 1, "C", false, 0, "")
		
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(280, 6, company.AddressEn, "", 1, "C", false, 0, "")
		
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(280, 8, "Salary Increments Report", "", 1, "C", false, 0, "")
		
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(280, 6, fmt.Sprintf("Report Date: %s", time.Now().Format("02-01-2006")), "", 1, "C", false, 0, "")
		pdf.Ln(5)

		pdf.SetFont("Arial", "B", 9)
		headers := []string{"Emp ID", "Name", "Designation", "Current Gross", "Inc. Amount", "New Gross", "Eff. Date", "Status"}
		widths := []float64{20, 45, 45, 30, 30, 30, 30, 20}

		for i, str := range headers {
			pdf.CellFormat(widths[i], 8, str, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-30)
		pdf.SetFont("Arial", "", 9)
		
		// Signatures
		pdf.CellFormat(93, 10, "_________________________", "", 0, "C", false, 0, "")
		pdf.CellFormat(93, 10, "_________________________", "", 0, "C", false, 0, "")
		pdf.CellFormat(93, 10, "_________________________", "", 1, "C", false, 0, "")
		
		pdf.CellFormat(93, 5, "Prepared By", "", 0, "C", false, 0, "")
		pdf.CellFormat(93, 5, "Checked By", "", 0, "C", false, 0, "")
		pdf.CellFormat(93, 5, "Authorized By", "", 1, "C", false, 0, "")

		// Page number
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont("Arial", "", 9)
	widths := []float64{20, 45, 45, 30, 30, 30, 30, 20}

	for _, inc := range increments {
		name := inc.Employee.NameEn
		designation := "-"
		if inc.Employee.DesignationRef != nil {
			designation = inc.Employee.DesignationRef.Name
		}

		pdf.CellFormat(widths[0], 8, inc.EmployeeID, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[1], 8, name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], 8, designation, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[3], 8, fmt.Sprintf("%.2f", inc.PreviousGross), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[4], 8, fmt.Sprintf("%.2f", inc.IncrementAmount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[5], 8, fmt.Sprintf("%.2f", inc.NewGross), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[6], 8, utils.FormatBillDate(inc.EffectiveDate), "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[7], 8, inc.Status, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=increments_%s.pdf", time.Now().Format("20060102_150405")))
	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	}
}
