package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(filepath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "backup failed: " + err.Error()})
		return
	}

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
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required field 'file' is missing. Please upload a .sql file using a multipart/form-data request with field name 'file'."})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".sql") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .sql files are supported. Received: " + header.Filename})
		return
	}

	tmpFile, err := os.CreateTemp("", "import_*.sql")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file"})
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
		return
	}
	tmpFile.Close()

	cmd := exec.Command("psql", h.psqlArgs(tmpFile.Name())...)
	cmd.Env = append(os.Environ(), h.buildEnv()...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "import failed: " + err.Error(),
			"output": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database import completed successfully",
		"output":  string(output),
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
//	@Description  Drop all tables and re-run auto-migration
//	@Tags         Database
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /database/reset [post]
func (h *DatabaseHandler) Reset(c *gin.Context) {
	db := database.DB

	// Drop all tables
	if err := db.Migrator().DropTable(
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
		&models.EidBonus{}, &models.NightBill{},
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to drop tables: " + err.Error()})
		return
	}

	// Re-run auto-migration
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
		&models.EidBonus{}, &models.NightBill{},
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "migration failed: " + err.Error()})
		return
	}

	// Re-apply all post-migration fixes (column types, missing columns, indexes)
	alterCol := func(table, col string) {
		db.Exec("ALTER TABLE " + table + " ALTER COLUMN " + col + " TYPE varchar(50) USING " + col + "::varchar(50)")
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
	alterCol("night_bills", "employee_id")
	alterCol("eid_bonuses", "employee_id")
	alterCol("id_cards", "employee_id")
	alterCol("separations", "employee_id")

	db.Exec("ALTER TABLE separations ADD COLUMN IF NOT EXISTS company_id uuid")
	db.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS nid varchar(50)")
	db.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS present_post_office varchar(100)")
	db.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS present_post_code varchar(20)")
	db.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS permanent_post_office varchar(100)")
	db.Exec("ALTER TABLE employees ADD COLUMN IF NOT EXISTS permanent_post_code varchar(20)")
	db.Exec("ALTER TABLE requirements ADD COLUMN IF NOT EXISTS section_id uuid")
	db.Exec("ALTER TABLE requirements ADD COLUMN IF NOT EXISTS designation_id uuid")
	db.Exec("ALTER TABLE requirements ADD COLUMN IF NOT EXISTS group_type varchar(20) DEFAULT 'Worker'")

	db.Exec(`
		ALTER TABLE attendances ALTER COLUMN check_in TYPE timestamp USING CASE
			WHEN check_in IS NOT NULL AND length(check_in::text) <= 5 THEN (date || ' ' || check_in)::timestamp
			WHEN check_in IS NOT NULL THEN check_in::timestamp
			ELSE NULL
		END
	`)
	db.Exec(`
		ALTER TABLE attendances ALTER COLUMN check_out TYPE timestamp USING CASE
			WHEN check_out IS NOT NULL AND length(check_out::text) <= 5 THEN (date || ' ' || check_out)::timestamp
			WHEN check_out IS NOT NULL THEN check_out::timestamp
			ELSE NULL
		END
	`)

	db.Exec("CREATE INDEX IF NOT EXISTS idx_salaries_company_month_year ON salaries(company_id, year, month)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_employees_company_status ON employees(company_id, status) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_leave_allocations_emp_year ON leave_allocations(employee_id, year)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_temporary_shifts_company_date ON temporary_shifts(company_id, date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_data_logs_date_processed ON data_logs(date, processed) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_attendances_date_status ON attendances(date, status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_leaves_status_dates ON leaves(status, from_date, to_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_level ON system_logs(level)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_source ON system_logs(source)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_user ON system_logs(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_system_logs_created ON system_logs(created_at)")

	// Seed superadmin user so the system remains accessible after reset
	seedSuperadmin(db)

	c.JSON(http.StatusOK, gin.H{
		"message": "Database reset completed — all tables dropped and re-created",
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
		password = "superadmin1234"
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
		user = models.User{
			Email:              email,
			PasswordHash:       hash,
			Name:               name,
			Status:             "active",
			EmailVerifiedAt:    &now,
			ForcePasswordChange: false,
		}
		db.Create(&user)
	}

	var count int64
	db.Model(&models.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count)
	if count == 0 {
		db.Create(&models.UserRole{UserID: user.ID, RoleID: role.ID})
	}
}
