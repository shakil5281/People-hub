package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

type GeoDivision struct {
	Name      string        `json:"name"`
	BnName    string        `json:"bn_name"`
	Districts []GeoDistrict `json:"districts"`
}

type GeoDistrict struct {
	Name      string      `json:"name"`
	BnName    string      `json:"bn_name"`
	Upazilas  []GeoUpazila `json:"upazilas"`
}

type GeoUpazila struct {
	Name   string    `json:"name"`
	BnName string    `json:"bn_name"`
	Unions []GeoUnion `json:"unions"`
}

type GeoUnion struct {
	Name   string `json:"name"`
	BnName string `json:"bn_name"`
}

func main() {
	cfg := config.Load()

	database.Connect(cfg)
	db := database.DB

	url := "https://iqbalhasandev.github.io/bangladesh-geo-json/bangladesh-geo.json"
	fmt.Println("Fetching geo data from:", url)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal("Failed to fetch geo data:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response:", err)
	}

	var divisions []GeoDivision
	if err := json.Unmarshal(body, &divisions); err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	fmt.Printf("Found %d divisions\n", len(divisions))

	var totalDistricts, totalUpazilas, totalUnions int

	for _, d := range divisions {
		division := models.Division{
			Name:   d.Name,
			NameBn: d.BnName,
		}
		if err := db.Where("name = ?", d.Name).FirstOrCreate(&division).Error; err != nil {
			log.Printf("Error creating division %s: %v", d.Name, err)
			continue
		}
		fmt.Printf("  Division: %s\n", d.Name)

		for _, dist := range d.Districts {
			district := models.District{
				Name:       dist.Name,
				NameBn:     dist.BnName,
				DivisionID: division.ID,
			}
			if err := db.Where("name = ? AND division_id = ?", dist.Name, division.ID).FirstOrCreate(&district).Error; err != nil {
				log.Printf("Error creating district %s: %v", dist.Name, err)
				continue
			}
			totalDistricts++
			fmt.Printf("    District: %s\n", dist.Name)

			for _, upa := range dist.Upazilas {
				upazila := models.Upazila{
					Name:       upa.Name,
					NameBn:     upa.BnName,
					DistrictID: district.ID,
				}
				if err := db.Where("name = ? AND district_id = ?", upa.Name, district.ID).FirstOrCreate(&upazila).Error; err != nil {
					log.Printf("Error creating upazila %s: %v", upa.Name, err)
					continue
				}
				totalUpazilas++
				fmt.Printf("      Upazila: %s\n", upa.Name)

				for _, u := range upa.Unions {
					union := models.Union{
						Name:      u.Name,
						NameBn:    u.BnName,
						UpazilaID: upazila.ID,
					}
					if err := db.Where("name = ? AND upazila_id = ?", u.Name, upazila.ID).FirstOrCreate(&union).Error; err != nil {
						log.Printf("Error creating union %s: %v", u.Name, err)
						continue
					}
					totalUnions++
				}
			}
		}
	}

	fmt.Println("\n--- Seeding complete ---")
	fmt.Printf("Divisions: %d\n", len(divisions))
	fmt.Printf("Districts: %d\n", totalDistricts)
	fmt.Printf("Upazilas: %d\n", totalUpazilas)
	fmt.Printf("Unions: %d\n", totalUnions)
}

