package main

import (
	"fmt"
	"log"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	var company models.Company
	err := db.Where("status = ?", "active").First(&company).Error
	if err != nil {
		log.Fatal("No active company found. Run organization seed first.")
	}

	seedLeaveTypes(db, &company)

	fmt.Println("\n--- Leave type seeding complete ---")
}

func seedLeaveTypes(db *gorm.DB, company *models.Company) {
	type leaveDef struct {
		Name             string
		Code             string
		Days             int
		CarryForward     int
		ApplicableGender string
	}

	leaveTypes := []leaveDef{
		{"Annual Leave", "AL", 14, 5, "All"},
		{"Sick Leave", "SL", 10, 0, "All"},
		{"Casual Leave", "CL", 6, 0, "All"},
		{"Maternity Leave", "ML", 120, 0, "Female"},
		{"Paternity Leave", "PL", 5, 0, "Male"},
		{"Emergency Leave", "EL", 3, 0, "All"},
		{"Study Leave", "STL", 30, 0, "All"},
		{"Hajj Leave", "HL", 40, 0, "All"},
	}

	for _, lt := range leaveTypes {
		var existing models.LeaveType
		err := db.Where("company_id = ? AND code = ?", company.ID, lt.Code).First(&existing).Error
		if err == nil {
			fmt.Printf("  Leave type %s (%s) already exists, skipping\n", lt.Name, lt.Code)
			continue
		}

		leaveType := models.LeaveType{
			CompanyID:        company.ID,
			Name:             lt.Name,
			Code:             lt.Code,
			TotalDays:        lt.Days,
			CarryForwardDays: lt.CarryForward,
			ApplicableGender: lt.ApplicableGender,
			Status:           "active",
		}
		if err := db.Create(&leaveType).Error; err != nil {
			log.Printf("Failed to create leave type %s: %v", lt.Name, err)
		} else {
			fmt.Printf("  Leave Type: %s (%s) - %d days\n", lt.Name, lt.Code, lt.Days)
		}
	}
}
