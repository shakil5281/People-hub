package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	// Build same wiring as server/server.go for salary
	employeeRepo := repository.NewEmployeeRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	salaryRepo := repository.NewSalaryRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	otEarlyExitRepo := repository.NewOtEarlyExitRepository(db)
	holidayRepo := repository.NewHolidayRepository(db)
	otEarlyExitService := service.NewOtEarlyExitService(otEarlyExitRepo, holidayRepo)
	advanceRepo := repository.NewAdvanceSalaryRepository(db)
	separationRepo := repository.NewSeparationRepository(db)
	salaryIncrementRepo := repository.NewSalaryIncrementRepository(db)

	salaryService := service.NewSalaryService(employeeRepo, attendanceRepo, salaryRepo, groupRepo, otEarlyExitRepo, otEarlyExitService, advanceRepo, separationRepo)
	salaryService.SetIncrementRepo(salaryIncrementRepo)

	// Find company
	var companyID string
	db.Raw("SELECT id FROM companies WHERE deleted_at IS NULL LIMIT 1").Scan(&companyID)
	if companyID == "" {
		log.Fatal("no company found")
	}
	fmt.Printf("CompanyID: %s\n", companyID)

	// Superadmin user for audit
	var userID string
	db.Raw("SELECT id FROM users WHERE email='superadmin@peoplehub.com' LIMIT 1").Scan(&userID)
	if userID == "" {
		db.Raw("SELECT id FROM users LIMIT 1").Scan(&userID)
	}
	fmt.Printf("UserID: %s\n", userID)

	startYear, endYear := 2020, 2026
	totalMonths := 0
	totalProcessed := 0
	start := time.Now()

	for year := startYear; year <= endYear; year++ {
		for month := 1; month <= 12; month++ {
			// Skip future months beyond current
			now := time.Now()
			if year > now.Year() || (year == now.Year() && month > int(now.Month())) {
				fmt.Printf("Skip future %04d-%02d\n", year, month)
				continue
			}
			totalMonths++
			fmt.Printf("\n[%d/%d] Processing %04d-%02d ... ", totalMonths, (endYear-startYear+1)*12, year, month)
			mStart := time.Now()
			result, err := salaryService.ProcessMonth(companyID, month, year, userID, true)
			elapsed := time.Since(mStart).Truncate(time.Millisecond)
			if err != nil {
				fmt.Printf("FAILED (%v) elapsed %v\n", err, elapsed)
				log.Printf("failed %04d-%02d: %v", year, month, err)
				continue
			}
			fmt.Printf("OK processed %d/%d (total employees %d) elapsed %v\n", result.Processed, result.Total, result.Total, elapsed)
			totalProcessed += result.Processed
		}
	}

	fmt.Printf("\n=== BULK SALARY DONE ===\n")
	fmt.Printf("Months attempted: %d\n", totalMonths)
	fmt.Printf("Total salary rows processed (sum): %d\n", totalProcessed)
	fmt.Printf("Total time: %v\n", time.Since(start).Truncate(time.Millisecond))

	// Summary per year
	type row struct {
		Year  int
		Month int
		Cnt   int64
	}
	var rows []row
	db.Raw("SELECT year, month, count(*) as cnt FROM salaries WHERE deleted_at IS NULL AND company_id = ? GROUP BY year, month ORDER BY year, month", companyID).Scan(&rows)
	fmt.Printf("\nSalaries per month in DB:\n")
	for _, r := range rows {
		fmt.Printf("  %04d-%02d: %d\n", r.Year, r.Month, r.Cnt)
	}
}
