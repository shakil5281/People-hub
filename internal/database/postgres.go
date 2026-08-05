package database

import (
	"fmt"
	"log"

	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// GORM v1.31.2 forces UUID on *_id columns, overriding type:varchar(50) tags.
	// Use silent session for main AutoMigrate to suppress benign constraint management noise.
	silentMigrate := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	silentMigrate.AutoMigrate(
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
		&models.MissingAttendance{},
		&models.OtEarlyExitDeduction{},
	)
	// Ensure new tables were created; if not, create them explicitly.
	db.Exec("CREATE TABLE IF NOT EXISTS punishments (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS daily_schedules (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS tiffin_bills (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	// Drop removed tables
	db.Exec("DROP TABLE IF EXISTS night_bill_processes")
	db.Exec("DROP TABLE IF EXISTS night_bills")
	// Re-run AutoMigrate after ensuring tables exist so columns/indexes are added.
	db.AutoMigrate(
		&models.SalaryIncrement{},
		&models.Punishment{}, &models.DailySchedule{}, &models.TiffinBill{},
		&models.Holiday{},
		&models.SystemLog{},
		&models.Notification{},
		&models.EidBonus{},
	)

	// Use silent session for ALTER statements to avoid noisy ERROR logs when tables don't exist yet
	silentDB := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})

	alterCol := func(table, col string) {
		silentDB.Exec("ALTER TABLE " + table + " ALTER COLUMN " + col + " TYPE varchar(50) USING " + col + "::varchar(50)")
	}
	alterCol("employees", "employee_id")
	alterCol("attendances", "employee_id")
	alterCol("leaves", "employee_id")
	alterCol("leave_allocations", "employee_id")
	alterCol("salaries", "employee_id")
	alterCol("temporary_shifts", "employee_id")
	alterCol("salary_increments", "employee_id")
	alterCol("punishments", "employee_id")
	alterCol("daily_schedules", "employee_id")
	alterCol("tiffin_bills", "employee_id")
	alterCol("eid_bonuses", "employee_id")
	alterCol("id_cards", "employee_id")
	alterCol("separations", "employee_id")
	alterCol("missing_attendances", "employee_id")
	alterCol("ot_early_exit_deductions", "employee_id")
	silentDB.Exec("ALTER TABLE separations ADD COLUMN IF NOT EXISTS company_id uuid")
	silentDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS nid varchar(50)")
	silentDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS present_post_office varchar(100)")
	silentDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS present_post_code varchar(20)")
	silentDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS permanent_post_office varchar(100)")
	silentDB.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS permanent_post_code varchar(20)")
	silentDB.Exec("ALTER TABLE requirements ADD COLUMN IF NOT EXISTS section_id uuid")
	silentDB.Exec("ALTER TABLE requirements ADD COLUMN IF NOT EXISTS designation_id uuid")
	silentDB.Exec("ALTER TABLE requirements ADD COLUMN IF NOT EXISTS group_type varchar(20) DEFAULT 'Worker'")

	// Migrate check_in/check_out from varchar to timestamp.
	// Existing data may be "HH:mm" (time-only, length=5) or already a full datetime string.
	// Uses length() via ::text which works on both varchar and already-altered timestamp columns.
	silentDB.Exec(`
		ALTER TABLE attendances ALTER COLUMN check_in TYPE timestamp USING CASE
			WHEN check_in IS NOT NULL AND length(check_in::text) <= 5 THEN (date || ' ' || check_in)::timestamp
			WHEN check_in IS NOT NULL THEN check_in::timestamp
			ELSE NULL
		END
	`)
	silentDB.Exec(`
		ALTER TABLE attendances ALTER COLUMN check_out TYPE timestamp USING CASE
			WHEN check_out IS NOT NULL AND length(check_out::text) <= 5 THEN (date || ' ' || check_out)::timestamp
			WHEN check_out IS NOT NULL THEN check_out::timestamp
			ELSE NULL
		END
	`)

	// Add performance indexes for high-frequency queries
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_salaries_company_month_year ON salaries(company_id, year, month)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_employees_company_status ON employees(company_id, status) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_leave_allocations_emp_year ON leave_allocations(employee_id, year)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_temporary_shifts_company_date ON temporary_shifts(company_id, date)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_data_logs_date_processed ON data_logs(date, processed) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_attendances_date_status ON attendances(date, status)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_ot_early_exit_company_month ON ot_early_exit_deductions(company_id, year, month) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_leaves_status_dates ON leaves(status, from_date, to_date)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_level ON system_logs(level)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_source ON system_logs(source)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_user ON system_logs(user_id)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_created ON system_logs(created_at)")

	DB = db
	fmt.Println("Database connected successfully")
}
