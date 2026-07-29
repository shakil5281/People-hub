package service

import (
	"log"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"gorm.io/gorm"
)

type NotificationChecker struct {
	db               *gorm.DB
	notificationRepo *repository.NotificationRepository
	employeeRepo     *repository.EmployeeRepository
}

func NewNotificationChecker(db *gorm.DB, notificationRepo *repository.NotificationRepository, employeeRepo *repository.EmployeeRepository) *NotificationChecker {
	return &NotificationChecker{
		db:               db,
		notificationRepo: notificationRepo,
		employeeRepo:     employeeRepo,
	}
}

func (nc *NotificationChecker) Start(interval time.Duration) {
	go func() {
		nc.runChecks()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			nc.runChecks()
		}
	}()
	log.Printf("Notification checker started — running every %v", interval)
}

func (nc *NotificationChecker) runChecks() {
	log.Println("Notification checker: running checks...")

	adminIDs := nc.getAdminUserIDs()
	if len(adminIDs) == 0 {
		log.Println("Notification checker: no admin users found, skipping")
		return
	}

	today := time.Now().Format("2006-01-02")

	nc.checkMissingAttendance(today, adminIDs)
	nc.checkLateArrivals(today, adminIDs)
	nc.checkApprovedLeaveMismatch(today, adminIDs)
	nc.checkStalePendingLeaves(adminIDs)
}

func (nc *NotificationChecker) getAdminUserIDs() []string {
	var userIDs []string
	nc.db.Model(&models.User{}).
		Joins("INNER JOIN user_roles ON user_roles.user_id = users.id").
		Joins("INNER JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ? AND users.deleted_at IS NULL", "admin").
		Pluck("users.id", &userIDs)
	return userIDs
}

func (nc *NotificationChecker) checkMissingAttendance(today string, adminIDs []string) {
	var employees []struct {
		EmployeeID string
		NameEn     string
	}
	nc.db.Model(&models.Employee{}).
		Select("employees.employee_id, employees.name_en").
		Where("employees.status = ? AND employees.deleted_at IS NULL", "active").
		Where("NOT EXISTS (SELECT 1 FROM attendances WHERE attendances.employee_id = employees.employee_id AND attendances.date = ?)", today).
		Find(&employees)

	for _, emp := range employees {
		for _, uid := range adminIDs {
			title := "Missing Attendance"
			message := emp.NameEn + " (" + emp.EmployeeID + ") has no attendance record for today."
			nc.createNotification(uid, title, message, "warning", "missing_attendance")
		}
	}
}

func (nc *NotificationChecker) checkLateArrivals(today string, adminIDs []string) {
	var lateRecords []struct {
		EmployeeID   string
		NameEn       string
		LateMinutes  int
	}
	nc.db.Table("attendances").
		Select("attendances.employee_id, employees.name_en, attendances.late_minutes").
		Joins("INNER JOIN employees ON employees.employee_id = attendances.employee_id").
		Where("attendances.date = ? AND attendances.status = ?", today, "late").
		Where("attendances.late_minutes > 0").
		Scan(&lateRecords)

	for _, rec := range lateRecords {
		for _, uid := range adminIDs {
			title := "Late Arrival Alert"
			message := rec.NameEn + " (" + rec.EmployeeID + ") was late by " + formatLateMinutes(rec.LateMinutes) + " today."
			nc.createNotification(uid, title, message, "warning", "late_arrival")
		}
	}
}

func (nc *NotificationChecker) checkApprovedLeaveMismatch(today string, adminIDs []string) {
	var mismatches []struct {
		EmployeeID string
		NameEn     string
	}
	nc.db.Model(&models.Leave{}).
		Select("leaves.employee_id, employees.name_en").
		Joins("INNER JOIN employees ON employees.employee_id = leaves.employee_id").
		Where("leaves.status = ? AND leaves.from_date <= ? AND leaves.to_date >= ?", "approved", today, today).
		Where("NOT EXISTS (SELECT 1 FROM attendances WHERE attendances.employee_id = leaves.employee_id AND attendances.date = ? AND attendances.status = ?)", today, "on_leave").
		Find(&mismatches)

	for _, m := range mismatches {
		for _, uid := range adminIDs {
			title := "Leave Attendance Mismatch"
			message := m.NameEn + " (" + m.EmployeeID + ") has an approved leave today but attendance is not marked as on_leave."
			nc.createNotification(uid, title, message, "error", "leave_mismatch")
		}
	}
}

func (nc *NotificationChecker) checkStalePendingLeaves(adminIDs []string) {
	threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02")

	var staleLeaves []struct {
		EmployeeID string
		NameEn     string
		FromDate   string
		ToDate     string
	}
	nc.db.Model(&models.Leave{}).
		Select("leaves.employee_id, employees.name_en, leaves.from_date, leaves.to_date").
		Joins("INNER JOIN employees ON employees.employee_id = leaves.employee_id").
		Where("leaves.status = ? AND leaves.created_at <= ?", "pending", threeDaysAgo+"T00:00:00").
		Find(&staleLeaves)

	for _, l := range staleLeaves {
		for _, uid := range adminIDs {
			title := "Stale Pending Leave"
			message := l.NameEn + " (" + l.EmployeeID + ") has a leave request pending since " + l.FromDate + "."
			nc.createNotification(uid, title, message, "info", "stale_leave")
		}
	}
}

func (nc *NotificationChecker) createNotification(userID, title, message, notifType, metadata string) {
	notif := &models.Notification{
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     notifType,
		Metadata: &metadata,
	}
	if err := nc.notificationRepo.Create(notif); err != nil {
		log.Printf("Notification checker: failed to create notification: %v", err)
	}
}

func formatLateMinutes(m int) string {
	if m < 60 {
		return formatInt(m) + " min"
	}
	return formatInt(m/60) + "h " + formatInt(m%60) + "min"
}

func formatInt(n int) string {
	if n < 10 {
		return string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
