package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type importLookups struct {
	companies    map[string]string
	shifts       map[string]string
	departments  map[string]string
	sections     map[string]string
	designations map[string]string // key: sectionID|name
	lines        map[string]string // key: sectionID|name
	groups       map[string]string
	floors       map[string]string
	divisions    map[string]string
	districts    map[string]string
	upazilas     map[string]string
	unions       map[string]string
	employees    map[string]string // employee_id (business key) → UUID
}

func loadImportLookups(db *gorm.DB) *importLookups {
	l := &importLookups{
		companies:    make(map[string]string),
		shifts:       make(map[string]string),
		departments:  make(map[string]string),
		sections:     make(map[string]string),
		designations: make(map[string]string),
		lines:        make(map[string]string),
		groups:       make(map[string]string),
		floors:       make(map[string]string),
		divisions:    make(map[string]string),
		districts:    make(map[string]string),
		upazilas:     make(map[string]string),
		unions:       make(map[string]string),
		employees:    make(map[string]string),
	}

	var companies []models.Company
	db.Find(&companies)
	for _, c := range companies {
		name := strings.TrimSpace(strings.ToLower(c.CompanyNameEn))
		if name != "" {
			l.companies[name] = c.ID
		}
	}

	var shifts []models.Shift
	db.Find(&shifts)
	for _, s := range shifts {
		name := strings.TrimSpace(strings.ToLower(s.Name))
		if name != "" {
			l.shifts[name] = s.ID
		}
	}

	var depts []models.Department
	db.Find(&depts)
	for _, d := range depts {
		name := strings.TrimSpace(strings.ToLower(d.Name))
		if name != "" {
			l.departments[name] = d.ID
		}
	}

	var sections []models.Section
	db.Find(&sections)
	for _, s := range sections {
		name := strings.TrimSpace(strings.ToLower(s.Name))
		if name != "" {
			l.sections[name] = s.ID
		}
	}

	var desigs []models.Designation
	db.Find(&desigs)
	for _, d := range desigs {
		name := strings.TrimSpace(strings.ToLower(d.Name))
		if name != "" && d.SectionID != "" {
			l.designations[d.SectionID+"|"+name] = d.ID
		}
	}

	var lines []models.Line
	db.Find(&lines)
	for _, li := range lines {
		name := strings.TrimSpace(strings.ToLower(li.Name))
		if name != "" && li.SectionID != "" {
			l.lines[li.SectionID+"|"+name] = li.ID
		}
	}

	var grps []models.Group
	db.Find(&grps)
	for _, g := range grps {
		name := strings.TrimSpace(strings.ToLower(g.Name))
		if name != "" {
			l.groups[name] = g.ID
		}
	}

	var floors []models.Floor
	db.Find(&floors)
	for _, f := range floors {
		name := strings.TrimSpace(strings.ToLower(f.Name))
		if name != "" {
			l.floors[name] = f.ID
		}
	}

	var divs []models.Division
	db.Find(&divs)
	for _, d := range divs {
		name := strings.TrimSpace(strings.ToLower(d.Name))
		if name != "" {
			l.divisions[name] = d.ID
		}
	}

	var dists []models.District
	db.Find(&dists)
	for _, d := range dists {
		name := strings.TrimSpace(strings.ToLower(d.Name))
		if name != "" {
			l.districts[name] = d.ID
		}
	}

	var upazilas []models.Upazila
	db.Find(&upazilas)
	for _, u := range upazilas {
		name := strings.TrimSpace(strings.ToLower(u.Name))
		if name != "" {
			l.upazilas[name] = u.ID
		}
	}

	var unions []models.Union
	db.Find(&unions)
	for _, u := range unions {
		name := strings.TrimSpace(strings.ToLower(u.Name))
		if name != "" {
			l.unions[name] = u.ID
		}
	}

	var employees []models.Employee
	db.Select("id, employee_id").Find(&employees)
	for _, e := range employees {
		key := strings.TrimSpace(strings.ToLower(e.EmployeeID))
		if key != "" {
			l.employees[key] = e.ID
		}
	}

	return l
}

type EmployeeImportHandler struct {
	employeeRepo *repository.EmployeeRepository
}

func NewEmployeeImportHandler(employeeRepo *repository.EmployeeRepository) *EmployeeImportHandler {
	return &EmployeeImportHandler{employeeRepo: employeeRepo}
}

var importHeaders = []string{
	"Sl No",
	"EmployeeId", "PunchNumber",
	"Name (En)", "Name (Bn)",
	"fatherName", "motherName", "spouseName",
	"date_of_birth", "Gender", "Blood_group", "MaritalStatus", "Religion", "Nationality",
	"NidNumber", "Phone", "Email",
	"EmergencyContact", "EmergencyPhone", "NumberOfDependents",
	"Present_Division", "Present_District", "Present_Upazila", "Present_Union", "Present_Address",
	"Permanent_Division", "Permanent_District", "Permanent_Upazila", "Permanent_Union", "Permanent_Address",
	"CompanyName", "EmployeeType", "JoiningDate",
	"shiftName", "Department", "Section", "Designation", "ReportsTo", "Line", "Group", "Floor",
	"Grade", "Status", "OverTimeStatus",
	"Gross_Salary", "Basic_Salary", "House_Rent", "Transport_Allowance", "Food_Allowance", "Medical_Allowance", "Other_Allowance",
	"accountType", "AccountNumber",
}

var importDescriptions = []string{
	"Serial number (auto)",
	"Unique employee ID *", "Badge/punch number *",
	"Full name in English *", "Name in Bangla",
	"Father's name", "Mother's name", "Spouse's name",
	"Date of birth (YYYY-MM-DD)", "Male/Female/Other", "A+/A-/B+/B-/AB+/AB-/O+/O-", "Marital status", "Religion", "Nationality",
	"National ID number", "Phone number", "Email address",
	"Emergency contact person", "Emergency phone", "Number of dependents",
	"Present division name", "Present district name", "Present upazila name", "Present union name", "Present address",
	"Permanent division name", "Permanent district name", "Permanent upazila name", "Permanent union name", "Permanent address",
	"Company name *", "Regular/Lefty/Close/Resign", "Joining date (YYYY-MM-DD) *",
	"Shift name", "Department name", "Section name", "Designation name", "Manager employee ID", "Line name", "Group name", "Floor name",
	"Grade", "active/inactive", "true/false",
	"Gross salary", "Basic salary", "House rent", "Transport allowance", "Food allowance", "Medical allowance", "Other allowance",
	"mCash or Card", "Account number (12 for mCash, 17 for Card)",
}

func (h *EmployeeImportHandler) DownloadTemplate(c *gin.Context) {
	f := excelize.NewFile()

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	calcColWidth := func(texts ...string) float64 {
		max := 0.0
		for _, t := range texts {
			w := float64(len(t)) * 1.3
			if w < 12 {
				w = 12
			}
			if w > 40 {
				w = 40
			}
			if w > max {
				max = w
			}
		}
		return max
	}

	setupSheet := func(sheet string, maxRow int) {
		for r := 1; r <= maxRow; r++ {
			f.SetRowHeight(sheet, r, 25)
		}
		for i, hdr := range importHeaders {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheet, fmt.Sprintf("%s1", col), hdr)
			f.SetCellStyle(sheet, fmt.Sprintf("%s1", col), fmt.Sprintf("%s1", col), centerStyle)
		}
		f.SetPanes(sheet, &excelize.Panes{
			Freeze:      true,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		})
	}

	// Sheet 1: Data Entry
	dataEntry := "Data Entry"
	f.SetSheetName("Sheet1", dataEntry)
	setupSheet(dataEntry, 200)

	// Sheet 2: Demo
	demo := "Demo"
	f.NewSheet(demo)
	setupSheet(demo, 4)

	demoData := [][]interface{}{
		{
			"1", "DEMO001", "2001",
			"Shakil Ahmed", "শাকিল আহমেদ",
			"Abdul Halim", "Rahima Khatun", "Jorina Khatun",
			"1995-06-15", "Male", "B+", "Married", "Islam", "Bangladeshi",
			"1995123456", "01711111111", "shakil@example.com",
			"Karim Ahmed", "01711111112", "2",
			"Dhaka", "Dhaka", "Uttara", "Uttara West", "House 12, Road 5, Sector 3, Uttara",
			"Dhaka", "Gazipur", "Gazipur Sadar", "Bason", "Village: Uttarpara",
			"Ekushe Fashions", "Regular", "2023-01-01",
			"General", "IT", "Software", "Jr. Executive", "", "Line-2", "Group-A", "Floor-1",
			"Grade-5", "active", "true",
			"50000", "25000", "10000", "5000", "4500", "3000", "2500",
			"mCash", "012345678901",
		},
		{
			"2", "DEMO002", "2002",
			"Fatima Begum", "ফাতিমা বেগম",
			"Mohammad Ali", "Saleha Begum", "",
			"1998-09-22", "Female", "O+", "Unmarried", "Islam", "Bangladeshi",
			"1998123456", "01722222222", "fatima@example.com",
			"Mohammad Ali", "01722222223", "0",
			"Dhaka", "Dhaka", "Tejgaon", "Tejgaon Ind. Area", "Flat 3B, 45 Elephant Road",
			"Chattogram", "Cumilla", "Cumilla Sadar", "Jhawtala", "23/1 Old Town",
			"Ekushe Fashions", "Regular", "2023-06-01",
			"General", "HR", "Recruitment", "Jr. Executive", "DEMO001", "Line-1", "Group-B", "Floor-2",
			"Grade-3", "active", "false",
			"36000", "18000", "7200", "450", "1250", "750", "1800",
			"Card", "12345678901234567",
		},
		{
			"3", "DEMO003", "2003",
			"Rahim Uddin", "রহিম উদ্দিন",
			"Karim Uddin", "Jahanara Begum", "Saleha Begum",
			"1992-03-08", "Male", "AB-", "Married", "Islam", "Bangladeshi",
			"1992123456", "01733333333", "rahim@example.com",
			"Karim Uddin", "01733333334", "4",
			"Khulna", "Khulna", "Khulna Sadar", "Khalishpur", "456 New Market",
			"Barishal", "Barishal", "Barishal Sadar", "Kashipur", "87 Old Road",
			"Ekushe Fashions", "Regular", "2022-07-15",
			"General", "Production", "Cutting", "Manager", "DEMO002", "Line-3", "Group-C", "Floor-1",
			"Grade-8", "active", "true",
			"70000", "35000", "14000", "450", "1250", "750", "3500",
			"mCash", "987654321012",
		},
	}

	for rowIdx, rowData := range demoData {
		row := rowIdx + 2
		for colIdx, val := range rowData {
			col, _ := excelize.ColumnNumberToName(colIdx + 1)
			f.SetCellValue(demo, fmt.Sprintf("%s%d", col, row), val)
			f.SetCellStyle(demo, fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), centerStyle)
		}
		f.SetRowHeight(demo, row, 25)
	}

	// Auto-fit column widths
	for i, hdr := range importHeaders {
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
	c.Header("Content-Disposition", "attachment; filename=employee_import_template.xlsx")
	f.Write(c.Writer)
}

// ImportExcel godoc
//
// @Summary      Import employees from Excel
// @Description  Bulk import employees from uploaded Excel file. Matches on employee_id for upsert.
// @Tags         Employees
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Excel file (.xlsx)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /employees/import [post]
func (h *EmployeeImportHandler) ImportExcel(c *gin.Context) {
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

	requiredFields := []string{"EmployeeId", "Name (En)", "JoiningDate"}
	for _, rf := range requiredFields {
		if _, ok := colMap[rf]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("required column '%s' not found", rf)})
			return
		}
	}

	lookups := loadImportLookups(database.DB)

	var parsed []parsedRow
	var errors []string

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) == 0 || allEmpty(row) {
			continue
		}
		empID := getCell(row, colMap, "EmployeeId")
		if empID == "" {
			errors = append(errors, fmt.Sprintf("Row %d: EmployeeId is required", rowIdx+1))
			continue
		}

		dob := getCell(row, colMap, "date_of_birth")
		if dob != "" && !isValidDate(dob) {
			errors = append(errors, fmt.Sprintf("Row %d: invalid date_of_birth format '%s' (use YYYY-MM-DD)", rowIdx+1, dob))
			continue
		}

		joiningDate := getCell(row, colMap, "JoiningDate")
		if joiningDate != "" && !isValidDate(joiningDate) {
			errors = append(errors, fmt.Sprintf("Row %d: invalid JoiningDate format '%s' (use YYYY-MM-DD)", rowIdx+1, joiningDate))
			continue
		}

		parsed = append(parsed, parsedRow{
			row:     row,
			index:   rowIdx + 1,
			code:    empID,
			colMap:  colMap,
			lookups: lookups,
		})
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  fmt.Sprintf("Validation errors in %d row(s)", len(errors)),
			"errors": errors,
		})
		return
	}

	if len(parsed) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No data rows to import", "created": 0, "updated": 0, "total": 0})
		return
	}

	userID := c.GetString("user_id")
	// Company resolved from CompanyName column, fallback to auth context
	companyName := strings.TrimSpace(strings.ToLower(getCell(parsed[0].row, parsed[0].colMap, "CompanyName")))
	resolvedCompanyID := lookups.companies[companyName]
	if resolvedCompanyID == "" {
		resolvedCompanyID = c.GetString("company_id")
	}

	existingMap, err := h.employeeRepo.MapByID(resolvedCompanyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup existing employees"})
		return
	}

	var toCreate []models.Employee
	var toUpdate []models.Employee

	for _, p := range parsed {
		emp := rowToEmployee(p.row, p.colMap, resolvedCompanyID, userID, lookups)
		if _, exists := existingMap[p.code]; exists {
			emp.ID = existingMap[p.code].ID
			emp.UpdatedBy = &userID
			emp.CreatedBy = nil
			toUpdate = append(toUpdate, emp)
		} else {
			toCreate = append(toCreate, emp)
		}
	}

	created := len(toCreate)
	updated := len(toUpdate)

	if created > 0 {
		if err := h.employeeRepo.BatchCreate(toCreate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employees: " + err.Error()})
			return
		}
	}

	if updated > 0 {
		if err := h.employeeRepo.BulkUpdateByID(toUpdate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employees: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Import complete. %d created, %d updated", created, updated),
		"created": created,
		"updated": updated,
		"total":   created + updated,
	})
}

type parsedRow struct {
	row     []string
	index   int
	code    string
	colMap  map[string]int
	lookups *importLookups
}

func getCell(row []string, colMap map[string]int, key string) string {
	if idx, ok := colMap[key]; ok && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func getCellFloat(row []string, colMap map[string]int, key string) float64 {
	s := getCell(row, colMap, key)
	if s == "" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func allEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func lookupID(nameMap map[string]string, name string) *string {
	if name == "" {
		return nil
	}
	key := strings.TrimSpace(strings.ToLower(name))
	if id, ok := nameMap[key]; ok {
		return &id
	}
	return nil
}

func rowToEmployee(row []string, colMap map[string]int, companyID, userID string, lookups *importLookups) models.Employee {
	joiningDate := getCell(row, colMap, "JoiningDate")
	parsedJoining, _ := time.Parse("2006-01-02", joiningDate)

	dateOfBirth := getCell(row, colMap, "date_of_birth")

	overTimeStatus := strings.ToLower(getCell(row, colMap, "OverTimeStatus"))
	overTimeBool := overTimeStatus == "true" || overTimeStatus == "1" || overTimeStatus == "yes"

	dependentsStr := getCell(row, colMap, "NumberOfDependents")
	dependents, _ := strconv.Atoi(dependentsStr)

	sectionID := lookupID(lookups.sections, getCell(row, colMap, "Section"))

	desigName := getCell(row, colMap, "Designation")
	lineName := getCell(row, colMap, "Line")

	var desigID, lineID *string
	if sectionID != nil {
		if desigName != "" {
			key := *sectionID + "|" + strings.TrimSpace(strings.ToLower(desigName))
			if id, ok := lookups.designations[key]; ok {
				desigID = &id
			}
		}
		if lineName != "" {
			key := *sectionID + "|" + strings.TrimSpace(strings.ToLower(lineName))
			if id, ok := lookups.lines[key]; ok {
				lineID = &id
			}
		}
	}

	return models.Employee{
		CompanyID:           companyID,
		EmployeeID:          getCell(row, colMap, "EmployeeId"),
		NameEn:              getCell(row, colMap, "Name (En)"),
		NameBn:              getCell(row, colMap, "Name (Bn)"),
		FatherName:          getCell(row, colMap, "fatherName"),
		MotherName:          getCell(row, colMap, "motherName"),
		SpouseName:          getCell(row, colMap, "spouseName"),
		DateOfBirth:         dateOfBirth,
		Gender:              getCell(row, colMap, "Gender"),
		BloodGroup:          getCell(row, colMap, "Blood_group"),
		MaritalStatus:       getCell(row, colMap, "MaritalStatus"),
		Religion:            getCell(row, colMap, "Religion"),
		Nationality:         getCell(row, colMap, "Nationality"),
		NID:                 getCell(row, colMap, "NidNumber"),
		Phone:               getCell(row, colMap, "Phone"),
		Email:               getCell(row, colMap, "Email"),
		EmergencyContact:    getCell(row, colMap, "EmergencyContact"),
		EmergencyPhone:      getCell(row, colMap, "EmergencyPhone"),
		NumberOfDependents:  dependents,
		PresentAddress:      getCell(row, colMap, "Present_Address"),
		PermanentAddress:    getCell(row, colMap, "Permanent_Address"),
		PresentDivisionID:   lookupID(lookups.divisions, getCell(row, colMap, "Present_Division")),
		PresentDistrictID:   lookupID(lookups.districts, getCell(row, colMap, "Present_District")),
		PresentUpazilaID:    lookupID(lookups.upazilas, getCell(row, colMap, "Present_Upazila")),
		PresentUnionID:      lookupID(lookups.unions, getCell(row, colMap, "Present_Union")),
		PermanentDivisionID: lookupID(lookups.divisions, getCell(row, colMap, "Permanent_Division")),
		PermanentDistrictID: lookupID(lookups.districts, getCell(row, colMap, "Permanent_District")),
		PermanentUpazilaID:  lookupID(lookups.upazilas, getCell(row, colMap, "Permanent_Upazila")),
		PermanentUnionID:    lookupID(lookups.unions, getCell(row, colMap, "Permanent_Union")),
		PunchNumber:         getCell(row, colMap, "PunchNumber"),
		EmployeeType:        getCell(row, colMap, "EmployeeType"),
		Grade:               getCell(row, colMap, "Grade"),
		DepartmentID:        lookupID(lookups.departments, getCell(row, colMap, "Department")),
		SectionID:           sectionID,
		DesignationID:       desigID,
		LineID:              lineID,
		GroupID:             lookupID(lookups.groups, getCell(row, colMap, "Group")),
		FloorID:             lookupID(lookups.floors, getCell(row, colMap, "Floor")),
		ShiftID:             lookupID(lookups.shifts, getCell(row, colMap, "shiftName")),
		ReportsTo:           lookupID(lookups.employees, getCell(row, colMap, "ReportsTo")),
		JoiningDate:         parsedJoining,
		Status:              getCell(row, colMap, "Status"),
		OverTimeStatus:      overTimeBool,
		GrossSalary:         getCellFloat(row, colMap, "Gross_Salary"),
		BasicSalary:         getCellFloat(row, colMap, "Basic_Salary"),
		HouseRent:           getCellFloat(row, colMap, "House_Rent"),
		TransportAllowance:  getCellFloat(row, colMap, "Transport_Allowance"),
		FoodAllowance:       getCellFloat(row, colMap, "Food_Allowance"),
		MedicalAllowance:    getCellFloat(row, colMap, "Medical_Allowance"),
		OtherAllowance:      getCellFloat(row, colMap, "Other_Allowance"),
		AccountType:         getCell(row, colMap, "accountType"),
		AccountNumber:       getCell(row, colMap, "AccountNumber"),
		CreatedBy:           &userID,
	}
}
