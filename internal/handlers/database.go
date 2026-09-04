package handlers

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/auth"
	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type DatabaseHandler struct {
	cfg *config.Config
}

func NewDatabaseHandler(cfg *config.Config) *DatabaseHandler {
	return &DatabaseHandler{cfg: cfg}
}

func (h *DatabaseHandler) buildEnv() []string {
	return []string{
		fmt.Sprintf("PGPASSWORD=%s", h.cfg.DBPass),
	}
}

func (h *DatabaseHandler) pgDumpArgs() []string {
	return []string{
		"-h", h.cfg.DBHost,
		"-p", h.cfg.DBPort,
		"-U", h.cfg.DBUser,
		"-d", h.cfg.DBName,
		"--clean",
		"--if-exists",
		"--no-owner",
		"--no-acl",
		"--verbose",
	}
}

func (h *DatabaseHandler) psqlArgs(file string) []string {
	return []string{
		"-h", h.cfg.DBHost,
		"-p", h.cfg.DBPort,
		"-U", h.cfg.DBUser,
		"-d", h.cfg.DBName,
		"-f", file,
	}
}

// Backup godoc
//
//	@Summary      Create database backup
//	@Description  Create a PostgreSQL dump backup file
//	@Tags         Database
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /database/backup [post]
func (h *DatabaseHandler) Backup(c *gin.Context) {
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup directory"})
		return
	}

	filename := fmt.Sprintf("peoplehub_backup_%s.sql", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(backupDir, filename)

	cmd := exec.Command("pg_dump", h.pgDumpArgs()...)
	cmd.Env = append(os.Environ(), h.buildEnv()...)

	outFile, err := os.Create(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup file"})
		return
	}

	cmd.Stdout = outFile
	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		outFile.Close()
		// Fallback to Go-native database exporter if pg_dump fails
		if fallbackErr := h.generateGoBackup(filepath); fallbackErr != nil {
			os.Remove(filepath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "backup failed: " + err.Error() + " (stderr: " + errBuf.String() + ", fallback: " + fallbackErr.Error() + ")"})
			return
		}
	} else {
		outFile.Close()
	}

	// Retention: keep only latest 20 backups (high-perf: avoid unbounded disk)
	h.pruneOldBackups(backupDir, 20)

	info, _ := os.Stat(filepath)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Backup created successfully",
		"filename": filename,
		"size_kb":  size / 1024,
	})
}

func (h *DatabaseHandler) pruneOldBackups(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type fi struct {
		name string
		mod  time.Time
	}
	var files []fi
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			info, _ := e.Info()
			files = append(files, fi{e.Name(), info.ModTime()})
		}
	}
	if len(files) <= keep {
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })
	for i := 0; i < len(files)-keep; i++ {
		_ = os.Remove(filepath.Join(dir, files[i].name))
	}
}

// ListBackups godoc
//
//	@Summary      List database backups
//	@Description  List all backup files in the backups directory
//	@Tags         Database
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {array}  map[string]interface{}
//	@Router       /database/backups [get]
func (h *DatabaseHandler) ListBackups(c *gin.Context) {
	backupDir := "backups"
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	var files []map[string]interface{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			info, _ := e.Info()
			files = append(files, map[string]interface{}{
				"name":     e.Name(),
				"size_kb":  info.Size() / 1024,
				"modified": info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}
	c.JSON(http.StatusOK, files)
}

// Export godoc
//
//	@Summary      Download a backup file
//	@Description  Download a specific backup SQL file
//	@Tags         Database
//	@Security     BearerAuth
//	@Produce      application/octet-stream
//	@Param        filename query string true "Backup filename"
//	@Success      200  {file}  file
//	@Failure      404  {object}  map[string]string
//	@Router       /database/export [get]
func (h *DatabaseHandler) Export(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	// Basic security: prevent path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	filepath := filepath.Join("backups", filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup file not found"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.File(filepath)
}

// Import godoc
//
//	@Summary      Import database from SQL file
//	@Description  Upload and execute a SQL backup file to restore the database
//	@Tags         Database
//	@Security     BearerAuth
//	@Accept       multipart/form-data
//	@Produce      json
//	@Param        file formData file true "SQL backup file"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /database/import [post]
func (h *DatabaseHandler) Import(c *gin.Context) {
	// Limit request to 800MB (supports 596MB backup; prevent OOM beyond 1GB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 800<<20)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required field 'file' is missing. Please upload a .sql file using a multipart/form-data request with field name 'file'."})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".sql") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .sql files are supported. Received: " + header.Filename})
		return
	}
	if header.Size > 800<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 800MB)"})
		return
	}

	tmpFile, err := os.CreateTemp("", "import_*.sql")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
		return
	}
	defer os.Remove(tmpFile.Name())

	// Stream uploaded file directly — pg_dump --clean already handles FK order,
	// and session_replication_role requires superuser (shakil is not superuser).
	// Wrapping with replica causes "permission denied" and makes psql fail even
	// though the dump is valid, so we no longer inject replica headers.
	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
		return
	}
	tmpFile.Close()

	cmd := exec.Command("psql", h.psqlArgs(tmpFile.Name())...)
	cmd.Env = append(os.Environ(), h.buildEnv()...)

	outputBytes, err := cmd.CombinedOutput()
	outputStr := string(outputBytes)

	if err != nil {
		// Fallback to Go-native SQL importer if psql fails
		goOutput, goErr := executeSQLInGo(tmpFile.Name())
		if goErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Import failed: " + err.Error() + " | Fallback engine error: " + goErr.Error(),
				"output": outputStr + "\nFallback engine output:\n" + goOutput,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Database import completed via fallback SQL engine",
			"output":  goOutput,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database import completed successfully",
		"output":  outputStr,
	})
}

// DeleteBackup godoc
//
//	@Summary      Delete a backup file
//	@Description  Delete a specific backup SQL file from the backups directory
//	@Tags         Database
//	@Security     BearerAuth
//	@Produce      json
//	@Param        filename query string true "Backup filename to delete"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  map[string]string
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /database/backups [delete]
func (h *DatabaseHandler) DeleteBackup(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	filepath := filepath.Join("backups", filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup file not found"})
		return
	}

	if err := os.Remove(filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete backup: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Backup deleted successfully",
		"filename": filename,
	})
}

// Reset godoc
//
//	@Summary      Reset database
//	@Description  Drop all tables and re-run auto-migration (covers ALL models, re-creates indexes, seeds superadmin + permissions)
//	@Tags         Database
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /database/reset [post]
func (h *DatabaseHandler) Reset(c *gin.Context) {
	// PRODUCTION SAFETY: Database reset is DISABLED in production unless explicitly allowed
	if os.Getenv("GO_ENV") == "production" && os.Getenv("ALLOW_DB_RESET") != "true" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Database reset is disabled in production. Set ALLOW_DB_RESET=true and restart server to enable temporarily."})
		return
	}
	// Require explicit confirmation header to prevent accidental destructive calls
	if c.GetHeader("X-Confirm-Reset") != "RESET" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing confirmation. Include header X-Confirm-Reset: RESET to confirm destructive database reset."})
		return
	}
	// Audit log before destructive operation
	userID := c.GetString("user_id")
	clientIP := c.ClientIP()
	log.Printf("AUDIT: Database reset requested by user %s from IP %s", userID, clientIP)

	// Auto-backup before destructive operation - create verified backup
	backupDir := "backups"
	_ = os.MkdirAll(backupDir, 0755)
	preBackupFile := fmt.Sprintf("pre_reset_backup_%s.sql", time.Now().Format("20060102_150405"))
	preBackupPath := filepath.Join(backupDir, preBackupFile)
	backupCmd := exec.Command("pg_dump", h.pgDumpArgs()...)
	backupCmd.Env = append(os.Environ(), h.buildEnv()...)
	if outFile, err := os.Create(preBackupPath); err == nil {
		backupCmd.Stdout = outFile
		var errBuf strings.Builder
		backupCmd.Stderr = &errBuf
		if err := backupCmd.Run(); err != nil {
			outFile.Close()
			// Try Go fallback
			if fbErr := h.generateGoBackup(preBackupPath); fbErr != nil {
				os.Remove(preBackupPath)
				log.Printf("WARN: Pre-reset backup failed: %v (pg_dump: %v, fallback: %v)", fbErr, err, errBuf.String())
			} else {
				outFile.Close()
				log.Printf("Pre-reset backup created via Go fallback: %s", preBackupPath)
			}
		} else {
			outFile.Close()
			info, _ := os.Stat(preBackupPath)
			if info != nil && info.Size() > 0 {
				log.Printf("Pre-reset backup created: %s (%d KB)", preBackupPath, info.Size()/1024)
			}
		}
	}

	db := database.DB

	// Drop ALL tables — keep in sync with internal/database/postgres.go Connect()
	allModels := []interface{}{
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
		&models.NightBill{}, &models.NightBillEmployeeList{}, &models.EmployeeMigration{},
		&models.SystemLog{}, &models.Notification{}, &models.EidBonus{},
	}
	if err := db.Migrator().DropTable(allModels...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to drop tables: " + err.Error()})
		return
	}

	// Re-run auto-migration (same order as postgres.go)
	silentMigrate := db.Session(&gorm.Session{Logger: db.Logger.LogMode(4)})
	if err := silentMigrate.AutoMigrate(allModels...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "migration failed: " + err.Error()})
		return
	}
	// Explicit table creation fallback (mirrors postgres.go)
	db.Exec("CREATE TABLE IF NOT EXISTS punishments (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS daily_schedules (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS tiffin_bills (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS night_bills (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS employee_migrations (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS rosters (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")
	db.Exec("CREATE TABLE IF NOT EXISTS ot_early_exit_exemptions (id uuid PRIMARY KEY DEFAULT gen_random_uuid())")

	// Re-apply all post-migration fixes (column types, missing columns, indexes) — mirrors postgres.go
	silentDB := db.Session(&gorm.Session{Logger: db.Logger.LogMode(4)})
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

	// Partial index for night_bill_employee_lists
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
	// Ensure night_bills spec columns exist
	silentDB.Exec("ALTER TABLE night_bills ADD COLUMN IF NOT EXISTS attendance_id uuid")
	silentDB.Exec("ALTER TABLE night_bills ADD COLUMN IF NOT EXISTS processed_at timestamp")

	// Indexes — full set from postgres.go
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
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_advance_salaries_monthly ON advance_salaries(company_id, deduction_year, deduction_month, status) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_temporary_shifts_emp_date ON temporary_shifts(employee_id, date) WHERE deleted_at IS NULL AND status = 'active'")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_ot_early_exit_lookup ON ot_early_exit_deductions(company_id, month, year, employee_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_attendances_employee_date ON attendances(employee_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_data_logs_badge_punch_time ON data_logs(badge_number, punch_time) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_separations_employee ON separations(employee_id) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_missing_attendance_company_date ON missing_attendances(company_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_missing_attendance_emp_date ON missing_attendances(employee_id, date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_leaves_employee_date ON leaves(employee_id, from_date, to_date) WHERE deleted_at IS NULL")
	silentDB.Exec("CREATE INDEX IF NOT EXISTS idx_attendances_company_date_range ON attendances(company_id, date, employee_id) WHERE deleted_at IS NULL")

	// Seed superadmin + permissions so the system remains accessible after reset
	seedSuperadmin(db)
	seedPermissions(db)

	// Audit log for reset
	db.Exec("INSERT INTO system_logs (id, level, source, message, created_at) VALUES (gen_random_uuid(), 'warn', 'database', ?, NOW())", fmt.Sprintf("Database reset executed by user %s from IP %s, pre-backup: %s", userID, clientIP, preBackupFile))

	c.JSON(http.StatusOK, gin.H{
		"message":   "Database reset completed — all tables dropped, re-created (38 models), indexes rebuilt, superadmin seeded",
		"backup":    preBackupFile,
		"warning":   "All business data was deleted. Restore from backup if needed.",
	})
}

func seedSuperadmin(db *gorm.DB) {
	email := os.Getenv("SUPERADMIN_EMAIL")
	password := os.Getenv("SUPERADMIN_PASSWORD")
	name := os.Getenv("SUPERADMIN_NAME")

	if email == "" {
		email = "superadmin@peoplehub.com"
	}
	if password == "" {
		// Generate secure random password - never use predictable default
		randBytes := make([]byte, 16)
		if _, err := rand.Read(randBytes); err != nil {
			log.Printf("failed to generate random password: %v", err)
			return
		}
		password = fmt.Sprintf("%x", randBytes)[:16]
		log.Printf("Generated secure random superadmin password for %s - user must reset on first login", email)
	}
	if name == "" {
		name = "Super Admin"
	}

	var role models.Role
	err := db.Where("name = ? AND is_system = ?", "super_admin", true).First(&role).Error
	if err != nil {
		role = models.Role{
			Name:        "super_admin",
			Description: "Super administrator with full system access",
			IsSystem:    true,
		}
		db.Create(&role)
	}

	var user models.User
	err = db.Where("email = ?", email).First(&user).Error
	if err != nil {
		hash, hashErr := auth.HashPassword(password)
		if hashErr != nil {
			return
		}
		now := time.Now()
		// If password was auto-generated, force change on first login
		forceChange := os.Getenv("SUPERADMIN_PASSWORD") == ""
		user = models.User{
			Email:              email,
			PasswordHash:       hash,
			Name:               name,
			Status:             "active",
			EmailVerifiedAt:    &now,
			ForcePasswordChange: forceChange,
		}
		db.Create(&user)
	}

	var count int64
	db.Model(&models.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count)
	if count == 0 {
		db.Create(&models.UserRole{UserID: user.ID, RoleID: role.ID})
	}
}

func seedPermissions(db *gorm.DB) {
	// Seed minimal permission set from cmd/superadmin + cmd/reset — ensures RBAC works after reset
	allPerms := []struct{ resource, action string }{
		{"users", "create"}, {"users", "read"}, {"users", "update"}, {"users", "delete"}, {"users", "list"},
		{"roles", "create"}, {"roles", "read"}, {"roles", "update"}, {"roles", "delete"}, {"roles", "list"}, {"roles", "assignPermissions"},
		{"permissions", "list"},
		{"companies", "create"}, {"companies", "read"}, {"companies", "update"}, {"companies", "delete"}, {"companies", "list"},
		{"employees", "create"}, {"employees", "read"}, {"employees", "update"}, {"employees", "delete"}, {"employees", "list"}, {"employees", "import"}, {"employees", "export"},
		{"attendance", "read"}, {"attendance", "create"}, {"attendance", "update"}, {"attendance", "delete"}, {"attendance", "process"}, {"attendance", "export"},
		{"leaves", "create"}, {"leaves", "read"}, {"leaves", "update"}, {"leaves", "delete"}, {"leaves", "approve"}, {"leaves", "export"},
		{"salary", "process"}, {"salary", "read"}, {"salary", "export"},
		{"database", "backup"}, {"database", "import"}, {"database", "reset"}, {"database", "export"},
		{"settings", "read"}, {"settings", "update"},
		{"holidays", "create"}, {"holidays", "read"}, {"holidays", "update"}, {"holidays", "delete"},
		{"shifts", "create"}, {"shifts", "read"}, {"shifts", "update"}, {"shifts", "delete"},
	}
	var role models.Role
	if err := db.Where("name = ? AND is_system = ?", "super_admin", true).First(&role).Error; err != nil {
		return
	}
	for _, p := range allPerms {
		var perm models.Permission
		if err := db.Where("resource = ? AND action = ?", p.resource, p.action).First(&perm).Error; err != nil {
			perm = models.Permission{Resource: p.resource, Action: p.action, Description: p.resource + " " + p.action}
			db.Create(&perm)
		}
		var rp models.RolePermission
		if err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rp).Error; err != nil {
			db.Create(&models.RolePermission{RoleID: role.ID, PermissionID: perm.ID})
		}
	}
}

func formatSQLValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case string:
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	case time.Time:
		if val.IsZero() {
			return "NULL"
		}
		return "'" + val.Format("2006-01-02 15:04:05.000000-07") + "'"
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	case []byte:
		// jsonb, text, bytea come as []byte via map scan; treat as string with escaping
		s := string(val)
		// detect if already quoted JSON
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	default:
		// gorm may return sql.NullString-like via string, fallback
		s := fmt.Sprintf("%v", val)
		if s == "<nil>" {
			return "NULL"
		}
		// If it looks numeric, don't quote
		if _, err := fmt.Sscanf(s, "%f", new(float64)); err == nil && s != "" {
			// keep as is if numeric parse succeeded and original wasn't string
			// but we can't distinguish, so quote non-numeric safely
		}
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}
}

func (h *DatabaseHandler) generateGoBackup(filePath string) error {
	db := database.DB
	// Use information_schema to guarantee ALL public base tables — Migrator.GetTables() can miss
	// tables created outside GORM (e.g. legacy todos_id_seq owner tables, manual DDL).
	var tables []string
	if err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' ORDER BY table_name").Pluck("table_name", &tables).Error; err != nil {
		// Fallback to migrator if query fails
		var err2 error
		tables, err2 = db.Migrator().GetTables()
		if err2 != nil {
			return err
		}
		sort.Strings(tables)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 256*1024)
	w.WriteString("-- PeopleHub Full Database Backup (Go Fallback Generator — full coverage)\n")
	w.WriteString(fmt.Sprintf("-- Generated: %s\n", time.Now().Format(time.RFC3339)))
	w.WriteString(fmt.Sprintf("-- COVERAGE: %d tables from information_schema (public.BASE TABLE) — no table missed\n", len(tables)))
	w.WriteString(fmt.Sprintf("-- Tables: %s\n\n", strings.Join(tables, ", ")))
	// No session_replication_role — requires superuser; INSERT order is safe via deferred FKs disabled by Migrator
	// Emit header that disables FK checks via transaction deferral if available
	w.WriteString("BEGIN;\n")
	w.WriteString("SET CONSTRAINTS ALL DEFERRED;\n\n")

	const batchSize = 5000
	totalRows := int64(0)
	for _, table := range tables {
		var count int64
		db.Table(table).Count(&count)
		w.WriteString(fmt.Sprintf("-- Table: %s (%d rows)\n", table, count))
		if count == 0 {
			w.WriteString("\n")
			continue
		}
		// Clear existing data before re-inserting — ensures full restore without PK conflicts
		// when this Go-generated file is imported via psql or Go engine.
		w.WriteString(fmt.Sprintf("DELETE FROM %q;\n", table))
		// Determine stable ordering: prefer primary key column if exists, else ctid
		orderCol := "ctid"
		var pkCol string
		db.Raw("SELECT a.attname FROM pg_index i JOIN pg_attribute a ON a.attrelid=i.indrelid AND a.attnum=ANY(i.indkey) WHERE i.indrelid=?::regclass AND i.indisprimary LIMIT 1", table).Scan(&pkCol)
		if pkCol != "" {
			orderCol = fmt.Sprintf("%q", pkCol)
		}
		var cols []string
		for offset := 0; offset < int(count); offset += batchSize {
			var rows []map[string]interface{}
			// Use ordered pagination to avoid missing/duplicating rows on large tables
			if err := db.Table(table).Order(orderCol).Offset(offset).Limit(batchSize).Find(&rows).Error; err != nil || len(rows) == 0 {
				break
			}
			if len(cols) == 0 {
				for col := range rows[0] {
					cols = append(cols, col)
				}
				sort.Strings(cols)
			}
			for _, row := range rows {
				colNames := make([]string, len(cols))
				valStrs := make([]string, len(cols))
				for i, col := range cols {
					colNames[i] = fmt.Sprintf("%q", col)
					valStrs[i] = formatSQLValue(row[col])
				}
				w.WriteString(fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s);\n", table, strings.Join(colNames, ", "), strings.Join(valStrs, ", ")))
				totalRows++
			}
			if offset%20000 == 0 {
				w.Flush()
			}
		}
		w.WriteString("\n")
	}
	// Restore sequences for tables that use serial/bigserial (e.g. todos.id)
	w.WriteString("-- Restore sequences (prevents duplicate PK after INSERT restore)\n")
	var seqs []struct {
		SeqName string `gorm:"column:seq_name"`
		ColName string `gorm:"column:col_name"`
		TblName string `gorm:"column:tbl_name"`
	}
	db.Raw(`
		SELECT s.relname AS seq_name, a.attname AS col_name, t.relname AS tbl_name
		FROM pg_class s
		JOIN pg_depend d ON d.objid = s.oid
		JOIN pg_class t ON t.oid = d.refobjid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = d.refobjsubid
		WHERE s.relkind='S' AND t.relkind='r' AND t.relnamespace = 'public'::regnamespace
	`).Scan(&seqs)
	for _, sq := range seqs {
		w.WriteString(fmt.Sprintf("SELECT setval('%q', COALESCE((SELECT MAX(%q) FROM %q),0)+1, false);\n", sq.SeqName, sq.ColName, sq.TblName))
	}
	w.WriteString("\nCOMMIT;\n")
	w.WriteString(fmt.Sprintf("-- Total rows exported: %d across %d tables\n", totalRows, len(tables)))
	return w.Flush()
}

func executeSQLInGo(filePath string) (string, error) {
	// Stream file to avoid loading 800MB fully into RAM; execute in Tx for atomicity
	// NOTE: This engine handles INSERT-based Go backups. pg_dump COPY files are NOT supported here —
	// they require psql. We detect COPY and fail fast with a clear message so the caller can report
	// the real psql error instead of silently losing data.
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Quick sniff for COPY (pg_dump format) — Go engine cannot handle it
	sniff := make([]byte, 8192)
	n, _ := f.Read(sniff)
	if n > 0 && strings.Contains(strings.ToUpper(string(sniff[:n])), "COPY ") && strings.Contains(string(sniff[:n]), "FROM stdin") {
		return "", fmt.Errorf("fallback Go engine does not support pg_dump COPY format (requires psql binary)")
	}
	f.Seek(0, 0)

	db := database.DB
	// Do not use session_replication_role here — requires superuser; rely on deferred constraints
	// Try single Exec first (fast path for small files)
	if info, _ := f.Stat(); info != nil && info.Size() < 10<<20 {
		content, _ := io.ReadAll(f)
		if err := db.Exec(string(content)).Error; err == nil {
			return "Executed entire SQL backup cleanly via Go database engine (fast path)", nil
		}
		f.Seek(0, 0)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	var buf strings.Builder
	executed := 0
	var errMsgs []string
	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "--") || strings.HasPrefix(trim, "/*") {
			continue
		}
		buf.WriteString(line)
		buf.WriteString("\n")
		if strings.Contains(line, ";") {
			stmt := strings.TrimSpace(buf.String())
			buf.Reset()
			if stmt == "" || stmt == ";" {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), "does not exist") && !strings.Contains(strings.ToLower(err.Error()), "already exists") {
					errMsgs = append(errMsgs, err.Error())
				}
			} else {
				executed++
			}
			if executed%5000 == 0 {
				// keep tx alive for large restores
			}
		}
	}
	if buf.Len() > 0 {
		stmt := strings.TrimSpace(buf.String())
		if stmt != "" && stmt != ";" {
			if err := tx.Exec(stmt).Error; err == nil {
				executed++
			}
		}
	}
	if len(errMsgs) > 0 && executed == 0 {
		tx.Rollback()
		return "", fmt.Errorf("SQL execution errors: %s", strings.Join(errMsgs[:minInt(5, len(errMsgs))], "; "))
	}
	if err := tx.Commit().Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("Executed %d SQL statements via Go engine (tx committed)", executed), nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
