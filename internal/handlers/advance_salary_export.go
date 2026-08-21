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
// @Summary      Export advance salary list to Excel
// @Description  Download advance salary report as Excel file
// @Tags         Salary
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Router       /salary/advances/export/excel [get]
func (h *AdvanceSalaryHandler) ExportExcel(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	advances, err := h.advanceRepo.List(repository.AdvanceFilter{
		CompanyID:     companyID,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
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
	sheetName := "Advances"
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

	f.MergeCell(sheetName, "A1", "G1")
	f.SetCellValue(sheetName, "A1", company.CompanyNameEn)
	f.SetCellStyle(sheetName, "A1", "A1", styleTitle)

	f.MergeCell(sheetName, "A2", "G2")
	f.SetCellValue(sheetName, "A2", company.AddressEn)
	f.SetCellStyle(sheetName, "A2", "A2", styleSubtitle)

	f.MergeCell(sheetName, "A3", "G3")
	f.SetCellValue(sheetName, "A3", "Advance Salary Report")
	f.SetCellStyle(sheetName, "A3", "A3", styleSubtitle)

	f.MergeCell(sheetName, "A4", "G4")
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Report Date: %s", time.Now().Format("02-01-2006")))
	f.SetCellStyle(sheetName, "A4", "A4", styleSubtitle)

	headers := []string{"Employee ID", "Name", "Designation", "Department", "Amount", "Advance Date", "Status"}
	row := 6
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styleHeader)
	}

	for _, adv := range advances {
		row++
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), adv.EmployeeID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), adv.Employee.NameEn)
		if adv.Employee.DesignationRef != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), adv.Employee.DesignationRef.Name)
		}
		if adv.Employee.Department != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), adv.Employee.Department.Name)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), adv.Amount)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), utils.FormatBillDate(adv.AdvanceDate))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), adv.Status)
	}

	row += 4
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row+1), "Prepared By")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row+1), "Checked By")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", row+1), "Authorized By")

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=advances_%s.xlsx", time.Now().Format("20060102_150405")))
	f.Write(c.Writer)
}

// ExportPDF godoc
// @Summary      Export advance salary list to PDF
// @Router       /salary/advances/export/pdf [get]
func (h *AdvanceSalaryHandler) ExportPDF(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	advances, err := h.advanceRepo.List(repository.AdvanceFilter{
		CompanyID:     companyID,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
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

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 14)
		pdf.CellFormat(190, 8, company.CompanyNameEn, "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(190, 6, company.AddressEn, "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(190, 8, "Advance Salary Report", "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(190, 6, fmt.Sprintf("Report Date: %s", time.Now().Format("02-01-2006")), "", 1, "C", false, 0, "")
		pdf.Ln(5)

		pdf.SetFont("Arial", "B", 9)
		headers := []string{"Emp ID", "Name", "Designation", "Amount", "Date", "Status"}
		widths := []float64{25, 45, 45, 25, 25, 25}
		for i, str := range headers {
			pdf.CellFormat(widths[i], 8, str, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-30)
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(63, 10, "__________________", "", 0, "C", false, 0, "")
		pdf.CellFormat(63, 10, "__________________", "", 0, "C", false, 0, "")
		pdf.CellFormat(63, 10, "__________________", "", 1, "C", false, 0, "")
		pdf.CellFormat(63, 5, "Prepared By", "", 0, "C", false, 0, "")
		pdf.CellFormat(63, 5, "Checked By", "", 0, "C", false, 0, "")
		pdf.CellFormat(63, 5, "Authorized By", "", 1, "C", false, 0, "")
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont("Arial", "", 9)
	widths := []float64{25, 45, 45, 25, 25, 25}

	for _, adv := range advances {
		name := adv.Employee.NameEn
		designation := "-"
		if adv.Employee.DesignationRef != nil {
			designation = adv.Employee.DesignationRef.Name
		}
		pdf.CellFormat(widths[0], 8, adv.EmployeeID, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[1], 8, name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], 8, designation, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[3], 8, fmt.Sprintf("%.2f", adv.Amount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[4], 8, utils.FormatBillDate(adv.AdvanceDate), "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[5], 8, adv.Status, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=advances_%s.pdf", time.Now().Format("20060102_150405")))
	pdf.Output(c.Writer)
}
