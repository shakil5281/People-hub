package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type SalaryAccountImportHandler struct{}

func NewSalaryAccountImportHandler() *SalaryAccountImportHandler {
	return &SalaryAccountImportHandler{}
}

// DownloadSalaryAccountTemplate godoc
//
// @Summary      Download salary account Excel template
// @Description  Download a demo Excel template formatted with EmployeeId, AccountType, and AccountNumber
// @Tags         Employees
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success      200 {file} binary
// @Failure      500 {object} map[string]string
// @Router       /employees/salary-account/template [get]
func (h *SalaryAccountImportHandler) DownloadSalaryAccountTemplate(c *gin.Context) {
	f := excelize.NewFile()
	sheet := "SalaryAccounts"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"EmployeeId", "AccountType", "AccountNumber"}
	for i, hdr := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(sheet, cell, hdr)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#1E3A8A"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)
	f.SetRowHeight(sheet, 1, 26)

	// Demo sample data rows (only mCash and Card)
	demoData := [][]string{
		{"EMP-1001", "mCash", "017123456789"},
		{"EMP-1002", "Card", "12345678901234567"},
		{"EMP-1003", "mCash", "018987654321"},
		{"EMP-1004", "Card", "98765432101234567"},
	}

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
		},
	})

	for rIdx, row := range demoData {
		rowNum := rIdx + 2
		for cIdx, val := range row {
			col, _ := excelize.ColumnNumberToName(cIdx + 1)
			cell := fmt.Sprintf("%s%d", col, rowNum)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet, cell, cell, dataStyle)
		}
		f.SetRowHeight(sheet, rowNum, 20)
	}

	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 20)
	f.SetColWidth(sheet, "C", "C", 30)

	buf, err := f.WriteToBuffer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel template"})
		return
	}

	filename := "employee_salary_account_template.xlsx"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// ImportSalaryAccountExcel godoc
//
// @Summary      Import salary accounts from Excel
// @Description  Bulk update only the account_type and account_number fields for matched EmployeeId
// @Tags         Employees
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Excel file (.xlsx, .xls)"
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} map[string]string
// @Failure      500  {object} map[string]string
// @Router       /employees/salary-account/import [post]
func (h *SalaryAccountImportHandler) ImportSalaryAccountExcel(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Please upload a valid Excel file."})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file content"})
		return
	}

	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Excel file format"})
		return
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file has no sheets"})
		return
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read rows from sheet"})
		return
	}

	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file is empty or missing data rows"})
		return
	}

	// Identify column indexes
	headerRow := rows[0]
	empIDCol := -1
	accTypeCol := -1
	accNumCol := -1

	normalize := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "_", "")
		s = strings.ReplaceAll(s, "-", "")
		return s
	}

	for i, hVal := range headerRow {
		norm := normalize(hVal)
		if norm == "employeeid" || norm == "empid" || norm == "id" {
			empIDCol = i
		} else if norm == "accounttype" || norm == "acctype" || norm == "type" {
			accTypeCol = i
		} else if norm == "accountnumber" || norm == "accnumber" || norm == "accountno" || norm == "accno" || norm == "number" {
			accNumCol = i
		}
	}

	// Fallback to positional columns if header names did not match exactly
	if empIDCol == -1 && len(headerRow) > 0 {
		empIDCol = 0
	}
	if accTypeCol == -1 && len(headerRow) > 1 {
		accTypeCol = 1
	}
	if accNumCol == -1 && len(headerRow) > 2 {
		accNumCol = 2
	}

	if empIDCol == -1 || accTypeCol == -1 || accNumCol == -1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not identify required columns: EmployeeId, AccountType, AccountNumber",
		})
		return
	}

	// Fetch all active/existing employees to cache IDs
	var allEmployees []models.Employee
	database.DB.Select("id, employee_id").Find(&allEmployees)
	empMap := make(map[string]string, len(allEmployees)) // key: lower(employee_id) -> id
	for _, e := range allEmployees {
		key := strings.ToLower(strings.TrimSpace(e.EmployeeID))
		if key != "" {
			empMap[key] = e.ID
		}
	}

	userID := c.GetString("user_id")
	now := time.Now()

	updatedCount := 0
	skippedCount := 0
	var errorMessages []string

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for rIdx := 1; rIdx < len(rows); rIdx++ {
			row := rows[rIdx]
			if len(row) == 0 {
				continue
			}

			getVal := func(colIdx int) string {
				if colIdx >= 0 && colIdx < len(row) {
					return strings.TrimSpace(row[colIdx])
				}
				return ""
			}

			empID := getVal(empIDCol)
			accType := getVal(accTypeCol)
			accNum := getVal(accNumCol)

			if empID == "" {
				// Empty row or missing ID
				continue
			}

			targetID, exists := empMap[strings.ToLower(empID)]
			if !exists {
				skippedCount++
				errorMessages = append(errorMessages, fmt.Sprintf("Row %d: Employee ID %q not found", rIdx+1, empID))
				continue
			}

			// Clean and standardize account type (only mCash and Card allowed)
			if strings.EqualFold(accType, "mcash") {
				accType = "mCash"
			} else if strings.EqualFold(accType, "card") {
				accType = "Card"
			} else if accType != "" {
				skippedCount++
				errorMessages = append(errorMessages, fmt.Sprintf("Row %d: Invalid account_type %q for employee %q (must be mCash or Card)", rIdx+1, accType, empID))
				continue
			}

			if msg := validateAccount(accType, accNum); msg != "" {
				skippedCount++
				errorMessages = append(errorMessages, fmt.Sprintf("Row %d (%s): %s", rIdx+1, empID, msg))
				continue
			}

			updates := map[string]interface{}{
				"account_type":   accType,
				"account_number": accNum,
				"updated_at":     now,
			}
			if userID != "" {
				updates["updated_by"] = userID
			}

			// Strictly update ONLY account_type and account_number
			if updErr := tx.Model(&models.Employee{}).Where("id = ?", targetID).Updates(updates).Error; updErr != nil {
				skippedCount++
				errorMessages = append(errorMessages, fmt.Sprintf("Row %d: Failed to update employee %q: %v", rIdx+1, empID, updErr))
				continue
			}

			updatedCount++
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    fmt.Sprintf("Salary account import completed. %d records updated, %d skipped.", updatedCount, skippedCount),
		"total_rows": len(rows) - 1,
		"updated":    updatedCount,
		"skipped":    skippedCount,
		"errors":     errorMessages,
	})
}

// ExportSalaryAccountExcel godoc
//
// @Summary      Export salary accounts to Excel
// @Description  Export filtered employee salary accounts list to Excel
// @Tags         Employees
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success      200 {file} binary
// @Failure      500 {object} map[string]string
// @Router       /employees/salary-account/export [get]
func (h *SalaryAccountImportHandler) ExportSalaryAccountExcel(c *gin.Context) {
	var employees []models.Employee
	query := database.DB.
		Preload("Company").
		Preload("Department").
		Preload("DesignationRef").
		Preload("SectionRef").
		Preload("LineRef").
		Preload("GroupRef").
		Preload("FloorRef").
		Preload("Shift")

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
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("employee_id"); v != "" {
		query = query.Where("employee_id ILIKE ?", "%"+v+"%")
	}
	if v := c.Query("account_type"); v != "" {
		if strings.EqualFold(v, "none") || strings.EqualFold(v, "unassigned") || strings.EqualFold(v, "hold") {
			query = query.Where("account_type IS NULL OR TRIM(account_type) = '' OR LOWER(account_type) IN ('none', 'hold')")
		} else {
			query = query.Where("LOWER(account_type) = LOWER(?)", v)
		}
	}

	if err := query.Order("LENGTH(employee_id) ASC, employee_id ASC").Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "SalaryAccounts"
	f.SetSheetName("Sheet1", sheet)

	companyName := "PeopleHub"
	if len(employees) > 0 && employees[0].Company.CompanyNameEn != "" {
		companyName = employees[0].Company.CompanyNameEn
	}

	// Title block
	f.MergeCell(sheet, "A1", "K1")
	f.SetCellValue(sheet, "A1", companyName)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 15, Color: "#1E3A8A"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "K1", titleStyle)
	f.SetRowHeight(sheet, 1, 28)

	f.MergeCell(sheet, "A2", "K2")
	f.SetCellValue(sheet, "A2", "Employee Salary Account Statement")
	subTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#4B5563"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A2", "K2", subTitleStyle)
	f.SetRowHeight(sheet, 2, 20)

	f.MergeCell(sheet, "A3", "K3")
	f.SetCellValue(sheet, "A3", fmt.Sprintf("Report Generated: %s | Total Records: %d", time.Now().Format("02-Jan-2006 03:04 PM"), len(employees)))
	dateStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Size: 9, Color: "#6B7280"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A3", "K3", dateStyle)
	f.SetRowHeight(sheet, 3, 18)

	// Header row at row 5
	headers := []string{
		"Sl No", "Employee ID", "Punch No", "Employee Name", "Designation",
		"Department", "Section", "Gross Salary", "Account Type", "Account Number", "Status",
	}

	for i, hVal := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s5", col)
		f.SetCellValue(sheet, cell, hVal)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E3A8A"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	f.SetRowStyle(sheet, 5, 5, headerStyle)
	f.SetRowHeight(sheet, 5, 24)

	rowStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
		},
	})

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
		},
	})

	rightStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
		},
	})

	for i, emp := range employees {
		rowNum := i + 6

		desig := ""
		if emp.DesignationRef != nil {
			desig = emp.DesignationRef.Name
		}
		dept := ""
		if emp.Department != nil {
			dept = emp.Department.Name
		}
		sec := ""
		if emp.SectionRef != nil {
			sec = emp.SectionRef.Name
		}

		accType := emp.AccountType
		if accType == "" {
			accType = "Unassigned"
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), emp.EmployeeID)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), emp.PunchNumber)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), emp.NameEn)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowNum), desig)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), dept)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowNum), sec)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowNum), emp.GrossSalary)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", rowNum), accType)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", rowNum), emp.AccountNumber)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", rowNum), emp.Status)

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("A%d", rowNum), centerStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("C%d", rowNum), centerStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", rowNum), fmt.Sprintf("G%d", rowNum), rowStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("H%d", rowNum), fmt.Sprintf("H%d", rowNum), rightStyle)
		f.SetCellStyle(sheet, fmt.Sprintf("I%d", rowNum), fmt.Sprintf("K%d", rowNum), centerStyle)

		f.SetRowHeight(sheet, rowNum, 20)
	}

	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", 14)
	f.SetColWidth(sheet, "C", "C", 14)
	f.SetColWidth(sheet, "D", "D", 26)
	f.SetColWidth(sheet, "E", "E", 20)
	f.SetColWidth(sheet, "F", "F", 18)
	f.SetColWidth(sheet, "G", "G", 18)
	f.SetColWidth(sheet, "H", "H", 14)
	f.SetColWidth(sheet, "I", "I", 14)
	f.SetColWidth(sheet, "J", "J", 22)
	f.SetColWidth(sheet, "K", "K", 12)

	buf, err := f.WriteToBuffer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}

	filename := fmt.Sprintf("salary_accounts_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
