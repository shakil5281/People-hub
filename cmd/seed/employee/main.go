package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)

	// Fix employee_id column type (AutoMigrate may create it as uuid)
	database.DB.Exec("ALTER TABLE employees ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)")

	companyID := ""
	if err := database.DB.Model(&models.Company{}).Where("slug = 'peoplehub-technologies'").Select("id").First(&companyID).Error; err != nil {
		log.Fatal("Company not found:", err)
	}

	employees := []struct {
		EmployeeID  string
		PunchNumber string
		Name        string
	}{
		{EmployeeID: "38", PunchNumber: "38", Name: "Employee 38"},
		{EmployeeID: "11", PunchNumber: "11", Name: "Employee 11"},
		{EmployeeID: "1733", PunchNumber: "1733", Name: "Employee 1733"},
		{EmployeeID: "39", PunchNumber: "39", Name: "Employee 39"},
		{EmployeeID: "8", PunchNumber: "8", Name: "Employee 8"},
		{EmployeeID: "10", PunchNumber: "10", Name: "Employee 10"},
	}

	for _, e := range employees {
		var count int64
		database.DB.Raw("SELECT count(*) FROM employees WHERE employee_id = ? AND company_id = ?", e.EmployeeID, companyID).Scan(&count)
		if count > 0 {
			fmt.Printf("  Employee %s already exists, skipping\n", e.EmployeeID)
			continue
		}

		emp := models.Employee{
			CompanyID:    companyID,
		EmployeeID: e.EmployeeID,
		PunchNumber:  e.PunchNumber,
			NameEn:       e.Name,
			JoiningDate:  time.Now(),
			Status:       "active",
		}
		if err := database.DB.Create(&emp).Error; err != nil {
			fmt.Printf("  Error creating employee %s: %v\n", e.EmployeeID, err)
		} else {
			fmt.Printf("  Created employee: %s (code: %s, punch: %s)\n", emp.NameEn, emp.EmployeeID, emp.PunchNumber)
		}
	}

	fmt.Println("\n--- Employee seeding complete ---")
}
