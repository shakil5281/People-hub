package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"github.com/xuri/excelize/v2"
)

type NightBillHandler struct {
	nightBillService *service.NightBillService
	nightBillRepo    *repository.NightBillRepository
	companyRepo      *repository.CompanyRepository
}

func NewNightBillHandler(nightBillService *service.NightBillService, nightBillRepo *repository.NightBillRepository, companyRepo *repository.CompanyRepository) *NightBillHandler {
	return &NightBillHandler{
		nightBillService: nightBillService,
		nightBillRepo:    nightBillRepo,
		companyRepo:      companyRepo,
	}
}

// ListNightBills godoc
//
//	@Summary      List night bill records
//	@Description  Get night bill records with pagination and filters
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /attendance/night-bill [get]
func (h *NightBillHandler) List(c *gin.Context) {
	p := utils.ParsePagination(c)
	f := repository.NightBillFilter{
		Date:          c.Query("date"),
		StartDate:     c.Query("start_date"),
		EndDate:       c.Query("end_date"),
		CompanyID:     c.Query("company_id"),
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		BillType:      c.Query("bill_type"),
		EmployeeID:    c.Query("employee_id"),
	}

	list, total, err := h.nightBillRepo.ListFiltered(f, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

// ProcessNightBill godoc
//
//	@Summary      Process daily night bills
//	@Description  Calculate and generate night bills for employees based on attendance
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill/process [post]
func (h *NightBillHandler) Process(c *gin.Context) {
	var req service.ProcessNightBillParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	req.UserID = userID

	res, err := h.nightBillService.Process(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ProcessNightBillsFromConfig godoc
//
//	@Summary      Process night bills from Employee Night Bill config
//	@Description  Generate night bills for configured employees using attendance, skip leaves/holidays/weekends
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        payload body service.ProcessConfigParams true "Process config"
//	@Router       /night-bill/process [post]
func (h *NightBillHandler) ProcessFromConfig(c *gin.Context) {
	var req service.ProcessConfigParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = c.GetString("user_id")
	res, err := h.nightBillService.ProcessFromConfig(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

type CreateNightBillRequest struct {
	CompanyID      string  `json:"company_id" binding:"required"`
	EmployeeID     string  `json:"employee_id" binding:"required"`
	AttendanceDate string  `json:"attendance_date" binding:"required"`
	BillType       string  `json:"bill_type"`
	EligibleHours  float64 `json:"eligible_hours"`
	Rate           float64 `json:"rate"`
	Amount         float64 `json:"amount" binding:"required"`
	Remarks        string  `json:"remarks"`
}

// CreateNightBill godoc
//
//	@Summary      Create manual night bill
//	@Description  Manually add a night bill for an employee
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill [post]
func (h *NightBillHandler) Create(c *gin.Context) {
	var req CreateNightBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	billType := req.BillType
	if billType == "" {
		billType = "manual"
	}

	nb := &models.NightBill{
		CompanyID:      req.CompanyID,
		EmployeeID:     req.EmployeeID,
		AttendanceDate: req.AttendanceDate,
		BillType:       billType,
		EligibleHours:  req.EligibleHours,
		Rate:           req.Rate,
		Amount:         req.Amount,
		Remarks:        req.Remarks,
		CreatedBy:      &userID,
		UpdatedBy:      &userID,
	}

	if err := h.nightBillRepo.Upsert(nb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, nb)
}

// UpdateNightBill godoc
//
//	@Summary      Update night bill
//	@Description  Update night bill record details
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill/{id} [put]
func (h *NightBillHandler) Update(c *gin.Context) {
	id := c.Param("id")
	nb, err := h.nightBillRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Night bill record not found"})
		return
	}

	var req CreateNightBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	nb.Amount = req.Amount
	nb.Rate = req.Rate
	nb.EligibleHours = req.EligibleHours
	if req.BillType != "" {
		nb.BillType = req.BillType
	}
	if req.Remarks != "" {
		nb.Remarks = req.Remarks
	}
	nb.UpdatedBy = &userID

	if err := h.nightBillRepo.Update(nb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nb)
}

// DeleteNightBill godoc
//
//	@Summary      Delete night bill
//	@Description  Soft delete a night bill record
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /attendance/night-bill/{id} [delete]
func (h *NightBillHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.nightBillRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Night bill record deleted"})
}

type BulkDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// BulkDeleteNightBill godoc
//
//	@Summary      Bulk delete night bills
//	@Description  Delete multiple night bill records
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill/bulk-delete [post]
func (h *NightBillHandler) DeleteBulk(c *gin.Context) {
	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.nightBillRepo.DeleteBulk(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d night bill record(s) deleted", len(req.IDs))})
}

// nightBillInOut returns the actual attendance clock-in/out times for display,
// preferring the linked attendance record over the stored night-bill times.
func nightBillInOut(b models.NightBill) (string, string) {
	inTime, outTime := "-", "-"
	if b.Attendance.ID != "" && b.Attendance.CheckIn != nil {
		inTime = b.Attendance.CheckIn.Format("15:04")
	} else if b.InTime != nil {
		inTime = b.InTime.Format("15:04")
	}
	if b.Attendance.ID != "" && b.Attendance.CheckOut != nil {
		outTime = b.Attendance.CheckOut.Format("15:04")
	} else if b.OutTime != nil {
		outTime = b.OutTime.Format("15:04")
	}
	return inTime, outTime
}

// nightBillExportHeader resolves the company name/address and report period from DB for exports.
func (h *NightBillHandler) nightBillExportHeader(companyID, startDate, endDate, date string) (string, string, string) {
	companyName := "Company Name"
	companyAddress := "Company Address"
	if companyID != "" {
		if company, err := h.companyRepo.FindByID(companyID); err == nil && company != nil {
			if company.CompanyNameEn != "" {
				companyName = company.CompanyNameEn
			}
			if company.AddressEn != "" {
				companyAddress = company.AddressEn
			}
		}
	} else if company, _, err := h.companyRepo.List(1, 1); err == nil && len(company) > 0 {
		if company[0].CompanyNameEn != "" {
			companyName = company[0].CompanyNameEn
		}
		if company[0].AddressEn != "" {
			companyAddress = company[0].AddressEn
		}
	}

	period := ""
	switch {
	case startDate != "" && endDate != "":
		period = fmt.Sprintf("Period: %s To %s", utils.FormatBillDate(startDate), utils.FormatBillDate(endDate))
	case date != "":
		period = "Date: " + utils.FormatBillDate(date)
	}
	return companyName, companyAddress, period
}

// ExportExcel godoc
//
//	@Summary      Export night bills to Excel
//	@Description  Generate Excel sheet of night bill records
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Produce      octet-stream
//	@Router       /attendance/night-bill/export/excel [get]
func (h *NightBillHandler) ExportExcel(c *gin.Context) {
	fFilter := repository.NightBillFilter{
		Date:          c.Query("date"),
		StartDate:     c.Query("start_date"),
		EndDate:       c.Query("end_date"),
		CompanyID:     c.Query("company_id"),
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		BillType:      c.Query("bill_type"),
		EmployeeID:    c.Query("employee_id"),
	}

	list, err := h.nightBillRepo.ListAllFiltered(fFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	companyName, companyAddress, period := h.nightBillExportHeader(fFilter.CompanyID, fFilter.StartDate, fFilter.EndDate, fFilter.Date)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Night Bill"
	f.SetSheetName("Sheet1", sheet)

	f.SetColWidth(sheet, "A", "A", 10)
	f.SetColWidth(sheet, "B", "B", 22)
	f.SetColWidth(sheet, "C", "C", 14)
	f.SetColWidth(sheet, "D", "D", 9)
	f.SetColWidth(sheet, "E", "E", 9)
	f.SetColWidth(sheet, "F", "F", 10)
	f.SetColWidth(sheet, "G", "G", 18)

	borderColor := "262626"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	companyNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	subHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	reportNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "DC2626"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
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
	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9E2F3"}, Pattern: 1},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	inWordsStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})


	headers := []string{"Employee ID", "Employee Name", "Designation", "In Time", "Out Time", "Amount", "Signature"}
	cols := []string{"A", "B", "C", "D", "E", "F", "G"}
	lastCol := "G"

	// Header: company name, address, report name, period
	f.MergeCell(sheet, "A1", lastCol+"1")
	f.SetCellValue(sheet, "A1", companyName)
	f.SetCellStyle(sheet, "A1", lastCol+"1", companyNameStyle)
	f.SetRowHeight(sheet, 1, 32)

	f.MergeCell(sheet, "A2", lastCol+"2")
	f.SetCellValue(sheet, "A2", companyAddress)
	f.SetCellStyle(sheet, "A2", lastCol+"2", subHeaderStyle)
	f.SetRowHeight(sheet, 2, 20)

	f.MergeCell(sheet, "A3", lastCol+"3")
	f.SetCellValue(sheet, "A3", "NIGHT BILL REPORT")
	f.SetCellStyle(sheet, "A3", lastCol+"3", reportNameStyle)
	f.SetRowHeight(sheet, 3, 22)

	f.MergeCell(sheet, "A4", lastCol+"4")
	f.SetCellValue(sheet, "A4", period)
	f.SetCellStyle(sheet, "A4", lastCol+"4", subHeaderStyle)
	f.SetRowHeight(sheet, 4, 18)

	// Table header
	headerRow := 6
	f.SetRowHeight(sheet, headerRow, 30)
	for i, h := range headers {
		cell := fmt.Sprintf("%s%d", cols[i], headerRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	showGrid := false
	f.SetSheetView(sheet, -1, &excelize.ViewOptions{ShowGridLines: &showGrid})

	rowIdx := headerRow + 1
	totalAmt := 0.0
	for _, b := range list {
		f.SetRowHeight(sheet, rowIdx, 30)
		desig := ""
		if b.Employee.DesignationRef != nil {
			desig = b.Employee.DesignationRef.Name
		}
		inTime, outTime := nightBillInOut(b)

		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), b.EmployeeID)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("A%d", rowIdx), cellNormal)

		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIdx), b.Employee.NameEn)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowIdx), fmt.Sprintf("B%d", rowIdx), cellLeft)

		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIdx), desig)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", rowIdx), fmt.Sprintf("C%d", rowIdx), cellLeft)

		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowIdx), inTime)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", rowIdx), fmt.Sprintf("D%d", rowIdx), cellNormal)

		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowIdx), outTime)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", rowIdx), fmt.Sprintf("E%d", rowIdx), cellNormal)

		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), b.Amount)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", rowIdx), fmt.Sprintf("F%d", rowIdx), cellNormal)

		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), "")
		f.SetCellStyle(sheet, fmt.Sprintf("G%d", rowIdx), fmt.Sprintf("G%d", rowIdx), cellNormal)

		totalAmt += b.Amount
		rowIdx++
	}

	// Grand Total Row
	f.SetRowHeight(sheet, rowIdx, 30)
	f.MergeCell(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("E%d", rowIdx))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), "Grand Total")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("E%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), totalAmt)
	f.SetCellStyle(sheet, fmt.Sprintf("F%d", rowIdx), fmt.Sprintf("F%d", rowIdx), grandTotalStyle)

	f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), "")
	f.SetCellStyle(sheet, fmt.Sprintf("G%d", rowIdx), fmt.Sprintf("G%d", rowIdx), grandTotalStyle)

	// In Words Row (next row after Grand Total)
	rowIdx++
	f.SetRowHeight(sheet, rowIdx, 26)
	f.MergeCell(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("G%d", rowIdx))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), utils.FormatTakaInWords(totalAmt))
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("G%d", rowIdx), inWordsStyle)

	// Page Layout: A4 paper size & Portrait orientation
	paperSize := 9 // A4
	orientation := "portrait"
	f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
		Size:        &paperSize,
		Orientation: &orientation,
	})

	// Page Margins: Narrow (Left: 0.25", Right: 0.25", Top: 0.75", Bottom: 0.75")
	marginLeft := 0.25
	marginRight := 0.25
	marginTop := 0.75
	marginBottom := 0.75
	marginHeader := 0.3
	marginFooter := 0.3
	f.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{
		Left:   &marginLeft,
		Right:  &marginRight,
		Top:    &marginTop,
		Bottom: &marginBottom,
		Header: &marginHeader,
		Footer: &marginFooter,
	})

	// Repeat header rows on every printed page
	if err := f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Titles",
		RefersTo: fmt.Sprintf("'%s'!$1:$%d", sheet, headerRow),
		Scope:    sheet,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Signature footer on every printed page (Page Setup Custom Footer)
	f.SetHeaderFooter(sheet, &excelize.HeaderFooterOptions{
		OddFooter: "&LPrepared By&CAdmin (A.G.M)                     Asst. General Manager&RApproved By",
	})

	filename := fmt.Sprintf("night_bill_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}

// ExportPDF godoc
//
//	@Summary      Export night bills to PDF
//	@Description  Generate PDF document of night bill records
//	@Tags         Night Bill
//	@Security     BearerAuth
//	@Produce      pdf
//	@Router       /attendance/night-bill/export/pdf [get]
func (h *NightBillHandler) ExportPDF(c *gin.Context) {
	fFilter := repository.NightBillFilter{
		Date:          c.Query("date"),
		StartDate:     c.Query("start_date"),
		EndDate:       c.Query("end_date"),
		CompanyID:     c.Query("company_id"),
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		BillType:      c.Query("bill_type"),
		EmployeeID:    c.Query("employee_id"),
	}

	list, err := h.nightBillRepo.ListAllFiltered(fFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	companyName, companyAddress, period := h.nightBillExportHeader(fFilter.CompanyID, fFilter.StartDate, fFilter.EndDate, fFilter.Date)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 10, 15)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	font := "Helvetica"
	pageW := 180.0
	x := 15.0
	curY := 10.0
	rowH := 10.0
	tableBottom := 270.0
	sigY := 278.0

	headers := []string{"Employee ID", "Employee Name", "Designation", "In Time", "Out Time", "Amount", "Signature"}
	widths := []float64{18, 45, 32, 18, 18, 24, 25}

	drawPageHeader := func() {
		curY = 10.0
		pdf.SetFont(font, "B", 20)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(x, curY)
		pdf.CellFormat(pageW, 8, companyName, "", 0, "C", false, 0, "")
		curY += 8

		pdf.SetFont(font, "", 11)
		pdf.SetXY(x, curY)
		pdf.CellFormat(pageW, 6, companyAddress, "", 0, "C", false, 0, "")
		curY += 6

		pdf.SetFont(font, "B", 11)
		pdf.SetTextColor(220, 38, 38)
		pdf.SetXY(x, curY)
		pdf.CellFormat(pageW, 7, "NIGHT BILL REPORT", "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
		curY += 7

		pdf.SetFont(font, "", 10)
		pdf.SetXY(x, curY)
		pdf.CellFormat(pageW, 5, period, "", 0, "C", false, 0, "")
		curY += 8

		pdf.SetFillColor(217, 226, 243)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont(font, "B", 8)
		pdf.SetDrawColor(38, 38, 38)
		cx := x
		for i, hStr := range headers {
			pdf.Rect(cx, curY, widths[i], rowH, "DF")
			pdf.SetXY(cx, curY+(rowH-3.5)/2.0)
			pdf.CellFormat(widths[i], 3.5, hStr, "", 0, "C", false, 0, "")
			cx += widths[i]
		}
		curY += rowH
	}

	drawSignature := func() {
		pdf.SetDrawColor(0, 0, 0)
		pdf.SetFont(font, "B", 9)
		sigCols := []struct {
			start float64
			width float64
			label string
		}{
			{x, 45, "Prepared By"},
			{x + 45, 45, "Admin (A.G.M)"},
			{x + 90, 45, "Asst. General Manager"},
			{x + 135, 45, "Approved By"},
		}
		for _, s := range sigCols {
			pdf.SetDashPattern([]float64{1.5, 1.2}, 0)
			pdf.Line(s.start, sigY, s.start+s.width-3, sigY)
			pdf.SetDashPattern([]float64{0, 0}, 0)
			pdf.SetXY(s.start, sigY+3)
			pdf.CellFormat(s.width, 5, s.label, "", 0, "C", false, 0, "")
		}
	}

	ensureSpace := func() {
		if curY+rowH > tableBottom {
			drawSignature()
			pdf.AddPage()
			drawPageHeader()
		}
	}

	drawPageHeader()
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont(font, "", 8)

	totalAmt := 0.0
	for _, b := range list {
		ensureSpace()
		desig := ""
		if b.Employee.DesignationRef != nil {
			desig = b.Employee.DesignationRef.Name
		}
		inTime, outTime := nightBillInOut(b)

		vals := []string{
			b.EmployeeID,
			b.Employee.NameEn,
			desig,
			inTime,
			outTime,
			fmt.Sprintf("%.2f", b.Amount),
			"",
		}

		cx := x
		for j, vStr := range vals {
			vertY := curY + (rowH-3.5)/2.0
			if j == 1 {
				pdf.Rect(cx, curY, widths[j], rowH, "D")
				pdf.SetXY(cx+1, vertY)
				pdf.CellFormat(widths[j]-1.0, 3.5, vStr, "", 0, "L", false, 0, "")
			} else {
				pdf.Rect(cx, curY, widths[j], rowH, "D")
				pdf.SetXY(cx+0.5, vertY)
				pdf.CellFormat(widths[j]-1.0, 3.5, vStr, "", 0, "C", false, 0, "")
			}
			cx += widths[j]
		}
		totalAmt += b.Amount
		curY += rowH
	}

	// Grand Total Row
	ensureSpace()
	pdf.SetFillColor(217, 226, 243)
	pdf.SetFont(font, "B", 8.5)
	grandW := widths[0] + widths[1] + widths[2] + widths[3] + widths[4]
	pdf.Rect(x, curY, grandW, rowH, "DF")
	pdf.SetXY(x+2, curY+(rowH-3.5)/2.0)
	pdf.CellFormat(grandW-3, 3.5, "Grand Total", "", 0, "C", false, 0, "")

	pdf.Rect(x+grandW, curY, widths[5], rowH, "DF")
	pdf.SetXY(x+grandW+0.5, curY+1.8)
	pdf.CellFormat(widths[5]-1.0, 3.5, fmt.Sprintf("%.2f", totalAmt), "", 0, "R", false, 0, "")

	pdf.Rect(x+grandW+widths[5], curY, widths[6], rowH, "DF")
	curY += rowH

	// In Words Row (next row after Grand Total)
	ensureSpace()
	pdf.SetFillColor(255, 255, 255)
	pdf.SetFont(font, "B", 8.5)
	pdf.Rect(x, curY, pageW, rowH, "D")
	pdf.SetXY(x+2, curY+(rowH-3.5)/2.0)
	pdf.CellFormat(pageW-4, 3.5, utils.FormatTakaInWords(totalAmt), "", 0, "L", false, 0, "")
	curY += rowH

	drawSignature()

	filename := fmt.Sprintf("night_bill_%s.pdf", time.Now().Format("20060102"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	pdf.Output(c.Writer)
}
