package main

import (
	"fmt"
	"log"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	var employees []models.Employee
	db.Where("designation_id IS NOT NULL AND section_id IS NOT NULL AND deleted_at IS NULL").Find(&employees)

	type desigKey struct {
		SectionID string
		Name      string
	}
	desigBySectionAndName := make(map[desigKey]string)
	var allDesigs []models.Designation
	db.Find(&allDesigs)
	for _, d := range allDesigs {
		if d.SectionID != "" && d.Name != "" {
			key := desigKey{SectionID: d.SectionID, Name: d.Name}
			desigBySectionAndName[key] = d.ID
		}
	}

	// Map designation_id → name
	desigNameByID := make(map[string]string)
	for _, d := range allDesigs {
		desigNameByID[d.ID] = d.Name
	}

	fixed := 0
	skipped := 0

	for _, emp := range employees {
		if emp.DesignationID == nil || emp.SectionID == nil {
			continue
		}
		currentDesigID := *emp.DesignationID
		sectionID := *emp.SectionID

		// Find the designation for this section with the same name
		name, ok := desigNameByID[currentDesigID]
		if !ok || name == "" {
			skipped++
			continue
		}

		correctID, exists := desigBySectionAndName[desigKey{SectionID: sectionID, Name: name}]
		if !exists {
			skipped++
			continue
		}

		if correctID == currentDesigID {
			skipped++
			continue
		}

		// Update employee
		if err := db.Model(&emp).Update("designation_id", correctID).Error; err != nil {
			log.Printf("Failed to update employee %s: %v", emp.EmployeeID, err)
			continue
		}
		fmt.Printf("Fixed employee %s: designation %s -> %s (name: %s)\n", emp.EmployeeID, currentDesigID, correctID, name)
		fixed++
	}

	fmt.Printf("\nDone. Fixed: %d, Skipped: %d\n", fixed, skipped)
}
