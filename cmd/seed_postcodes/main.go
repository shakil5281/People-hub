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

func main() {
	log.Println("Starting Postcode Seeding...")
	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB
	db.AutoMigrate(&models.PostOffice{})

	// Load JSON file
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
	for _, entry := range data {
		divNameEn := strings.TrimSpace(entry.En.Division)
		divNameBn := strings.TrimSpace(entry.Bn.Division)
		distNameEn := strings.TrimSpace(entry.En.District)
		distNameBn := strings.TrimSpace(entry.Bn.District)
		thanaNameEn := strings.TrimSpace(entry.En.Thana)
		thanaNameBn := strings.TrimSpace(entry.Bn.Thana)
		subOfficeNameEn := strings.TrimSpace(entry.En.SubOffice)
		subOfficeNameBn := strings.TrimSpace(entry.Bn.SubOffice)
		postCode := strings.TrimSpace(entry.En.PostCode)

		if postCode == "" || distNameEn == "" {
			continue
		}

		// Division
		var division models.Division
		if divNameEn != "" {
			err = db.Where("name = ?", divNameEn).FirstOrCreate(&division, models.Division{
				Name:   divNameEn,
				NameBn: divNameBn,
			}).Error
			if err != nil {
				log.Printf("Error with division %s: %v", divNameEn, err)
				continue
			}
		}

		// District
		var district models.District
		db.Where("name = ?", distNameEn).FirstOrCreate(&district, models.District{
			Name:       distNameEn,
			NameBn:     distNameBn,
			DivisionID: division.ID,
		})

		// Upazila (Thana)
		var upazila models.Upazila
		if thanaNameEn != "" {
			db.Where("name = ? AND district_id = ?", thanaNameEn, district.ID).FirstOrCreate(&upazila, models.Upazila{
				Name:       thanaNameEn,
				NameBn:     thanaNameBn,
				DistrictID: district.ID,
			})
		}

		// Post Office
		var postOffice models.PostOffice
		db.Where("postal_code = ? AND district_id = ?", postCode, district.ID).FirstOrCreate(&postOffice, models.PostOffice{
			Name:       subOfficeNameEn,
			NameBn:     subOfficeNameBn,
			PostalCode: postCode,
			DistrictID: district.ID,
			UpazilaID:  upazila.ID,
		})
		count++
	}

	fmt.Printf("Successfully processed %d postcodes.\n", count)
}
