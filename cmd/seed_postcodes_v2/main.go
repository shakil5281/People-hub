package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

type PostcodeEntry struct {
	En struct {
		Division  string `json:"division"`
		District  string `json:"district"`
		Thana     string `json:"thana"`
		SubOffice string `json:"suboffice"`
		PostCode  string `json:"postcode"`
	} `json:"en"`
	Bn struct {
		Division  string `json:"division"`
		District  string `json:"district"`
		Thana     string `json:"thana"`
		SubOffice string `json:"suboffice"`
		PostCode  string `json:"postcode"`
	} `json:"bn"`
}

var divisionNameMap = map[string]string{
	"Chattogram": "Chattagram",
	"Barishal":   "Barisal",
}

var districtNameMap = map[string]string{
	"Cumilla":          "Comilla",
	"Coxsbazar":        "Coxsbazar",
	"Cox's Bazar":      "Coxsbazar",
	"Chapai Nawabganj": "Chapainawabganj",
	"Khagrachari":      "Khagrachhari",
	"Netrokona":        "Netrakona",
	"Jhalokati":        "Jhalakathi",
	"Barishal":         "Barisal",
}

func main() {
	log.Println("Starting Postcode Seeding (v3)...")
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	file, err := os.Open("postcode.json")
	if err != nil {
		log.Fatalf("Failed to open postcode.json: %v", err)
	}
	defer file.Close()

	var data map[string]PostcodeEntry
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	count := 0
	skipped := 0
	for _, entry := range data {
		divNameEn := strings.TrimSpace(entry.En.Division)
		distNameEn := strings.TrimSpace(entry.En.District)
		thanaNameEn := strings.TrimSpace(entry.En.Thana)
		subOfficeNameEn := strings.TrimSpace(entry.En.SubOffice)
		subOfficeNameBn := strings.TrimSpace(entry.Bn.SubOffice)
		postCode := strings.TrimSpace(entry.En.PostCode)

		if postCode == "" || distNameEn == "" {
			skipped++
			continue
		}

		if mapped, ok := divisionNameMap[divNameEn]; ok {
			divNameEn = mapped
		}
		if mapped, ok := districtNameMap[distNameEn]; ok {
			distNameEn = mapped
		}

		var division models.Division
		if err := db.Where("name = ?", divNameEn).First(&division).Error; err != nil {
			skipped++
			continue
		}

		var district models.District
		if err := db.Where("name = ? AND division_id = ?", distNameEn, division.ID).First(&district).Error; err != nil {
			if err := db.Where("name = ?", distNameEn).First(&district).Error; err != nil {
				skipped++
				continue
			}
		}

		postOffice := models.PostOffice{
			Name:       subOfficeNameEn,
			NameBn:     subOfficeNameBn,
			PostalCode: postCode,
			DistrictID: district.ID,
		}

		var existing models.PostOffice
		if err := db.Where("postal_code = ? AND district_id = ?", postCode, district.ID).First(&existing).Error; err == nil {
			count++
			continue
		}

		if thanaNameEn != "" {
			var upazila models.Upazila
			if err := db.Where("name = ? AND district_id = ?", thanaNameEn, district.ID).First(&upazila).Error; err == nil {
				postOffice.UpazilaID = upazila.ID
			}
		}

		if postOffice.UpazilaID != "" {
			if err := db.Create(&postOffice).Error; err != nil {
				continue
			}
		} else {
			if err := db.Omit("upazila_id").Create(&postOffice).Error; err != nil {
				continue
			}
		}
		count++
	}

	fmt.Printf("Successfully processed %d postcodes (skipped %d).\n", count, skipped)
}
