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

type EmpInfo struct {
	EmpID       string     `gorm:"column:emp_id"`
	CardNo      string     `gorm:"column:card_no"`
	NatiID      string     `gorm:"column:nati_id"`
	EmpFname    string     `gorm:"column:emp_fname"`
	EmpLname    string     `gorm:"column:emp_lname"`
	NameBang    string     `gorm:"column:namebang"`
	Bdate       *time.Time `gorm:"column:bdate"`
	FatherName  string     `gorm:"column:father_name"`
	MotherName  string     `gorm:"column:mother_name"`
	PresentAdd  string     `gorm:"column:present_add"`
	PermanentAdd string    `gorm:"column:permanent_add"`
	Phone       string     `gorm:"column:phone"`
	BloodGroup  string     `gorm:"column:bloodgroup"`
	Sex         string     `gorm:"column:sex"`
	Email       string     `gorm:"column:email"`
	Religion    string     `gorm:"column:religion"`
	Mstatus     string     `gorm:"column:mstatus"`
	NoOfChildren *int      `gorm:"column:no_of_children"`
	JoinDate    *time.Time `gorm:"column:join_date"`
	Gross       *float64   `gorm:"column:gross"`
	Basic       *float64   `gorm:"column:basic"`
	HouseRent   *float64   `gorm:"column:house_rent"`
	Medical     *float64   `gorm:"column:medical"`
	Fooding     *float64   `gorm:"column:fooding"`
	Conveince   *float64   `gorm:"column:conveince"`
	OtherAllowance *float64 `gorm:"column:other_allowance"`
	Status      string     `gorm:"column:status"`
	OtStatus    string     `gorm:"column:ot_status"`
	BankAc      string     `gorm:"column:bank_ac"`
	PresentPob  string     `gorm:"column:present_pob"`
	PresentPcb  string     `gorm:"column:present_pcb"`
	PerPob      string     `gorm:"column:per_pob"`
	PerPcb      string     `gorm:"column:per_pcb"`
	SpouseName  string     `gorm:"column:spouse_name"`
	EmptypeID   *int       `gorm:"column:emptype_id"`
}

func (EmpInfo) TableName() string { return "EmpInfo" }

type EmpTrans struct {
	TrnID     int64      `gorm:"column:trn_id"`
	EmpID     string     `gorm:"column:emp_id"`
	TransType string     `gorm:"column:trans_type"`
	EffDate   *time.Time `gorm:"column:eff_date"`
	DDate     *time.Time `gorm:"column:d_date"`
	EmptypeID *int       `gorm:"column:emptype_id"`
}

func (EmpTrans) TableName() string { return "EmpTrans" }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	godotenv.Load()
	godotenv.Load(".env.local")

	dryRun := flag.Bool("dry-run", false, "dry run without inserting")
	onlyEmployees := flag.Bool("only-employees", false, "only migrate employees")
	onlySeparations := flag.Bool("only-separations", false, "only migrate separations")
	flag.Parse()

	companyID := "b0b60d1f-1bd8-4803-98ab-ecd40d8162f5"
	deptID := "" // will be fetched from peoplehub

	// Production department fallback for separations (NOT NULL)
	// Will fetch first department if not set
	pgHost := getEnv("DB_HOST", "localhost")
	pgPort := getEnv("DB_PORT", "5432")
	pgUser := getEnv("DB_USER", "postgres")
	pgPass := getEnv("DB_PASS", "123580")
	pgDB := getEnv("DB_NAME", "peoplehub")
	pgSSL := getEnv("DB_SSLMODE", "disable")

	hrhubDBName := getEnv("HRHUB_DB_NAME", "hrhub")
	hrhubUser := getEnv("HRHUB_DB_USER", "postgres")
	hrhubPass := getEnv("HRHUB_DB_PASS", "123580")
	// Force peoplehub to postgres if env shakil fails: allow override via env
	if getEnv("FORCE_PG_USER", "") != "" {
		pgUser = getEnv("FORCE_PG_USER", pgUser)
		pgPass = getEnv("FORCE_PG_PASS", pgPass)
	}

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

	// Fetch default department ID for separations
	if deptID == "" {
		var dept struct {
			ID string `gorm:"column:id"`
		}
		if err := peopleDB.Raw("SELECT id FROM departments WHERE deleted_at IS NULL LIMIT 1").Scan(&dept).Error; err == nil && dept.ID != "" {
			deptID = dept.ID
		} else {
			log.Fatalf("no department found for separations: %v", err)
		}
	}
	fmt.Printf("Using company_id=%s department_id=%s\n", companyID, deptID)

	// Load existing employee_ids into map
	var existingIDs []string
	peopleDB.Raw("SELECT employee_id FROM employees WHERE deleted_at IS NULL").Scan(&existingIDs)
	existingMap := make(map[string]bool, len(existingIDs))
	for _, id := range existingIDs {
		existingMap[strings.TrimSpace(id)] = true
	}
	fmt.Printf("PeopleHub employees existing: %d\n", len(existingMap))

	// Load latest EmpTrans per emp_id for employee_type/status
	type latestTrans struct {
		EmpID     string
		TransType string
		EmptypeID *int
		EffDate   *time.Time
	}
	var latestList []latestTrans
	// Use DISTINCT ON equivalent via subquery
	hrhubDB.Raw(`
		SELECT DISTINCT ON (emp_id) emp_id, trans_type, emptype_id, eff_date
		FROM "EmpTrans" ORDER BY emp_id, eff_date DESC
	`).Scan(&latestList)
	transMap := make(map[string]latestTrans, len(latestList))
	for _, lt := range latestList {
		transMap[strings.TrimSpace(lt.EmpID)] = lt
	}
	fmt.Printf("EmpTrans latest map: %d distinct emp_id\n", len(transMap))

	if !*onlySeparations {
		// --- EMPLOYEES MIGRATION ---
		var empInfos []EmpInfo
		if err := hrhubDB.Table(`"EmpInfo"`).Find(&empInfos).Error; err != nil {
			log.Fatalf("load EmpInfo failed: %v", err)
		}
		fmt.Printf("HrHub EmpInfo total: %d\n", len(empInfos))

		toInsert := 0
		skippedExists := 0
		skippedInvalid := 0
		invalidRows := []string{}

		// Prepare batch
		type EmployeeRow struct {
			ID                 string  `gorm:"column:id"`
			CompanyID          string  `gorm:"column:company_id"`
			EmployeeID         string  `gorm:"column:employee_id"`
			PunchNumber        string  `gorm:"column:punch_number"`
			NameEn             string  `gorm:"column:name_en"`
			NameBn             string  `gorm:"column:name_bn"`
			FatherName         string  `gorm:"column:father_name"`
			MotherName         string  `gorm:"column:mother_name"`
			DateOfBirth        string  `gorm:"column:date_of_birth"`
			NID                string  `gorm:"column:nid"`
			Phone              string  `gorm:"column:phone"`
			BloodGroup         string  `gorm:"column:blood_group"`
			Gender             string  `gorm:"column:gender"`
			Email              string  `gorm:"column:email"`
			Religion           string  `gorm:"column:religion"`
			MaritalStatus      string  `gorm:"column:marital_status"`
			PresentAddress     string  `gorm:"column:present_address"`
			PermanentAddress   string  `gorm:"column:permanent_address"`
			SpouseName         string  `gorm:"column:spouse_name"`
			NumberOfDependents int     `gorm:"column:number_of_dependents"`
			JoiningDate        time.Time `gorm:"column:joining_date"`
			GrossSalary        float64 `gorm:"column:gross_salary"`
			BasicSalary        float64 `gorm:"column:basic_salary"`
			HouseRent          float64 `gorm:"column:house_rent"`
			MedicalAllowance   float64 `gorm:"column:medical_allowance"`
			FoodAllowance      float64 `gorm:"column:food_allowance"`
			TransportAllowance float64 `gorm:"column:transport_allowance"`
			OtherAllowance     float64 `gorm:"column:other_allowance"`
			AccountNumber      string  `gorm:"column:account_number"`
			PresentPostOffice  *string `gorm:"column:present_post_office"`
			PresentPostCode    *string `gorm:"column:present_post_code"`
			PermanentPostOffice *string `gorm:"column:permanent_post_office"`
			PermanentPostCode   *string `gorm:"column:permanent_post_code"`
			Status             string  `gorm:"column:status"`
			OverTimeStatus     bool    `gorm:"column:over_time_status"`
			EmployeeType       string  `gorm:"column:employee_type"`
		}

		var batch []EmployeeRow
		const batchSize = 500

		flush := func() error {
			if len(batch) == 0 {
				return nil
			}
			if *dryRun {
				fmt.Printf("[DRY-RUN] would insert %d employees (sample employee_id=%s)\n", len(batch), batch[0].EmployeeID)
				batch = batch[:0]
				return nil
			}
			// Use raw insert with ON CONFLICT DO NOTHING to respect existing
			tx := peopleDB.Begin()
			for _, r := range batch {
				// Double-check exists to avoid punch_number conflict
				var cnt int64
				tx.Raw("SELECT count(*) FROM employees WHERE employee_id = ? AND deleted_at IS NULL", r.EmployeeID).Scan(&cnt)
				if cnt > 0 {
					skippedExists++
					continue
				}
				// Also check punch_number collision
				tx.Raw("SELECT count(*) FROM employees WHERE punch_number = ? AND deleted_at IS NULL", r.PunchNumber).Scan(&cnt)
				punch := r.PunchNumber
				if cnt > 0 {
					// make unique by appending emp_id suffix? skip instead to avoid violation
					// try alternative punch = employee_id + "_HR"
					alt := r.EmployeeID + "_HR"
					tx.Raw("SELECT count(*) FROM employees WHERE punch_number = ? AND deleted_at IS NULL", alt).Scan(&cnt)
					if cnt > 0 {
						skippedExists++
						continue
					}
					punch = alt
				}
				err := tx.Exec(`
					INSERT INTO employees (id, company_id, employee_id, punch_number, name_en, name_bn, father_name, mother_name, date_of_birth, nid, phone, blood_group, gender, email, religion, marital_status, present_address, permanent_address, spouse_name, number_of_dependents, joining_date, gross_salary, basic_salary, house_rent, medical_allowance, food_allowance, transport_allowance, other_allowance, account_number, present_post_office, present_post_code, permanent_post_office, permanent_post_code, status, over_time_status, employee_type, created_at, updated_at)
					VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), now())
					ON CONFLICT (employee_id) DO NOTHING
				`, r.CompanyID, r.EmployeeID, punch, r.NameEn, r.NameBn, r.FatherName, r.MotherName, r.DateOfBirth, r.NID, r.Phone, r.BloodGroup, r.Gender, r.Email, r.Religion, r.MaritalStatus, r.PresentAddress, r.PermanentAddress, r.SpouseName, r.NumberOfDependents, r.JoiningDate, r.GrossSalary, r.BasicSalary, r.HouseRent, r.MedicalAllowance, r.FoodAllowance, r.TransportAllowance, r.OtherAllowance, r.AccountNumber, r.PresentPostOffice, r.PresentPostCode, r.PermanentPostOffice, r.PermanentPostCode, r.Status, r.OverTimeStatus, r.EmployeeType).Error
				if err != nil {
					tx.Rollback()
					return err
				}
				toInsert++
			}
			if err := tx.Commit().Error; err != nil {
				return err
			}
			batch = batch[:0]
			return nil
		}

		for _, e := range empInfos {
			empID := strings.TrimSpace(e.EmpID)
			if empID == "" {
				skippedInvalid++
				invalidRows = append(invalidRows, fmt.Sprintf("empty emp_id card_no=%s", e.CardNo))
				continue
			}
			if existingMap[empID] {
				skippedExists++
				continue
			}
			// punch_number
			punch := strings.TrimSpace(e.CardNo)
			if punch == "" {
				punch = empID
			}
			// name_en
			nameEn := strings.TrimSpace(strings.TrimSpace(e.EmpFname) + " " + strings.TrimSpace(e.EmpLname))
			if nameEn == "" {
				nameEn = empID
			}
			// date_of_birth
			dob := ""
			if e.Bdate != nil {
				dob = e.Bdate.Format("2006-01-02")
			}
			// blood_group
			bg := strings.TrimSpace(e.BloodGroup)
			if bg == "N/A" {
				bg = ""
			}
			// gender
			gender := ""
			switch strings.ToUpper(strings.TrimSpace(e.Sex)) {
			case "M":
				gender = "Male"
			case "F":
				gender = "Female"
			default:
				if e.Sex != "" {
					gender = strings.TrimSpace(e.Sex)
				}
			}
			// joining_date required
			var joinDate time.Time
			if e.JoinDate != nil && !e.JoinDate.IsZero() {
				joinDate = *e.JoinDate
			} else {
				skippedInvalid++
				invalidRows = append(invalidRows, empID)
				continue
			}
			// salaries
			gross := 0.0
			if e.Gross != nil {
				gross = *e.Gross
			}
			basic := 0.0
			if e.Basic != nil {
				basic = *e.Basic
			}
			hr := 0.0
			if e.HouseRent != nil {
				hr = *e.HouseRent
			}
			med := 750.0
			if e.Medical != nil && *e.Medical != 0 {
				med = *e.Medical
			}
			food := 1250.0
			if e.Fooding != nil && *e.Fooding != 0 {
				food = *e.Fooding
			}
			conv := 450.0
			if e.Conveince != nil && *e.Conveince != 0 {
				conv = *e.Conveince
			}
			other := 0.0
			if e.OtherAllowance != nil {
				other = *e.OtherAllowance
			}
			// status from EmpTrans + EmpInfo
			status := "active"
			employeeType := "Regular"
			if lt, ok := transMap[empID]; ok {
				// trans_type Left/Resign etc indicates inactive
				tt := strings.TrimSpace(lt.TransType)
				if tt == "Resign" || tt == "Left" || tt == "Dismiss" || tt == "Terminated" || tt == "Did Not Join" {
					status = "inactive"
					if tt == "Resign" {
						employeeType = "Resign"
					} else if tt == "Left" {
						employeeType = "Close"
					}
				} else if tt == "Rejoin" {
					status = "active"
					employeeType = "Regular"
				}
				// emptype_id could influence employee_type but keep Regular as default
				// If emptype_id exists and employee is inactive, preserve type
			} else {
				// fallback to EmpInfo status
				s := strings.TrimSpace(e.Status)
				if s == "Left" || s == "Resign" || s == "Resigne" || s == "Dismiss" || s == "6" || s == "8" {
					status = "inactive"
				}
			}
			ot := strings.ToUpper(strings.TrimSpace(e.OtStatus)) == "Y"
			// post office fields
			var presentPob *string
			if v := strings.TrimSpace(e.PresentPob); v != "" {
				presentPob = &v
			}
			var presentPcb *string
			if v := strings.TrimSpace(e.PresentPcb); v != "" {
				presentPcb = &v
			}
			var perPob *string
			if v := strings.TrimSpace(e.PerPob); v != "" {
				perPob = &v
			}
			var perPcb *string
			if v := strings.TrimSpace(e.PerPcb); v != "" {
				perPcb = &v
			}
			nid := strings.TrimSpace(e.NatiID)
			phone := strings.TrimSpace(e.Phone)
			email := strings.TrimSpace(e.Email)
			dependents := 0
			if e.NoOfChildren != nil {
				dependents = *e.NoOfChildren
			}

			batch = append(batch, EmployeeRow{
				CompanyID:          companyID,
				EmployeeID:         empID,
				PunchNumber:        punch,
				NameEn:             nameEn,
				NameBn:             strings.TrimSpace(e.NameBang),
				FatherName:         strings.TrimSpace(e.FatherName),
				MotherName:         strings.TrimSpace(e.MotherName),
				DateOfBirth:        dob,
				NID:                nid,
				Phone:              phone,
				BloodGroup:         bg,
				Gender:             gender,
				Email:              email,
				Religion:           strings.TrimSpace(e.Religion),
				MaritalStatus:      strings.TrimSpace(e.Mstatus),
				PresentAddress:     strings.TrimSpace(e.PresentAdd),
				PermanentAddress:   strings.TrimSpace(e.PermanentAdd),
				SpouseName:         strings.TrimSpace(e.SpouseName),
				NumberOfDependents: dependents,
				JoiningDate:        joinDate,
				GrossSalary:        gross,
				BasicSalary:        basic,
				HouseRent:          hr,
				MedicalAllowance:   med,
				FoodAllowance:      food,
				TransportAllowance: conv,
				OtherAllowance:     other,
				AccountNumber:      strings.TrimSpace(e.BankAc),
				PresentPostOffice:  presentPob,
				PresentPostCode:    presentPcb,
				PermanentPostOffice: perPob,
				PermanentPostCode:   perPcb,
				Status:             status,
				OverTimeStatus:     ot,
				EmployeeType:       employeeType,
			})
			if len(batch) >= batchSize {
				if err := flush(); err != nil {
					log.Fatalf("flush employees failed: %v", err)
				}
			}
		}
		if err := flush(); err != nil {
			log.Fatalf("final flush employees failed: %v", err)
		}
		fmt.Printf("Employees migration done: inserted=%d skipped_exists=%d skipped_invalid=%d\n", toInsert, skippedExists, skippedInvalid)
		if len(invalidRows) > 0 && len(invalidRows) < 20 {
			fmt.Printf("Invalid rows: %v\n", invalidRows)
		}
	}

	if !*onlyEmployees {
		// --- SEPARATIONS MIGRATION ---
		// Check existing separations to avoid duplicates: key (employee_id, type, date)
		var existingSeps []struct {
			EmployeeID string `gorm:"column:employee_id"`
			Type       string `gorm:"column:type"`
			Date       string `gorm:"column:date"`
		}
		peopleDB.Raw("SELECT employee_id, type, date FROM separations WHERE deleted_at IS NULL").Scan(&existingSeps)
		sepMap := make(map[string]bool, len(existingSeps))
		for _, s := range existingSeps {
			key := strings.TrimSpace(s.EmployeeID) + "|" + strings.TrimSpace(s.Type) + "|" + strings.TrimSpace(s.Date)
			sepMap[key] = true
		}
		fmt.Printf("Existing separations: %d\n", len(sepMap))

		var transList []EmpTrans
		if err := hrhubDB.Table(`"EmpTrans"`).Order("eff_date ASC").Find(&transList).Error; err != nil {
			log.Fatalf("load EmpTrans failed: %v", err)
		}
		fmt.Printf("EmpTrans total: %d\n", len(transList))

		// Need employee names for separations: load EmpInfo name map
		var empInfos []EmpInfo
		hrhubDB.Table(`"EmpInfo"`).Select("emp_id, emp_fname, emp_lname").Find(&empInfos)
		nameMap := make(map[string]string, len(empInfos))
		for _, ei := range empInfos {
			nameMap[strings.TrimSpace(ei.EmpID)] = strings.TrimSpace(strings.TrimSpace(ei.EmpFname) + " " + strings.TrimSpace(ei.EmpLname))
		}

		insertedSep := 0
		skippedSepExists := 0
		skippedSepInvalid := 0
		batchCount := 0
		tx := peopleDB.Begin()
		// Build set of employee_ids that will exist after employees migration (existing + to-be-inserted)
		willExist := make(map[string]bool, len(existingMap)+3000)
		for k := range existingMap {
			willExist[k] = true
		}
		// Include EmpInfo ids that would be inserted (dry-run or not)
		for _, e := range empInfos {
			id := strings.TrimSpace(e.EmpID)
			if id != "" && !existingMap[id] {
				willExist[id] = true
			}
		}
		allowedSepTypes := map[string]bool{"Left": true, "Resign": true, "Dismiss": true, "Terminated": true, "Did Not Join": true}
		for _, tr := range transList {
			empID := strings.TrimSpace(tr.EmpID)
			if empID == "" {
				skippedSepInvalid++
				continue
			}
			tt := strings.TrimSpace(tr.TransType)
			if tt == "" {
				continue
			}
			if !allowedSepTypes[tt] {
				continue
			}
			sepType := tt
			// Date from eff_date or d_date
			var d *time.Time
			if tr.EffDate != nil {
				d = tr.EffDate
			} else if tr.DDate != nil {
				d = tr.DDate
			} else {
				skippedSepInvalid++
				continue
			}
			dateStr := d.Format("2006-01-02")
			key := empID + "|" + sepType + "|" + dateStr
			if sepMap[key] {
				skippedSepExists++
				continue
			}
			// Check if employee will exist (existing or to-be-inserted)
			if !willExist[empID] {
				skippedSepInvalid++
				continue
			}
			empName := nameMap[empID]
			if empName == "" {
				empName = empID
			}
			if *dryRun {
				if batchCount < 3 {
					fmt.Printf("[DRY-RUN] separation emp_id=%s type=%s date=%s name=%s\n", empID, sepType, dateStr, empName)
				}
				batchCount++
				insertedSep++
				continue
			}
			// DepartmentID is NOT NULL - use default deptID
			err := tx.Exec(`
				INSERT INTO separations (id, employee, employee_id, company_id, department_id, type, date, status, reason, created_at, updated_at)
				VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, 'Pending', '', now(), now())
				ON CONFLICT DO NOTHING
			`, empName, empID, companyID, deptID, sepType, dateStr).Error
			if err != nil {
				tx.Rollback()
				log.Fatalf("insert separation failed emp_id=%s: %v", empID, err)
			}
			sepMap[key] = true
			insertedSep++
			batchCount++
			// Commit per 500 to avoid huge tx
			if batchCount%500 == 0 {
				if err := tx.Commit().Error; err != nil {
					log.Fatalf("commit separations batch failed: %v", err)
				}
				tx = peopleDB.Begin()
			}
		}
		if !*dryRun {
			if err := tx.Commit().Error; err != nil {
				log.Fatalf("final commit separations failed: %v", err)
			}
		}
		fmt.Printf("Separations migration done: inserted=%d skipped_exists=%d skipped_invalid=%d\n", insertedSep, skippedSepExists, skippedSepInvalid)
	}

	fmt.Println("Migration completed")
}
