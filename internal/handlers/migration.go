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
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"github.com/xuri/excelize/v2"
)

type MigrationHandler struct {
	repo    *repository.MigrationRepository
	service *service.MigrationService
}

func NewMigrationHandler(repo *repository.MigrationRepository, svc *service.MigrationService) *MigrationHandler {
	return &MigrationHandler{repo: repo, service: svc}
}

// List godoc
//
//	@Summary      List employee migrations
//	@Tags         Migrations
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /migrations [get]
func (h *MigrationHandler) List(c *gin.Context) {
	employee := c.Query("employee")
	employeeID := c.Query("employee_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	migrationType := c.Query("migration_type")
	status := c.Query("status")
	companyID := c.Query("company_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	p := utils.ParsePagination(c)

	items, total, err := h.repo.ListFiltered(employee, employeeID, departmentID, sectionID, designationID, lineID, migrationType, status, companyID, dateFrom, dateTo, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(items, total, p))
}

// GetByID godoc
//
//	@Summary      Get migration details by ID
//	@Tags         Migrations
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /migrations/{id} [get]
func (h *MigrationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "migration record not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// GetSummary godoc
//
//	@Summary      Get migration summary metrics with Department, Section, Designation, Line breakdowns
//	@Tags         Migrations
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /migrations/summary [get]
func (h *MigrationHandler) GetSummary(c *gin.Context) {
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")

	summary, err := h.repo.GetFullSummary(dateFrom, dateTo, companyID, departmentID, sectionID, designationID, lineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// Create godoc
//
//	@Summary      Record an employee migration/transfer
//	@Tags         Migrations
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /migrations [post]
func (h *MigrationHandler) Create(c *gin.Context) {
	var req service.CreateMigrationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID != "" {
		req.CreatedBy = &userID
	}

	item, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// ExportExcel godoc
//
//	@Summary      Export employee migration summary to Excel
//	@Tags         Migrations
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Router       /migrations/export/excel [get]
func (h *MigrationHandler) ExportExcel(c *gin.Context) {
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	companyID := c.Query("company_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	summary, err := h.repo.GetFullSummary(dateFrom, dateTo, companyID, departmentID, sectionID, designationID, lineID)
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

	renderFullMigrationExcel(c, company, periodStr, summary)
}

// ExportPDF godoc
//
//	@Summary      Export employee migration summary to PDF
//	@Tags         Migrations
//	@Security     BearerAuth
//	@Produce      application/pdf
//	@Router       /migrations/export/pdf [get]
func (h *MigrationHandler) ExportPDF(c *gin.Context) {
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	companyID := c.Query("company_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	lang := c.DefaultQuery("lang", "en")

	isBn := strings.ToLower(lang) == "bn" || strings.ToLower(lang) == "bangla"

	summary, err := h.repo.GetFullSummary(dateFrom, dateTo, companyID, departmentID, sectionID, designationID, lineID)
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

	renderFullMigrationPDF(c, company, periodStr, summary, isBn)
}

func renderFullMigrationExcel(c *gin.Context, company models.Company, periodStr string, s *repository.FullMigrationSummary) {
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

	borderColor := "262626"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "1D4ED8"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F3F4F6"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9E2F3"}, Pattern: 1},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
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
	headerStyleA1, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	renderSheet := func(sheetName, subTitle, nameColHeader string, rows []repository.OrganizationalSummaryRow) {
		orientation := "landscape"
		paperSize := 9 // A4
		_ = f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
		})

		margin05 := 0.5
		_ = f.SetPageMargins(sheetName, &excelize.PageLayoutMarginsOptions{
			Left:   &margin05,
			Right:  &margin05,
			Top:    &margin05,
			Bottom: &margin05,
			Header: &margin05,
			Footer: &margin05,
		})

		_ = f.SetColWidth(sheetName, "A", "A", 8)
		_ = f.SetColWidth(sheetName, "B", "B", 35)
		_ = f.SetColWidth(sheetName, "C", "C", 20)
		_ = f.SetColWidth(sheetName, "D", "D", 20)
		_ = f.SetColWidth(sheetName, "E", "E", 20)
		_ = f.SetColWidth(sheetName, "F", "F", 20)

		headerRuns := []excelize.RichTextRun{
			{Text: companyName, Font: &excelize.Font{Bold: true, Size: 20}},
			{Text: "\n" + companyAddress, Font: &excelize.Font{Size: 11}},
			{Text: "\nEMPLOYEE MIGRATION SUMMARY (" + strings.ToUpper(subTitle) + ")", Font: &excelize.Font{Bold: true, Size: 12, Color: "1D4ED8"}},
			{Text: "\n" + periodStr, Font: &excelize.Font{Size: 11}},
		}

		_ = f.SetCellRichText(sheetName, "A1", headerRuns)
		_ = f.MergeCell(sheetName, "A1", "F1")
		_ = f.SetCellStyle(sheetName, "A1", "F1", headerStyleA1)
		_ = f.SetRowHeight(sheetName, 1, 75)

		// Top Metric Summary
		_ = f.SetCellValue(sheetName, "A3", "OVERALL METRICS SUMMARY")
		_ = f.MergeCell(sheetName, "A3", "F3")
		_ = f.SetCellStyle(sheetName, "A3", "F3", titleStyle)
		_ = f.SetRowHeight(sheetName, 3, 22)

		mHeaders := []string{"Total Joining", "Total Left/Close", "Total Resignation", "Active Total"}
		mCols := []string{"B", "C", "D", "E"}
		for i, mh := range mHeaders {
			cell := mCols[i] + "4"
			_ = f.SetCellValue(sheetName, cell, mh)
			_ = f.SetCellStyle(sheetName, cell, cell, tableHeaderStyle)
		}
		_ = f.SetRowHeight(sheetName, 4, 22)

		_ = f.SetCellValue(sheetName, "B5", s.TotalJoining)
		_ = f.SetCellValue(sheetName, "C5", s.TotalLeft)
		_ = f.SetCellValue(sheetName, "D5", s.TotalResign)
		_ = f.SetCellValue(sheetName, "E5", s.TotalActive)
		for _, mc := range mCols {
			_ = f.SetCellStyle(sheetName, mc+"5", mc+"5", cellNormal)
		}
		_ = f.SetRowHeight(sheetName, 5, 20)

		// Table Title
		_ = f.SetCellValue(sheetName, "A7", subTitle+" BREAKDOWN")
		_ = f.MergeCell(sheetName, "A7", "F7")
		_ = f.SetCellStyle(sheetName, "A7", "F7", titleStyle)
		_ = f.SetRowHeight(sheetName, 7, 22)

		// Table Header
		tCols := []string{"A", "B", "C", "D", "E", "F"}
		tNames := []string{"SL", nameColHeader, "New Joining", "Total Left/Close", "Resignation", "Active Total"}
		for i, tn := range tNames {
			cell := tCols[i] + "8"
			_ = f.SetCellValue(sheetName, cell, tn)
			_ = f.SetCellStyle(sheetName, cell, cell, tableHeaderStyle)
		}
		_ = f.SetRowHeight(sheetName, 8, 24)

		// Table Rows
		currRow := 9
		for idx, row := range rows {
			rStr := strconv.Itoa(currRow)
			_ = f.SetRowHeight(sheetName, currRow, 20)
			_ = f.SetCellValue(sheetName, "A"+rStr, idx+1)
			_ = f.SetCellValue(sheetName, "B"+rStr, row.Name)
			_ = f.SetCellValue(sheetName, "C"+rStr, row.NewJoining)
			_ = f.SetCellValue(sheetName, "D"+rStr, row.TotalLeft)
			_ = f.SetCellValue(sheetName, "E"+rStr, row.TotalResign)
			_ = f.SetCellValue(sheetName, "F"+rStr, row.ActiveTotal)

			_ = f.SetCellStyle(sheetName, "A"+rStr, "A"+rStr, cellNormal)
			_ = f.SetCellStyle(sheetName, "B"+rStr, "B"+rStr, cellLeft)
			_ = f.SetCellStyle(sheetName, "C"+rStr, "C"+rStr, cellNormal)
			_ = f.SetCellStyle(sheetName, "D"+rStr, "D"+rStr, cellNormal)
			_ = f.SetCellStyle(sheetName, "E"+rStr, "E"+rStr, cellNormal)
			_ = f.SetCellStyle(sheetName, "F"+rStr, "F"+rStr, cellNormal)
			currRow++
		}
	}

	// 1. Department Summary Sheet
	_ = f.SetSheetName("Sheet1", "Department Summary")
	renderSheet("Department Summary", "Department Summary", "Department Name", s.ByDepartment)

	// 2. Section Summary Sheet
	_, _ = f.NewSheet("Section Summary")
	renderSheet("Section Summary", "Section Summary", "Section Name", s.BySection)

	// 3. Designation Summary Sheet
	_, _ = f.NewSheet("Designation Summary")
	renderSheet("Designation Summary", "Designation Summary", "Designation Name", s.ByDesignation)

	// 4. Line Summary Sheet
	_, _ = f.NewSheet("Line Summary")
	renderSheet("Line Summary", "Line Summary", "Line Name", s.ByLine)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=migration_summary_%s.xlsx", time.Now().Format("20060102_150405")))

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func renderFullMigrationPDF(c *gin.Context, company models.Company, periodStr string, s *repository.FullMigrationSummary, isBn bool) {
	companyName := companyDisplayName(company, isBn)
	companyAddress := company.AddressEn
	if isBn && company.AddressBn != "" {
		companyAddress = utils.UnicodeToBijoy(company.AddressBn)
	}
	if companyAddress == "" {
		companyAddress = "Masterbari, Gazipur City, Gazipur."
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 12)

	font := "Arial"
	if isBn {
		font = loadBanglaFont(pdf)
	}

	pdf.AddPage()

	pdf.SetFont(font, "B", 14)
	pdf.SetTextColor(18, 58, 99)
	pdf.CellFormat(190, 7, companyName, "", 1, "C", false, 0, "")

	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(190, 5, companyAddress, "", 1, "C", false, 0, "")

	pdf.SetFont(font, "B", 11)
	pdf.SetTextColor(29, 78, 216)
	reportTitle := "EMPLOYEE MIGRATION SUMMARY REPORT"
	if isBn {
		reportTitle = utils.UnicodeToBijoy("কর্মচারী মাইগ্রেশন ও বিভাগওয়ারী সারসংক্ষেপ")
	}
	pdf.CellFormat(190, 6, reportTitle, "", 1, "C", false, 0, "")

	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(190, 5, periodStr, "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Top Summary Cards Box
	pdf.SetFillColor(243, 244, 246)
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)

	pdf.SetFont(font, "B", 9)
	pdf.SetTextColor(31, 41, 55)

	cardW := 45.0
	pdf.CellFormat(cardW, 6, "Total Joining", "1", 0, "C", true, 0, "")
	pdf.CellFormat(cardW, 6, "Total Left/Close", "1", 0, "C", true, 0, "")
	pdf.CellFormat(cardW, 6, "Total Resign", "1", 0, "C", true, 0, "")
	pdf.CellFormat(cardW, 6, "Total Active", "1", 1, "C", true, 0, "")

	pdf.SetFont(font, "B", 11)
	pdf.SetTextColor(16, 185, 129)
	pdf.CellFormat(cardW, 7, strconv.FormatInt(s.TotalJoining, 10), "1", 0, "C", false, 0, "")
	pdf.SetTextColor(239, 68, 68)
	pdf.CellFormat(cardW, 7, strconv.FormatInt(s.TotalLeft, 10), "1", 0, "C", false, 0, "")
	pdf.SetTextColor(245, 158, 11)
	pdf.CellFormat(cardW, 7, strconv.FormatInt(s.TotalResign, 10), "1", 0, "C", false, 0, "")
	pdf.SetTextColor(59, 130, 246)
	pdf.CellFormat(cardW, 7, strconv.FormatInt(s.TotalActive, 10), "1", 1, "C", false, 0, "")
	pdf.Ln(5)

	renderPdfSection := func(title string, rows []repository.OrganizationalSummaryRow) {
		pdf.SetFont(font, "B", 10)
		pdf.SetTextColor(29, 78, 216)
		pdf.CellFormat(190, 6, title, "", 1, "L", false, 0, "")

		wSL := 12.0
		wName := 78.0
		wJoin := 25.0
		wLeft := 25.0
		wRes := 25.0
		wAct := 25.0

		pdf.SetFillColor(217, 226, 243)
		pdf.SetDrawColor(38, 38, 38)
		pdf.SetFont(font, "B", 8.5)
		pdf.SetTextColor(0, 0, 0)

		pdf.CellFormat(wSL, 6, "SL", "1", 0, "C", true, 0, "")
		pdf.CellFormat(wName, 6, "Name", "1", 0, "C", true, 0, "")
		pdf.CellFormat(wJoin, 6, "New Joining", "1", 0, "C", true, 0, "")
		pdf.CellFormat(wLeft, 6, "Total Left", "1", 0, "C", true, 0, "")
		pdf.CellFormat(wRes, 6, "Resignation", "1", 0, "C", true, 0, "")
		pdf.CellFormat(wAct, 6, "Active Total", "1", 1, "C", true, 0, "")

		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(20, 20, 20)

		for idx, row := range rows {
			nameStr := row.Name
			if isBn {
				nameStr = utils.UnicodeToBijoy(nameStr)
			}
			pdf.CellFormat(wSL, 6, strconv.Itoa(idx+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wName, 6, truncateString(nameStr, 35), "1", 0, "L", false, 0, "")
			pdf.CellFormat(wJoin, 6, strconv.FormatInt(row.NewJoining, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wLeft, 6, strconv.FormatInt(row.TotalLeft, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wRes, 6, strconv.FormatInt(row.TotalResign, 10), "1", 0, "C", false, 0, "")
			pdf.CellFormat(wAct, 6, strconv.FormatInt(row.ActiveTotal, 10), "1", 1, "C", false, 0, "")
		}
		pdf.Ln(4)
	}

	renderPdfSection("1. Department Wise Summary", s.ByDepartment)
	renderPdfSection("2. Section Wise Summary", s.BySection)
	renderPdfSection("3. Designation Wise Summary", s.ByDesignation)
	renderPdfSection("4. Line Wise Summary", s.ByLine)

	pdf.AliasNbPages("{nb}")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-10)
		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(190, 5, fmt.Sprintf("Page %d of {nb}   |   Print Date: %s", pdf.PageNo(), time.Now().Format("02-01-2006 15:04")), "", 0, "C", false, 0, "")
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
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=migration_summary_%s.pdf", time.Now().Format("20060102_150405")))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
