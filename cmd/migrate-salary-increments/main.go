package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type HrSalaryRow struct {
	EmpID       string  `gorm:"column:emp_id"`
	MMonth      int     `gorm:"column:mmonth"`
	YYear       int     `gorm:"column:yyear"`
	GrossSalary float64 `gorm:"column:gross_salary"`
	Basic       float64 `gorm:"column:basic"`
	Hrent       float64 `gorm:"column:hrent"`
	Medical     float64 `gorm:"column:medical"`
}

func normalizeEmpID(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	return s
}

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	// Connect to hrhub source DB - use same host/port but different DB and credentials
	hrDsn := fmt.Sprintf("host=%s port=%s user=postgres password=123580 dbname=hrhub sslmode=disable",
		cfg.DBHost, cfg.DBPort)
	hrDB, err := gorm.Open(postgres.Open(hrDsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect to hrhub: %v", err)
	}
	fmt.Println("Connected to hrhub and peoplehub")

	// Get company_id from peoplehub
	var company models.Company
	if err := db.Where("deleted_at IS NULL").First(&company).Error; err != nil {
		log.Fatalf("no company found in peoplehub: %v", err)
	}
	fmt.Printf("Using company: %s (%s)\n", company.CompanyNameEn, company.ID)

	// Build employee map for validation and company resolution
	var emps []models.Employee
	if err := db.Where("deleted_at IS NULL").Find(&emps).Error; err != nil {
		log.Fatalf("failed to load employees: %v", err)
	}
	empMap := make(map[string]models.Employee, len(emps))
	for _, e := range emps {
		empMap[e.EmployeeID] = e
		// also map normalized
		n := normalizeEmpID(e.EmployeeID)
		empMap[n] = e
	}
	fmt.Printf("PeopleHub employees loaded: %d\n", len(empMap))

	// Check existing increments to avoid duplicates
	var existingCount int64
	db.Model(&models.SalaryIncrement{}).Where("deleted_at IS NULL").Count(&existingCount)
	fmt.Printf("Existing salary_increments: %d\n", existingCount)

	// Load hrhub salary history from empSalaryCom (preferred, has full breakdown)
	var rows []HrSalaryRow
	// Use raw SQL to get ordered history; filter out zero gross if needed but keep all for delta calc
	if err := hrDB.Raw(`
		SELECT emp_id, mmonth, yyear, gross_salary, basic, hrent, medical
		FROM "empSalaryCom"
		WHERE gross_salary IS NOT NULL AND gross_salary > 0
		ORDER BY emp_id, yyear, mmonth
	`).Scan(&rows).Error; err != nil {
		log.Fatalf("failed to load empSalaryCom: %v", err)
	}
	fmt.Printf("Loaded hrhub empSalaryCom rows: %d\n", len(rows))

	// Also try empSalaryUnique as fallback if Com is empty
	if len(rows) == 0 {
		if err := hrDB.Raw(`SELECT emp_id, mmonth, yyear, gross_salary, 0 as basic, 0 as hrent, 0 as medical FROM "empSalaryUnique" ORDER BY emp_id, yyear, mmonth`).Scan(&rows).Error; err != nil {
			log.Fatalf("failed to load empSalaryUnique: %v", err)
		}
		fmt.Printf("Fallback loaded empSalaryUnique rows: %d\n", len(rows))
	}

	// Group by normalized emp_id
	type hist struct {
		rows []HrSalaryRow
	}
	grouped := make(map[string][]HrSalaryRow)
	for _, r := range rows {
		nid := normalizeEmpID(r.EmpID)
		grouped[nid] = append(grouped[nid], r)
	}

	var toInsert []models.SalaryIncrement
	skippedNoEmp := 0
	skippedNoChange := 0
	skippedDuplicate := 0

	// Cache existing effective dates per employee to avoid duplicate inserts
	existingMap := make(map[string]bool)
	var existing []models.SalaryIncrement
	db.Where("deleted_at IS NULL").Find(&existing)
	for _, e := range existing {
		key := e.EmployeeID + "|" + e.EffectiveDate
		existingMap[key] = true
	}

	// Get superadmin user for created_by
	var superadmin models.User
	db.Where("email = ?", "superadmin@peoplehub.com").First(&superadmin)
	createdBy := ""
	if company.CreatedBy != nil {
		createdBy = *company.CreatedBy
	}
	if superadmin.ID != "" {
		createdBy = superadmin.ID
	}

	now := time.Now()

	for empNormID, histRows := range grouped {
		emp, ok := empMap[empNormID]
		if !ok {
			// Try original ID without normalize
			// Already normalized, so skip
			skippedNoEmp++
			continue
		}
		// Deduplicate same yyear+mmonth keeping last (should not happen for Com but safe)
		seen := make(map[string]HrSalaryRow)
		orderedKeys := []string{}
		for _, r := range histRows {
			k := fmt.Sprintf("%04d-%02d", r.YYear, r.MMonth)
			if _, exists := seen[k]; !exists {
				orderedKeys = append(orderedKeys, k)
			}
			seen[k] = r
		}
		// Rebuild sorted unique rows
		var uniq []HrSalaryRow
		for _, k := range orderedKeys {
			uniq = append(uniq, seen[k])
		}
		// Detect increments: compare each month to previous
		var prev *HrSalaryRow
		for _, cur := range uniq {
			if prev == nil {
				prev = &cur
				continue
			}
			// Only create increment when gross changes
			if cur.GrossSalary == prev.GrossSalary {
				prev = &cur
				skippedNoChange++
				continue
			}
			effectiveDate := fmt.Sprintf("%04d-%02d-01", cur.YYear, cur.MMonth)
			key := emp.EmployeeID + "|" + effectiveDate
			if existingMap[key] {
				skippedDuplicate++
				prev = &cur
				continue
			}
			// Validate month/year
			if cur.MMonth < 1 || cur.MMonth > 12 || cur.YYear < 2000 || cur.YYear > 2030 {
				prev = &cur
				continue
			}
			inc := models.SalaryIncrement{
				CompanyID:       emp.CompanyID,
				EmployeeID:      emp.EmployeeID,
				IncrementType:   "increment",
				CalculationType: "fixed",
				CalculationValue: cur.GrossSalary - prev.GrossSalary,
				PreviousGross:   prev.GrossSalary,
				PreviousBasic:   prev.Basic,
				PreviousHouse:   prev.Hrent,
				PreviousMedical: prev.Medical,
				IncrementAmount: cur.GrossSalary - prev.GrossSalary,
				NewGross:        cur.GrossSalary,
				NewBasic:        cur.Basic,
				NewHouse:        cur.Hrent,
				NewMedical:      cur.Medical,
				IncrementDate:   effectiveDate,
				EffectiveDate:   effectiveDate,
				Status:          "approved",
				Remarks:         fmt.Sprintf("Auto-migrated from hrhub empSalaryCom %04d-%02d (prev: %.0f → new: %.0f)", cur.YYear, cur.MMonth, prev.GrossSalary, cur.GrossSalary),
				CreatedAt:       now,
				UpdatedAt:       now,
				CreatedBy:       createdBy,
			}
			toInsert = append(toInsert, inc)
			existingMap[key] = true
			prev = &cur
		}
	}

	fmt.Printf("Prepared increments: %d (skipped noEmp:%d, noChange:%d, duplicate:%d)\n", len(toInsert), skippedNoEmp, skippedNoChange, skippedDuplicate)

	if len(toInsert) == 0 {
		fmt.Println("Nothing to migrate")
		os.Exit(0)
	}

	// Batch insert
	batchSize := 100
	created := 0
	for i := 0; i < len(toInsert); i += batchSize {
		end := i + batchSize
		if end > len(toInsert) {
			end = len(toInsert)
		}
		batch := toInsert[i:end]
		if err := db.CreateInBatches(batch, 100).Error; err != nil {
			log.Fatalf("batch insert failed at %d: %v", i, err)
		}
		created += len(batch)
		fmt.Printf("Inserted batch %d-%d\n", i, end)
	}

	fmt.Printf("Migration completed: %d salary_increments inserted\n", created)

	// Verify
	var finalCount int64
	db.Model(&models.SalaryIncrement{}).Where("deleted_at IS NULL").Count(&finalCount)
	fmt.Printf("Final salary_increments count: %d\n", finalCount)
}
