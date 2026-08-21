package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/auth"
	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	email := os.Getenv("SUPERADMIN_EMAIL")
	password := os.Getenv("SUPERADMIN_PASSWORD")
	name := os.Getenv("SUPERADMIN_NAME")

	if email == "" {
		email = "superadmin@peoplehub.com"
	}
	if password == "" {
		password = "superadmin1234"
	}
	if name == "" {
		name = "Super Admin"
	}

	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	db.Logger = db.Logger.LogMode(logger.Info)

	fmt.Println("\n========================================")
	fmt.Println("  DROPPING ALL TABLES...")
	fmt.Println("========================================")
	dropAllTables(db)

	fmt.Println("\n========================================")
	fmt.Println("  RUNNING AUTO-MIGRATION...")
	fmt.Println("========================================")
	runMigration(db)

	fmt.Println("\n========================================")
	fmt.Println("  RUNNING ALTER FIXES & INDEXES...")
	fmt.Println("========================================")
	runPostMigrationFixes(db)

	fmt.Println("\n========================================")
	fmt.Println("  SEEDING SUPERADMIN...")
	fmt.Println("========================================")
	role := seedRole(db)
	seedPermissions(db, role)
	user := seedUser(db, email, password, name)
	assignRole(db, user, role)

	fmt.Println("\n========================================")
	fmt.Println("  RESET COMPLETE")
	fmt.Println("========================================")
	fmt.Printf("  Email:    %s\n", email)
	fmt.Printf("  Password: %s\n", password)
	fmt.Println("  All other data has been deleted.")
	fmt.Println()
}

func dropAllTables(db *gorm.DB) {
	tables := []interface{}{
		&models.User{}, &models.Role{}, &models.Permission{}, &models.UserRole{},
		&models.RolePermission{}, &models.RefreshToken{}, &models.LoginHistory{},
		&models.PasswordHistory{}, &models.AuditLog{}, &models.EmailVerification{},
		&models.PasswordReset{}, &models.Company{},
		&models.Department{}, &models.Section{}, &models.Designation{}, &models.Line{},
		&models.Group{}, &models.Floor{}, &models.Division{}, &models.District{},
		&models.Upazila{}, &models.Union{}, &models.Employee{}, &models.Requirement{},
		&models.Separation{}, &models.IdCard{}, &models.Shift{}, &models.LeaveType{},
		&models.LeaveAllocation{}, &models.Leave{}, &models.TemporaryShift{},
		&models.Attendance{}, &models.DataLog{}, &models.Salary{}, &models.Session{},
		&models.SystemSetting{}, &models.SalaryIncrement{},
		&models.Punishment{}, &models.DailySchedule{}, &models.TiffinBill{},
		&models.Holiday{},
		&models.SystemLog{},
		&models.Notification{},
		&models.EidBonus{},
	}
	if err := db.Migrator().DropTable(tables...); err != nil {
		log.Fatal("Failed to drop tables:", err)
	}
	fmt.Println("All tables dropped successfully")
}

func runMigration(db *gorm.DB) {
	if err := db.AutoMigrate(
		&models.User{}, &models.Role{}, &models.Permission{}, &models.UserRole{},
		&models.RolePermission{}, &models.RefreshToken{}, &models.LoginHistory{},
		&models.PasswordHistory{}, &models.AuditLog{}, &models.EmailVerification{},
		&models.PasswordReset{}, &models.Company{},
		&models.Department{}, &models.Section{}, &models.Designation{}, &models.Line{},
		&models.Group{}, &models.Floor{}, &models.Division{}, &models.District{},
		&models.Upazila{}, &models.Union{}, &models.Employee{}, &models.Requirement{},
		&models.Separation{}, &models.IdCard{}, &models.Shift{}, &models.LeaveType{},
		&models.LeaveAllocation{}, &models.Leave{}, &models.TemporaryShift{},
		&models.Attendance{}, &models.DataLog{}, &models.Salary{}, &models.Session{},
		&models.SystemSetting{}, &models.SalaryIncrement{},
		&models.Punishment{}, &models.DailySchedule{}, &models.TiffinBill{},
		&models.Holiday{},
		&models.SystemLog{},
		&models.Notification{},
		&models.EidBonus{},
	); err != nil {
		log.Fatal("Migration failed:", err)
	}
	fmt.Println("Auto-migration completed successfully")
}

func runPostMigrationFixes(db *gorm.DB) {
	fixes := []string{
		"ALTER TABLE employees ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE attendances ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE leaves ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE leave_allocations ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE salaries ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE temporary_shifts ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE salary_increments ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE punishments ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE daily_schedules ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE tiffin_bills ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE eid_bonuses ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE id_cards ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"ALTER TABLE separations ALTER COLUMN employee_id TYPE varchar(50) USING employee_id::varchar(50)",
		"CREATE INDEX IF NOT EXISTS idx_salaries_company_month_year ON salaries(company_id, year, month)",
		"CREATE INDEX IF NOT EXISTS idx_employees_company_status ON employees(company_id, status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_leave_allocations_emp_year ON leave_allocations(employee_id, year)",
		"CREATE INDEX IF NOT EXISTS idx_temporary_shifts_company_date ON temporary_shifts(company_id, date)",
		"CREATE INDEX IF NOT EXISTS idx_data_logs_date_processed ON data_logs(date, processed) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_attendances_date_status ON attendances(date, status)",
		"CREATE INDEX IF NOT EXISTS idx_leaves_status_dates ON leaves(status, from_date, to_date)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id) WHERE deleted_at IS NULL",
	}
	for _, fix := range fixes {
		if err := db.Exec(fix).Error; err != nil {
			log.Printf("Warning: %s: %v", fix, err)
		}
	}
	fmt.Println("Post-migration fixes and indexes applied")
}

func seedRole(db *gorm.DB) *models.Role {
	var role models.Role
	err := db.Where("name = ? AND is_system = ?", "super_admin", true).First(&role).Error
	if err == nil {
		fmt.Println("Superadmin role already exists, skipping creation")
		return &role
	}

	role = models.Role{
		Name:        "super_admin",
		Description: "Super administrator with full system access",
		IsSystem:    true,
	}
	if err := db.Create(&role).Error; err != nil {
		log.Fatal("Failed to create superadmin role:", err)
	}
	fmt.Println("Created super_admin role")
	return &role
}

func seedPermissions(db *gorm.DB, role *models.Role) {
	resources := []string{
		"users", "roles", "permissions",
		"companies", "departments", "sections", "designations", "lines", "groups", "floors",
		"employees",
		"shifts",
		"attendance",
		"data_logs",
		"divisions", "districts", "upazilas", "unions",
		"dashboard", "audit_logs", "settings",
	}
	actions := []string{"create", "read", "update", "delete", "manage"}

	count := 0
	for _, res := range resources {
		for _, act := range actions {
			permName := res + "." + act
			var perm models.Permission
			err := db.Where("resource = ? AND action = ?", res, act).First(&perm).Error
			if err != nil {
				perm = models.Permission{
					Resource:    res,
					Action:      act,
					Description: fmt.Sprintf("Can %s %s", act, res),
				}
				if err := db.Create(&perm).Error; err != nil {
					log.Printf("Failed to create permission %s: %v", permName, err)
					continue
				}
			}

			var rp models.RolePermission
			err = db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rp).Error
			if err != nil {
				rp = models.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				}
				if err := db.Create(&rp).Error; err != nil {
					log.Printf("Failed to assign permission %s to role: %v", permName, err)
					continue
				}
				count++
			}
		}
	}
	fmt.Printf("Seeded %d permissions and assigned to super_admin role\n", count)
}

func seedUser(db *gorm.DB, email, password, name string) *models.User {
	var existing models.User
	err := db.Where("email = ?", email).First(&existing).Error
	if err == nil {
		fmt.Printf("User %s already exists, skipping creation\n", email)
		return &existing
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	now := time.Now()
	user := models.User{
		Email:              email,
		PasswordHash:       hash,
		Name:               name,
		Status:             "active",
		EmailVerifiedAt:    &now,
		ForcePasswordChange: false,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatal("Failed to create superadmin user:", err)
	}
	fmt.Printf("Created superadmin user: %s\n", email)
	return &user
}

func assignRole(db *gorm.DB, user *models.User, role *models.Role) {
	var ur models.UserRole
	err := db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&ur).Error
	if err == nil {
		fmt.Println("Role already assigned to user")
		return
	}

	ur = models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}
	if err := db.Create(&ur).Error; err != nil {
		log.Fatal("Failed to assign super_admin role to user:", err)
	}
	fmt.Printf("Assigned super_admin role to user %s\n", user.Email)
}
