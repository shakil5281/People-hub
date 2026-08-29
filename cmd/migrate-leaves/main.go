package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type LeaveTrans struct {
	TrnID     int64      `gorm:"column:trn_id"`
	EmpID     string     `gorm:"column:emp_id"`
	LID       int        `gorm:"column:l_id"`
	StartDate *time.Time `gorm:"column:start_date"`
	EndDate   *time.Time `gorm:"column:end_date"`
	NoOfDay   *float64   `gorm:"column:no_of_day"`
	Reason    string     `gorm:"column:reason"`
	Tdate     *time.Time `gorm:"column:tdate"`
}

func (LeaveTrans) TableName() string { return "LeaveTrans" }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	godotenv.Load()
	godotenv.Load(".env.local")

	dryRun := flag.Bool("dry-run", false, "dry run")
	limit := flag.Int("limit", 0, "limit rows for testing (0 = all)")
	flag.Parse()

	companyID := "b0b60d1f-1bd8-4803-98ab-ecd40d8162f5"

	pgHost := getEnv("DB_HOST", "localhost")
	pgPort := getEnv("DB_PORT", "5432")
	pgUser := getEnv("DB_USER", "postgres")
	pgPass := getEnv("DB_PASS", "123580")
	pgDB := getEnv("DB_NAME", "peoplehub")
	pgSSL := getEnv("DB_SSLMODE", "disable")
	if getEnv("FORCE_PG_USER", "") != "" {
		pgUser = getEnv("FORCE_PG_USER", pgUser)
		pgPass = getEnv("FORCE_PG_PASS", pgPass)
	}
	hrhubDBName := getEnv("HRHUB_DB_NAME", "hrhub")
	hrhubUser := getEnv("HRHUB_DB_USER", "postgres")
	hrhubPass := getEnv("HRHUB_DB_PASS", "123580")

	peopleDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", pgHost, pgPort, pgUser, pgPass, pgDB, pgSSL)
	hrhubDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", pgHost, pgPort, hrhubUser, hrhubPass, hrhubDBName, pgSSL)

	peopleDB, err := gorm.Open(postgres.Open(peopleDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("peoplehub connect failed: %v", err)
	}
	hrhubDB, err := gorm.Open(postgres.Open(hrhubDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("hrhub connect failed: %v", err)
	}

	// Load leave_types mapping
	type LeaveType struct {
		ID   string `gorm:"column:id"`
		Code string `gorm:"column:code"`
		Name string `gorm:"column:name"`
	}
	var leaveTypes []LeaveType
	peopleDB.Raw("SELECT id, code, name FROM leave_types WHERE deleted_at IS NULL").Scan(&leaveTypes)
	codeToID := make(map[string]string)
	for _, lt := range leaveTypes {
		codeToID[strings.ToUpper(lt.Code)] = lt.ID
	}
	fmt.Printf("LeaveTypes: %d\n", len(leaveTypes))
	for _, lt := range leaveTypes {
		fmt.Printf("  %s -> %s\n", lt.Code, lt.ID[:8])
	}

	// Map hrhub l_id -> peoplehub leave_type_id
	// Based on distribution: l_id 1 (2210) -> CL, l_id 2 (110) -> SL, l_id 3 (10) -> AL, others -> CL fallback
	// User can adjust this map if needed
	lidMap := map[int]string{
		1: codeToID["CL"], // Casual (most common)
		2: codeToID["SL"], // Sick (second)
		3: codeToID["AL"], // Annual
		5: codeToID["ML"], // Maternity guess
		6: codeToID["PL"], // Paternity guess
		7: codeToID["EL"], // Emergency
		8: codeToID["SL"], // fallback Sick
	}
	fallbackCL := codeToID["CL"]
	for lid, id := range lidMap {
		if id == "" {
			lidMap[lid] = fallbackCL
		}
	}
	fmt.Printf("l_id mapping: %+v (fallback CL=%s)\n", lidMap, fallbackCL[:8])

	// Load existing employees for FK check (3144)
	var empIDs []string
	peopleDB.Raw("SELECT employee_id FROM employees WHERE deleted_at IS NULL").Scan(&empIDs)
	empMap := make(map[string]bool, len(empIDs))
	for _, id := range empIDs {
		empMap[strings.TrimSpace(id)] = true
	}
	fmt.Printf("Employees existing: %d\n", len(empMap))

	// Load existing leaves deduplication key: employee_id|leave_type_id|from_date|to_date
	var existingLeaves []struct {
		EmployeeID  string `gorm:"column:employee_id"`
		LeaveTypeID string `gorm:"column:leave_type_id"`
		FromDate    string `gorm:"column:from_date"`
		ToDate      string `gorm:"column:to_date"`
	}
	peopleDB.Raw("SELECT employee_id, leave_type_id, from_date::text as from_date, to_date::text as to_date FROM leaves WHERE deleted_at IS NULL").Scan(&existingLeaves)
	existingMap := make(map[string]bool, len(existingLeaves))
	for _, l := range existingLeaves {
		key := strings.TrimSpace(l.EmployeeID) + "|" + l.LeaveTypeID + "|" + l.FromDate + "|" + l.ToDate
		existingMap[key] = true
	}
	fmt.Printf("Existing leaves: %d\n", len(existingMap))

	// Load LeaveTrans
	var leaveTrans []LeaveTrans
	q := hrhubDB.Table(`"LeaveTrans"`).Order("start_date ASC")
	if *limit > 0 {
		q = q.Limit(*limit)
	}
	if err := q.Find(&leaveTrans).Error; err != nil {
		log.Fatalf("load LeaveTrans failed: %v", err)
	}
	fmt.Printf("LeaveTrans total: %d (limit %d)\n", len(leaveTrans), *limit)

	inserted := 0
	skippedExists := 0
	skippedInvalid := 0
	skippedEmpNotFound := 0
	batchCount := 0

	// Use transaction batch 500
	tx := peopleDB.Begin()
	if tx.Error != nil {
		log.Fatalf("begin tx failed: %v", tx.Error)
	}

	for _, lt := range leaveTrans {
		empID := strings.TrimSpace(lt.EmpID)
		if empID == "" {
			skippedInvalid++
			continue
		}
		if !empMap[empID] {
			skippedEmpNotFound++
			continue
		}
		if lt.StartDate == nil || lt.EndDate == nil {
			skippedInvalid++
			continue
		}
		// Normalize dates to date only
		fromDate := lt.StartDate.Format("2006-01-02")
		toDate := lt.EndDate.Format("2006-01-02")
		// Validate date order
		if lt.EndDate.Before(*lt.StartDate) {
			// swap
			fromDate, toDate = toDate, fromDate
		}
		// l_id -> leave_type_id
		leaveTypeID, ok := lidMap[lt.LID]
		if !ok || leaveTypeID == "" {
			leaveTypeID = fallbackCL
		}
		if leaveTypeID == "" {
			skippedInvalid++
			continue
		}
		// total_days
		totalDays := 0
		if lt.NoOfDay != nil && *lt.NoOfDay > 0 {
			totalDays = int(*lt.NoOfDay)
		} else {
			// compute from dates inclusive
			days := int(lt.EndDate.Sub(*lt.StartDate).Hours()/24) + 1
			if days < 1 {
				days = 1
			}
			totalDays = days
		}
		if totalDays <= 0 {
			totalDays = 1
		}
		key := empID + "|" + leaveTypeID + "|" + fromDate + "|" + toDate
		if existingMap[key] {
			skippedExists++
			continue
		}
		reason := strings.TrimSpace(lt.Reason)
		if reason == "" {
			reason = "Migrated from hrhub"
		} else if reason == "1" {
			reason = "Migrated - reason 1"
		}

		if *dryRun {
			if batchCount < 3 {
				fmt.Printf("[DRY-RUN] leave emp=%s l_id=%d -> %s from=%s to=%s days=%d reason=%s\n", empID, lt.LID, leaveTypeID[:8], fromDate, toDate, totalDays, reason)
			}
			batchCount++
			inserted++
			existingMap[key] = true
			continue
		}

		// Insert with ON CONFLICT DO NOTHING (unique not enforced but safe)
		// leaves table has no unique constraint on these 4, but we dedupe via map + check
		// Check again live to avoid race
		var cnt int64
		tx.Raw("SELECT count(*) FROM leaves WHERE employee_id = ? AND leave_type_id = ? AND from_date = ?::date AND to_date = ?::date AND deleted_at IS NULL", empID, leaveTypeID, fromDate, toDate).Scan(&cnt)
		if cnt > 0 {
			skippedExists++
			existingMap[key] = true
			continue
		}

		err := tx.Exec(`
			INSERT INTO leaves (id, company_id, employee_id, leave_type_id, from_date, to_date, total_days, reason, status, created_at, updated_at)
			VALUES (gen_random_uuid(), ?, ?, ?, ?::date, ?::date, ?, ?, 'approved', now(), now())
		`, companyID, empID, leaveTypeID, fromDate, toDate, totalDays, reason).Error
		if err != nil {
			tx.Rollback()
			log.Fatalf("insert leave failed emp=%s: %v", empID, err)
		}
		existingMap[key] = true
		inserted++
		batchCount++

		if batchCount%500 == 0 {
			if err := tx.Commit().Error; err != nil {
				log.Fatalf("commit batch failed: %v", err)
			}
			tx = peopleDB.Begin()
		}
	}

	if !*dryRun {
		if err := tx.Commit().Error; err != nil {
			log.Fatalf("final commit failed: %v", err)
		}
	}
	fmt.Printf("Leaves migration done: inserted=%d skipped_exists=%d skipped_invalid=%d skipped_emp_not_found=%d\n", inserted, skippedExists, skippedInvalid, skippedEmpNotFound)
	if *dryRun {
		fmt.Println("DRY-RUN completed - no data written")
	} else {
		fmt.Println("Migration completed - data written")
	}
}
