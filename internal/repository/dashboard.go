package repository

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type GenderCount struct {
	Gender string `json:"gender"`
	Count  int64  `json:"count"`
}

type DeptCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type MonthlyAttendanceCount struct {
	Month   string `json:"month"`
	Present int64  `json:"present"`
	Absent  int64  `json:"absent"`
	Late    int64  `json:"late"`
}

type DailyAttendanceCount struct {
	Date    string `json:"date"`
	Present int64  `json:"present"`
	Absent  int64  `json:"absent"`
	Late    int64  `json:"late"`
}

type ActivityItem struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) CountTotalEmployees() int64 {
	var count int64
	r.db.Model(&models.Employee{}).Where("deleted_at IS NULL").Count(&count)
	return count
}

func (r *DashboardRepository) CountActiveEmployees() int64 {
	var count int64
	r.db.Model(&models.Employee{}).Where("status = ? AND deleted_at IS NULL", "active").Count(&count)
	return count
}

func (r *DashboardRepository) CountDepartments() int64 {
	var count int64
	r.db.Model(&models.Department{}).Where("deleted_at IS NULL").Count(&count)
	return count
}

func (r *DashboardRepository) CountSections() int64 {
	var count int64
	r.db.Model(&models.Section{}).Where("deleted_at IS NULL").Count(&count)
	return count
}

func (r *DashboardRepository) CountTodayAttendance(today string) int64 {
	var count int64
	r.db.Model(&models.Attendance{}).Where("date = ? AND status IN ?", today, []string{"present", "late"}).Count(&count)
	return count
}

func (r *DashboardRepository) CountTodayLogs(today string) int64 {
	var count int64
	r.db.Model(&models.DataLog{}).Where("date = ?", today).Count(&count)
	return count
}

func (r *DashboardRepository) CountPendingLeaves() int64 {
	var count int64
	r.db.Model(&models.Leave{}).Where("status = ?", "pending").Count(&count)
	return count
}

func (r *DashboardRepository) CountNewHiresMonth(firstOfMonth string) int64 {
	var count int64
	r.db.Model(&models.Employee{}).Where("joining_date >= ? AND deleted_at IS NULL", firstOfMonth).Count(&count)
	return count
}

func (r *DashboardRepository) CountSeparationsMonth(firstOfMonth string) int64 {
	var count int64
	r.db.Raw("SELECT COUNT(*) FROM separations WHERE date >= ? AND deleted_at IS NULL", firstOfMonth).Scan(&count)
	return count
}

func (r *DashboardRepository) GenderDistribution() []GenderCount {
	var result []GenderCount
	r.db.Model(&models.Employee{}).
		Select("gender, COUNT(*) as count").
		Where("deleted_at IS NULL AND gender != ''").
		Group("gender").
		Find(&result)
	return result
}

func (r *DashboardRepository) DepartmentCounts() []DeptCount {
	var result []DeptCount
	r.db.Model(&models.Employee{}).
		Select("departments.name, COUNT(*) as count").
		Joins("LEFT JOIN departments ON departments.id = employees.department_id").
		Where("employees.deleted_at IS NULL AND departments.deleted_at IS NULL").
		Group("departments.name").
		Order("count DESC").
		Limit(10).
		Find(&result)
	return result
}

func (r *DashboardRepository) MonthlyAttendanceTrend(monthsBack int) []MonthlyAttendanceCount {
	var result []MonthlyAttendanceCount
	for i := monthsBack; i >= 0; i-- {
		m := time.Now().AddDate(0, -i, 0)
		monthStr := m.Format("2006-01")
		start := monthStr + "-01"
		end := m.AddDate(0, 1, -1).Format("2006-01-02")

		var present, absent, late int64
		r.db.Model(&models.Attendance{}).
			Where("date BETWEEN ? AND ? AND status = ?", start, end, "present").
			Count(&present)
		r.db.Model(&models.Attendance{}).
			Where("date BETWEEN ? AND ? AND status = ?", start, end, "absent").
			Count(&absent)
		r.db.Model(&models.Attendance{}).
			Where("date BETWEEN ? AND ? AND late_minutes > 0", start, end).
			Count(&late)

		result = append(result, MonthlyAttendanceCount{
			Month:   monthStr,
			Present: present,
			Absent:  absent,
			Late:    late,
		})
	}
	return result
}

func (r *DashboardRepository) Last7DaysAttendance() []DailyAttendanceCount {
	var result []DailyAttendanceCount
	today := time.Now()
	for i := 6; i >= 0; i-- {
		d := today.AddDate(0, 0, -i)
		dateStr := d.Format("2006-01-02")

		var present, absent, late int64
		r.db.Model(&models.Attendance{}).
			Where("date = ? AND status = ?", dateStr, "present").
			Count(&present)
		r.db.Model(&models.Attendance{}).
			Where("date = ? AND status = ?", dateStr, "absent").
			Count(&absent)
		r.db.Model(&models.Attendance{}).
			Where("date = ? AND late_minutes > 0", dateStr).
			Count(&late)

		result = append(result, DailyAttendanceCount{
			Date:    d.Format("Jan 02"),
			Present: present,
			Absent:  absent,
			Late:    late,
		})
	}
	return result
}

func (r *DashboardRepository) RecentActivity() []ActivityItem {
	var items []ActivityItem

	var recentLeaves []models.Leave
	r.db.Where("status = ?", "pending").Order("created_at DESC").Limit(5).Find(&recentLeaves)
	for _, l := range recentLeaves {
		items = append(items, ActivityItem{
			Type:        "leave_request",
			Description: "Leave request #" + l.ID[:8],
			Date:        l.CreatedAt.Format("2006-01-02"),
		})
	}

	var recentEmployees []models.Employee
	r.db.Order("created_at DESC").Limit(5).Find(&recentEmployees)
	for _, e := range recentEmployees {
		items = append(items, ActivityItem{
			Type:        "new_employee",
			Description: "New employee: " + e.NameEn,
			Date:        e.CreatedAt.Format("2006-01-02"),
		})
	}

	return items
}
