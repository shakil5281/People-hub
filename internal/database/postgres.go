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

	// Night Bill: migrate legacy columns to spec names before AutoMigrate maps the new model.
	// Guarded with information_schema checks so they are idempotent across restarts and fresh DBs.
	db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills') THEN
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='date') AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='attendance_date') THEN
					ALTER TABLE night_bills RENAME COLUMN "date" TO attendance_date;
				END IF;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='check_in') AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='in_time') THEN
					ALTER TABLE night_bills RENAME COLUMN check_in TO in_time;
				END IF;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='check_out') AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='out_time') THEN
					ALTER TABLE night_bills RENAME COLUMN check_out TO out_time;
				END IF;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='mode') AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='bill_type') THEN
					ALTER TABLE night_bills RENAME COLUMN mode TO bill_type;
				END IF;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='extra_hours') AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='eligible_hours') THEN
					ALTER TABLE night_bills RENAME COLUMN extra_hours TO eligible_hours;
				END IF;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='in_time') THEN
					ALTER TABLE night_bills ALTER COLUMN in_time TYPE timestamp WITHOUT TIME ZONE USING in_time AT TIME ZONE 'UTC';
				END IF;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=CURRENT_SCHEMA() AND table_name='night_bills' AND column_name='out_time') THEN
					ALTER TABLE night_bills ALTER COLUMN out_time TYPE timestamp WITHOUT TIME ZONE USING out_time AT TIME ZONE 'UTC';
				END IF;
			END IF;
		END $$;
	`)
	db.Exec("ALTER TABLE night_bills ADD COLUMN IF NOT EXISTS attendance_id uuid")
	db.Exec("ALTER TABLE night_bills ADD COLUMN IF NOT EXISTS processed_at timestamp")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_night_bills_employee_date_type ON night_bills(employee_id, attendance_date, bill_type)")
	db.Exec(`
		UPDATE night_bills nb
		SET in_time = a.check_in,
		    out_time = a.check_out
		FROM attendances a
		WHERE nb.attendance_id = a.id
		  AND a.deleted_at IS NULL
		  AND (nb.in_time IS NULL OR nb.out_time IS NULL OR nb.in_time != a.check_in OR nb.out_time != a.check_out)
	`)

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
		&models.Roster{},
		&models.Attendance{}, &models.DataLog{}, &models.Salary{}, &models.Session{},
		&models.SystemSetting{}, &models.SalaryIncrement{}, &models.AdvanceSalary{},
		&models.Punishment{}, &models.DailySchedule{}, &models.TiffinBill{},
		&models.Holiday{}, &models.PostOffice{},
		&models.MissingAttendance{},
		&models.OtEarlyExitDeduction{},
		&models.OtEarlyExitExemption{},
		&models.NightBill{},
		&models.NightBillEmployeeList{},
		&models.EmployeeMigration{},
	)
	// Ensure new tables were created; if not, create them explicitly.
	db.Exec("CREATE TABLE IF NOT EXISTS punishments (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS daily_schedules (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS tiffin_bills (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS night_bills (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS employee_migrations (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS rosters (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS ot_early_exit_exemptions (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	// Re-run AutoMigrate after ensuring tables exist so columns/indexes are added.
	db.AutoMigrate(
		&models.SalaryIncrement{}, &models.AdvanceSalary{},
		&models.Punishment{}, &models.DailySchedule{}, &models.TiffinBill{},
		&models.Holiday{},
		&models.SystemLog{},
		&models.Notification{},
		&models.EidBonus{},
		&models.NightBill{},
		&models.NightBillEmployeeList{},
		&models.EmployeeMigration{},
		&models.Roster{},
		&models.OtEarlyExitExemption{},
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
	alterCol("night_bill_employee_lists", "employee_id")
	alterCol("rosters", "employee_id")
	// Recreate employee_id unique index as partial so soft-deleted records don't block re-adding an employee.
	silentDB.Exec("DROP INDEX IF EXISTS idx_night_bill_employee_lists_employee_id")
	silentDB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_night_bill_employee_lists_employee_id ON night_bill_employee_lists(employee_id) WHERE deleted_at IS NULL")
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
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_attendances_company_date ON attendances(company_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_attendances_employee_date_company ON attendances(employee_id, date, company_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_attendances_date_status ON attendances(date, status)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_ot_early_exit_company_month ON ot_early_exit_deductions(company_id, year, month) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_leaves_status_dates ON leaves(status, from_date, to_date)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_rosters_company_date ON rosters(company_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_rosters_employee_date ON rosters(employee_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_holidays_company_date ON holidays(company_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_holidays_weekend_date ON holidays(company_id, weekend_date) WHERE deleted_at IS NULL AND weekend_date IS NOT NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_holidays_type_status ON holidays(type, status) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_level ON system_logs(level)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_source ON system_logs(source)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_user ON system_logs(user_id)")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_created ON system_logs(created_at)")

	// Salary-process optimization indexes
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_advance_salaries_monthly ON advance_salaries(company_id, deduction_year, deduction_month, status) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_temporary_shifts_emp_date ON temporary_shifts(employee_id, date) WHERE deleted_at IS NULL AND status = 'active'")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_ot_early_exit_lookup ON ot_early_exit_deductions(company_id, month, year, employee_id) WHERE deleted_at IS NULL")

	// Daily Process — high-performance indexes & idempotency
	// Unique partial index prevents duplicate attendance; enables UPSERT ON CONFLICT
	silentDB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_attendances_employee_date ON attendances(employee_id, date) WHERE deleted_at IS NULL")
	// Leave-lock design: attendance.status = 'on_leave' (Lv) is authoritative
	// only via daily process (attendance_processor.SyncLeaveLockedStatus) and
	// revertible only via DeleteLeave (ClearOnLeaveStatus). No other API may
	// write Lv — enforced at handler (403) and repository guard.
	// Punch lookup — primary query for daily process is badge_number + punch_time range
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_data_logs_badge_punch_time ON data_logs(badge_number, punch_time) WHERE deleted_at IS NULL")
	// Separation lookup for eligibility (batch by employee_id)
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_separations_employee ON separations(employee_id) WHERE deleted_at IS NULL")
	// Missing attendance — bulk fetch by company+date
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_missing_attendance_company_date ON missing_attendances(company_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_missing_attendance_emp_date ON missing_attendances(employee_id, date) WHERE deleted_at IS NULL")
	// Leave range scan
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_leaves_employee_date ON leaves(employee_id, from_date, to_date) WHERE deleted_at IS NULL")
	// Attendance bulk fetch by company+date range
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_attendances_company_date_range ON attendances(company_id, date, employee_id) WHERE deleted_at IS NULL")

	// Salary increment promotion hierarchy (dept→sec→desig→line) — fast lookup for promo target validation
	silentDB.Exec("ALTER TABLE salary_increments ADD COLUMN IF NOT EXISTS promo_department_id uuid")
	silentDB.Exec("ALTER TABLE salary_increments ADD COLUMN IF NOT EXISTS promo_section_id uuid")
	silentDB.Exec("ALTER TABLE salary_increments ADD COLUMN IF NOT EXISTS promo_line_id uuid")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_salary_increments_promo_dept ON salary_increments(promo_department_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_salary_increments_promo_sec ON salary_increments(promo_section_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_salary_increments_promo_line ON salary_increments(promo_line_id) WHERE deleted_at IS NULL")

	// Salary increment effective-date lookup — supports temporal salary (effective month → next increment)
	// Used by salary ProcessMonth to resolve gross as of month end: latest approved increment where effective_date <= endStr
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_salary_increments_effective_lookup ON salary_increments(employee_id, effective_date) WHERE status = 'approved' AND deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_salary_increments_company_effective ON salary_increments(company_id, effective_date) WHERE status = 'approved' AND deleted_at IS NULL")

	// Tune connection pool — prevents salary process blocking on single connection (was 1m30s due to pool starvation)
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(30 * 60 * 1_000_000_000) // 30m
		sqlDB.SetConnMaxIdleTime(5 * 60 * 1_000_000_000)  // 5m
	}

	DB = db
	fmt.Println("Database connected successfully")
}
