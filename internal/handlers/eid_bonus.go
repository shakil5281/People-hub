package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/xuri/excelize/v2"
)

type EidBonusHandler struct {
	eidBonusService *service.EidBonusService
	eidBonusRepo    *repository.EidBonusRepository
}

func NewEidBonusHandler(
	eidBonusService *service.EidBonusService,
	eidBonusRepo *repository.EidBonusRepository,
) *EidBonusHandler {
	return &EidBonusHandler{
		eidBonusService: eidBonusService,
		eidBonusRepo:    eidBonusRepo,
	}
}

type ProcessEidBonusRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
	Year      int    `json:"year" binding:"required"`
}

// ProcessEidBonus godoc
//
//	@Summary      Process Eid bonus
//	@Description  Calculate and generate eid bonus for all active employees for a given year
//	@Tags         Eid Bonus
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body ProcessEidBonusRequest true "Eid bonus process params"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /eid-bonus/process [post]
func (h *EidBonusHandler) Process(c *gin.Context) {
	var req ProcessEidBonusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	result, err := h.eidBonusService.ProcessYear(req.CompanyID, req.Year, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Eid bonus processed for %d employees", result.Processed),
		"processed": result.Processed,
		"total":     result.Total,
		"year":      req.Year,
	})
}

// Sheet godoc
//
//	@Summary      Eid bonus sheet
//	@Description  Get eid bonus records for all employees for a given year
//	@Tags         Eid Bonus
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id query string true  "Company ID"
//	@Param        year       query int    true  "Year"
//	@Param        bonus_type query string false "Filter by bonus type"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /eid-bonus/sheet [get]
func (h *EidBonusHandler) Sheet(c *gin.Context) {
	companyID := c.Query("company_id")
	yearStr := c.Query("year")
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id and year are required"})
		return
	}

	bonuses, err := h.eidBonusRepo.ListAllByYear(repository.EidBonusFilter{
		CompanyID: companyID,
		Year:      year,
		BonusType: c.Query("bonus_type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totals := map[string]float64{
		"gross_salary": 0,
		"bonus_amount": 0,
	}
	for _, b := range bonuses {
		totals["gross_salary"] += b.GrossSalary
		totals["bonus_amount"] += b.BonusAmount
	}

	c.JSON(http.StatusOK, gin.H{
		"bonuses": bonuses,
		"total":   len(bonuses),
		"totals":  totals,
		"year":    year,
	})
}

// Summary godoc
//
//	@Summary      Eid bonus summary by department/line
//	@Description  Get eid bonus summary grouped by department or line
//	@Tags         Eid Bonus
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id  query string true  "Company ID"
//	@Param        year        query int    true  "Year"
//	@Param        group_by    query string false "Group by: department|line (default: department)"
//	@Param        bonus_type  query string false "Filter by bonus type"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /eid-bonus/summary [get]
func (h *EidBonusHandler) Summary(c *gin.Context) {
	companyID := c.Query("company_id")
	yearStr := c.Query("year")
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id and year are required"})
		return
	}

	bonuses, err := h.eidBonusRepo.ListAllByYear(repository.EidBonusFilter{
		CompanyID: companyID,
		Year:      year,
		BonusType: c.Query("bonus_type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	groupBy := c.DefaultQuery("group_by", "department")

	type groupKey struct {
		Name string
		ID   string
	}
	type groupData struct {
		Employees   int
		GrossSalary float64
		BonusAmount float64
	}

	groupMap := make(map[groupKey]*groupData)
	for _, b := range bonuses {
		var key groupKey
		switch groupBy {
		case "line":
			if b.Employee.LineRef != nil {
				key = groupKey{Name: b.Employee.LineRef.Name, ID: b.Employee.LineRef.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		default:
			if b.Employee.Department != nil {
				key = groupKey{Name: b.Employee.Department.Name, ID: b.Employee.Department.ID}
			} else {
				key = groupKey{Name: "Unknown", ID: ""}
			}
		}
		if groupMap[key] == nil {
			groupMap[key] = &groupData{}
		}
		d := groupMap[key]
		d.Employees++
		d.GrossSalary += b.GrossSalary
		d.BonusAmount += b.BonusAmount
	}

	var summaries []map[string]interface{}
	for key, d := range groupMap {
		summaries = append(summaries, map[string]interface{}{
			"group_key":    key.Name,
			"group_id":     key.ID,
			"employees":    d.Employees,
			"gross_salary": d.GrossSalary,
			"bonus_amount": d.BonusAmount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries": summaries,
		"total":     len(summaries),
	})
}

// BankSheet godoc
//
//	@Summary      Eid bonus bank sheet
//	@Description  Get eid bonus records for bank transfer
//	@Tags         Eid Bonus
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id query string true  "Company ID"
//	@Param        year       query int    true  "Year"
//	@Param        bonus_type query string false "Filter by bonus type"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /eid-bonus/bank-sheet [get]
func (h *EidBonusHandler) BankSheet(c *gin.Context) {
	companyID := c.Query("company_id")
	yearStr := c.Query("year")
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id and year are required"})
		return
	}

	bonuses, err := h.eidBonusRepo.ListAllByYear(repository.EidBonusFilter{
		CompanyID: companyID,
		Year:      year,
		BonusType: c.Query("bonus_type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totals := map[string]float64{
		"gross_salary": 0,
		"bonus_amount": 0,
	}
	for _, b := range bonuses {
		totals["gross_salary"] += b.GrossSalary
		totals["bonus_amount"] += b.BonusAmount
	}

	c.JSON(http.StatusOK, gin.H{
		"bonuses": bonuses,
		"total":   len(bonuses),
		"totals":  totals,
		"year":    year,
	})
}

func eidBonusSheetStyles(f *excelize.File) *eidStyles {
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
	return &eidStyles{header, data, line, subtotal, money, moneyBold}
}

type eidStyles struct {
	header    int
	data      int
	line      int
	subtotal  int
	money     int
	moneyBold int
}

// ExportExcel godoc
//
//	@Summary      Export eid bonus to Excel
//	@Description  Download eid bonus sheet as Excel file
//	@Tags         Eid Bonus
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        company_id query string true  "Company ID"
//	@Param        year       query int    true  "Year"
//	@Param        bonus_type query string false "Filter by bonus type"
//	@Success      200  {file}  file
//	@Failure      500  {object}  map[string]string
//	@Router       /eid-bonus/export/excel [get]
func (h *EidBonusHandler) ExportExcel(c *gin.Context) {
	companyID := c.Query("company_id")
	yearStr := c.Query("year")
	year, _ := strconv.Atoi(yearStr)

	if companyID == "" || year == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id and year are required"})
		return
	}

	bonuses, err := h.eidBonusRepo.ListAllByYear(repository.EidBonusFilter{
		CompanyID: companyID,
		Year:      year,
		BonusType: c.Query("bonus_type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	styles := eidBonusSheetStyles(f)

	// Sheet 1: Summary grouped by line
	lineMap := make(map[string][]models.EidBonus)
	for _, b := range bonuses {
		lineName := "No Line"
		if b.Employee.LineRef != nil {
			lineName = b.Employee.LineRef.Name
		}
		lineMap[lineName] = append(lineMap[lineName], b)
	}

	sortedNames := make([]string, 0, len(lineMap))
	for k := range lineMap {
		sortedNames = append(sortedNames, k)
	}
	sort.Strings(sortedNames)

	f.SetSheetName("Sheet1", "Summary")
	for i, h := range []string{"Sl", "Line", "Employee ID", "Name", "Account Number", "Bonus Amount"} {
		col := string(rune('A' + i))
		f.SetCellValue("Summary", fmt.Sprintf("%s1", col), h)
		f.SetColWidth("Summary", col, col, []float64{6, 16, 16, 30, 22, 16}[i])
		f.SetCellStyle("Summary", fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), styles.header)
	}
	f.SetRowHeight("Summary", 1, 30)

	row := 2
	sl := 0
	for _, lineName := range sortedNames {
		items := lineMap[lineName]
		var lineTotal float64
		for _, it := range items {
			lineTotal += it.BonusAmount
		}

		f.SetCellValue("Summary", fmt.Sprintf("A%d", row), "")
		f.SetCellValue("Summary", fmt.Sprintf("B%d", row), fmt.Sprintf("%s  (%d employees)", lineName, len(items)))
		f.MergeCell("Summary", fmt.Sprintf("B%d", row), fmt.Sprintf("E%d", row))
		f.SetCellValue("Summary", fmt.Sprintf("F%d", row), "")
		for i := 0; i < 6; i++ {
			col := string(rune('A' + i))
			f.SetCellStyle("Summary", fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), styles.line)
		}
		f.SetRowHeight("Summary", row, 22)
		row++

		for _, it := range items {
			sl++
			f.SetCellValue("Summary", fmt.Sprintf("A%d", row), sl)
			f.SetCellStyle("Summary", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles.data)
			f.SetCellValue("Summary", fmt.Sprintf("B%d", row), lineName)
			f.SetCellStyle("Summary", fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles.data)
			f.SetCellValue("Summary", fmt.Sprintf("C%d", row), it.Employee.EmployeeID)
			f.SetCellStyle("Summary", fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styles.data)
			f.SetCellValue("Summary", fmt.Sprintf("D%d", row), it.Employee.NameEn)
			f.SetCellStyle("Summary", fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styles.data)
			f.SetCellValue("Summary", fmt.Sprintf("E%d", row), it.Employee.AccountNumber)
			f.SetCellStyle("Summary", fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styles.data)
			f.SetCellValue("Summary", fmt.Sprintf("F%d", row), it.BonusAmount)
			f.SetCellStyle("Summary", fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styles.money)
			f.SetRowHeight("Summary", row, 20)
			row++
		}

		f.SetCellValue("Summary", fmt.Sprintf("A%d", row), "")
		f.SetCellValue("Summary", fmt.Sprintf("B%d", row), "")
		f.SetCellValue("Summary", fmt.Sprintf("C%d", row), "")
		f.SetCellValue("Summary", fmt.Sprintf("D%d", row), "")
		f.SetCellValue("Summary", fmt.Sprintf("E%d", row), "Line Total")
		for i := 0; i < 5; i++ {
			col := string(rune('A' + i))
			f.SetCellStyle("Summary", fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), styles.subtotal)
		}
		f.SetCellValue("Summary", fmt.Sprintf("F%d", row), lineTotal)
		f.SetCellStyle("Summary", fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styles.moneyBold)
		f.SetRowHeight("Summary", row, 22)
		row++
	}

	f.SetSheetView("Summary", -1, &excelize.ViewOptions{
		ShowGridLines: func(b bool) *bool { return &b }(false),
	})

	filename := fmt.Sprintf("eid_bonus_%d.xlsx", year)

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(c.Writer)
}
