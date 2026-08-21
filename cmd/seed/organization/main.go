package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	company := seedCompany(db)
	seedDepartments(db, company)
	seedGroups(db)
	seedFloors(db)
	seedShifts(db, company)

	fmt.Println("\n--- Organization seeding complete ---")
}

func seedCompany(db *gorm.DB) *models.Company {
	slug := "peoplehub-technologies"
	var existing models.Company
	err := db.Where("slug = ?", slug).First(&existing).Error
	if err == nil {
		fmt.Println("Company already exists, skipping")
		return &existing
	}

	settingsJSON, _ := json.Marshal(map[string]any{
		"timezone":    "Asia/Dhaka",
		"date_format": "d/m/Y",
		"currency":    "BDT",
		"language":    "en",
	})

	company := models.Company{
		CompanyNameBn: "পিপলহাব টেকনোলজিস লিমিটেড",
		CompanyNameEn: "PeopleHub Technologies Ltd.",
		Slug:          slug,
		AddressEn:     "House 12, Road 5, Sector 3, Uttara, Dhaka 1230",
		Phone:         "+880-2-9876543",
		Status:        "active",
		Settings:      datatypes.JSON(settingsJSON),
	}
	if err := db.Create(&company).Error; err != nil {
		log.Fatal("Failed to create company:", err)
	}
	fmt.Printf("Created company: %s\n", company.CompanyNameEn)
	return &company
}

func seedDepartments(db *gorm.DB, company *models.Company) {
	type deptDef struct {
		Name   string
		NameBn string
	}

	depts := []struct {
		Name   string
		NameBn string
	}{
		{"Human Resources", "মানব সম্পদ"},
		{"Administration", "প্রশাসন"},
		{"Information Technology", "তথ্য প্রযুক্তি"},
		{"Finance & Accounts", "অর্থ ও হিসাব"},
		{"Marketing", "বিপণন"},
		{"Sales", "বিক্রয়"},
		{"Production", "উৎপাদন"},
		{"Quality Control", "গুণগত মান নিয়ন্ত্রণ"},
		{"Store & Inventory", "স্টোর ও ইনভেন্টরি"},
		{"Maintenance", "রক্ষণাবেক্ষণ"},
		{"Production 2", "উৎপাদন ২"},
		{"Quality Control 2", "গুণগত মান নিয়ন্ত্রণ ২"},
	}

	for _, d := range depts {
		var existing models.Department
		err := db.Where("company_id = ? AND name = ?", company.ID, d.Name).First(&existing).Error
		if err == nil {
			continue
		}
		dept := models.Department{
			CompanyID: &company.ID,
			Name:      d.Name,
			NameBn:    d.NameBn,
			Status:    "active",
		}
		if err := db.Create(&dept).Error; err != nil {
			log.Printf("Failed to create department %s: %v", d.Name, err)
		} else {
			fmt.Printf("    Department: %s\n", d.Name)
		}
	}

	seedSections(db, company)
}

func seedSections(db *gorm.DB, company *models.Company) {
	type secDef struct {
		Dept   string
		Name   string
		NameBn string
	}

	var depts []models.Department
	db.Where("company_id = ?", company.ID).Find(&depts)
	deptMap := map[string]string{}
	for _, d := range depts {
		deptMap[d.Name] = d.ID
	}

	sections := []secDef{
		{"Human Resources", "Recruitment", "নিয়োগ"},
		{"Human Resources", "Training", "প্রশিক্ষণ"},
		{"Human Resources", "Payroll", "বেতন"},
		{"Information Technology", "Software Development", "সফটওয়্যার ডেভেলপমেন্ট"},
		{"Information Technology", "Network & Infrastructure", "নেটওয়ার্ক ও অবকাঠামো"},
		{"Information Technology", "IT Support", "আইটি সাপোর্ট"},
		{"Finance & Accounts", "Accounts", "হিসাব"},
		{"Finance & Accounts", "Audit", "অডিট"},
		{"Production", "Cutting", "কাটিং"},
		{"Production", "Sewing", "সেলাই"},
		{"Production", "Finishing", "ফিনিশিং"},
		{"Production", "Packing", "প্যাকিং"},
		{"Quality Control", "Inspection", "পরিদর্শন"},
		{"Quality Control", "Testing", "পরীক্ষণ"},
		{"Store & Inventory", "Raw Materials", "কাঁচামাল"},
		{"Store & Inventory", "Finished Goods", "প্রস্তুত পণ্য"},
		{"Maintenance", "Electrical", "বৈদ্যুতিক"},
		{"Maintenance", "Mechanical", "যান্ত্রিক"},
		{"Production 2", "Cutting 2", "কাটিং ২"},
		{"Production 2", "Sewing 2", "সেলাই ২"},
	}

	for _, s := range sections {
		deptID := deptMap[s.Dept]
		var existing models.Section
		err := db.Where("department_id = ? AND name = ?", deptID, s.Name).First(&existing).Error
		if err == nil {
			continue
		}
		sec := models.Section{
			DepartmentID: deptID,
			Name:         s.Name,
			NameBn:       s.NameBn,
		}
		if err := db.Create(&sec).Error; err != nil {
			log.Printf("Failed to create section %s: %v", s.Name, err)
		} else {
			fmt.Printf("      Section: %s (%s)\n", s.Name, s.Dept)
		}
	}

	seedDesignations(db, company)
	seedLines(db, company)
}

func seedDesignations(db *gorm.DB, company *models.Company) {
	type desigDef struct {
		Section string
		Name    string
		NameBn  string
	}

	var sections []models.Section
	db.Model(&models.Section{}).
		Joins("JOIN departments ON departments.id = sections.department_id").
		Joins("JOIN companies ON companies.id = departments.company_id").
		Where("companies.id = ?", company.ID).
		Find(&sections)

	secMap := map[string]string{}
	for _, s := range sections {
		secMap[s.Name] = s.ID
	}

	designations := []desigDef{
		{"Software Development", "Software Engineer", "সফটওয়্যার ইঞ্জিনিয়ার"},
		{"Software Development", "Senior Software Engineer", "সিনিয়র সফটওয়্যার ইঞ্জিনিয়ার"},
		{"Software Development", "Team Lead", "টিম লিড"},
		{"Network & Infrastructure", "Network Administrator", "নেটওয়ার্ক অ্যাডমিনিস্ট্রেটর"},
		{"IT Support", "IT Support Engineer", "আইটি সাপোর্ট ইঞ্জিনিয়ার"},
		{"Accounts", "Accounts Officer", "অ্যাকাউন্টস অফিসার"},
		{"Accounts", "Senior Accountant", "সিনিয়র অ্যাকাউন্ট্যান্ট"},
		{"Payroll", "Payroll Officer", "পেরোল অফিসার"},
		{"Recruitment", "HR Executive", "এইচআর এক্সিকিউটিভ"},
		{"Recruitment", "HR Manager", "এইচআর ম্যানেজার"},
		{"Cutting", "Cutting Master", "কাটিং মাস্টার"},
		{"Cutting", "Cutting Operator", "কাটিং অপারেটর"},
		{"Sewing", "Sewing Supervisor", "সেলাই সুপারভাইজার"},
		{"Sewing", "Sewing Operator", "সেলাই অপারেটর"},
		{"Finishing", "Finishing Supervisor", "ফিনিশিং সুপারভাইজার"},
		{"Finishing", "Finishing Operator", "ফিনিশিং অপারেটর"},
		{"Packing", "Packing Supervisor", "প্যাকিং সুপারভাইজার"},
		{"Packing", "Packing Operator", "প্যাকিং অপারেটর"},
		{"Inspection", "QC Inspector", "কিউসি ইন্সপেক্টর"},
		{"Testing", "QC Tester", "কিউসি টেস্টার"},
		{"Raw Materials", "Store Keeper", "স্টোর কিপার"},
		{"Finished Goods", "Store Officer", "স্টোর অফিসার"},
		{"Electrical", "Electrician", "ইলেকট্রিশিয়ান"},
		{"Mechanical", "Mechanical Engineer", "মেকানিক্যাল ইঞ্জিনিয়ার"},
		{"Administration", "Admin Officer", "অ্যাডমিন অফিসার"},
		{"Marketing", "Marketing Executive", "মার্কেটিং এক্সিকিউটিভ"},
		{"Sales", "Sales Executive", "সেলস এক্সিকিউটিভ"},
	}

	for _, d := range designations {
		secID := secMap[d.Section]
		if secID == "" {
			continue
		}
		var existing models.Designation
		err := db.Where("section_id = ? AND name = ?", secID, d.Name).First(&existing).Error
		if err == nil {
			continue
		}
		desig := models.Designation{
			SectionID: secID,
			Name:      d.Name,
			NameBn:    d.NameBn,
		}
		if err := db.Create(&desig).Error; err != nil {
			log.Printf("Failed to create designation %s: %v", d.Name, err)
		}
	}
	fmt.Printf("      Designations seeded\n")
}

func seedLines(db *gorm.DB, company *models.Company) {
	var sections []models.Section
	db.Model(&models.Section{}).
		Joins("JOIN departments ON departments.id = sections.department_id").
		Joins("JOIN companies ON companies.id = departments.company_id").
		Where("companies.id = ?", company.ID).
		Where("sections.name IN ?", []string{"Cutting", "Sewing", "Finishing", "Packing"}).
		Find(&sections)

	for _, sec := range sections {
		for i := 1; i <= 3; i++ {
			name := fmt.Sprintf("%s Line %d", sec.Name, i)
			var existing models.Line
			err := db.Where("section_id = ? AND name = ?", sec.ID, name).First(&existing).Error
			if err == nil {
				continue
			}
			line := models.Line{
				SectionID: sec.ID,
				Name:      name,
			}
			if err := db.Create(&line).Error; err != nil {
				log.Printf("Failed to create line %s: %v", name, err)
			}
		}
	}
	fmt.Printf("      Lines seeded\n")
}

func seedGroups(db *gorm.DB) {
	groups := []string{"Group A", "Group B", "Group C", "Group D"}
	for _, g := range groups {
		var existing models.Group
		err := db.Where("name = ?", g).First(&existing).Error
		if err == nil {
			continue
		}
		if err := db.Create(&models.Group{Name: g}).Error; err != nil {
			log.Printf("Failed to create group %s: %v", g, err)
		} else {
			fmt.Printf("  Group: %s\n", g)
		}
	}
}

func seedFloors(db *gorm.DB) {
	floors := []string{"Floor 1", "Floor 2", "Floor 3", "Floor 4", "Floor 5", "Floor 6", "Floor 7", "Floor 8", "Floor 9", "Floor 10"}
	for _, f := range floors {
		var existing models.Floor
		err := db.Where("name = ?", f).First(&existing).Error
		if err == nil {
			continue
		}
		if err := db.Create(&models.Floor{Name: f}).Error; err != nil {
			log.Printf("Failed to create floor %s: %v", f, err)
		} else {
			fmt.Printf("  Floor: %s\n", f)
		}
	}
}

func seedShifts(db *gorm.DB, company *models.Company) {
	type shiftDef struct {
		Name      string
		ShiftType string
		Start     string
		End       string
		Grace     int
		Weekend   string
	}
	shifts := []shiftDef{
		{"General Shift", "day", "09:00", "18:00", 15, "Fri"},
		{"Morning Shift", "day", "06:00", "14:00", 10, "Fri"},
		{"Evening Shift", "evening", "14:00", "22:00", 10, "Fri"},
		{"Night Shift", "night", "22:00", "06:00", 10, "Fri"},
		{"General Half Day", "day", "09:00", "13:00", 0, "Fri"},
	}

	for _, s := range shifts {
		var existing models.Shift
		err := db.Where("company_id = ? AND name = ?", company.ID, s.Name).First(&existing).Error
		if err == nil {
			continue
		}
		shift := models.Shift{
			CompanyID:        company.ID,
			Name:             s.Name,
			ShiftType:        s.ShiftType,
			StartTime:        s.Start,
			EndTime:          s.End,
			LateGraceMinutes: s.Grace,
			WeekendDays:      s.Weekend,
			Status:           "active",
		}
		if err := db.Create(&shift).Error; err != nil {
			log.Printf("Failed to create shift %s: %v", s.Name, err)
		} else {
			fmt.Printf("  Shift: %s (%s-%s)\n", s.Name, s.Start, s.End)
		}
	}
}
