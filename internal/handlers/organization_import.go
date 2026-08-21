package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type OrganizationImportHandler struct{}

func NewOrganizationImportHandler() *OrganizationImportHandler {
	return &OrganizationImportHandler{}
}

var orgImportHeaders = []string{
	"Company(En)",
	"Company(Bn)",
	"Department(En)",
	"Department(Bn)",
	"Section(En)",
	"Section(Bn)",
	"Designation(En)",
	"Designation(Bn)",
	"Line(En)",
	"Line(Bn)",
}

// DownloadOrganizationTemplate godoc
//
//	@Summary      Download organization import template
//	@Description  Download Excel template for bulk organization import with Company(En/Bn), Department(En/Bn), Section(En/Bn), Designation(En/Bn), Line(En/Bn)
//	@Tags         Organization
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Success      200  {file}  file
//	@Router       /organization/template [get]
func (h *OrganizationImportHandler) DownloadTemplate(c *gin.Context) {
	f := excelize.NewFile()

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#4472C4"},
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
			Size:  12,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
			{Type: "top", Color: "FFFFFF", Style: 1},
			{Type: "bottom", Color: "FFFFFF", Style: 1},
		},
	})

	calcColWidth := func(texts ...string) float64 {
		max := 0.0
		for _, t := range texts {
			w := float64(len(t)) * 1.4
			if w < 14 {
				w = 14
			}
			if w > 45 {
				w = 45
			}
			if w > max {
				max = w
			}
		}
		return max
	}

	setupSheet := func(sheet string, maxRow int) {
		for r := 1; r <= maxRow; r++ {
			f.SetRowHeight(sheet, r, 26)
		}
		for i, hdr := range orgImportHeaders {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheet, fmt.Sprintf("%s1", col), hdr)
			f.SetCellStyle(sheet, fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), headerStyle)
		}
		f.SetPanes(sheet, &excelize.Panes{
			Freeze:      true,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		})
	}

	dataEntry := "Data Entry"
	f.SetSheetName("Sheet1", dataEntry)
	setupSheet(dataEntry, 200)

	demo := "Demo"
	f.NewSheet(demo)
	setupSheet(demo, 6)

	demoData := [][]interface{}{
		{"PeopleHub Technologies Ltd.", "পিপলহাব টেকনোলজিস লি.", "Human Resources", "মানব সম্পদ", "Recruitment", "নিয়োগ", "HR Executive", "এইচআর এক্সিকিউটিভ", "Sewing Line 1", "সেলাই লাইন ১"},
		{"PeopleHub Technologies Ltd.", "পিপলহাব টেকনোলজিস লি.", "Human Resources", "মানব সম্পদ", "Training", "প্রশিক্ষণ", "Training Officer", "প্রশিক্ষণ কর্মকর্তা", "", ""},
		{"PeopleHub Technologies Ltd.", "পিপলহাব টেকনোলজিস লি.", "Finance", "অর্থ", "Accounts", "হিসাব", "Accountant", "হিসাবরক্ষক", "", ""},
		{"PeopleHub Technologies Ltd.", "পিপলহাব টেকনোলজিস লি.", "Production", "উত্পাদন", "Cutting", "কাটিং", "Cutting Master", "কাটিং মাস্টার", "Cutting Line 1", "কাটিং লাইন ১"},
		{"PeopleHub Technologies Ltd.", "পিপলহাব টেকনোলজিস লি.", "Production", "উত্পাদন", "Sewing", "সেলাই", "Sewing Operator", "সেলাই অপারেটর", "Sewing Line 2", "সেলাই লাইন ২"},
	}

	for rowIdx, rowData := range demoData {
		row := rowIdx + 2
		for colIdx, val := range rowData {
			col, _ := excelize.ColumnNumberToName(colIdx + 1)
			f.SetCellValue(demo, fmt.Sprintf("%s%d", col, row), val)
			f.SetCellStyle(demo, fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), centerStyle)
		}
		f.SetRowHeight(demo, row, 26)
	}

	for i, hdr := range orgImportHeaders {
		col, _ := excelize.ColumnNumberToName(i + 1)
		longest := hdr
		for _, rowData := range demoData {
			if i < len(rowData) {
				if val, ok := rowData[i].(string); ok && len(val) > len(longest) {
					longest = val
				}
			}
		}
		w := calcColWidth(hdr, longest)
		f.SetColWidth(dataEntry, col, col, w)
		f.SetColWidth(demo, col, col, w)
	}

	if idx, _ := f.GetSheetIndex(dataEntry); idx > 0 {
		f.SetActiveSheet(idx)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=organization_import_template.xlsx")
	f.Write(c.Writer)
}

// ImportOrganizationExcel godoc
//
//	@Summary      Bulk import organization entities
//	@Description  Bulk create/update companies, departments, sections, designations, lines from Excel with flat hierarchy columns. Companies are looked up by name or auto-created.
//	@Tags         Organization
//	@Security     BearerAuth
//	@Accept       multipart/form-data
//	@Produce      json
//	@Param        file formData file true "Excel file (.xlsx)"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /organization/import [post]
func (h *OrganizationImportHandler) ImportExcel(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required field 'file' is missing. Please upload a .xlsx file using a multipart/form-data request with field name 'file'."})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".xlsx") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .xlsx files are supported. Received: " + header.Filename})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse Excel file"})
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())
	if idx, _ := f.GetSheetIndex("Data Entry"); idx > 0 {
		sheetName = "Data Entry"
	}

	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no data rows found in sheet"})
		return
	}

	headerRow := rows[0]
	colMap := make(map[string]int)
	for i, h := range headerRow {
		h = strings.TrimSpace(h)
		colMap[h] = i
	}

	requiredCols := []string{"Company(En)", "Department(En)", "Section(En)"}
	for _, col := range requiredCols {
		if _, ok := colMap[col]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("required column '%s' not found", col)})
			return
		}
	}

	type entityResult struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Action string `json:"action"`
	}

	var results []entityResult
	var errs []string

	companyCache := map[string]string{}
	deptCache := map[string]string{}

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) == 0 || allEmptyOrg(row) {
			continue
		}

		companyName := getColOrg(row, colMap, "Company(En)")
		companyNameBn := getColOrg(row, colMap, "Company(Bn)")
		deptName := getColOrg(row, colMap, "Department(En)")
		deptNameBn := getColOrg(row, colMap, "Department(Bn)")
		secName := getColOrg(row, colMap, "Section(En)")
		secNameBn := getColOrg(row, colMap, "Section(Bn)")
		desigName := getColOrg(row, colMap, "Designation(En)")
		desigNameBn := getColOrg(row, colMap, "Designation(Bn)")
		lineName := getColOrg(row, colMap, "Line(En)")
		lineNameBn := getColOrg(row, colMap, "Line(Bn)")

		if companyName == "" {
			errs = append(errs, fmt.Sprintf("Row %d: Company(En) is required", rowIdx+1))
			continue
		}
		if deptName == "" {
			errs = append(errs, fmt.Sprintf("Row %d: Department(En) is required", rowIdx+1))
			continue
		}
		if secName == "" {
			errs = append(errs, fmt.Sprintf("Row %d: Section(En) is required", rowIdx+1))
			continue
		}

		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			// Resolve Company
			companyKey := strings.ToLower(strings.TrimSpace(companyName))
			companyID, ok := companyCache[companyKey]
			if !ok {
				var company models.Company
				if err := tx.Where("LOWER(company_name_en) = ?", companyKey).First(&company).Error; err != nil {
					slug := strings.ToLower(strings.ReplaceAll(companyName, " ", "-"))
					slug = strings.NewReplacer(".", "", ",", "", "(", "", ")", "", "'", "", "&", "and").Replace(slug)
					company = models.Company{
						CompanyNameEn: companyName,
						CompanyNameBn: companyNameBn,
						Slug:          slug,
						Status:        "active",
					}
					if err := tx.Create(&company).Error; err != nil {
						return fmt.Errorf("failed to create company '%s': %s", companyName, err)
					}
					results = append(results, entityResult{Type: "Company", Name: companyName, Action: "created"})
				}
				companyID = company.ID
				companyCache[companyKey] = companyID
			}

			// Upsert Department (lookup by company_id + name)
			deptKey := companyID + "|" + strings.ToLower(deptName)
			deptID, deptExists := deptCache[deptKey]
			if !deptExists {
				var dept models.Department
				if err := tx.Where("company_id = ? AND LOWER(name) = ?", companyID, strings.ToLower(deptName)).First(&dept).Error; err != nil {
					dept = models.Department{
						CompanyID: &companyID,
						Name:      deptName,
						NameBn:    deptNameBn,
						Status:    "active",
					}
					if err := tx.Create(&dept).Error; err != nil {
						return fmt.Errorf("failed to create department '%s': %s", deptName, err)
					}
					results = append(results, entityResult{Type: "Department", Name: deptName, Action: "created"})
				} else if deptNameBn != "" {
					tx.Model(&models.Department{}).Where("id = ?", dept.ID).Update("name_bn", deptNameBn)
				}
				deptID = dept.ID
				deptCache[deptKey] = deptID
			}

			// Upsert Section (lookup by department_id + name)
			var section models.Section
			secExists := tx.Where("name = ? AND department_id = ?", secName, deptID).First(&section).Error == nil
			var secID string
			if !secExists {
				sec := &models.Section{
					Name:         secName,
					NameBn:       secNameBn,
					DepartmentID: deptID,
				}
				if err := tx.Create(sec).Error; err != nil {
					return fmt.Errorf("failed to create section '%s': %s", secName, err)
				}
				secID = sec.ID
				results = append(results, entityResult{Type: "Section", Name: secName, Action: "created"})
			} else {
				secID = section.ID
				if secNameBn != "" {
					tx.Model(&models.Section{}).Where("id = ?", secID).Update("name_bn", secNameBn)
				}
			}

			// Upsert Designation (lookup by section_id + name)
			if desigName != "" {
				var desig models.Designation
				desigExists := tx.Where("name = ? AND section_id = ?", desigName, secID).First(&desig).Error == nil
				if !desigExists {
					d := &models.Designation{
						Name:      desigName,
						NameBn:    desigNameBn,
						SectionID: secID,
					}
					if err := tx.Create(d).Error; err != nil {
						return fmt.Errorf("failed to create designation '%s': %s", desigName, err)
					}
					results = append(results, entityResult{Type: "Designation", Name: desigName, Action: "created"})
				} else if desigNameBn != "" {
					tx.Model(&models.Designation{}).Where("id = ?", desig.ID).Update("name_bn", desigNameBn)
				}
			}

			// Upsert Line (lookup by section_id + name)
			if lineName != "" {
				var line models.Line
				lineExists := tx.Where("name = ? AND section_id = ?", lineName, secID).First(&line).Error == nil
				if !lineExists {
					l := &models.Line{
						Name:      lineName,
						NameBn:    lineNameBn,
						SectionID: secID,
					}
					if err := tx.Create(l).Error; err != nil {
						return fmt.Errorf("failed to create line '%s': %s", lineName, err)
					}
					results = append(results, entityResult{Type: "Line", Name: lineName, Action: "created"})
				} else if lineNameBn != "" {
					tx.Model(&models.Line{}).Where("id = ?", line.ID).Update("name_bn", lineNameBn)
				}
			}

			return nil
		}); err != nil {
			errs = append(errs, fmt.Sprintf("Row %d: %s", rowIdx+1, err.Error()))
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   fmt.Sprintf("Errors in %d row(s)", len(errs)),
			"errors":  errs,
			"results": results,
		})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No data rows to import", "total": 0})
		return
	}

	counts := map[string]int{}
	for _, r := range results {
		counts[r.Type+"_"+r.Action]++
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Import complete. %d companies, %d departments, %d sections, %d designations, %d lines",
			counts["Company_created"],
			counts["Department_created"]+counts["Department_updated"],
			counts["Section_created"]+counts["Section_updated"],
			counts["Designation_created"]+counts["Designation_updated"],
			counts["Line_created"]+counts["Line_updated"]),
		"total":   len(results),
		"results": results,
	})
}

func getColOrg(row []string, colMap map[string]int, key string) string {
	if idx, ok := colMap[key]; ok && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func allEmptyOrg(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
