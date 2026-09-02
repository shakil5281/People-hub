package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
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

	// Page setup - landscape A4, fit to width, margins
	orientation := "landscape"
	paperSize := 9 // A4
	fitW := 1
	fitH := 0
	_ = f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
		Orientation: &orientation,
		Size:        &paperSize,
		FitToWidth:  &fitW,
		FitToHeight: &fitH,
	})
	left := 0.4
	right := 0.4
	top := 0.5
	bottom := 0.5
	_ = f.SetPageMargins(sheetName, &excelize.PageLayoutMarginsOptions{
		Left: &left, Right: &right, Top: &top, Bottom: &bottom,
		Header: func(v float64) *float64 { return &v }(0.2),
		Footer: func(v float64) *float64 { return &v }(0.2),
	})
	showGrid := false
	_ = f.SetSheetView(sheetName, -1, &excelize.ViewOptions{ShowGridLines: &showGrid})
	_ = f.SetHeaderFooter(sheetName, &excelize.HeaderFooterOptions{
		OddFooter: "&LProduction manager&CAdmin (A.G.M)&RApproved By",
	})

	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleSubtitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleCenter, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleLeft, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	styleSection, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"5B9BD5"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleTotal, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFF2CC"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleTotalLeft, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFF2CC"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleGrand, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleGrandLeft, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 22)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 15)
	f.SetColWidth(sheetName, "F", "F", 15)
	f.SetColWidth(sheetName, "G", "G", 18)
	f.SetColWidth(sheetName, "H", "H", 15)
	f.SetColWidth(sheetName, "I", "I", 15)

	f.MergeCell(sheetName, "A1", "I1")
	f.SetCellValue(sheetName, "A1", company.CompanyNameEn)
	f.SetCellStyle(sheetName, "A1", "A1", styleTitle)
	f.SetRowHeight(sheetName, 1, 26)

	f.MergeCell(sheetName, "A2", "I2")
	f.SetCellValue(sheetName, "A2", company.AddressEn)
	f.SetCellStyle(sheetName, "A2", "A2", styleSubtitle)
	f.SetRowHeight(sheetName, 2, 20)

	f.MergeCell(sheetName, "A3", "I3")
	f.SetCellValue(sheetName, "A3", "Salary Increments Report")
	f.SetCellStyle(sheetName, "A3", "A3", styleSubtitle)
	f.SetRowHeight(sheetName, 3, 20)

	f.MergeCell(sheetName, "A4", "I4")
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Report Date: %s", time.Now().Format("02-01-2006")))
	f.SetCellStyle(sheetName, "A4", "A4", styleSubtitle)
	f.SetRowHeight(sheetName, 4, 18)

	// Group by section
	sectionMap := make(map[string][]models.SalaryIncrement)
	for _, inc := range increments {
		secName := "No Section"
		if inc.Employee.SectionRef != nil && strings.TrimSpace(inc.Employee.SectionRef.Name) != "" {
			secName = strings.TrimSpace(inc.Employee.SectionRef.Name)
		}
		sectionMap[secName] = append(sectionMap[secName], inc)
	}
	var sections []string
	for k := range sectionMap {
		sections = append(sections, k)
	}
	sort.Slice(sections, func(i, j int) bool {
		if sections[i] == "No Section" {
			return false
		}
		if sections[j] == "No Section" {
			return true
		}
		return strings.ToLower(sections[i]) < strings.ToLower(sections[j])
	})

	headers := []string{"Employee ID", "Name", "Designation", "Department", "Joining Date", "Current Gross", "Increment Amount", "New Gross", "Increment Date"}

	row := 6
	// If no grouping (empty), still render single header
	if len(increments) == 0 {
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			f.SetCellValue(sheetName, cell, h)
			f.SetCellStyle(sheetName, cell, cell, styleHeader)
		}
		f.SetRowHeight(sheetName, row, 25)
		row++
	} else {
		var grandPrev, grandIncAmt, grandNew float64
		grandCount := 0
		for _, secName := range sections {
			list := sectionMap[secName]
			// Section header row
			f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("I%d", row))
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("Section: %s (%d)", secName, len(list)))
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("I%d", row), styleSection)
			f.SetRowHeight(sheetName, row, 25)
			row++

			// Column headers for this section
			for i, h := range headers {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				f.SetCellValue(sheetName, cell, h)
				f.SetCellStyle(sheetName, cell, cell, styleHeader)
			}
			f.SetRowHeight(sheetName, row, 25)
			row++

			var secPrev, secIncAmt, secNew float64
			for _, inc := range list {
				joiningDate := ""
				if !inc.Employee.JoiningDate.IsZero() {
					joiningDate = inc.Employee.JoiningDate.Format("02-01-2006")
				}
				f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), inc.EmployeeID)
				f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), inc.Employee.NameEn)
				f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleLeft)
				if inc.Employee.DesignationRef != nil {
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), inc.Employee.DesignationRef.Name)
				}
				f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleLeft)
				if inc.Employee.Department != nil {
					f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), inc.Employee.Department.Name)
				}
				f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleLeft)
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), joiningDate)
				f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleCenter)
				f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), inc.PreviousGross)
				f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styleRight)
				f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), inc.IncrementAmount)
				f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), styleRight)
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), inc.NewGross)
				f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), styleRight)
				f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), utils.FormatBillDate(inc.IncrementDate))
				f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), styleCenter)
				f.SetRowHeight(sheetName, row, 25)
				row++
				secPrev += inc.PreviousGross
				secIncAmt += inc.IncrementAmount
				secNew += inc.NewGross
			}
			// Section Total row
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "")
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("Total (%d)", len(list)))
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleTotalLeft)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "")
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), "")
			f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "")
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), secPrev)
			f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styleTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), secIncAmt)
			f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), styleTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), secNew)
			f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), styleTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "")
			f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), styleTotal)
			f.SetRowHeight(sheetName, row, 25)
			row++
			grandPrev += secPrev
			grandIncAmt += secIncAmt
			grandNew += secNew
			grandCount += len(list)
			// blank spacer row between sections
			row++
		}
		// Grand Total
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleGrand)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("Grand Total (%d)", grandCount))
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleGrandLeft)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "")
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleGrand)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), "")
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleGrand)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "")
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleGrand)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), grandPrev)
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styleGrand)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), grandIncAmt)
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), styleGrand)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), grandNew)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), styleGrand)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "")
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), styleGrand)
		f.SetRowHeight(sheetName, row, 26)
		row++
	}

	// Signatures - adjusted for 9 cols: B, E, I
	row += 2
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row+1), "Production manager")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row+1), "Admin (A.G.M)")
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "______________________")
	f.SetCellValue(sheetName, fmt.Sprintf("I%d", row+1), "Approved By")

	filename := fmt.Sprintf("increments_%s.xlsx", time.Now().Format("20060102_150405"))
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}
	// Prevent IDM / browser double-download: send exact Content-Length and no-cache headers
	// IDM duplicates when Content-Length missing or attachment is re-fetched without auth. Using buffer + auth-protected XHR blob ensures single download.
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", filename, url.PathEscape(filename)))
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Header("Cache-Control", "private, max-age=0, must-revalidate")
	c.Header("Pragma", "public")
	c.Header("Expires", "0")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
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

	// Group by section
	sectionMap := make(map[string][]models.SalaryIncrement)
	for _, inc := range increments {
		secName := "No Section"
		if inc.Employee.SectionRef != nil && strings.TrimSpace(inc.Employee.SectionRef.Name) != "" {
			secName = strings.TrimSpace(inc.Employee.SectionRef.Name)
		}
		sectionMap[secName] = append(sectionMap[secName], inc)
	}
	var sections []string
	for k := range sectionMap {
		sections = append(sections, k)
	}
	sort.Slice(sections, func(i, j int) bool {
		if sections[i] == "No Section" {
			return false
		}
		if sections[j] == "No Section" {
			return true
		}
		return strings.ToLower(sections[i]) < strings.ToLower(sections[j])
	})

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 14)
		pdf.CellFormat(277, 8, company.CompanyNameEn, "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(277, 6, company.AddressEn, "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(277, 8, "Salary Increments Report", "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(277, 6, fmt.Sprintf("Report Date: %s", time.Now().Format("02-01-2006")), "", 1, "C", false, 0, "")
		pdf.Ln(3)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-25)
		pdf.SetFont("Arial", "", 9)

		// Signatures - Production manager, Admin (A.G.M), Approved By
		pdf.CellFormat(92, 8, "_________________________", "", 0, "C", false, 0, "")
		pdf.CellFormat(92, 8, "_________________________", "", 0, "C", false, 0, "")
		pdf.CellFormat(93, 8, "_________________________", "", 1, "C", false, 0, "")

		pdf.CellFormat(92, 5, "Production manager", "", 0, "C", false, 0, "")
		pdf.CellFormat(92, 5, "Admin (A.G.M)", "", 0, "C", false, 0, "")
		pdf.CellFormat(93, 5, "Approved By", "", 1, "C", false, 0, "")

		// Page number
		pdf.SetY(-12)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont("Arial", "", 9)
	// 8 columns: added Joining Date
	headers := []string{"Emp ID", "Name", "Designation", "Joining Date", "Current Gross", "Inc. Amount", "New Gross", "Inc. Date"}
	widths := []float64{22, 48, 42, 28, 32, 32, 33, 40} // sum 277

	// Helper to render table header
	renderHeader := func() {
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(68, 114, 196)
		pdf.SetTextColor(255, 255, 255)
		for i, str := range headers {
			pdf.CellFormat(widths[i], 9, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetTextColor(0, 0, 0)
	}

	// Helper to render section title
	renderSection := func(secName string, count int) {
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(91, 155, 213)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(277, 9, fmt.Sprintf("Section: %s (%d)", secName, count), "1", 1, "L", true, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	if len(increments) == 0 {
		renderHeader()
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(277, 9, "No records found", "1", 1, "C", false, 0, "")
	} else {
		var grandPrev, grandIncAmt, grandNew float64
		grandCount := 0
		for _, secName := range sections {
			list := sectionMap[secName]
			// section banner
			renderSection(secName, len(list))
			renderHeader()
			pdf.SetFont("Arial", "", 9)
			rowH := 9.0 // increased for readability (was 8), Excel uses 25pt approx 8.8mm
			var secPrev, secIncAmt, secNew float64
			for _, inc := range list {
				// check page break before row
				if pdf.GetY()+rowH > 180 {
					pdf.AddPage()
					renderSection(secName+" (contd.)", len(list))
					renderHeader()
					pdf.SetFont("Arial", "", 9)
				}
				name := inc.Employee.NameEn
				designation := "-"
				if inc.Employee.DesignationRef != nil {
					designation = inc.Employee.DesignationRef.Name
				}
				joiningDate := ""
				if !inc.Employee.JoiningDate.IsZero() {
					joiningDate = inc.Employee.JoiningDate.Format("02-01-2006")
				}

				pdf.CellFormat(widths[0], rowH, inc.EmployeeID, "1", 0, "C", false, 0, "")
				pdf.CellFormat(widths[1], rowH, name, "1", 0, "L", false, 0, "")
				pdf.CellFormat(widths[2], rowH, designation, "1", 0, "L", false, 0, "")
				pdf.CellFormat(widths[3], rowH, joiningDate, "1", 0, "C", false, 0, "")
				pdf.CellFormat(widths[4], rowH, fmt.Sprintf("%.2f", inc.PreviousGross), "1", 0, "R", false, 0, "")
				pdf.CellFormat(widths[5], rowH, fmt.Sprintf("%.2f", inc.IncrementAmount), "1", 0, "R", false, 0, "")
				pdf.CellFormat(widths[6], rowH, fmt.Sprintf("%.2f", inc.NewGross), "1", 0, "R", false, 0, "")
				pdf.CellFormat(widths[7], rowH, utils.FormatBillDate(inc.IncrementDate), "1", 0, "C", false, 0, "")
				pdf.Ln(-1)
				secPrev += inc.PreviousGross
				secIncAmt += inc.IncrementAmount
				secNew += inc.NewGross
			}
			// Section Total
			if pdf.GetY()+rowH > 180 {
				pdf.AddPage()
				renderHeader()
				pdf.SetFont("Arial", "", 9)
			}
			pdf.SetFont("Arial", "B", 9)
			pdf.SetFillColor(255, 242, 204)
			pdf.CellFormat(widths[0], rowH, "", "1", 0, "C", true, 0, "")
			pdf.CellFormat(widths[1], rowH, fmt.Sprintf("Total (%d)", len(list)), "1", 0, "L", true, 0, "")
			pdf.CellFormat(widths[2], rowH, "", "1", 0, "L", true, 0, "")
			pdf.CellFormat(widths[3], rowH, "", "1", 0, "C", true, 0, "")
			pdf.CellFormat(widths[4], rowH, fmt.Sprintf("%.2f", secPrev), "1", 0, "R", true, 0, "")
			pdf.CellFormat(widths[5], rowH, fmt.Sprintf("%.2f", secIncAmt), "1", 0, "R", true, 0, "")
			pdf.CellFormat(widths[6], rowH, fmt.Sprintf("%.2f", secNew), "1", 0, "R", true, 0, "")
			pdf.CellFormat(widths[7], rowH, "", "1", 0, "C", true, 0, "")
			pdf.Ln(-1)
			pdf.SetFont("Arial", "", 9)
			grandPrev += secPrev
			grandIncAmt += secIncAmt
			grandNew += secNew
			grandCount += len(list)
			pdf.Ln(4)
		}
		// Grand Total
		if pdf.GetY()+10 > 180 {
			pdf.AddPage()
		}
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(68, 114, 196)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(widths[0], 10, "", "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[1], 10, fmt.Sprintf("Grand Total (%d)", grandCount), "1", 0, "L", true, 0, "")
		pdf.CellFormat(widths[2], 10, "", "1", 0, "L", true, 0, "")
		pdf.CellFormat(widths[3], 10, "", "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[4], 10, fmt.Sprintf("%.2f", grandPrev), "1", 0, "R", true, 0, "")
		pdf.CellFormat(widths[5], 10, fmt.Sprintf("%.2f", grandIncAmt), "1", 0, "R", true, 0, "")
		pdf.CellFormat(widths[6], 10, fmt.Sprintf("%.2f", grandNew), "1", 0, "R", true, 0, "")
		pdf.CellFormat(widths[7], 10, "", "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
		pdf.SetTextColor(0, 0, 0)
	}

	filename := fmt.Sprintf("increments_%s.pdf", time.Now().Format("20060102_150405"))
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", filename, url.PathEscape(filename)))
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Header("Cache-Control", "private, max-age=0, must-revalidate")
	c.Header("Pragma", "public")
	c.Header("Expires", "0")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
