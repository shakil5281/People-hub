package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"github.com/xuri/excelize/v2"
)

type AttendanceHandler struct {
	attendanceRepo *repository.AttendanceRepository
	employeeRepo   *repository.EmployeeRepository
	dataLogRepo    *repository.DataLogRepository
	separationRepo *repository.SeparationRepository
}

func NewAttendanceHandler(attendanceRepo *repository.AttendanceRepository, employeeRepo *repository.EmployeeRepository, dataLogRepo *repository.DataLogRepository, separationRepo *repository.SeparationRepository) *AttendanceHandler {
	return &AttendanceHandler{attendanceRepo: attendanceRepo, employeeRepo: employeeRepo, dataLogRepo: dataLogRepo, separationRepo: separationRepo}
}

type CreateAttendanceRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	CompanyID  string `json:"company_id" binding:"required"`
	ShiftID    string `json:"shift_id"`
	Date       string `json:"date" binding:"required"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	Status     string `json:"status"`
}

type ClockInRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	ShiftID    string `json:"shift_id"`
}

type ClockOutRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
}

type AttendanceRow struct {
	ID           string  `json:"id"`
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	Designation  string  `json:"designation"`
	Date         string  `json:"date"`
	CheckIn      *string `json:"check_in"`
	CheckOut     *string `json:"check_out"`
	TotalHours   *string `json:"total_hours"`
	OverTime     *string `json:"over_time"`
	Status       string  `json:"status"`
	LateMinutes  int     `json:"late_minutes"`
	ShiftName    string  `json:"shift_name"`
	PunchNumber  *string `json:"punch_number"`
	CompanyID    string  `json:"company_id"`
}

func toAttendanceRow(a models.Attendance) AttendanceRow {
	r := AttendanceRow{
		ID:          a.ID,
		EmployeeID:  a.EmployeeID,
		Date:        a.Date,
		Status:      a.Status,
		LateMinutes: a.LateMinutes,
		PunchNumber: a.PunchNumber,
		CompanyID:   a.CompanyID,
		TotalHours:  a.TotalHours,
		OverTime:    a.OverTime,
	}
	if a.Employee.NameEn != "" {
		r.EmployeeName = a.Employee.NameEn
	}
	if a.Employee.DesignationRef != nil {
		r.Designation = a.Employee.DesignationRef.Name
	}
	if a.Shift != nil {
		r.ShiftName = a.Shift.Name
	}
	if a.CheckIn != nil {
		s := a.CheckIn.Format("2006-01-02 15:04:05")
		r.CheckIn = &s
	}
	if a.CheckOut != nil {
		s := a.CheckOut.Format("2006-01-02 15:04:05")
		r.CheckOut = &s
	}
	return r
}

func toAttendanceRows(list []models.Attendance) []AttendanceRow {
	res := make([]AttendanceRow, len(list))
	for i, a := range list {
		res[i] = toAttendanceRow(a)
	}
	return res
}

type JobCardRow struct {
	ID          string  `json:"id"`
	EmployeeID  string  `json:"employee_id"`
	Date        string  `json:"date"`
	CheckIn     *string `json:"check_in"`
	CheckOut    *string `json:"check_out"`
	TotalHours  *string `json:"total_hours"`
	OverTime    *string `json:"over_time"`
	Status      string  `json:"status"`
	LateMinutes int     `json:"late_minutes"`
	ShiftName   string  `json:"shift_name"`
}

func toJobCardRow(a models.Attendance) JobCardRow {
	r := JobCardRow{
		ID:          a.ID,
		EmployeeID:  a.EmployeeID,
		Date:        a.Date,
		Status:      a.Status,
		LateMinutes: a.LateMinutes,
		TotalHours:  a.TotalHours,
		OverTime:    a.OverTime,
	}
	if a.Shift != nil {
		r.ShiftName = a.Shift.Name
	}
	if a.CheckIn != nil {
		s := a.CheckIn.Format("2006-01-02 15:04:05")
		r.CheckIn = &s
	}
	if a.CheckOut != nil {
		s := a.CheckOut.Format("2006-01-02 15:04:05")
		r.CheckOut = &s
	}
	return r
}

func toJobCardRows(list []models.Attendance) []JobCardRow {
	res := make([]JobCardRow, len(list))
	for i, a := range list {
		res[i] = toJobCardRow(a)
	}
	return res
}

// ListAttendances godoc
//
// @Summary      List attendances
// @Description  Get attendances by date (default: today)
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        date           query string false "Date (YYYY-MM-DD)"
// @Param        company_id     query string false "Filter by company"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        shift_id       query string false "Filter by shift"
// @Param        status         query string false "Filter by status"
// @Param        employee_id    query string false "Filter by employee"
// @Param        page           query int    false "Page number (default: 1)"
// @Param        limit          query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /attendance [get]
func (h *AttendanceHandler) List(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	date := c.Query("date")

	if date == "" {
		if startDate != "" {
			date = startDate
		} else {
			date = time.Now().Format("2006-01-02")
		}
	}
	if startDate == "" {
		startDate = date
	}
	if endDate == "" {
		endDate = date
	}

	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	shiftID := c.Query("shift_id")
	status := c.Query("status")
	employeeID := c.Query("employee_id")

	p := utils.ParsePagination(c)

	hasFilters := companyID != "" || departmentID != "" || sectionID != "" || designationID != "" || lineID != "" || groupID != "" || shiftID != "" || status != "" || employeeID != "" || startDate != endDate
	if hasFilters {
		realCount, err := h.attendanceRepo.CountFilteredByDateRange(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if realCount == 0 {
			c.JSON(http.StatusOK, utils.NewPaginatedResponse([]AttendanceRow{}, 0, p))
			return
		}

		attendances, total, err := h.attendanceRepo.ListByDateRangeFiltered(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID, p.Page, p.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, utils.NewPaginatedResponse(toAttendanceRows(attendances), total, p))
		return
	}

	realCount, err := h.attendanceRepo.CountByDate(date, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if realCount == 0 {
		c.JSON(http.StatusOK, utils.NewPaginatedResponse([]AttendanceRow{}, 0, p))
		return
	}

	attendances, total, err := h.attendanceRepo.ListByDate(date, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(toAttendanceRows(attendances), total, p))
}

// GetAttendance godoc
//
// @Summary      Get attendance by ID
// @Description  Get an attendance record by its ID
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Attendance ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /attendance/{id} [get]
func (h *AttendanceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	attendance, err := h.attendanceRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attendance record not found"})
		return
	}
	c.JSON(http.StatusOK, attendance)
}

// CreateAttendance godoc
//
// @Summary      Create attendance
// @Description  Create a new attendance record
// @Tags         Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateAttendanceRequest true "Attendance details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /attendance [post]
func (h *AttendanceHandler) Create(c *gin.Context) {
	var req CreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "present"
	}

	userID := c.GetString("user_id")
	var checkIn, checkOut *time.Time
	if req.CheckIn != "" {
		if t, err := utils.ParseDateTime(req.CheckIn, req.Date); err == nil {
			checkIn = &t
		}
	}
	if req.CheckOut != "" {
		if t, err := utils.ParseDateTime(req.CheckOut, req.Date); err == nil {
			checkOut = &t
		}
	}

	attendance := &models.Attendance{
		EmployeeID: req.EmployeeID,
		CompanyID:  req.CompanyID,
		Date:       req.Date,
		CheckIn:    checkIn,
		CheckOut:   checkOut,
		Status:     status,
		CreatedBy:  &userID,
	}

	if req.ShiftID != "" {
		attendance.ShiftID = &req.ShiftID
	}

	existing, err := h.attendanceRepo.FindByEmployeeAndDate(req.EmployeeID, req.Date)
	if err == nil && existing != nil && existing.ID != "" {
		existing.CheckIn = checkIn
		existing.CheckOut = checkOut
		existing.Status = status
		existing.CompanyID = req.CompanyID
		if req.ShiftID != "" {
			existing.ShiftID = &req.ShiftID
		}
		existing.UpdatedBy = &userID
		existing.CalculateHours()
		if err := h.attendanceRepo.Update(existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existing)
		return
	}

	attendance.CalculateHours()
	if err := h.attendanceRepo.Create(attendance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

// UpdateAttendance godoc
//
// @Summary      Update attendance
// @Description  Update an attendance record
// @Tags         Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Attendance ID"
// @Param        request body CreateAttendanceRequest true "Updated attendance details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /attendance/{id} [put]
func (h *AttendanceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	attendance, err := h.attendanceRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attendance record not found"})
		return
	}

	var req CreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	attendance.EmployeeID = req.EmployeeID
	attendance.CompanyID = req.CompanyID
	attendance.Date = req.Date
	attendance.Status = req.Status
	attendance.UpdatedBy = &userID

	if req.CheckIn != "" {
		if t, err := utils.ParseDateTime(req.CheckIn, req.Date); err == nil {
			attendance.CheckIn = &t
		}
	}
	if req.CheckOut != "" {
		if t, err := utils.ParseDateTime(req.CheckOut, req.Date); err == nil {
			attendance.CheckOut = &t
		}
	}
	if req.ShiftID != "" {
		attendance.ShiftID = &req.ShiftID
	}

	attendance.CalculateHours()
	if err := h.attendanceRepo.Update(attendance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

// DeleteAttendance godoc
//
// @Summary      Delete attendance
// @Description  Soft delete an attendance record
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Attendance ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /attendance/{id} [delete]
func (h *AttendanceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.attendanceRepo.FindByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attendance record not found"})
		return
	}

	if err := h.attendanceRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "attendance record deleted"})
}

// ListJobCard godoc
//
// @Summary      Job card data
// @Description  Get attendance records for job card with advanced filters
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date    query string false "Start date (YYYY-MM-DD)"
// @Param        end_date      query string false "End date (YYYY-MM-DD)"
// @Param        company_id    query string false "Filter by company"
// @Param        employee_id   query string false "Filter by employee"
// @Param        department_id query string false "Filter by department"
// @Param        status        query string false "Filter by status (present|late|absent|half-day)"
// @Param        page          query int    false "Page number (default: 1)"
// @Param        limit         query int    false "Page size (default: 20, max: 100)"
// @Param        list_mode     query string false "Set to 'true' to get employee list only"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      500  {object}  map[string]string
// @Router       /attendance/job-card [get]
func (h *AttendanceHandler) ListJobCard(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	companyID := c.Query("company_id")
	employeeID := c.Query("employee_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	shiftID := c.Query("shift_id")
	status := c.Query("status")
	listMode := c.Query("list_mode")

	if employeeID != "" {
		emp, err := h.employeeRepo.FindByEmployeeID(employeeID)
		if err != nil {
			emp, err = h.employeeRepo.FindByPunchNumber(employeeID)
		}
		if err == nil && emp != nil {
			employeeID = emp.EmployeeID
		}
	}

	// Cap end_date to separation date if employee has been separated
	endDate = capEndDateToSeparation(h, employeeID, endDate)

	// List mode: return only distinct employee IDs/names for navigation
	if listMode == "true" {
		employees, err := h.attendanceRepo.ListJobCardEmployees(startDate, endDate, companyID, employeeID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		type EmpInfo struct {
			EmployeeID  string `json:"employee_id"`
			NameEn      string `json:"name_en"`
			Designation string `json:"designation"`
			Department  string `json:"department"`
			Company     string `json:"company"`
			Phone       string `json:"phone"`
			JoiningDate string `json:"joining_date"`
		}
		var result []EmpInfo
		for _, e := range employees {
			d := ""
			if e.DesignationRef != nil {
				d = e.DesignationRef.Name
			}
			dp := ""
			if e.Department != nil {
				dp = e.Department.Name
			}
			cn := ""
			if e.Company.CompanyNameEn != "" {
				cn = e.Company.CompanyNameEn
			}
			jo := ""
			if !e.JoiningDate.IsZero() {
				jo = e.JoiningDate.Format("2006-01-02")
			}
			result = append(result, EmpInfo{
				EmployeeID: e.EmployeeID, NameEn: e.NameEn,
				Designation: d, Department: dp,
				Company: cn, Phone: e.Phone, JoiningDate: jo,
			})
		}
		c.JSON(http.StatusOK, gin.H{"data": result, "total": len(result)})
		return
	}

	page := 1
	limit := 366
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 366 {
			limit = v
		}
	}

	attendances, total, err := h.attendanceRepo.ListJobCard(startDate, endDate, companyID, employeeID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	p := utils.Pagination{Page: page, Limit: limit}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(toJobCardRows(attendances), total, p))
}

// ClockIn godoc
//
// @Summary      Clock in
// @Description  Mark clock-in for an employee today
// @Tags         Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body ClockInRequest true "Clock-in details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /attendance/clock-in [post]
func (h *AttendanceHandler) ClockIn(c *gin.Context) {
	var req ClockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	existing, err := h.attendanceRepo.FindByEmployeeAndDate(req.EmployeeID, today)
	if err == nil && existing != nil && existing.ID != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "already clocked in today"})
		return
	}

	userID := c.GetString("user_id")
	attendance := &models.Attendance{
		EmployeeID: req.EmployeeID,
		Date:       today,
		CheckIn:    &now,
		Status:     "present",
		CreatedBy:  &userID,
	}

	if req.ShiftID != "" {
		attendance.ShiftID = &req.ShiftID
	}

	if err := h.attendanceRepo.Create(attendance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

// ClockOut godoc
//
// @Summary      Clock out
// @Description  Mark clock-out for an employee today
// @Tags         Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body ClockOutRequest true "Clock-out details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /attendance/clock-out [post]
func (h *AttendanceHandler) ClockOut(c *gin.Context) {
	var req ClockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	attendance, err := h.attendanceRepo.FindByEmployeeAndDate(req.EmployeeID, today)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no clock-in found for today"})
		return
	}

	if attendance.CheckOut != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already clocked out today"})
		return
	}

	userID := c.GetString("user_id")
	attendance.CheckOut = &now
	attendance.UpdatedBy = &userID

	// Calculate total hours
	if attendance.CheckIn != nil {
		duration := now.Sub(*attendance.CheckIn)
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		totalHours := fmt.Sprintf("%02d:%02d", hours, minutes)
		attendance.TotalHours = &totalHours
	}

	if err := h.attendanceRepo.Update(attendance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

// DeleteAllAttendances godoc
//
// @Summary      Delete all attendances
// @Description  Permanently delete all attendance records
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /attendance/delete-all [delete]
func (h *AttendanceHandler) DeleteAll(c *gin.Context) {
	if err := h.attendanceRepo.DeleteAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All attendance records deleted permanently"})
}

type BulkDeleteAttendanceRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// DeleteBulk godoc
//
// @Summary      Bulk delete attendance records
// @Description  Delete multiple attendance records by ID array
// @Tags         Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkDeleteAttendanceRequest true "Attendance IDs to delete"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /attendance/bulk-delete [post]
func (h *AttendanceHandler) DeleteBulk(c *gin.Context) {
	var req BulkDeleteAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no attendance IDs provided"})
		return
	}
	if err := h.attendanceRepo.DeleteBulk(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d attendance record(s) deleted successfully", len(req.IDs))})
}

// AttendanceStats godoc
//
// @Summary      Attendance statistics
// @Description  Get count of attendance records
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /attendance/stats [get]
func (h *AttendanceHandler) Stats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	todayCount, _ := h.attendanceRepo.CountByDateOnly(today)
	c.JSON(http.StatusOK, gin.H{
		"today_count": todayCount,
		"today_date":  today,
	})
}

// Summary godoc
//
// @Summary      Daily attendance summary
// @Description  Get aggregated attendance summary by date/company/department
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date    query string false "Start date (YYYY-MM-DD)"
// @Param        end_date      query string false "End date (YYYY-MM-DD)"
// @Param        company_id    query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /attendance/summary [get]
func (h *AttendanceHandler) Summary(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", startDate)
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	shiftID := c.Query("shift_id")
	statusFilter := c.Query("status")
	employeeID := c.Query("employee_id")
	groupBy := c.Query("group_by")

	realCount, err := h.attendanceRepo.CountRangeByDate(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, statusFilter, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []map[string]interface{}

	if realCount == 0 {
		if result == nil {
			result = []map[string]interface{}{}
		}
	} else if groupBy != "" {
		result, err = h.attendanceRepo.SummaryByGroup(startDate, endDate, groupBy, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, statusFilter, employeeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		result, err = h.attendanceRepo.Summary(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, statusFilter, employeeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries":  result,
		"total":      len(result),
		"group_by":   groupBy,
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// Overtime godoc
//
// @Summary      Overtime sheet
// @Description  Get employee overtime records for a date range
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date    query string false "Start date (YYYY-MM-DD)"
// @Param        end_date      query string false "End date (YYYY-MM-DD)"
// @Param        company_id    query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /attendance/overtime [get]
func (h *AttendanceHandler) Overtime(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", startDate)
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	shiftID := c.Query("shift_id")
	statusFilter := c.Query("status")

	records, err := h.attendanceRepo.Overtime(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records":    records,
		"total":      len(records),
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// OvertimeSummary godoc
//
// @Summary      Overtime summary by department
// @Description  Get department-level overtime aggregation
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date    query string false "Start date (YYYY-MM-DD)"
// @Param        end_date      query string false "End date (YYYY-MM-DD)"
// @Param        company_id    query string false "Filter by company"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /attendance/overtime-summary [get]
func (h *AttendanceHandler) OvertimeSummary(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", startDate)
	companyID := c.Query("company_id")
	groupBy := c.DefaultQuery("group_by", "department")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")

	result, err := h.attendanceRepo.OvertimeSummaryGrouped(startDate, endDate, companyID, groupBy, departmentID, sectionID, designationID, lineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries":  result,
		"total":      len(result),
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// MissingAttendance godoc
//
// @Summary      Missing attendance
// @Description  Find attendance records where check_in or check_out is missing
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date    query string false "Start date (YYYY-MM-DD, default: today)"
// @Param        end_date      query string false "End date (YYYY-MM-DD, default: today)"
// @Param        company_id    query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Param        section_id    query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id       query string false "Filter by line"
// @Param        group_id      query string false "Filter by group"
// @Param        shift_id      query string false "Filter by shift"
// @Param        status        query string false "Filter by status"
// @Param        page          query int    false "Page number (default: 1)"
// @Param        limit         query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      500  {object}  map[string]string
// @Router       /attendance/missing [get]
func (h *AttendanceHandler) MissingAttendance(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	shiftID := c.Query("shift_id")
	status := c.Query("status")

	p := utils.ParsePagination(c)

	attendances, total, err := h.attendanceRepo.ListMissing(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(h.buildAttendanceRows(attendances, startDate, endDate), total, p))
}

// CustomAttendance godoc
//
// @Summary      Custom attendance report
// @Description  Get all attendance records in a date range (any status) with filters, for manual update
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date     query string false "Start date (YYYY-MM-DD)"
// @Param        end_date       query string false "End date (YYYY-MM-DD)"
// @Param        company_id     query string false "Filter by company"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        shift_id       query string false "Filter by shift"
// @Param        status         query string false "Filter by status"
// @Param        employee_id    query string false "Search by employee ID (partial match)"
// @Param        page           query int    false "Page number (default: 1)"
// @Param        limit          query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      500  {object}  map[string]string
// @Router       /attendance/custom [get]
func (h *AttendanceHandler) CustomAttendance(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	shiftID := c.Query("shift_id")
	status := c.Query("status")
	employeeID := c.Query("employee_id")

	p := utils.ParsePagination(c)

	attendances, total, err := h.attendanceRepo.ListCustom(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, status, employeeID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(h.buildAttendanceRows(attendances, startDate, endDate), total, p))
}

// buildAttendanceRows builds the API response rows for attendance lists, enriching
// each record with designation, shift name, formatted times and raw punch logs.
func (h *AttendanceHandler) buildAttendanceRows(attendances []models.Attendance, startDate, endDate string) []map[string]interface{} {
	badgeNumbers := make([]string, 0, len(attendances))
	empBadgeMap := make(map[string]string)
	for _, a := range attendances {
		if a.Employee.PunchNumber != "" {
			badgeNumbers = append(badgeNumbers, a.Employee.PunchNumber)
			empBadgeMap[a.EmployeeID] = a.Employee.PunchNumber
		}
	}

	punchLogs := make(map[string][]map[string]string)
	if len(badgeNumbers) > 0 {
		logs, err := h.dataLogRepo.ListByBadgeAndDateRange(badgeNumbers, startDate, endDate)
		if err == nil {
			for _, l := range logs {
				t := l.PunchTime.Format("2006-01-02 15:04:05")
				pt := "I"
				if l.PunchType == "O" {
					pt = "O"
				}
				punchLogs[l.BadgeNumber] = append(punchLogs[l.BadgeNumber], map[string]string{
					"time": t,
					"type": pt,
				})
			}
		}
	}

	var result []map[string]interface{}
	for _, a := range attendances {
		desig := ""
		if a.Employee.DesignationRef != nil {
			desig = a.Employee.DesignationRef.Name
		}
		shiftName := ""
		if a.Shift != nil {
			shiftName = a.Shift.Name
		}
		inTime := ""
		if a.CheckIn != nil {
			inTime = a.CheckIn.Format("2006-01-02 15:04:05")
		}
		outTime := ""
		if a.CheckOut != nil {
			outTime = a.CheckOut.Format("2006-01-02 15:04:05")
		}
		punches := punchLogs[empBadgeMap[a.EmployeeID]]
		result = append(result, map[string]interface{}{
			"id":            a.ID,
			"employee_id":   a.EmployeeID,
			"employee_name": a.Employee.NameEn,
			"designation":   desig,
			"shift_name":    shiftName,
			"check_in":      inTime,
			"check_out":     outTime,
			"status":        a.Status,
			"date":          a.Date,
			"company_id":    a.CompanyID,
			"punches":       punches,
		})
	}

	return result
}

// MonthlyReport godoc
//
// @Summary      Monthly attendance report
// @Description  Get per-employee monthly attendance summary (present, absent, leave, weekend, late, half_day, holiday, over_time)
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        year           query int    true  "Year"
// @Param        month          query int    true  "Month (1-12)"
// @Param        company_id     query string true  "Company ID"
// @Param        department_id  query string false "Filter by department"
// @Param        section_id     query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Param        line_id        query string false "Filter by line"
// @Param        group_id       query string false "Filter by group"
// @Param        shift_id       query string false "Filter by shift"
// @Param        employee_id    query string false "Search by employee ID (partial match)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /attendance/monthly-report [get]
func (h *AttendanceHandler) MonthlyReport(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")
	companyID := c.Query("company_id")

	if year == "" || month == "" || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year, month, and company_id are required"})
		return
	}

	y, err := time.Parse("2006", year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	m, err := time.Parse("1", month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return
	}

	startDate := time.Date(y.Year(), m.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	results, err := h.attendanceRepo.MonthlyReport(
		startStr, endStr, companyID,
		c.Query("department_id"),
		c.Query("section_id"),
		c.Query("designation_id"),
		c.Query("line_id"),
		c.Query("group_id"),
		c.Query("shift_id"),
		c.Query("employee_id"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totals := map[string]int{"present": 0, "absent": 0, "late": 0, "leave": 0, "weekend": 0, "half_day": 0, "holiday": 0, "over_time": 0}
	for _, r := range results {
		for k := range totals {
			if v, ok := r[k].(int64); ok {
				totals[k] += int(v)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"records":    results,
		"total":      len(results),
		"start_date": startStr,
		"end_date":   endStr,
		"year":       year,
		"month":      month,
		"totals":     totals,
	})
}

// AbsentAttendance godoc
//
// @Summary      Absent attendance
// @Description  Get employees marked as absent for a date range
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      json
// @Param        start_date query string false "Start date (YYYY-MM-DD)"
// @Param        end_date      query string false "End date (YYYY-MM-DD)"
// @Param        company_id    query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Param        page          query int    false "Page number (default: 1)"
// @Param        limit         query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      500  {object}  map[string]string
// @Router       /attendance/absent [get]
func (h *AttendanceHandler) AbsentAttendance(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", startDate)
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")
	shiftID := c.Query("shift_id")
	employeeID := c.Query("employee_id")

	p := utils.ParsePagination(c)
	attendances, total, err := h.attendanceRepo.ListByStatus(startDate, endDate, "absent", companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, employeeID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich each row with how many total absent days that employee has in the range.
	absentCounts, _ := h.attendanceRepo.CountAbsencesByEmployee(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, shiftID, employeeID)
	rows := make([]gin.H, 0, len(attendances))
	for _, a := range attendances {
		rows = append(rows, gin.H{
			"id":            a.ID,
			"employee_id":   a.EmployeeID,
			"date":          a.Date,
			"status":        a.Status,
			"check_in":      a.CheckIn,
			"check_out":     a.CheckOut,
			"total_absent":  absentCounts[a.EmployeeID],
			"employee":      a.Employee,
		})
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(rows, total, p))
}

func colNameAttendance(n int) string {
	name, _ := excelize.ColumnNumberToName(n)
	return name
}

func statusMap(s string) string {
	switch s {
	case "present":
		return "P"
	case "late":
		return "L"
	case "absent":
		return "A"
	case "on_leave":
		return "V"
	case "leave":
		return "V"
	case "weekend":
		return "W"
	case "half_day":
		return "H"
	default:
		return s
	}
}

func addGroupedSheet(f *excelize.File, sheetName, companyName, companyAddress, dateDisplay string, attendances []models.Attendance, groupFn func(models.Attendance) string) {
	f.NewSheet(sheetName)

	nCols := 8
	cols := []struct {
		header string
		width  float64
	}{
		{"Employee ID", 14},
		{"Name", 30},
		{"Designation", 24},
		{"In Time", 12},
		{"Out Time", 12},
		{"Late (Min)", 11},
		{"OT (Hr)", 11},
		{"Status", 10},
	}

	thinBorder := []excelize.Border{
		{Type: "left", Color: "333333", Style: 1},
		{Type: "top", Color: "333333", Style: 1},
		{Type: "bottom", Color: "333333", Style: 1},
		{Type: "right", Color: "333333", Style: 1},
	}

	companyNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	normalCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	groupHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})

	dataCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	dataLeft, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	endCol := colNameAttendance(nCols)

	// Row 1-4: Header
	f.SetCellValue(sheetName, "A1", companyName)
	f.MergeCell(sheetName, "A1", endCol+"1")
	f.SetCellStyle(sheetName, "A1", endCol+"1", companyNameStyle)
	f.SetRowHeight(sheetName, 1, 32)

	f.SetCellValue(sheetName, "A2", companyAddress)
	f.MergeCell(sheetName, "A2", endCol+"2")
	f.SetCellStyle(sheetName, "A2", endCol+"2", normalCenter)
	f.SetRowHeight(sheetName, 2, 20)

	f.SetCellValue(sheetName, "A3", sheetName+" ATTENDANCE")
	f.MergeCell(sheetName, "A3", endCol+"3")
	f.SetCellStyle(sheetName, "A3", endCol+"3", normalCenter)
	f.SetRowHeight(sheetName, 3, 20)

	f.SetCellValue(sheetName, "A4", "Date: "+dateDisplay)
	f.MergeCell(sheetName, "A4", endCol+"4")
	f.SetCellStyle(sheetName, "A4", endCol+"4", normalCenter)
	f.SetRowHeight(sheetName, 4, 20)

	// Row 5: Column Headers
	for i, c := range cols {
		cell := colNameAttendance(i+1) + "5"
		f.SetCellValue(sheetName, cell, c.header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
		f.SetColWidth(sheetName, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheetName, 5, 24)

	// Group attendances
	grouped := make(map[string][]models.Attendance)
	var groupOrder []string
	for _, a := range attendances {
		name := groupFn(a)
		if name == "" {
			name = "-"
		}
		if _, ok := grouped[name]; !ok {
			groupOrder = append(groupOrder, name)
		}
		grouped[name] = append(grouped[name], a)
	}

	row := 6
	sl := 0
	totalPres := 0
	totalAbs := 0
	totalLate := 0
	totalLeave := 0

	for _, groupName := range groupOrder {
		list := grouped[groupName]

		// Group header row
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), groupName+" ("+fmt.Sprintf("%d", len(list))+")")
		f.MergeCell(sheetName, "A"+strconv.Itoa(row), endCol+strconv.Itoa(row))
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), endCol+strconv.Itoa(row), groupHeaderStyle)
		f.SetRowHeight(sheetName, row, 22)
		row++

		// Attendance rows
		for _, a := range list {
			sl++
			svc := func(c int, v string) {
				f.SetCellValue(sheetName, colNameAttendance(c)+strconv.Itoa(row), v)
				f.SetCellStyle(sheetName, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataCenter)
			}
			svl := func(c int, v string) {
				f.SetCellValue(sheetName, colNameAttendance(c)+strconv.Itoa(row), v)
				f.SetCellStyle(sheetName, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataLeft)
			}

			svc(1, a.EmployeeID)
			svl(2, a.Employee.NameEn)

			designation := ""
			if a.Employee.DesignationRef != nil {
				designation = a.Employee.DesignationRef.Name
			}
			svl(3, designation)

			checkIn := ""
			if a.CheckIn != nil {
				checkIn = a.CheckIn.Format("2006-01-02 15:04:05")
			}
			svc(4, checkIn)

			checkOut := ""
			if a.CheckOut != nil {
				checkOut = a.CheckOut.Format("2006-01-02 15:04:05")
			}
			svc(5, checkOut)

			svc(6, fmt.Sprintf("%d", a.LateMinutes))

			overTime := ""
			if a.OverTime != nil {
				overTime = *a.OverTime
			}
			svc(7, overTime)

			status := a.Status
			if status == "" {
				status = "present"
			}
			statusCode := statusMap(status)
			if status == "absent" {
				redStyle, _ := f.NewStyle(&excelize.Style{
					Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "FF0000"},
					Border:    thinBorder,
					Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
				})
				f.SetCellValue(sheetName, colNameAttendance(8)+strconv.Itoa(row), statusCode)
				f.SetCellStyle(sheetName, colNameAttendance(8)+strconv.Itoa(row), colNameAttendance(8)+strconv.Itoa(row), redStyle)
			} else {
				svc(8, statusCode)
			}

			switch status {
			case "present":
				totalPres++
			case "absent":
				totalAbs++
			case "late":
				totalLate++
			case "on_leave", "leave":
				totalLeave++
			}

			f.SetRowHeight(sheetName, row, 20)
			row++
		}
		_ = sl
	}

	// Footer
	footerRow := row + 1
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	f.SetCellValue(sheetName, "A"+strconv.Itoa(footerRow), fmt.Sprintf("Total: %d Employees | Present: %d | Absent: %d | Late: %d | Leave: %d", len(attendances), totalPres, totalAbs, totalLate, totalLeave))
	f.MergeCell(sheetName, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow))
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow), footerStyle)
	f.SetRowHeight(sheetName, footerRow, 22)
}

// ExportAttendanceExcel godoc
//
//	@Summary      Export attendances to Excel
//	@Description  Export attendance data to Excel with company header, report info, and summary footer
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        date query string true "Date (YYYY-MM-DD)"
//	@Success      200  {file}  binary
//	@Router       /attendance/export/excel [get]
func (h *AttendanceHandler) ExportExcel(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var company models.Company
	database.DB.First(&company)

	query := database.DB.
		Preload("Employee.DesignationRef").
		Preload("Employee.Department").
		Preload("Employee.SectionRef").
		Preload("Employee.LineRef").
		Preload("Employee").
		Where("attendances.date = ? AND attendances.deleted_at IS NULL", date)

	if empID := c.Query("employee_id"); empID != "" {
		query = query.Where("attendances.employee_id = ?", empID)
	}
	if companyID := c.Query("company_id"); companyID != "" {
		query = query.Where("attendances.company_id = ?", companyID)
	}
	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("employees.department_id = ?", deptID)
	}
	if sectionID := c.Query("section_id"); sectionID != "" {
		query = query.Where("employees.section_id = ?", sectionID)
	}
	if desigID := c.Query("designation_id"); desigID != "" {
		query = query.Where("employees.designation_id = ?", desigID)
	}
	if lineID := c.Query("line_id"); lineID != "" {
		query = query.Where("employees.line_id = ?", lineID)
	}
	if groupID := c.Query("group_id"); groupID != "" {
		query = query.Where("employees.group_id = ?", groupID)
	}
	if shiftID := c.Query("shift_id"); shiftID != "" {
		query = query.Where("attendances.shift_id = ?", shiftID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("attendances.status = ?", status)
	}

	var attendances []models.Attendance
	if err := query.
		Joins("LEFT JOIN employees ON employees.employee_id = attendances.employee_id AND employees.deleted_at IS NULL").
		Order("LENGTH(attendances.employee_id) ASC, attendances.employee_id ASC").
		Find(&attendances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "Daily Attendance"
	index, _ := f.GetSheetIndex("Sheet1")
	f.SetActiveSheet(index)
	f.SetSheetName("Sheet1", sheet)

	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	// -- Styles --
	borderColor := "262626"

	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	headerCellStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})

	dataFont := &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"}
	styleData, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: thinBorder, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}})
	styleDataCenter, _ := f.NewStyle(&excelize.Style{Font: dataFont, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})

	cols := []struct {
		header string
		width  float64
		center bool
	}{
		{"Employee ID", 14, true},
		{"Name", 32, false},
		{"Designation", 26, false},
		{"In Time", 12, true},
		{"Out Time", 12, true},
		{"Late (Min)", 11, true},
		{"OT (Hr)", 11, true},
		{"Status", 12, true},
	}

	nCols := len(cols)
	headerRow := 5
	dataStartRow := headerRow + 1

	// --- Header Rows 1-4: separate rows with individual styling ---
	parsedDate, _ := time.Parse("2006-01-02", date)
	dateDisplay := parsedDate.Format("02 January, 2006")

	normalCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	companyNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	// Row 1: Company Name (Bold, Size 20)
	f.SetCellValue(sheet, "A1", companyName)
	f.MergeCell(sheet, "A1", colNameAttendance(nCols)+"1")
	f.SetCellStyle(sheet, "A1", colNameAttendance(nCols)+"1", companyNameStyle)
	f.SetRowHeight(sheet, 1, 32)

	// Row 2: Address
	f.SetCellValue(sheet, "A2", companyAddress)
	f.MergeCell(sheet, "A2", colNameAttendance(nCols)+"2")
	f.SetCellStyle(sheet, "A2", colNameAttendance(nCols)+"2", normalCenter)
	f.SetRowHeight(sheet, 2, 20)

	// Row 3: Report Name
	f.SetCellValue(sheet, "A3", "DAILY ATTENDANCE REPORT")
	f.MergeCell(sheet, "A3", colNameAttendance(nCols)+"3")
	f.SetCellStyle(sheet, "A3", colNameAttendance(nCols)+"3", normalCenter)
	f.SetRowHeight(sheet, 3, 20)

	// Row 4: Date
	f.SetCellValue(sheet, "A4", "Date: "+dateDisplay)
	f.MergeCell(sheet, "A4", colNameAttendance(nCols)+"4")
	f.SetCellStyle(sheet, "A4", colNameAttendance(nCols)+"4", normalCenter)
	f.SetRowHeight(sheet, 4, 20)

	// --- Row 5: Column Headers (no border) ---
	for i, c := range cols {
		cell := colNameAttendance(i+1) + strconv.Itoa(headerRow)
		f.SetCellValue(sheet, cell, c.header)
		f.SetCellStyle(sheet, cell, cell, headerCellStyle)
		f.SetColWidth(sheet, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheet, headerRow, 24)

	// --- Data Rows ---
	summary := map[string]int{"present": 0, "absent": 0, "late": 0, "weekend": 0, "half_day": 0, "on_leave": 0}

	for rowIdx, att := range attendances {
		row := rowIdx + dataStartRow

		svc := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), styleDataCenter)
		}
		sv := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), styleData)
		}

		svc(1, att.EmployeeID)
		sv(2, att.Employee.NameEn)

		designation := ""
		if att.Employee.DesignationRef != nil {
			designation = att.Employee.DesignationRef.Name
		}
		sv(3, designation)

		checkIn := ""
		if att.CheckIn != nil {
			checkIn = att.CheckIn.Format("15:04")
		}
		svc(4, checkIn)

		checkOut := ""
		if att.CheckOut != nil {
			checkOut = att.CheckOut.Format("15:04")
		}
		svc(5, checkOut)

		svc(6, fmt.Sprintf("%d", att.LateMinutes))

		overTime := ""
		if att.OverTime != nil {
			overTime = *att.OverTime
		}
		svc(7, overTime)

		status := att.Status
		if status == "" {
			status = "present"
		}
		statusCode := statusMap(status)
		if status == "absent" {
			redStyle, _ := f.NewStyle(&excelize.Style{
				Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "FF0000"},
				Border:    thinBorder,
				Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
			})
			f.SetCellValue(sheet, colNameAttendance(8)+strconv.Itoa(row), statusCode)
			f.SetCellStyle(sheet, colNameAttendance(8)+strconv.Itoa(row), colNameAttendance(8)+strconv.Itoa(row), redStyle)
		} else {
			svc(8, statusCode)
		}

		summary[att.Status]++
		f.SetRowHeight(sheet, row, 25)
	}

	// --- Footer: Summary (no border, full row merge) ---
	lastDataRow := dataStartRow + len(attendances) - 1
	footerRow := lastDataRow + 2

	totalEmployees := len(attendances)
	totalLeave := summary["on_leave"]

	footerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	drawFooter := func(label, value string) {
		f.SetCellValue(sheet, "A"+strconv.Itoa(footerRow), label+" "+value)
		f.MergeCell(sheet, "A"+strconv.Itoa(footerRow), colNameAttendance(nCols)+strconv.Itoa(footerRow))
		f.SetCellStyle(sheet, "A"+strconv.Itoa(footerRow), colNameAttendance(nCols)+strconv.Itoa(footerRow), footerStyle)
		f.SetRowHeight(sheet, footerRow, 20)
		footerRow++
	}

	drawFooter("Total Employees:", fmt.Sprintf("%d", totalEmployees))
	drawFooter("Total Present:", fmt.Sprintf("%d", summary["present"]))
	drawFooter("Total Absent:", fmt.Sprintf("%d", summary["absent"]))
	drawFooter("Total Late:", fmt.Sprintf("%d", summary["late"]))
	drawFooter("Total Leave:", fmt.Sprintf("%d", totalLeave))

	// --- Additional Summary Sheets ---
	addGroupedSheet(f, "Department Wise", companyName, companyAddress, dateDisplay, attendances, func(a models.Attendance) string {
		if a.Employee.Department != nil {
			return a.Employee.Department.Name
		}
		return "-"
	})

	addGroupedSheet(f, "Section Wise", companyName, companyAddress, dateDisplay, attendances, func(a models.Attendance) string {
		if a.Employee.SectionRef != nil {
			return a.Employee.SectionRef.Name
		}
		return "-"
	})

	addGroupedSheet(f, "Designation Wise", companyName, companyAddress, dateDisplay, attendances, func(a models.Attendance) string {
		if a.Employee.DesignationRef != nil {
			return a.Employee.DesignationRef.Name
		}
		return "-"
	})

	addGroupedSheet(f, "Line Wise", companyName, companyAddress, dateDisplay, attendances, func(a models.Attendance) string {
		if a.Employee.LineRef != nil {
			return a.Employee.LineRef.Name
		}
		return "-"
	})

	// --- Page Setup: A4 Portrait + No Gridlines for all sheets ---
	for _, s := range f.GetSheetList() {
		orientation := "portrait"
		paperSize := 9
		fitWidth := 1
		fitHeight := 0
		f.SetPageLayout(s, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitWidth,
			FitToHeight: &fitHeight,
		})

		f.SetPageMargins(s, &excelize.PageLayoutMarginsOptions{
			Left:   func(f float64) *float64 { return &f }(0.3),
			Right:  func(f float64) *float64 { return &f }(0.3),
			Top:    func(f float64) *float64 { return &f }(0.4),
			Bottom: func(f float64) *float64 { return &f }(0.4),
			Header: func(f float64) *float64 { return &f }(0),
			Footer: func(f float64) *float64 { return &f }(0),
		})

		f.SetSheetView(s, -1, &excelize.ViewOptions{ShowGridLines: func(b bool) *bool { return &b }(false)})
	}

	// Freeze panes for daily attendance sheet
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      headerRow,
		TopLeftCell: "A" + strconv.Itoa(dataStartRow),
		ActivePane:  "bottomLeft",
	})

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=attendance_%s.xlsx", date))
	f.Write(c.Writer)
}

// ExportAbsentExcel godoc
//
//	@Summary      Export absent attendance to Excel
//	@Description  Export absent employee attendance data to Excel with company header, report info, and summary footer
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        start_date query string true "Start date (YYYY-MM-DD)"
//	@Param        end_date query string true "End date (YYYY-MM-DD)"
//	@Success      200  {file}  binary
//	@Router       /attendance/absent/export/excel [get]
func (h *AttendanceHandler) ExportAbsentExcel(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if endDate == "" {
		endDate = startDate
	}

	companyFilter := c.Query("company_id")
	departmentFilter := c.Query("department_id")
	sectionFilter := c.Query("section_id")
	designationFilter := c.Query("designation_id")
	lineFilter := c.Query("line_id")
	groupFilter := c.Query("group_id")

	var company models.Company
	database.DB.First(&company)

	baseQuery := database.DB.Model(&models.Attendance{}).
		Preload("Employee.DesignationRef").
		Preload("Employee.Department").
		Preload("Employee.SectionRef").
		Preload("Employee.LineRef").
		Preload("Employee").
		Where("attendances.date BETWEEN ? AND ? AND attendances.status = ? AND attendances.deleted_at IS NULL", startDate, endDate, "absent")

	if companyFilter != "" {
		baseQuery = baseQuery.Where("attendances.company_id = ?", companyFilter)
	}
	if departmentFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.department_id = ?", departmentFilter)
	}
	if sectionFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.section_id = ?", sectionFilter)
	}
	if designationFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.designation_id = ?", designationFilter)
	}
	if lineFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.line_id = ?", lineFilter)
	}
	if groupFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.group_id = ?", groupFilter)
	}

	var attendances []models.Attendance
	if err := baseQuery.Order("LENGTH(attendances.employee_id) ASC, attendances.employee_id ASC, attendances.date ASC").
		Find(&attendances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build last continuous absent count per employee:
	// Starting from the latest date in the range, count consecutive absent days backwards.
	lastContAbsentMap := h.buildLastContinuousAbsentMap(startDate, endDate, companyFilter)

	f := excelize.NewFile()
	sheet := "Absent Report"
	index, _ := f.GetSheetIndex("Sheet1")
	f.SetActiveSheet(index)
	f.SetSheetName("Sheet1", sheet)

	nCols := 5
	cols := []struct {
		header string
		width  float64
	}{
		{"Employee ID", 15},
		{"Name", 28},
		{"Designation", 20},
		{"Last Cont. Absent", 11},
		{"Status", 11},
	}

	borderColor := "808080"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	companyNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	normalCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})
	dataCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	dataLeft, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	redStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "FF0000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	parsedDate, _ := time.Parse("2006-01-02", startDate)
	dateDisplay := parsedDate.Format("02 Jan 2006")

	endCol := colNameAttendance(nCols)

	// Row 1: Company Name
	f.SetCellValue(sheet, "A1", companyName)
	f.MergeCell(sheet, "A1", endCol+"1")
	f.SetCellStyle(sheet, "A1", endCol+"1", companyNameStyle)
	f.SetRowHeight(sheet, 1, 32)

	// Row 2: Address
	f.SetCellValue(sheet, "A2", companyAddress)
	f.MergeCell(sheet, "A2", endCol+"2")
	f.SetCellStyle(sheet, "A2", endCol+"2", normalCenter)
	f.SetRowHeight(sheet, 2, 20)

	// Row 3: Report Name
	f.SetCellValue(sheet, "A3", "ABSENT ATTENDANCE REPORT")
	f.MergeCell(sheet, "A3", endCol+"3")
	f.SetCellStyle(sheet, "A3", endCol+"3", normalCenter)
	f.SetRowHeight(sheet, 3, 20)

	// Row 4: Date Range
	f.SetCellValue(sheet, "A4", "Date: "+dateDisplay)
	f.MergeCell(sheet, "A4", endCol+"4")
	f.SetCellStyle(sheet, "A4", endCol+"4", normalCenter)
	f.SetRowHeight(sheet, 4, 20)

	// Row 5: Column Headers
	for i, c := range cols {
		cell := colNameAttendance(i+1) + "5"
		f.SetCellValue(sheet, cell, c.header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		f.SetColWidth(sheet, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheet, 5, 36)

	// Data rows
	for rowIdx, a := range attendances {
		row := rowIdx + 6
		svc := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataCenter)
		}
		svl := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataLeft)
		}

		svc(1, a.EmployeeID)
		svl(2, a.Employee.NameEn)

		designation := ""
		if a.Employee.DesignationRef != nil {
			designation = a.Employee.DesignationRef.Name
		}
		svl(3, designation)

		svc(4, fmt.Sprintf("%d", lastContAbsentMap[a.EmployeeID]))

		f.SetCellValue(sheet, colNameAttendance(5)+strconv.Itoa(row), "A")
		f.SetCellStyle(sheet, colNameAttendance(5)+strconv.Itoa(row), colNameAttendance(5)+strconv.Itoa(row), redStyle)

		f.SetRowHeight(sheet, row, 25)
	}

	// Footer
	lastRow := len(attendances) + 5
	footerRow := lastRow + 2
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	uniqueCount := len(lastContAbsentMap)
	f.SetCellValue(sheet, "A"+strconv.Itoa(footerRow), fmt.Sprintf("Total Absent: %d Employees", uniqueCount))
	f.MergeCell(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow))
	f.SetCellStyle(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow), footerStyle)
	f.SetRowHeight(sheet, footerRow, 22)

	// --- Grouped Sheets ---
	addGroupedAbsentSheet(f, "Department Wise", companyName, companyAddress, dateDisplay, attendances, lastContAbsentMap, func(a models.Attendance) string {
		if a.Employee.Department != nil {
			return a.Employee.Department.Name
		}
		return "-"
	})
	addGroupedAbsentSheet(f, "Section Wise", companyName, companyAddress, dateDisplay, attendances, lastContAbsentMap, func(a models.Attendance) string {
		if a.Employee.SectionRef != nil {
			return a.Employee.SectionRef.Name
		}
		return "-"
	})
	addGroupedAbsentSheet(f, "Designation Wise", companyName, companyAddress, dateDisplay, attendances, lastContAbsentMap, func(a models.Attendance) string {
		if a.Employee.DesignationRef != nil {
			return a.Employee.DesignationRef.Name
		}
		return "-"
	})
	addGroupedAbsentSheet(f, "Line Wise", companyName, companyAddress, dateDisplay, attendances, lastContAbsentMap, func(a models.Attendance) string {
		if a.Employee.LineRef != nil {
			return a.Employee.LineRef.Name
		}
		return "-"
	})

	// --- Page Setup: A4 Portrait + No Gridlines for all sheets ---
	for _, s := range f.GetSheetList() {
		orientation := "portrait"
		paperSize := 9
		fitWidth := 1
		fitHeight := 0
		f.SetPageLayout(s, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitWidth,
			FitToHeight: &fitHeight,
		})
		f.SetPageMargins(s, &excelize.PageLayoutMarginsOptions{
			Left:   func(f float64) *float64 { return &f }(0.3),
			Right:  func(f float64) *float64 { return &f }(0.3),
			Top:    func(f float64) *float64 { return &f }(0.4),
			Bottom: func(f float64) *float64 { return &f }(0.4),
			Header: func(f float64) *float64 { return &f }(0),
			Footer: func(f float64) *float64 { return &f }(0),
		})
		f.SetSheetView(s, -1, &excelize.ViewOptions{ShowGridLines: func(b bool) *bool { return &b }(false)})
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=absent_report_%s.xlsx", startDate))
	f.Write(c.Writer)
}

// ExportMissingAttendanceExcel godoc
//
//	@Summary      Export missing attendance to Excel
//	@Description  Export employees with missing punches (only check_in or only check_out) to Excel
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        start_date query string true "Start date (YYYY-MM-DD)"
//	@Param        end_date query string true "End date (YYYY-MM-DD)"
//	@Param        company_id  query string false "Filter by company"
//	@Param        department_id  query string false "Filter by department"
//	@Param        section_id  query string false "Filter by section"
//	@Param        designation_id  query string false "Filter by designation"
//	@Param        line_id  query string false "Filter by line"
//	@Param        group_id  query string false "Filter by group"
//	@Success      200  {file}  binary
//	@Router       /attendance/missing/export/excel [get]
func (h *AttendanceHandler) ExportMissingAttendanceExcel(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if endDate == "" {
		endDate = startDate
	}

	companyFilter := c.Query("company_id")
	departmentFilter := c.Query("department_id")
	sectionFilter := c.Query("section_id")
	designationFilter := c.Query("designation_id")
	lineFilter := c.Query("line_id")
	groupFilter := c.Query("group_id")

	var company models.Company
	database.DB.First(&company)

	baseQuery := database.DB.Model(&models.Attendance{}).
		Preload("Employee.DesignationRef").
		Preload("Employee.Department").
		Preload("Employee.SectionRef").
		Preload("Employee.LineRef").
		Preload("Employee").
		Where(`(
			(attendances.check_in IS NULL AND attendances.check_out IS NOT NULL)
			OR
			(attendances.check_in IS NOT NULL AND attendances.check_out IS NULL)
			OR
			(
			  attendances.check_in IS NOT NULL
			  AND attendances.check_out IS NOT NULL
			  AND attendances.shift_id IS NOT NULL
			  AND attendances.check_out < (attendances.date::text || ' ' || (SELECT s.start_time FROM shifts s WHERE s.id = attendances.shift_id))::timestamp + INTERVAL '2 hours'
			)
			OR
			(
			  attendances.check_in IS NOT NULL
			  AND attendances.check_out IS NOT NULL
			  AND attendances.shift_id IS NOT NULL
			  AND attendances.check_out < (attendances.date::text || ' ' || (SELECT s.end_time FROM shifts s WHERE s.id = attendances.shift_id))::timestamp
			)
		) AND attendances.date BETWEEN ? AND ? AND attendances.deleted_at IS NULL`, startDate, endDate)

	if companyFilter != "" {
		baseQuery = baseQuery.Where("attendances.company_id = ?", companyFilter)
	}
	if departmentFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.department_id = ?", departmentFilter)
	}
	if sectionFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.section_id = ?", sectionFilter)
	}
	if designationFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.designation_id = ?", designationFilter)
	}
	if lineFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.line_id = ?", lineFilter)
	}
	if groupFilter != "" {
		baseQuery = baseQuery.Joins("JOIN employees ON employees.employee_id = attendances.employee_id").
			Where("employees.group_id = ?", groupFilter)
	}

	var attendances []models.Attendance
	if err := baseQuery.Order("LENGTH(attendances.employee_id) ASC, attendances.employee_id ASC, attendances.date ASC").
		Find(&attendances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "Missing Attendance"
	index, _ := f.GetSheetIndex("Sheet1")
	f.SetActiveSheet(index)
	f.SetSheetName("Sheet1", sheet)

	nCols := 5
	cols := []struct {
		header string
		width  float64
	}{
		{"Employee ID", 15},
		{"Name", 25},
		{"Designation", 25},
		{"Check In", 14},
		{"Check Out", 14},
	}

	borderColor := "808080"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	companyNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	normalCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})
	dataCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	dataLeft, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})

	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	parsedDate, _ := time.Parse("2006-01-02", startDate)
	dateDisplay := parsedDate.Format("02 Jan 2006")

	endCol := colNameAttendance(nCols)

	f.SetCellValue(sheet, "A1", companyName)
	f.MergeCell(sheet, "A1", endCol+"1")
	f.SetCellStyle(sheet, "A1", endCol+"1", companyNameStyle)
	f.SetRowHeight(sheet, 1, 32)

	f.SetCellValue(sheet, "A2", companyAddress)
	f.MergeCell(sheet, "A2", endCol+"2")
	f.SetCellStyle(sheet, "A2", endCol+"2", normalCenter)
	f.SetRowHeight(sheet, 2, 20)

	f.SetCellValue(sheet, "A3", "MISSING ATTENDANCE REPORT")
	f.MergeCell(sheet, "A3", endCol+"3")
	f.SetCellStyle(sheet, "A3", endCol+"3", normalCenter)
	f.SetRowHeight(sheet, 3, 20)

	f.SetCellValue(sheet, "A4", "Date: "+dateDisplay)
	f.MergeCell(sheet, "A4", endCol+"4")
	f.SetCellStyle(sheet, "A4", endCol+"4", normalCenter)
	f.SetRowHeight(sheet, 4, 20)

	for i, c := range cols {
		cell := colNameAttendance(i+1) + "5"
		f.SetCellValue(sheet, cell, c.header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		f.SetColWidth(sheet, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheet, 5, 36)

	for rowIdx, a := range attendances {
		row := rowIdx + 6
		svc := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataCenter)
		}
		svl := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataLeft)
		}

		svc(1, a.EmployeeID)
		svl(2, a.Employee.NameEn)

		designation := ""
		if a.Employee.DesignationRef != nil {
			designation = a.Employee.DesignationRef.Name
		}
		svl(3, designation)

		checkIn := "-"
		if a.CheckIn != nil {
			checkIn = a.CheckIn.Format("15:04")
		} else if a.CheckOut != nil {
			checkIn = "07:55"
		}
		svc(4, checkIn)

		checkOut := "-"
		if a.CheckOut != nil {
			checkOut = a.CheckOut.Format("15:04")
		}
		svc(5, checkOut)

		f.SetRowHeight(sheet, row, 30)
	}

	lastRow := len(attendances) + 5
	footerRow := lastRow + 2
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	uniqueCount := len(attendances)
	f.SetCellValue(sheet, "A"+strconv.Itoa(footerRow), fmt.Sprintf("Total Missing: %d Records", uniqueCount))
	f.MergeCell(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow))
	f.SetCellStyle(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow), footerStyle)
	f.SetRowHeight(sheet, footerRow, 22)

	for _, s := range f.GetSheetList() {
		orientation := "portrait"
		paperSize := 9
		fitWidth := 1
		fitHeight := 0
		f.SetPageLayout(s, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitWidth,
			FitToHeight: &fitHeight,
		})
		f.SetPageMargins(s, &excelize.PageLayoutMarginsOptions{
			Left:   func(f float64) *float64 { return &f }(0.25),
			Right:  func(f float64) *float64 { return &f }(0.25),
			Top:    func(f float64) *float64 { return &f }(0.75),
			Bottom: func(f float64) *float64 { return &f }(0.75),
			Header: func(f float64) *float64 { return &f }(0.3),
			Footer: func(f float64) *float64 { return &f }(0.3),
		})
		f.SetHeaderFooter(s, &excelize.HeaderFooterOptions{
			OddFooter: "&LProduction manager&CAdmin (A.G.M)&RApproved By",
		})
		f.SetSheetView(s, -1, &excelize.ViewOptions{ShowGridLines: func(b bool) *bool { return &b }(false)})
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=missing_attendance_%s.xlsx", startDate))
	f.Write(c.Writer)
}

// ExportOvertimeExcel godoc
//
//	@Summary      Export overtime sheet to Excel
//	@Description  Export employee overtime records for a date range to Excel
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        start_date    query string true "Start date (YYYY-MM-DD)"
//	@Param        end_date      query string true "End date (YYYY-MM-DD)"
//	@Param        company_id    query string false "Filter by company"
//	@Param        department_id query string false "Filter by department"
//	@Param        section_id    query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id       query string false "Filter by line"
//	@Param        group_id      query string false "Filter by group"
//	@Success      200  {file}  binary
//	@Router       /attendance/overtime/export/excel [get]
func (h *AttendanceHandler) ExportOvertimeExcel(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", startDate)
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")
	groupID := c.Query("group_id")

	records, err := h.attendanceRepo.Overtime(startDate, endDate, companyID, departmentID, sectionID, designationID, lineID, groupID, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	database.DB.First(&company)

	f := excelize.NewFile()
	sheet := "Over Time Sheet"
	index, _ := f.GetSheetIndex("Sheet1")
	f.SetActiveSheet(index)
	f.SetSheetName("Sheet1", sheet)

	nCols := 7
	cols := []struct {
		header string
		width  float64
	}{
		{"Sl No", 8},
		{"Employee Name", 28},
		{"Employee ID", 15},
		{"Date", 14},
		{"In Time", 14},
		{"Out Time", 14},
		{"Over Time (Hr)", 13},
	}

	borderColor := "808080"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	companyNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	normalCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})
	dataCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	dataLeft, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})

	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	parsedHeaderDate, _ := time.Parse("2006-01-02", startDate)
	dateDisplay := parsedHeaderDate.Format("02-01-2006")

	endCol := colNameAttendance(nCols)

	f.SetCellValue(sheet, "A1", companyName)
	f.MergeCell(sheet, "A1", endCol+"1")
	f.SetCellStyle(sheet, "A1", endCol+"1", companyNameStyle)
	f.SetRowHeight(sheet, 1, 32)

	f.SetCellValue(sheet, "A2", companyAddress)
	f.MergeCell(sheet, "A2", endCol+"2")
	f.SetCellStyle(sheet, "A2", endCol+"2", normalCenter)
	f.SetRowHeight(sheet, 2, 20)

	f.SetCellValue(sheet, "A3", "OVER TIME SHEET")
	f.MergeCell(sheet, "A3", endCol+"3")
	f.SetCellStyle(sheet, "A3", endCol+"3", normalCenter)
	f.SetRowHeight(sheet, 3, 20)

	f.SetCellValue(sheet, "A4", "Date: "+dateDisplay)
	f.MergeCell(sheet, "A4", endCol+"4")
	f.SetCellStyle(sheet, "A4", endCol+"4", normalCenter)
	f.SetRowHeight(sheet, 4, 20)

	for i, c := range cols {
		cell := colNameAttendance(i+1) + "5"
		f.SetCellValue(sheet, cell, c.header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		f.SetColWidth(sheet, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheet, 5, 36)

	for rowIdx, rec := range records {
		row := rowIdx + 6
		svc := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataCenter)
		}
		svl := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataLeft)
		}

		svc(1, strconv.Itoa(rowIdx+1))
		svl(2, toString(rec["employee_name"]))
		svc(3, toString(rec["employee_id"]))
		dateStr := toString(rec["date"])
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			dateStr = t.Format("02-01-2006")
		}
		svc(4, dateStr)

		checkIn := "-"
		if v, ok := rec["check_in"].(string); ok && v != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				checkIn = t.Format("15:04")
			}
		}
		svc(5, checkIn)

		checkOut := "-"
		if v, ok := rec["check_out"].(string); ok && v != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				checkOut = t.Format("15:04")
			}
		}
		svc(6, checkOut)

		svc(7, toString(rec["over_time"]))

		f.SetRowHeight(sheet, row, 25)
	}

	lastRow := len(records) + 5
	footerRow := lastRow + 2
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellValue(sheet, "A"+strconv.Itoa(footerRow), fmt.Sprintf("Total Records: %d", len(records)))
	f.MergeCell(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow))
	f.SetCellStyle(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow), footerStyle)
	f.SetRowHeight(sheet, footerRow, 22)

	for _, s := range f.GetSheetList() {
		orientation := "portrait"
		paperSize := 9
		fitWidth := 1
		fitHeight := 0
		f.SetPageLayout(s, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitWidth,
			FitToHeight: &fitHeight,
		})
		f.SetPageMargins(s, &excelize.PageLayoutMarginsOptions{
			Left:   func(f float64) *float64 { return &f }(0.25),
			Right:  func(f float64) *float64 { return &f }(0.25),
			Top:    func(f float64) *float64 { return &f }(0.75),
			Bottom: func(f float64) *float64 { return &f }(0.75),
			Header: func(f float64) *float64 { return &f }(0.3),
			Footer: func(f float64) *float64 { return &f }(0.3),
		})
		f.SetSheetView(s, -1, &excelize.ViewOptions{ShowGridLines: func(b bool) *bool { return &b }(false)})
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=overtime_sheet_%s.xlsx", startDate))
	f.Write(c.Writer)
}

// ExportOvertimeSummaryExcel godoc
//
//	@Summary      Export overtime summary to Excel
//	@Description  Export department-level overtime summary to Excel
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        start_date    query string true "Start date (YYYY-MM-DD)"
//	@Param        end_date      query string true "End date (YYYY-MM-DD)"
//	@Param        company_id    query string false "Filter by company"
//	@Param        department_id query string false "Filter by department"
//	@Param        section_id    query string false "Filter by section"
//	@Param        designation_id query string false "Filter by designation"
//	@Param        line_id       query string false "Filter by line"
//	@Param        group_id      query string false "Filter by group"
//	@Success      200  {file}  binary
//	@Router       /attendance/overtime-summary/export/excel [get]
func (h *AttendanceHandler) ExportOvertimeSummaryExcel(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", startDate)
	companyID := c.Query("company_id")
	groupBy := c.DefaultQuery("group_by", "department")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")
	lineID := c.Query("line_id")

	summaries, err := h.attendanceRepo.OvertimeSummaryGrouped(startDate, endDate, companyID, groupBy, departmentID, sectionID, designationID, lineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	groupLabel := "Department"
	switch groupBy {
	case "section":
		groupLabel = "Section"
	case "designation":
		groupLabel = "Designation"
	case "line":
		groupLabel = "Line"
	}

	var company models.Company
	database.DB.First(&company)

	f := excelize.NewFile()
	sheet := "Over Time Summary"
	index, _ := f.GetSheetIndex("Sheet1")
	f.SetActiveSheet(index)
	f.SetSheetName("Sheet1", sheet)

	nCols := 4
	cols := []struct {
		header string
		width  float64
	}{
		{"Sl No", 8},
		{groupLabel, 32},
		{"Employees", 14},
		{"Total Hours", 14},
	}

	borderColor := "808080"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	companyNameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	normalCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    thinBorder,
	})
	dataCenter, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	dataLeft, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"},
		Border:    thinBorder,
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})

	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	parsedHeaderDate, _ := time.Parse("2006-01-02", startDate)
	dateDisplay := parsedHeaderDate.Format("02-01-2006")

	endCol := colNameAttendance(nCols)

	f.SetCellValue(sheet, "A1", companyName)
	f.MergeCell(sheet, "A1", endCol+"1")
	f.SetCellStyle(sheet, "A1", endCol+"1", companyNameStyle)
	f.SetRowHeight(sheet, 1, 32)

	f.SetCellValue(sheet, "A2", companyAddress)
	f.MergeCell(sheet, "A2", endCol+"2")
	f.SetCellStyle(sheet, "A2", endCol+"2", normalCenter)
	f.SetRowHeight(sheet, 2, 20)

	f.SetCellValue(sheet, "A3", "OVER TIME SUMMARY ("+groupLabel+")")
	f.MergeCell(sheet, "A3", endCol+"3")
	f.SetCellStyle(sheet, "A3", endCol+"3", normalCenter)
	f.SetRowHeight(sheet, 3, 20)

	f.SetCellValue(sheet, "A4", "Date: "+dateDisplay)
	f.MergeCell(sheet, "A4", endCol+"4")
	f.SetCellStyle(sheet, "A4", endCol+"4", normalCenter)
	f.SetRowHeight(sheet, 4, 20)

	for i, c := range cols {
		cell := colNameAttendance(i+1) + "5"
		f.SetCellValue(sheet, cell, c.header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		f.SetColWidth(sheet, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheet, 5, 36)

	for rowIdx, rec := range summaries {
		row := rowIdx + 6
		svc := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataCenter)
		}
		svl := func(c int, v string) {
			f.SetCellValue(sheet, colNameAttendance(c)+strconv.Itoa(row), v)
			f.SetCellStyle(sheet, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataLeft)
		}

		svc(1, strconv.Itoa(rowIdx+1))
		svl(2, toString(rec["name"]))
		employeeCount := toString(fmt.Sprintf("%v", rec["employee_count"]))
		if employeeCount == "<nil>" {
			employeeCount = "0"
		}
		svc(3, employeeCount)
		if totalHours, ok := rec["total_hours"].(float64); ok {
			svc(4, fmt.Sprintf("%.2f", totalHours))
		} else {
			svc(4, "0.00")
		}

		f.SetRowHeight(sheet, row, 25)
	}

	lastRow := len(summaries) + 5
	footerRow := lastRow + 2
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	totalEmp := 0
	for _, rec := range summaries {
		if c, ok := rec["employee_count"].(float64); ok {
			totalEmp += int(c)
		}
	}
	f.SetCellValue(sheet, "A"+strconv.Itoa(footerRow), fmt.Sprintf("Total Departments: %d | Total Employees: %d", len(summaries), totalEmp))
	f.MergeCell(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow))
	f.SetCellStyle(sheet, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow), footerStyle)
	f.SetRowHeight(sheet, footerRow, 22)

	for _, s := range f.GetSheetList() {
		orientation := "portrait"
		paperSize := 9
		fitWidth := 1
		fitHeight := 0
		f.SetPageLayout(s, &excelize.PageLayoutOptions{
			Orientation: &orientation,
			Size:        &paperSize,
			FitToWidth:  &fitWidth,
			FitToHeight: &fitHeight,
		})
		f.SetPageMargins(s, &excelize.PageLayoutMarginsOptions{
			Left:   func(f float64) *float64 { return &f }(0.25),
			Right:  func(f float64) *float64 { return &f }(0.25),
			Top:    func(f float64) *float64 { return &f }(0.75),
			Bottom: func(f float64) *float64 { return &f }(0.75),
			Header: func(f float64) *float64 { return &f }(0.3),
			Footer: func(f float64) *float64 { return &f }(0.3),
		})
		f.SetSheetView(s, -1, &excelize.ViewOptions{ShowGridLines: func(b bool) *bool { return &b }(false)})
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=overtime_summary_%s.xlsx", startDate))
	f.Write(c.Writer)
}

// buildLastContinuousAbsentMap counts the consecutive absent days extending backward
// from endDate for each employee. It finds each employee's last present date (any date <= endDate)
// and counts absent days after that date up to endDate. The search is unbounded backward — if
// the employee has no present date in the record, ALL absent days up to endDate are counted.
// If companyID is non-empty, only employees of that company are considered.
//
// The key difference from using a BETWEEN-bound approach: this correctly handles single-day
// ranges (e.g., today-only reports) by looking at ALL absent days after the last present,
// not just those within an arbitrary start/end window.
func (h *AttendanceHandler) buildLastContinuousAbsentMap(startDate, endDate, companyID string) map[string]int {
	type countRow struct {
		EmployeeID string `gorm:"column:employee_id"`
		Count      int    `gorm:"column:count"`
	}
	var counts []countRow

	query := `
		SELECT a.employee_id, COUNT(*)::int AS count
		FROM attendances a
		LEFT JOIN LATERAL (
			SELECT MAX(date) AS last_present
			FROM attendances
			WHERE employee_id = a.employee_id
			  AND date <= ?
			  AND status IN ('present','late','half_day')
			  AND deleted_at IS NULL
		) lp ON true
		WHERE a.date <= ?
		  AND a.status = 'absent'
		  AND a.deleted_at IS NULL
		  AND (lp.last_present IS NULL OR a.date > lp.last_present)
	`
	args := []interface{}{endDate, endDate}
	if companyID != "" {
		query += ` AND a.company_id = ?`
		args = append(args, companyID)
	}
	query += ` GROUP BY a.employee_id`

	database.DB.Raw(query, args...).Scan(&counts)

	result := make(map[string]int, len(counts))
	for _, c := range counts {
		result[c.EmployeeID] = c.Count
	}
	return result
}

func addGroupedAbsentSheet(f *excelize.File, sheetName, companyName, companyAddress, dateDisplay string, attendances []models.Attendance, lastContAbsentMap map[string]int, groupFn func(models.Attendance) string) {
	f.NewSheet(sheetName)

	nCols := 5
	cols := []struct {
		header string
		width  float64
	}{
		{"Employee ID", 15},
		{"Name", 28},
		{"Designation", 20},
		{"Last Cont. Absent", 11},
		{"Status", 11},
	}

	thinBorder := []excelize.Border{
		{Type: "left", Color: "808080", Style: 1},
		{Type: "top", Color: "808080", Style: 1},
		{Type: "bottom", Color: "808080", Style: 1},
		{Type: "right", Color: "808080", Style: 1},
	}

	companyNameStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	normalCenter, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}, Border: thinBorder})
	dataCenter, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"}, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}})
	dataLeft, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"}, Border: thinBorder, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}})
	redStyleG, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "FF0000"}, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}})
	groupHeaderStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"}})

	endCol := colNameAttendance(nCols)

	// Header rows
	f.SetCellValue(sheetName, "A1", companyName)
	f.MergeCell(sheetName, "A1", endCol+"1")
	f.SetCellStyle(sheetName, "A1", endCol+"1", companyNameStyle)
	f.SetRowHeight(sheetName, 1, 32)

	f.SetCellValue(sheetName, "A2", companyAddress)
	f.MergeCell(sheetName, "A2", endCol+"2")
	f.SetCellStyle(sheetName, "A2", endCol+"2", normalCenter)
	f.SetRowHeight(sheetName, 2, 20)

	f.SetCellValue(sheetName, "A3", sheetName+" ABSENT REPORT")
	f.MergeCell(sheetName, "A3", endCol+"3")
	f.SetCellStyle(sheetName, "A3", endCol+"3", normalCenter)
	f.SetRowHeight(sheetName, 3, 20)

	f.SetCellValue(sheetName, "A4", "Date: "+dateDisplay)
	f.MergeCell(sheetName, "A4", endCol+"4")
	f.SetCellStyle(sheetName, "A4", endCol+"4", normalCenter)
	f.SetRowHeight(sheetName, 4, 20)

	// Column headers
	for i, c := range cols {
		cell := colNameAttendance(i+1) + "5"
		f.SetCellValue(sheetName, cell, c.header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
		f.SetColWidth(sheetName, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheetName, 5, 36)

	// Group data
	grouped := make(map[string][]models.Attendance)
	var groupOrder []string
	for _, a := range attendances {
		name := groupFn(a)
		if name == "" {
			name = "-"
		}
		if _, ok := grouped[name]; !ok {
			groupOrder = append(groupOrder, name)
		}
		grouped[name] = append(grouped[name], a)
	}

	row := 6
	totalAbsent := 0
	for _, groupName := range groupOrder {
		list := grouped[groupName]

		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), groupName+" ("+fmt.Sprintf("%d", len(list))+")")
		f.MergeCell(sheetName, "A"+strconv.Itoa(row), endCol+strconv.Itoa(row))
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), endCol+strconv.Itoa(row), groupHeaderStyle)
		f.SetRowHeight(sheetName, row, 22)
		row++

		for _, a := range list {
			svc := func(c int, v string) {
				f.SetCellValue(sheetName, colNameAttendance(c)+strconv.Itoa(row), v)
				f.SetCellStyle(sheetName, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataCenter)
			}
			svl := func(c int, v string) {
				f.SetCellValue(sheetName, colNameAttendance(c)+strconv.Itoa(row), v)
				f.SetCellStyle(sheetName, colNameAttendance(c)+strconv.Itoa(row), colNameAttendance(c)+strconv.Itoa(row), dataLeft)
			}

			svc(1, a.EmployeeID)
			svl(2, a.Employee.NameEn)
			d := ""
			if a.Employee.DesignationRef != nil {
				d = a.Employee.DesignationRef.Name
			}
			svl(3, d)
			svc(4, fmt.Sprintf("%d", lastContAbsentMap[a.EmployeeID]))
			f.SetCellValue(sheetName, colNameAttendance(5)+strconv.Itoa(row), "A")
			f.SetCellStyle(sheetName, colNameAttendance(5)+strconv.Itoa(row), colNameAttendance(5)+strconv.Itoa(row), redStyleG)

			totalAbsent++
			f.SetRowHeight(sheetName, row, 25)
			row++
		}
	}

	footerRow := row + 1
	footerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"}})
	f.SetCellValue(sheetName, "A"+strconv.Itoa(footerRow), fmt.Sprintf("Total Absent: %d", totalAbsent))
	f.MergeCell(sheetName, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow))
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(footerRow), endCol+strconv.Itoa(footerRow), footerStyle)
	f.SetRowHeight(sheetName, footerRow, 22)
}

// ExportSummaryExcel godoc
//
//	@Summary      Export daily summary to Excel
//
// # CustomSummaryReport godoc
//
// @Summary      Custom summary report
// @Description  Generate a custom mixed summary report with configurable sections
// @Tags         Attendance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body repository.CustomSectionFilter true "Filter params"
// @Param        company_id query string true "Company ID"
// @Param        date query string true "Date (YYYY-MM-DD)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /attendance/custom-summary [post]
func (h *AttendanceHandler) CustomSummaryReport(c *gin.Context) {
	companyID := c.Query("company_id")
	date := c.Query("date")
	if companyID == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id and date are required"})
		return
	}

	var sections []repository.CustomSectionFilter
	if err := c.ShouldBindJSON(&sections); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	type sectionResult struct {
		Name    string                   `json:"name"`
		Type    string                   `json:"type"`
		SubRows []map[string]interface{} `json:"sub_rows"`
		Present int64                    `json:"present"`
		Absent  int64                    `json:"absent"`
		Leave   int64                    `json:"leave"`
		Others  int64                    `json:"others"`
		Total   int64                    `json:"total"`
	}

	var report []sectionResult
	var grandPresent, grandAbsent, grandLeave, grandOthers, grandTotal int64

	for _, sec := range sections {
		rows, err := h.attendanceRepo.CustomSummarySection(companyID, date, date, sec)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var secPresent, secAbsent, secLeave, secOthers, secTotal int64
		var subRows []map[string]interface{}

		for _, row := range rows {
			name := ""
			if n, ok := row["name"]; ok {
				name = fmt.Sprintf("%v", n)
			}
			p := toInt64(row["present"])
			a := toInt64(row["absent"])
			l := toInt64(row["on_leave"])
			o := toInt64(row["others"])
			t := toInt64(row["total"])

			if sec.GroupByLine && name != "" && name != "Total" {
				subRows = append(subRows, map[string]interface{}{
					"name":    name,
					"present": p, "absent": a, "leave": l, "others": o, "total": t,
				})
			}
			secPresent += p
			secAbsent += a
			secLeave += l
			secOthers += o
			secTotal += t
		}

		report = append(report, sectionResult{
			Name:    sec.Name,
			Type:    sec.Type,
			SubRows: subRows,
			Present: secPresent, Absent: secAbsent,
			Leave: secLeave, Others: secOthers, Total: secTotal,
		})
		grandPresent += secPresent
		grandAbsent += secAbsent
		grandLeave += secLeave
		grandOthers += secOthers
		grandTotal += secTotal
	}

	c.JSON(http.StatusOK, gin.H{
		"sections": report,
		"grand_totals": map[string]int64{
			"present": grandPresent, "absent": grandAbsent,
			"leave": grandLeave, "others": grandOthers, "total": grandTotal,
		},
	})
}

// @Description  Export attendance summary grouped by Department, Section, Designation, and Line
// @Tags         Attendance
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        start_date query string true "Start date (YYYY-MM-DD)"
// @Param        end_date query string true "End date (YYYY-MM-DD)"
// @Success      200  {file}  binary
// @Router       /attendance/summary/export/excel [get]
func (h *AttendanceHandler) ExportSummaryExcel(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if endDate == "" {
		endDate = startDate
	}

	var company models.Company
	database.DB.First(&company)

	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "Company Name"
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Company Address"
	}

	parsedStart, _ := time.Parse("2006-01-02", startDate)
	parsedEnd, _ := time.Parse("2006-01-02", endDate)
	dateRange := parsedStart.Format("02 Jan 2006") + " to " + parsedEnd.Format("02 Jan 2006")

	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Department Wise")

	groups := []struct{ key, label string }{
		{"department", "Department Wise"},
		{"section", "Section Wise"},
		{"designation", "Designation Wise"},
		{"line", "Line Wise"},
	}

	for idx, g := range groups {
		if idx > 0 {
			f.NewSheet(g.label)
		}
		result, err := h.attendanceRepo.SummaryByGroup(startDate, endDate, g.key, "", "", "", "", "", "", "", "", "")
		if err != nil {
			continue
		}
		addDailySummarySheet(f, g.label, companyName, companyAddress, dateRange, result)
	}

	for _, s := range f.GetSheetList() {
		o := "portrait"
		ps := 9
		fw := 1
		fh := 0
		f.SetPageLayout(s, &excelize.PageLayoutOptions{Orientation: &o, Size: &ps, FitToWidth: &fw, FitToHeight: &fh})
		f.SetPageMargins(s, &excelize.PageLayoutMarginsOptions{Left: ptr(0.3), Right: ptr(0.3), Top: ptr(0.4), Bottom: ptr(0.4)})
		f.SetSheetView(s, -1, &excelize.ViewOptions{ShowGridLines: ptrBool(false)})
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=daily_summary_%s_%s.xlsx", startDate, endDate))
	f.Write(c.Writer)
}

func ptr(v float64) *float64 { return &v }
func ptrBool(v bool) *bool   { return &v }

func addDailySummarySheet(f *excelize.File, sheetName, companyName, companyAddress, dateRange string, data []map[string]interface{}) {
	nCols := 9
	cols := []struct {
		header string
		width  float64
	}{
		{"SL", 5}, {"Name", 30}, {"Present", 12}, {"Late", 12}, {"Absent", 12},
		{"Half Day", 12}, {"On Leave", 12}, {"Weekend", 12}, {"Total", 12},
	}

	thinBorder := []excelize.Border{
		{Type: "left", Color: "333333", Style: 1}, {Type: "top", Color: "333333", Style: 1},
		{Type: "bottom", Color: "333333", Style: 1}, {Type: "right", Color: "333333", Style: 1},
	}

	cnStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 20, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	ncStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 11, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	hdStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 11, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}, Border: thinBorder})
	dcStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"}, Border: thinBorder, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	dlStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"}, Border: thinBorder, Alignment: &excelize.Alignment{Vertical: "center"}})
	ftStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"}, Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"}})

	endCol := colNameAttendance(nCols)

	f.SetCellValue(sheetName, "A1", companyName)
	f.MergeCell(sheetName, "A1", endCol+"1")
	f.SetCellStyle(sheetName, "A1", endCol+"1", cnStyle)
	f.SetRowHeight(sheetName, 1, 32)
	f.SetCellValue(sheetName, "A2", companyAddress)
	f.MergeCell(sheetName, "A2", endCol+"2")
	f.SetCellStyle(sheetName, "A2", endCol+"2", ncStyle)
	f.SetRowHeight(sheetName, 2, 20)
	f.SetCellValue(sheetName, "A3", sheetName+" SUMMARY")
	f.MergeCell(sheetName, "A3", endCol+"3")
	f.SetCellStyle(sheetName, "A3", endCol+"3", ncStyle)
	f.SetRowHeight(sheetName, 3, 20)
	f.SetCellValue(sheetName, "A4", "Date: "+dateRange)
	f.MergeCell(sheetName, "A4", endCol+"4")
	f.SetCellStyle(sheetName, "A4", endCol+"4", ncStyle)
	f.SetRowHeight(sheetName, 4, 20)

	for i, c := range cols {
		cell := colNameAttendance(i+1) + "5"
		f.SetCellValue(sheetName, cell, c.header)
		f.SetCellStyle(sheetName, cell, cell, hdStyle)
		f.SetColWidth(sheetName, colNameAttendance(i+1), colNameAttendance(i+1), c.width)
	}
	f.SetRowHeight(sheetName, 5, 24)

	var gp, gl, ga, ghd, gol, gw, gt int64
	for i, row := range data {
		r := i + 6
		svc := func(c int, v string) {
			f.SetCellValue(sheetName, colNameAttendance(c)+strconv.Itoa(r), v)
			f.SetCellStyle(sheetName, colNameAttendance(c)+strconv.Itoa(r), colNameAttendance(c)+strconv.Itoa(r), dcStyle)
		}
		svl := func(c int, v string) {
			f.SetCellValue(sheetName, colNameAttendance(c)+strconv.Itoa(r), v)
			f.SetCellStyle(sheetName, colNameAttendance(c)+strconv.Itoa(r), colNameAttendance(c)+strconv.Itoa(r), dlStyle)
		}

		svc(1, fmt.Sprintf("%d", i+1))
		n := ""
		if v, ok := row["name"]; ok && v != nil {
			n = fmt.Sprintf("%v", v)
		}
		svl(2, n)

		gv := func(k string) int64 {
			if v, ok := row[k]; ok && v != nil {
				switch x := v.(type) {
				case int64:
					return x
				case float64:
					return int64(x)
				}
			}
			return 0
		}
		p := gv("present")
		l := gv("late")
		a := gv("absent")
		hd := gv("half_day")
		ol := gv("on_leave")
		we := gv("weekend")
		t := gv("total")

		svc(3, fmt.Sprintf("%d", p))
		svc(4, fmt.Sprintf("%d", l))
		svc(5, fmt.Sprintf("%d", a))
		svc(6, fmt.Sprintf("%d", hd))
		svc(7, fmt.Sprintf("%d", ol))
		svc(8, fmt.Sprintf("%d", we))
		svc(9, fmt.Sprintf("%d", t))

		gp += p
		gl += l
		ga += a
		ghd += hd
		gol += ol
		gw += we
		gt += t
		f.SetRowHeight(sheetName, r, 20)
	}

	fr := len(data) + 7
	f.SetCellValue(sheetName, "A"+strconv.Itoa(fr), fmt.Sprintf("Grand Total | Present: %d | Late: %d | Absent: %d | Half Day: %d | On Leave: %d | Weekend: %d | Total: %d", gp, gl, ga, ghd, gol, gw, gt))
	f.MergeCell(sheetName, "A"+strconv.Itoa(fr), endCol+strconv.Itoa(fr))
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(fr), endCol+strconv.Itoa(fr), ftStyle)
	f.SetRowHeight(sheetName, fr, 22)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	case int:
		return int64(x)
	default:
		return 0
	}
}

type bulkUpdateMissingRequest struct {
	Status        string   `json:"status"`
	InTime        string   `json:"inTime"`
	OutTime       string   `json:"outTime"`
	AttendanceIDs []string `json:"attendanceIds" binding:"required,min=1"`
}

// BulkUpdateMissing godoc
//
//	@Summary      Bulk update missing attendance
//	@Description  Update many missing attendance records in one transaction. Each record's original attendance date is taken from the database and combined with the entered In/Out times. Only missing fields are updated; existing times are preserved. All updates roll back if any record fails validation.
//	@Tags         Attendance
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body bulkUpdateMissingRequest true "Bulk missing attendance request"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /attendance/bulk-update-missing [post]
func (h *AttendanceHandler) BulkUpdateMissing(c *gin.Context) {
	var req bulkUpdateMissingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.InTime == "" && req.OutTime == "" && req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one of inTime, outTime or status must be provided"})
		return
	}

	userID := c.GetString("user_id")

	updated, err := h.attendanceRepo.BulkUpdateMissing(req.AttendanceIDs, req.InTime, req.OutTime, req.Status, userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"updated": updated,
		"failed":  0,
		"message": fmt.Sprintf("%d attendance records updated successfully", updated),
	})
}
