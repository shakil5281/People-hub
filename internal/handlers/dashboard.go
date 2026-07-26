package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type DashboardHandler struct {
	repo *repository.DashboardRepository
}

func NewDashboardHandler(repo *repository.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

type dashboardStats struct {
	TotalEmployees     int64                               `json:"total_employees"`
	ActiveEmployees    int64                               `json:"active_employees"`
	TotalDepartments   int64                               `json:"total_departments"`
	TotalSections      int64                               `json:"total_sections"`
	TodayAttendance    int64                               `json:"today_attendance"`
	TodayLogs          int64                               `json:"today_logs"`
	PendingLeaves      int64                               `json:"pending_leaves"`
	NewHiresMonth      int64                               `json:"new_hires_month"`
	SeparationsMonth   int64                               `json:"separations_month"`
	GenderDistribution []repository.GenderCount            `json:"gender_distribution"`
	DepartmentCounts   []repository.DeptCount              `json:"department_counts"`
	MonthlyAttendance  []repository.MonthlyAttendanceCount `json:"monthly_attendance"`
	RecentActivity     []repository.ActivityItem           `json:"recent_activity"`
}

// GetStats godoc
//
//	@Summary      Dashboard statistics
//	@Description  Get aggregated dashboard stats
//	@Tags         Dashboard
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {object}  dashboardStats
//	@Router       /dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	firstOfMonth := time.Now().Format("2006-01") + "-01"

	stats := dashboardStats{
		TotalEmployees:     h.repo.CountTotalEmployees(),
		ActiveEmployees:    h.repo.CountActiveEmployees(),
		TotalDepartments:   h.repo.CountDepartments(),
		TotalSections:      h.repo.CountSections(),
		TodayAttendance:    h.repo.CountTodayAttendance(today),
		TodayLogs:          h.repo.CountTodayLogs(today),
		PendingLeaves:      h.repo.CountPendingLeaves(),
		NewHiresMonth:      h.repo.CountNewHiresMonth(firstOfMonth),
		SeparationsMonth:   h.repo.CountSeparationsMonth(firstOfMonth),
		GenderDistribution: h.repo.GenderDistribution(),
		DepartmentCounts:   h.repo.DepartmentCounts(),
		MonthlyAttendance:  h.repo.MonthlyAttendanceTrend(5),
		RecentActivity:     h.repo.RecentActivity(),
	}

	c.JSON(http.StatusOK, stats)
}
