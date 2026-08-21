package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type MissingAttendanceHandler struct {
	repo        *repository.MissingAttendanceRepository
	empRepo     *repository.EmployeeRepository
	attRepo     *repository.AttendanceRepository
	companyRepo *repository.CompanyRepository
}

func NewMissingAttendanceHandler(
	repo *repository.MissingAttendanceRepository,
	empRepo *repository.EmployeeRepository,
	attRepo *repository.AttendanceRepository,
	companyRepo *repository.CompanyRepository,
) *MissingAttendanceHandler {
	return &MissingAttendanceHandler{repo: repo, empRepo: empRepo, attRepo: attRepo, companyRepo: companyRepo}
}

type createMissingAttRequest struct {
	EmployeeID string  `json:"employee_id" binding:"required"`
	CompanyID  string  `json:"company_id" binding:"required"`
	Date       string  `json:"date" binding:"required"`
	CheckIn    *string `json:"check_in"`
	CheckOut   *string `json:"check_out"`
	Status     string  `json:"status" binding:"required"`
	Notes      string  `json:"notes"`
}

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	formats := []string{"2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02T15:04:05"}
	for _, f := range formats {
		if t, err := time.Parse(f, *s); err == nil {
			return &t
		}
	}
	return nil
}

// Create godoc
//
// @Summary      Create missing attendance
// @Description  Create a missing attendance override record
// @Tags         Missing Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body createMissingAttRequest true "Missing attendance details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /missing-attendance [post]
func (h *MissingAttendanceHandler) Create(c *gin.Context) {
	var req createMissingAttRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "present"
	}

	ma := &models.MissingAttendance{
		EmployeeID: req.EmployeeID,
		CompanyID:  req.CompanyID,
		Date:       req.Date,
		CheckIn:    parseTimePtr(req.CheckIn),
		CheckOut:   parseTimePtr(req.CheckOut),
		Status:     status,
		Notes:      req.Notes,
	}

	userID := c.GetString("user_id")
	if userID != "" {
		ma.CreatedBy = &userID
	}

	if err := h.repo.Create(ma); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ma, "message": "Missing attendance created"})
}

// List godoc
//
// @Summary      List missing attendance
// @Description  List missing attendance records with filters
// @Tags         Missing Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date query string true "Start date"
// @Param        end_date query string true "End date"
// @Param        company_id query string false "Company ID"
// @Param        department_id query string false "Department ID"
// @Param        section_id query string false "Section ID"
// @Param        designation_id query string false "Designation ID"
// @Param        line_id query string false "Line ID"
// @Param        group_id query string false "Group ID"
// @Param        shift_id query string false "Shift ID"
// @Param        status query string false "Status"
// @Param        employee_id query string false "Employee ID"
// @Param        page query int false "Page" default(1)
// @Param        limit query int false "Limit" default(20)
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /missing-attendance [get]
func (h *MissingAttendanceHandler) List(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	p := utils.ParsePagination(c)
	data, total, err := h.repo.List(
		startDate, endDate,
		c.Query("company_id"), c.Query("department_id"),
		c.Query("section_id"), c.Query("designation_id"),
		c.Query("line_id"), c.Query("group_id"),
		c.Query("shift_id"), c.Query("status"),
		c.Query("employee_id"),
		p.Page, p.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(data, total, p))
}

// Delete godoc
//
// @Summary      Delete missing attendance
// @Description  Delete a missing attendance record
// @Tags         Missing Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Missing Attendance ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /missing-attendance/{id} [delete]
func (h *MissingAttendanceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// Update godoc
//
// @Summary      Update missing attendance
// @Description  Update a missing attendance record
// @Tags         Missing Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Missing Attendance ID"
// @Param        request body createMissingAttRequest true "Update details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /missing-attendance/{id} [put]
func (h *MissingAttendanceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	var req createMissingAttRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.CheckIn = parseTimePtr(req.CheckIn)
	existing.CheckOut = parseTimePtr(req.CheckOut)
	existing.Status = req.Status
	existing.Notes = req.Notes

	if err := h.repo.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": existing, "message": "Updated"})
}

// UpsertByEmployeeAndDate godoc
//
// @Summary      Upsert missing attendance by employee and date
// @Description  Create or update a missing attendance record by employee_id + date
// @Tags         Missing Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body createMissingAttRequest true "Missing attendance details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /missing-attendance/upsert [post]
func (h *MissingAttendanceHandler) UpsertByEmployeeAndDate(c *gin.Context) {
	var req createMissingAttRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "present"
	}

	ma := &models.MissingAttendance{
		EmployeeID: req.EmployeeID,
		CompanyID:  req.CompanyID,
		Date:       req.Date,
		CheckIn:    parseTimePtr(req.CheckIn),
		CheckOut:   parseTimePtr(req.CheckOut),
		Status:     status,
		Notes:      req.Notes,
	}

	userID := c.GetString("user_id")
	if userID != "" {
		ma.CreatedBy = &userID
	}

	if err := h.repo.UpsertByEmployeeAndDate(ma); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Also update the attendances table immediately so the missing report reflects the change
	dateStr := req.Date
	if len(dateStr) > 10 {
		dateStr = dateStr[:10]
	}
	existingAtt, _ := h.attRepo.FindByEmployeeAndDate(req.EmployeeID, dateStr)
	if existingAtt != nil {
		existingAtt.CheckIn = ma.CheckIn
		existingAtt.CheckOut = ma.CheckOut
		existingAtt.Status = ma.Status
		existingAtt.CalculateHours()
		h.attRepo.Update(existingAtt)
	} else {
		att := &models.Attendance{
			EmployeeID: req.EmployeeID,
			CompanyID:  req.CompanyID,
			Date:       dateStr,
			CheckIn:    ma.CheckIn,
			CheckOut:   ma.CheckOut,
			Status:     ma.Status,
		}
		att.CalculateHours()
		h.attRepo.Create(att)
	}

	c.JSON(http.StatusOK, gin.H{
		"employee_id": ma.EmployeeID,
		"date":        ma.Date,
		"check_in":    ma.CheckIn,
		"check_out":   ma.CheckOut,
		"status":      ma.Status,
		"message":     "Missing attendance saved",
	})
}

type bulkMissingAttRequest struct {
	Records []createMissingAttRequest `json:"records" binding:"required,min=1"`
}

// BulkUpsert godoc
//
// @Summary      Bulk upsert missing attendance
// @Description  Create or update multiple missing attendance records at once
// @Tags         Missing Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body bulkMissingAttRequest true "Bulk missing attendance records"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /missing-attendance/bulk [post]
func (h *MissingAttendanceHandler) BulkUpsert(c *gin.Context) {
	var req bulkMissingAttRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	records := make([]models.MissingAttendance, 0, len(req.Records))
	for _, r := range req.Records {
		status := r.Status
		if status == "" {
			status = "present"
		}
		ma := models.MissingAttendance{
			EmployeeID: r.EmployeeID,
			CompanyID:  r.CompanyID,
			Date:       r.Date,
			CheckIn:    parseTimePtr(r.CheckIn),
			CheckOut:   parseTimePtr(r.CheckOut),
			Status:     status,
			Notes:      r.Notes,
		}
		if userID != "" {
			ma.CreatedBy = &userID
		}
		records = append(records, ma)
	}

	if err := h.repo.BulkUpsert(records, h.attRepo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(records),
		"message": "Bulk missing attendance saved",
	})
}
