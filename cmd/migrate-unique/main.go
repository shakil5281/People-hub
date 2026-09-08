package main

import (
	"fmt"
	"log"
	"sort"
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

	// Source: migration_manager DB
	srcDsn := fmt.Sprintf("host=%s port=%s user=postgres password=123580 dbname=migration_manager sslmode=disable", cfg.DBHost, cfg.DBPort)
	srcDB, err := gorm.Open(postgres.Open(srcDsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("failed to connect to migration_manager: %v", err)
	}
	fmt.Println("Connected to migration_manager and peoplehub")

	var company models.Company
	if err := db.Where("deleted_at IS NULL").First(&company).Error; err != nil {
		log.Fatalf("no company found: %v", err)
	}
	fmt.Printf("Using company: %s (%s)\n", company.CompanyNameEn, company.ID)

	var emps []models.Employee
	if err := db.Where("deleted_at IS NULL").Find(&emps).Error; err != nil {
		log.Fatalf("failed to load employees: %v", err)
	}
	empMap := make(map[string]models.Employee, len(emps)*2)
	for _, e := range emps {
		empMap[e.EmployeeID] = e
		n := normalizeEmpID(e.EmployeeID)
		if _, ok := empMap[n]; !ok {
			empMap[n] = e
		}
	}
	fmt.Printf("PeopleHub employees loaded: %d\n", len(emps))

	var existingCount int64
	db.Model(&models.SalaryIncrement{}).Where("deleted_at IS NULL").Count(&existingCount)
	fmt.Printf("Existing salary_increments (active): %d\n", existingCount)

	var rows []HrSalaryRow
	// Load from empSalaryUnique, exclude august 2026, exclude zero gross, order for grouping
	if err := srcDB.Raw(`
		SELECT "emp_id", "mmonth", "yyear", "gross_salary"
		FROM "empSalaryUnique"
		WHERE "gross_salary" IS NOT NULL AND "gross_salary" > 0
		  AND NOT ("mmonth"=8 AND "yyear"=2026)
		ORDER BY "emp_id", "yyear", "mmonth"
	`).Scan(&rows).Error; err != nil {
		log.Fatalf("failed to load empSalaryUnique: %v", err)
	}
	fmt.Printf("Loaded empSalaryUnique rows (filtered): %d\n", len(rows))

	// Count before dedup
	grouped := make(map[string][]HrSalaryRow)
	for _, r := range rows {
		nid := normalizeEmpID(r.EmpID)
		grouped[nid] = append(grouped[nid], r)
	}
	fmt.Printf("Grouped by normalized emp_id: %d groups\n", len(grouped))

	// Cache existing effective dates - use raw ::text to ensure YYYY-MM-DD format matches generated key
	existingMap := make(map[string]bool)
	type existingRow struct {
		EmployeeID    string `gorm:"column:employee_id"`
		EffectiveDate string `gorm:"column:effective_date"`
	}
	var existingRows []existingRow
	db.Raw(`SELECT employee_id, effective_date::text as effective_date FROM salary_increments WHERE deleted_at IS NULL`).Scan(&existingRows)
	for _, e := range existingRows {
		key := e.EmployeeID + "|" + e.EffectiveDate
		existingMap[key] = true
	}
	// Also cache for dedup within this run
	toInsertMap := make(map[string]bool)

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

	var toInsert []models.SalaryIncrement
	skippedNoEmp := 0
	skippedNoChange := 0
	skippedDuplicate := 0
	skippedInvalid := 0
	skippedAug2026 := 0 // already filtered in SQL, but double-check

	// For deterministic output, sort emp keys
	keys := make([]string, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, empNormID := range keys {
		histRows := grouped[empNormID]
		emp, ok := empMap[empNormID]
		if !ok {
			skippedNoEmp++
			continue
		}
		// Deduplicate same yyear+mmonth keeping last
		seen := make(map[string]HrSalaryRow)
		orderedKeys := []string{}
		for _, r := range histRows {
			// Double-check aug 2026 filter
			if r.MMonth == 8 && r.YYear == 2026 {
				skippedAug2026++
				continue
			}
			if r.MMonth < 1 || r.MMonth > 12 || r.YYear < 2000 || r.YYear > 2030 {
				skippedInvalid++
				continue
			}
			k := fmt.Sprintf("%04d-%02d", r.YYear, r.MMonth)
			if _, exists := seen[k]; !exists {
				orderedKeys = append(orderedKeys, k)
			}
			seen[k] = r
		}
		sort.Strings(orderedKeys)
		var uniq []HrSalaryRow
		for _, k := range orderedKeys {
			uniq = append(uniq, seen[k])
		}
		var prev *HrSalaryRow
		for _, cur := range uniq {
			if prev == nil {
				prev = &cur
				continue
			}
			if cur.GrossSalary == prev.GrossSalary {
				skippedNoChange++
				prev = &cur
				continue
			}
			effectiveDate := fmt.Sprintf("%04d-%02d-01", cur.YYear, cur.MMonth)
			key := emp.EmployeeID + "|" + effectiveDate
			if existingMap[key] || toInsertMap[key] {
				skippedDuplicate++
				prev = &cur
				continue
			}
			diff := cur.GrossSalary - prev.GrossSalary
			// diff can be negative (decrement) — keep it, as user said salary distance
			inc := models.SalaryIncrement{
				CompanyID:       emp.CompanyID,
				EmployeeID:      emp.EmployeeID,
				IncrementType:   "increment",
				CalculationType: "fixed",
				CalculationValue: diff,
				PreviousGross:   prev.GrossSalary,
				PreviousBasic:   0,
				PreviousHouse:   0,
				PreviousMedical: 0,
				IncrementAmount: diff,
				NewGross:        cur.GrossSalary,
				NewBasic:        0,
				NewHouse:        0,
				NewMedical:      0,
				IncrementDate:   effectiveDate,
				EffectiveDate:   effectiveDate,
				Status:          "approved",
				Remarks:         fmt.Sprintf("Auto-migrated from migration_manager.empSalaryUnique %04d-%02d (prev: %.0f → new: %.0f)", cur.YYear, cur.MMonth, prev.GrossSalary, cur.GrossSalary),
				CreatedAt:       now,
				UpdatedAt:       now,
				CreatedBy:       createdBy,
			}
			toInsert = append(toInsert, inc)
			toInsertMap[key] = true
			prev = &cur
		}
	}

	fmt.Printf("Prepared increments: %d (skipped noEmp:%d, noChange:%d, duplicate:%d, invalid:%d, aug2026:%d)\n", len(toInsert), skippedNoEmp, skippedNoChange, skippedDuplicate, skippedInvalid, skippedAug2026)

	if len(toInsert) == 0 {
		fmt.Println("Nothing to migrate")
		return
	}

	// Show sample
	for i := 0; i < len(toInsert) && i < 5; i++ {
		t := toInsert[i]
		fmt.Printf(" Sample %d: emp=%s eff=%s prev=%.0f new=%.0f diff=%.0f\n", i+1, t.EmployeeID, t.EffectiveDate, t.PreviousGross, t.NewGross, t.IncrementAmount)
	}

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
	var finalCount int64
	db.Model(&models.SalaryIncrement{}).Where("deleted_at IS NULL").Count(&finalCount)
	fmt.Printf("Final salary_increments count: %d\n", finalCount)
}
