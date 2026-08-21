package main

import (
	"fmt"
	"log"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	type AbsentResult struct {
		EmployeeID string
		Count      int64
	}

	var results []AbsentResult

	err := db.Raw(`
		SELECT a.employee_id, COUNT(*) as count
		FROM attendances a
		JOIN employees e ON e.employee_id = a.employee_id
		WHERE a.status = ? AND a.deleted_at IS NULL
		  AND e.deleted_at IS NULL
		  AND e.status = ?
		  AND LOWER(e.employee_type) = ?
		GROUP BY a.employee_id
		HAVING COUNT(*) = 26
		ORDER BY a.employee_id
	`, "absent", "active", "regular").Scan(&results).Error
	if err != nil {
		log.Fatal("Query failed:", err)
	}

	if len(results) == 0 {
		fmt.Println("No employees found with exactly 26 absent days.")
		return
	}

	fmt.Printf("Found %d employees with exactly 26 absent days:\n\n", len(results))

	sepRepo := repository.NewSeparationRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	sepService := service.NewSeparationService(db, sepRepo, employeeRepo, attendanceRepo)

	created := 0
	skipped := 0
	errCount := 0

	for _, r := range results {
		var emp models.Employee
		if err := db.Where("employee_id = ?", r.EmployeeID).First(&emp).Error; err != nil {
			fmt.Printf("  SKIP  employee_id=%s — not found\n", r.EmployeeID)
			skipped++
			continue
		}

		fmt.Printf("  Employee: %s (ID=%s, Type=%s, Status=%s) — Absent count: %d\n",
			emp.NameEn, emp.EmployeeID, emp.EmployeeType, emp.Status, r.Count)

		exists, _ := sepRepo.ExistsPendingOrApproved(emp.EmployeeID)
		if exists {
			fmt.Printf("    SKIP — already has pending/approved separation\n")
			skipped++
			continue
		}

		var existingProc models.Separation
		if err := db.Where("employee_id = ? AND status = ? AND deleted_at IS NULL", emp.EmployeeID, "Processed").
			First(&existingProc).Error; err == nil {
			fmt.Printf("    SKIP — already processed (date=%s)\n", existingProc.Date)
			skipped++
			continue
		}

		input := service.CreateSeparationInput{
			EmployeeID: emp.EmployeeID,
			SepType:    "Close",
			Date:       "2026-07-01",
			Reason:     fmt.Sprintf("Auto: %d absent days", r.Count),
		}

		_, result, createErr := sepService.Create(input)
		if createErr != nil {
			fmt.Printf("    ERROR — %v\n", createErr)
			errCount++
			continue
		}

		fmt.Printf("    CREATED — processed=%v, new_type=%s, new_status=%s, attn_deleted=%d\n",
			result.NewType != "", result.NewType, result.NewStatus, result.AttendanceDeleted)
		created++
	}

	fmt.Printf("\nSummary: %d created, %d skipped, %d errors\n", created, skipped, errCount)
}
