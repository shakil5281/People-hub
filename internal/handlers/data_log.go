package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

const maxProcessDays = 31

type DataLogHandler struct {
	dataLogRepo         *repository.DataLogRepository
	dataLogService      *service.DataLogService
	attendanceProcessor *service.AttendanceProcessor
}

func NewDataLogHandler(
	dataLogRepo *repository.DataLogRepository,
	dataLogService *service.DataLogService,
	attendanceProcessor *service.AttendanceProcessor,
) *DataLogHandler {
	return &DataLogHandler{
		dataLogRepo:         dataLogRepo,
		dataLogService:      dataLogService,
		attendanceProcessor: attendanceProcessor,
	}
}

type ImportRequest struct {
	FilePath  string `json:"file_path"`
	StartDate string `json:"start_date"` // YYYY-MM-DD, optional
	EndDate   string `json:"end_date"`   // YYYY-MM-DD, optional
}

type ProcessRequest struct {
	Date      string `json:"date"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	CompanyID string `json:"company_id"`
}

// ImportDataLogs godoc
//
// @Summary      Import data logs from ZKTeco MDB
// @Description  Read and import raw punch data from ZKTeco Access MDB file
// @Tags         Data Logs
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body ImportRequest false "MDB file path (defaults to C:\\Program Files (x86)\\ZKTeco\\att2000.mdb)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /data-logs/import [post]
func (h *DataLogHandler) Import(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.dataLogService.ImportFromMDB(req.FilePath, req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Data logs imported successfully",
		"imported": result.Imported,
		"skipped":  result.Skipped,
	})
}

// ListDataLogs godoc
//
// @Summary      List data logs
// @Description  Get raw punch data logs by date range
// @Tags         Data Logs
// @Security     BearerAuth
// @Produce      json
// @Param        start query string false "Start date (YYYY-MM-DD)"
// @Param        end   query string false "End date (YYYY-MM-DD)"
// @Success      200  {array}   map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /data-logs [get]
func (h *DataLogHandler) List(c *gin.Context) {
	start := c.DefaultQuery("start", time.Now().Format("2006-01-02"))
	end := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

	logs, err := h.dataLogRepo.ListByDateRange(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// ProcessDataLogs godoc
//
// @Summary      Process data logs into attendance
// @Description  Convert unprocessed raw punch data into attendance records (bulk-optimized, idempotent)
// @Tags         Data Logs
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body ProcessRequest true "Date range and company to process"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /data-logs/process [post]
func (h *DataLogHandler) Process(c *gin.Context) {
	startTime := time.Now()
	var req ProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	// Mandatory company_id
	if strings.TrimSpace(req.CompanyID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}
	if !uuidRegex.MatchString(req.CompanyID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id must be a valid UUID"})
		return
	}
	// Company authorization: non-super_admin can only process own company
	if tokenCompanyID, exists := c.Get("company_id"); exists {
		if cid, ok := tokenCompanyID.(string); ok && cid != "" {
			rolesVal, _ := c.Get("roles")
			isSuper := false
			if roles, ok := rolesVal.([]string); ok {
				for _, r := range roles {
					if strings.EqualFold(r, "super_admin") {
						isSuper = true
						break
					}
				}
			}
			// also check comma-separated if stored differently
			if !isSuper && cid != req.CompanyID {
				c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to process for this company"})
				return
			}
		}
	}

	startDate, endDate := h.resolveDateRange(req)

	// Validate date formats
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected YYYY-MM-DD"})
		return
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected YYYY-MM-DD"})
		return
	}
	if startDate > endDate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be <= end_date"})
		return
	}
	// Enforce max range (spec §22, §26)
	if days := dateDiffDays(startDate, endDate); days > maxProcessDays {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("date range exceeds maximum %d days (requested %d days)", maxProcessDays, days)})
		return
	}

	result, err := h.attendanceProcessor.ProcessDateRange(startDate, endDate, req.CompanyID)
	if err != nil {
		// Concurrent lock
		if strings.Contains(err.Error(), "already running") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		log.Printf("[data-log] process failed company=%s range=%s..%s err=%v", req.CompanyID, startDate, endDate, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	duration := time.Since(startTime)
	log.Printf("[data-log] process ok company=%s range=%s..%s days=%d logs=%d created=%d updated=%d skipped=%d duration=%s",
		req.CompanyID, startDate, endDate, result.Days, result.TotalLogs, result.TotalCreated, result.TotalUpdated, result.TotalSkipped, duration)

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Processed %d employee attendances across %d days from %d raw logs", result.TotalProcessed, result.Days, result.TotalLogs),
		"start":      startDate,
		"end":        endDate,
		"days":       result.Days,
		"total_logs": result.TotalLogs,
		"processed":  result.TotalProcessed,
		"created":    result.TotalCreated,
		"updated":    result.TotalUpdated,
		"skipped":    result.TotalSkipped,
		"details":    result.Details,
		"duration_ms": duration.Milliseconds(),
	})
}

func dateDiffDays(start, end string) int {
	s, _ := time.Parse("2006-01-02", start)
	e, _ := time.Parse("2006-01-02", end)
	return int(e.Sub(s).Hours()/24) + 1
}

func (h *DataLogHandler) resolveDateRange(req ProcessRequest) (string, string) {
	startDate := req.StartDate
	endDate := req.EndDate

	if startDate == "" && endDate == "" {
		if req.Date != "" {
			startDate = req.Date
			endDate = req.Date
		} else {
			startDate = time.Now().Format("2006-01-02")
			endDate = startDate
		}
	} else if startDate == "" {
		startDate = endDate
	} else if endDate == "" {
		endDate = startDate
	}

	return startDate, endDate
}

// DataLogStats godoc
//
// @Summary      Data log statistics
// @Description  Get count of imported data logs
// @Tags         Data Logs
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /data-logs/stats [get]
func (h *DataLogHandler) Stats(c *gin.Context) {
	total, err := h.dataLogRepo.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	today := time.Now().Format("2006-01-02")
	todayCount, err := h.dataLogRepo.CountByDate(today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_logs": total,
		"today_logs": todayCount,
		"today_date": today,
	})
}

// DeleteAllDataLogs godoc
//
// @Summary      Delete all data logs
// @Description  Permanently delete all raw punch data logs
// @Tags         Data Logs
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /data-logs/delete-all [delete]
func (h *DataLogHandler) DeleteAll(c *gin.Context) {
	if err := h.dataLogRepo.DeleteAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All data logs deleted permanently"})
}
