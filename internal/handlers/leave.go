package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type LeaveHandler struct {
	leaveRepo      *repository.LeaveRepository
	employeeRepo   *repository.EmployeeRepository
	attendanceRepo *repository.AttendanceRepository
}

func NewLeaveHandler(leaveRepo *repository.LeaveRepository, employeeRepo *repository.EmployeeRepository, attendanceRepo *repository.AttendanceRepository) *LeaveHandler {
	return &LeaveHandler{leaveRepo: leaveRepo, employeeRepo: employeeRepo, attendanceRepo: attendanceRepo}
}

// --- Request types ---

type CreateLeaveTypeRequest struct {
	CompanyID        string `json:"company_id" binding:"required"`
	Name             string `json:"name" binding:"required"`
	Code             string `json:"code" binding:"required"`
	TotalDays        int    `json:"total_days" binding:"required"`
	CarryForwardDays int    `json:"carry_forward_days"`
	ApplicableGender string `json:"applicable_gender"`
}

type UpdateLeaveTypeRequest struct {
	Name             string `json:"name" binding:"required"`
	TotalDays        int    `json:"total_days" binding:"required"`
	CarryForwardDays int    `json:"carry_forward_days"`
	ApplicableGender string `json:"applicable_gender"`
	Status           string `json:"status"`
}

type ApplyLeaveRequest struct {
	CompanyID   string `json:"company_id" binding:"required"`
	EmployeeID  string `json:"employee_id" binding:"required"`
	LeaveTypeID string `json:"leave_type_id" binding:"required"`
	FromDate    string `json:"from_date" binding:"required"`
	ToDate      string `json:"to_date" binding:"required"`
	Reason      string `json:"reason"`
}

type UpdateLeaveRequest struct {
	LeaveTypeID string `json:"leave_type_id"`
	FromDate    string `json:"from_date"`
	ToDate      string `json:"to_date"`
	TotalDays   int    `json:"total_days"`
	Reason      string `json:"reason"`
}

type RejectLeaveRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required"`
}

// --- Leave Types ---

// ListLeaveTypes godoc
//
// @Summary      List leave types
// @Description  Get all leave types
// @Tags         Leave Types
// @Security     BearerAuth
// @Produce      json
// @Param        company_id query string false "Filter by company"
// @Param        page       query int    false "Page number (default: 1)"
// @Param        limit      query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /leave-types [get]
func (h *LeaveHandler) ListLeaveTypes(c *gin.Context) {
	companyID := c.Query("company_id")
	p := utils.ParsePagination(c)
	list, total, err := h.leaveRepo.ListLeaveTypes(companyID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

// GetLeaveType godoc
//
// @Summary      Get leave type by ID
// @Tags         Leave Types
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Leave Type ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /leave-types/{id} [get]
func (h *LeaveHandler) GetLeaveType(c *gin.Context) {
	id := c.Param("id")
	lt, err := h.leaveRepo.FindLeaveTypeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave type not found"})
		return
	}
	c.JSON(http.StatusOK, lt)
}

// CreateLeaveType godoc
//
// @Summary      Create leave type
// @Tags         Leave Types
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateLeaveTypeRequest true "Leave type details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /leave-types [post]
func (h *LeaveHandler) CreateLeaveType(c *gin.Context) {
	var req CreateLeaveTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetString("user_id")

	lt := &models.LeaveType{
		CompanyID:        req.CompanyID,
		Name:             req.Name,
		Code:             req.Code,
		TotalDays:        req.TotalDays,
		CarryForwardDays: req.CarryForwardDays,
		ApplicableGender: req.ApplicableGender,
		Status:           "active",
		CreatedBy:        &userID,
	}
	if lt.ApplicableGender == "" {
		lt.ApplicableGender = "All"
	}

	if err := h.leaveRepo.CreateLeaveType(lt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, lt)
}

// UpdateLeaveType godoc
//
// @Summary      Update leave type
// @Tags         Leave Types
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Leave Type ID"
// @Param        request body UpdateLeaveTypeRequest true "Updated leave type"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /leave-types/{id} [put]
func (h *LeaveHandler) UpdateLeaveType(c *gin.Context) {
	id := c.Param("id")
	lt, err := h.leaveRepo.FindLeaveTypeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave type not found"})
		return
	}
	var req UpdateLeaveTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetString("user_id")

	lt.Name = req.Name
	lt.TotalDays = req.TotalDays
	lt.CarryForwardDays = req.CarryForwardDays
	if req.ApplicableGender != "" {
		lt.ApplicableGender = req.ApplicableGender
	}
	if req.Status != "" {
		lt.Status = req.Status
	}
	lt.UpdatedBy = &userID

	if err := h.leaveRepo.UpdateLeaveType(lt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lt)
}

// DeleteLeaveType godoc
//
// @Summary      Delete leave type
// @Tags         Leave Types
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Leave Type ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /leave-types/{id} [delete]
func (h *LeaveHandler) DeleteLeaveType(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.leaveRepo.FindLeaveTypeByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave type not found"})
		return
	}
	if err := h.leaveRepo.DeleteLeaveType(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "leave type deleted"})
}

// --- Leave Applications ---

// ListLeaves godoc
//
// @Summary      List leave applications
// @Description  Get all leave applications with optional filters
// @Tags         Leaves
// @Security     BearerAuth
// @Produce      json
// @Param        company_id    query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Param        employee_id   query string false "Filter by employee"
// @Param        status        query string false "Filter by status (pending|approved|rejected|cancelled)"
// @Param        from_date     query string false "Start date (YYYY-MM-DD)"
// @Param        to_date       query string false "End date (YYYY-MM-DD)"
// @Param        page          query int    false "Page number (default: 1)"
// @Param        limit         query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /leaves [get]
func (h *LeaveHandler) ListLeaves(c *gin.Context) {
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	employeeID := c.Query("employee_id")
	status := c.Query("status")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	p := utils.ParsePagination(c)
	list, total, err := h.leaveRepo.ListLeaves(companyID, departmentID, employeeID, status, fromDate, toDate, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

// GetLeave godoc
//
// @Summary      Get leave by ID
// @Tags         Leaves
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Leave ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /leaves/{id} [get]
func (h *LeaveHandler) GetLeave(c *gin.Context) {
	id := c.Param("id")
	l, err := h.leaveRepo.FindLeaveByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave not found"})
		return
	}
	c.JSON(http.StatusOK, l)
}

// ApplyLeave godoc
//
// @Summary      Apply for leave
// @Description  Submit a new leave application. Validates allocation balance.
// @Tags         Leaves
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body ApplyLeaveRequest true "Leave application"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /leaves [post]
func (h *LeaveHandler) ApplyLeave(c *gin.Context) {
	var req ApplyLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetString("user_id")

	from, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date"})
		return
	}
	to, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_date"})
		return
	}
	if to.Before(from) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to_date must be after from_date"})
		return
	}
	totalDays := int(to.Sub(from).Hours()/24) + 1

	// Check allocation
	year := from.Year()
	alloc, err := h.leaveRepo.FindAllocation(req.EmployeeID, req.LeaveTypeID, year)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		alloc = nil
	} else {
		remaining := alloc.TotalDays - alloc.UsedDays - alloc.PendingDays
		if remaining < totalDays {
			c.JSON(http.StatusConflict, gin.H{"error": "insufficient leave balance", "remaining": remaining, "requested": totalDays})
			return
		}
	}

	l := &models.Leave{
		CompanyID:   req.CompanyID,
		EmployeeID:  req.EmployeeID,
		LeaveTypeID: req.LeaveTypeID,
		FromDate:    req.FromDate,
		ToDate:      req.ToDate,
		TotalDays:   totalDays,
		Reason:      req.Reason,
		Status:      "pending",
		CreatedBy:   &userID,
	}

	// Wrap leave creation + allocation update in a transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		leaveTx := h.leaveRepo.WithTx(tx)
		if err := leaveTx.CreateLeave(l); err != nil {
			return err
		}
		if alloc != nil {
			alloc.PendingDays += totalDays
			if err := leaveTx.UpsertAllocation(alloc); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, l)
}

// UpdateLeave godoc
//
// @Summary      Update leave application
// @Tags         Leaves
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Leave ID"
// @Param        request body UpdateLeaveRequest true "Updated leave"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /leaves/{id} [put]
func (h *LeaveHandler) UpdateLeave(c *gin.Context) {
	id := c.Param("id")
	l, err := h.leaveRepo.FindLeaveByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave not found"})
		return
	}
	if l.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only update pending leaves"})
		return
	}

	var req UpdateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetString("user_id")

	if req.LeaveTypeID != "" {
		l.LeaveTypeID = req.LeaveTypeID
	}
	if req.FromDate != "" {
		l.FromDate = req.FromDate
	}
	if req.ToDate != "" {
		l.ToDate = req.ToDate
	}
	if req.TotalDays > 0 {
		l.TotalDays = req.TotalDays
	}
	if req.Reason != "" {
		l.Reason = req.Reason
	}
	l.UpdatedBy = &userID

	if err := h.leaveRepo.UpdateLeave(l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, l)
}

// DeleteLeave godoc
//
// @Summary      Permanently delete leave application
// @Description  Hard delete a leave regardless of status and revert related attendance
// @Tags         Leaves
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Leave ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /leaves/{id} [delete]
func (h *LeaveHandler) DeleteLeave(c *gin.Context) {
	id := c.Param("id")
	l, err := h.leaveRepo.FindLeaveByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave not found"})
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		leaveTx := h.leaveRepo.WithTx(tx)
		attTx := h.attendanceRepo.WithTx(tx)

		// Revert leave allocation
		year := time.Now().Year()
		if yr, err := strconv.Atoi(l.FromDate[:4]); err == nil {
			year = yr
		}
		alloc, err := leaveTx.FindAllocation(l.EmployeeID, l.LeaveTypeID, year)
		if err == nil && alloc != nil {
			switch l.Status {
			case "approved":
				alloc.UsedDays -= l.TotalDays
				if alloc.UsedDays < 0 {
					alloc.UsedDays = 0
				}
			case "pending":
				alloc.PendingDays -= l.TotalDays
				if alloc.PendingDays < 0 {
					alloc.PendingDays = 0
				}
			}
			if err := leaveTx.UpsertAllocation(alloc); err != nil {
				return err
			}
		}

		// Remove on_leave attendance marker for the leave period
		if err := attTx.ClearOnLeaveStatus(l.EmployeeID, l.FromDate, l.ToDate); err != nil {
			return err
		}

		// Hard delete the leave
		return leaveTx.HardDeleteLeave(id)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "leave deleted"})
}

// ApproveLeave godoc
//
// @Summary      Approve leave application
// @Tags         Leaves
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Leave ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /leaves/{id}/approve [put]
func (h *LeaveHandler) ApproveLeave(c *gin.Context) {
	id := c.Param("id")
	l, err := h.leaveRepo.FindLeaveByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave not found"})
		return
	}
	if l.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leave is not pending"})
		return
	}

	userID := c.GetString("user_id")
	now := time.Now()
	l.Status = "approved"
	l.ApprovedBy = &userID
	l.ApprovedAt = &now

	// Wrap leave approval + allocation + attendance update in a transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		leaveTx := h.leaveRepo.WithTx(tx)
		attTx := h.attendanceRepo.WithTx(tx)

		if err := leaveTx.UpdateLeave(l); err != nil {
			return err
		}

		// Move pending → used in allocation
		year := time.Now().Year()
		alloc, err := leaveTx.FindAllocation(l.EmployeeID, l.LeaveTypeID, year)
		if err == nil && alloc != nil {
			alloc.PendingDays -= l.TotalDays
			alloc.UsedDays += l.TotalDays
			if err := leaveTx.UpsertAllocation(alloc); err != nil {
				return err
			}
		}

		// Mark attendance as on_leave for the leave period
		if err := attTx.UpdateStatusByEmployeeAndDateRange(l.EmployeeID, l.FromDate, l.ToDate, "on_leave"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

// RejectLeave godoc
//
// @Summary      Reject leave application
// @Tags         Leaves
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Leave ID"
// @Param        request body RejectLeaveRequest true "Rejection reason"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /leaves/{id}/reject [put]
func (h *LeaveHandler) RejectLeave(c *gin.Context) {
	id := c.Param("id")
	l, err := h.leaveRepo.FindLeaveByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave not found"})
		return
	}
	if l.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leave is not pending"})
		return
	}

	var req RejectLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	l.Status = "rejected"
	l.ApprovedBy = &userID
	l.RejectionReason = req.RejectionReason

	// Wrap leave rejection + allocation update in a transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		leaveTx := h.leaveRepo.WithTx(tx)
		if err := leaveTx.UpdateLeave(l); err != nil {
			return err
		}

		// Decrement pending in allocation
		year := time.Now().Year()
		alloc, err := leaveTx.FindAllocation(l.EmployeeID, l.LeaveTypeID, year)
		if err == nil && alloc != nil {
			alloc.PendingDays -= l.TotalDays
			if alloc.PendingDays < 0 {
				alloc.PendingDays = 0
			}
			if err := leaveTx.UpsertAllocation(alloc); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, l)
}

// --- Leave Balance ---

// ListLeaveBalance godoc
//
// @Summary      Get leave balance
// @Description  Get leave balance for employees
// @Tags         Leave Balance
// @Security     BearerAuth
// @Produce      json
// @Param        employee_id query string false "Filter by employee"
// @Param        year        query int    false "Year (default: current)"
// @Param        page        query int    false "Page number (default: 1)"
// @Param        limit       query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      500  {object}  map[string]string
// @Router       /leave-balance [get]
func (h *LeaveHandler) ListLeaveBalance(c *gin.Context) {
	employeeID := c.Query("employee_id")
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)
	if year == 0 {
		year = time.Now().Year()
	}

	p := utils.ParsePagination(c)
	list, total, err := h.leaveRepo.ListAllocations(employeeID, year, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type BalanceEntry struct {
		EmployeeID  string `json:"employee_id"`
		LeaveTypeID string `json:"leave_type_id"`
		LeaveType   string `json:"leave_type"`
		Year        int    `json:"year"`
		Total       int    `json:"total"`
		Used        int    `json:"used"`
		Pending     int    `json:"pending"`
		Remaining   int    `json:"remaining"`
	}

	var result []BalanceEntry
	for _, a := range list {
		result = append(result, BalanceEntry{
			EmployeeID:  a.EmployeeID,
			LeaveTypeID: a.LeaveTypeID,
			LeaveType:   a.LeaveType.Name,
			Year:        a.Year,
			Total:       a.TotalDays,
			Used:        a.UsedDays,
			Pending:     a.PendingDays,
			Remaining:   a.TotalDays - a.UsedDays - a.PendingDays,
		})
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(result, total, p))
}

// GetLeaveDetails godoc
//
// @Summary      Get employee leave details with balance report and leave history
// @Description  Get comprehensive leave details for an employee filtered by year and month
// @Tags         Leave Reports
// @Security     BearerAuth
// @Produce      json
// @Param        employee_id   query string false "Employee ID, Punch Number, or UUID"
// @Param        year          query int    false "Year (default: current)"
// @Param        month         query int    false "Month (1-12, optional)"
// @Param        company_id    query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Param        section_id    query string false "Filter by section"
// @Param        designation_id query string false "Filter by designation"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /leave-details [get]
func (h *LeaveHandler) GetLeaveDetails(c *gin.Context) {
	empCode := strings.TrimSpace(c.Query("employee_id"))
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.Query("month")
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	sectionID := c.Query("section_id")
	designationID := c.Query("designation_id")

	year, _ := strconv.Atoi(yearStr)
	if year == 0 {
		year = time.Now().Year()
	}
	month, _ := strconv.Atoi(monthStr)
	if month < 0 || month > 12 {
		month = 0
	}

	var emp *models.Employee

	// If employee_id is provided, search by EmployeeID, PunchNumber, or ID
	if empCode != "" {
		emp, _ = h.employeeRepo.FindWithDetails(empCode)
	}

	// If no employee specified or not found, find the first matching employee for company/department
	if emp == nil {
		filter := repository.EmployeeFilter{
			CompanyID:     companyID,
			DepartmentID:  departmentID,
			SectionID:     sectionID,
			DesignationID: designationID,
			EmployeeID:    empCode,
		}
		emp, _ = h.employeeRepo.FindFirstFilteredWithDetails(filter)
	}

	if emp == nil {
		c.JSON(http.StatusOK, gin.H{
			"employee": nil,
			"balances": []interface{}{},
			"summary": gin.H{
				"total_entitled":     0,
				"total_used":         0,
				"total_pending":      0,
				"total_remaining":    0,
				"total_applications": 0,
			},
			"leaves": []interface{}{},
		})
		return
	}

	// 1. Fetch employee's company leave types
	compID := emp.CompanyID
	if companyID != "" {
		compID = companyID
	}
	leaveTypes, _ := h.leaveRepo.ListActiveLeaveTypes(compID)

	// 2. Fetch employee's allocations for the year
	allocations, _ := h.leaveRepo.ListAllocationsByEmployees([]string{emp.EmployeeID}, year)
	allocMap := make(map[string]models.LeaveAllocation)
	for _, a := range allocations {
		allocMap[a.LeaveTypeID] = a
	}

	// Fetch all leaves for the full year to accurately compute approved and pending days per leave type
	yearLeaves, _ := h.leaveRepo.ListEmployeeLeavesByPeriod(emp.EmployeeID, year, 0)
	approvedDaysMap := make(map[string]int)
	pendingDaysMap := make(map[string]int)
	for _, l := range yearLeaves {
		if strings.EqualFold(l.Status, "approved") {
			approvedDaysMap[l.LeaveTypeID] += l.TotalDays
		} else if strings.EqualFold(l.Status, "pending") {
			pendingDaysMap[l.LeaveTypeID] += l.TotalDays
		}
	}

	// 3. Build balances list
	type BalanceItem struct {
		ID               string `json:"id"`
		LeaveTypeID      string `json:"leave_type_id"`
		LeaveType        string `json:"leave_type"`
		Code             string `json:"code"`
		Year             int    `json:"year"`
		Total            int    `json:"total"`
		Used             int    `json:"used"`
		Pending          int    `json:"pending"`
		Remaining        int    `json:"remaining"`
		ApplicableGender string `json:"applicable_gender"`
	}

	var balances []BalanceItem
	totalEntitled := 0
	totalUsed := 0
	totalPending := 0
	totalRemaining := 0

	for _, lt := range leaveTypes {
		// Gender check
		if lt.ApplicableGender != "" && !strings.EqualFold(lt.ApplicableGender, "all") {
			if emp.Gender != "" && !strings.EqualFold(emp.Gender, lt.ApplicableGender) {
				continue
			}
		}

		alloc, exists := allocMap[lt.ID]
		total := lt.TotalDays
		used := 0
		pending := 0

		if exists {
			total = alloc.TotalDays
			used = alloc.UsedDays
			pending = alloc.PendingDays
		}

		if approvedDaysMap[lt.ID] > used {
			used = approvedDaysMap[lt.ID]
		}
		if pendingDaysMap[lt.ID] > pending {
			pending = pendingDaysMap[lt.ID]
		}

		remaining := total - used - pending
		if remaining < 0 {
			remaining = 0
		}

		balances = append(balances, BalanceItem{
			ID:               lt.ID,
			LeaveTypeID:      lt.ID,
			LeaveType:        lt.Name,
			Code:             lt.Code,
			Year:             year,
			Total:            total,
			Used:             used,
			Pending:          pending,
			Remaining:        remaining,
			ApplicableGender: lt.ApplicableGender,
		})

		totalEntitled += total
		totalUsed += used
		totalPending += pending
		totalRemaining += remaining
	}

	// 4. Fetch leaves history for this employee matching year & month
	var leavesList []models.Leave
	if month > 0 {
		leavesList, _ = h.leaveRepo.ListEmployeeLeavesByPeriod(emp.EmployeeID, year, month)
	} else {
		leavesList = yearLeaves
	}

	type LeaveItem struct {
		ID              string `json:"id"`
		EmployeeID      string `json:"employee_id"`
		LeaveTypeID     string `json:"leave_type_id"`
		LeaveType       string `json:"leave_type"`
		LeaveCode       string `json:"leave_code"`
		FromDate        string `json:"from_date"`
		ToDate          string `json:"to_date"`
		TotalDays       int    `json:"total_days"`
		Reason          string `json:"reason"`
		Status          string `json:"status"`
		RejectionReason string `json:"rejection_reason,omitempty"`
		CreatedAt       string `json:"created_at"`
	}

	var leaves []LeaveItem
	for _, l := range leavesList {
		typeName := ""
		typeCode := ""
		if l.LeaveType.Name != "" {
			typeName = l.LeaveType.Name
			typeCode = l.LeaveType.Code
		}
		leaves = append(leaves, LeaveItem{
			ID:              l.ID,
			EmployeeID:      l.EmployeeID,
			LeaveTypeID:     l.LeaveTypeID,
			LeaveType:       typeName,
			LeaveCode:       typeCode,
			FromDate:        l.FromDate,
			ToDate:          l.ToDate,
			TotalDays:       l.TotalDays,
			Reason:          l.Reason,
			Status:          l.Status,
			RejectionReason: l.RejectionReason,
			CreatedAt:       l.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	// Format employee profile
	companyName := ""
	if emp.Company.CompanyNameEn != "" {
		companyName = emp.Company.CompanyNameEn
	}
	deptName := ""
	if emp.Department.Name != "" {
		deptName = emp.Department.Name
	}
	desigName := ""
	if emp.DesignationRef != nil {
		desigName = emp.DesignationRef.Name
	}
	secName := ""
	if emp.SectionRef != nil {
		secName = emp.SectionRef.Name
	}
	lineName := ""
	if emp.LineRef != nil {
		lineName = emp.LineRef.Name
	}
	groupName := ""
	if emp.GroupRef != nil {
		groupName = emp.GroupRef.Name
	}
	shiftName := ""
	if emp.Shift != nil {
		shiftName = emp.Shift.Name
	}
	joiningDateStr := ""
	if !emp.JoiningDate.IsZero() {
		joiningDateStr = emp.JoiningDate.Format("2006-01-02")
	}

	employeeData := gin.H{
		"id":               emp.ID,
		"employee_id":      emp.EmployeeID,
		"punch_number":     emp.PunchNumber,
		"name_en":          emp.NameEn,
		"name_bn":          emp.NameBn,
		"phone":            emp.Phone,
		"gender":           emp.Gender,
		"status":           emp.Status,
		"employee_type":    emp.EmployeeType,
		"joining_date":     joiningDateStr,
		"company_id":       emp.CompanyID,
		"company_name":     companyName,
		"department_id":    emp.DepartmentID,
		"department_name":  deptName,
		"designation_id":   emp.DesignationID,
		"designation_name": desigName,
		"section_id":       emp.SectionID,
		"section_name":     secName,
		"line_id":          emp.LineID,
		"line_name":        lineName,
		"group_id":         emp.GroupID,
		"group_name":       groupName,
		"shift_id":         emp.ShiftID,
		"shift_name":       shiftName,
		"photo_url":        emp.ImageURL,
	}

	c.JSON(http.StatusOK, gin.H{
		"employee": employeeData,
		"balances": balances,
		"summary": gin.H{
			"total_entitled":     totalEntitled,
			"total_used":         totalUsed,
			"total_pending":      totalPending,
			"total_remaining":    totalRemaining,
			"total_applications": len(leaves),
		},
		"leaves": leaves,
	})
}

// --- Monthly Report ---

// MonthlyLeaveReport godoc
//
// @Summary      Monthly leave report
// @Description  Get monthly leave report grouped by department
// @Tags         Leave Reports
// @Security     BearerAuth
// @Produce      json
// @Param        month       query int    false "Month (1-12, default: current)"
// @Param        year        query int    false "Year (default: current)"
// @Param        company_id  query string false "Filter by company"
// @Param        department_id query string false "Filter by department"
// @Success      200  {array}   map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /leave-reports/monthly [get]
func (h *LeaveHandler) MonthlyLeaveReport(c *gin.Context) {
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}
	if year == 0 {
		year = time.Now().Year()
	}

	results, err := h.leaveRepo.MonthlyReport(month, year, companyID, departmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// ExportLeaveFormPDF godoc
//
// @Summary      Export single leave application as PDF form
// @Description  Generate a leave application form PDF for a specific leave
// @Tags         Leaves
// @Security     BearerAuth
// @Produce      application/pdf
// @Param        id path string true "Leave ID"
// @Param        lang query string false "Language: en|bn (default en)"
// @Success      200  {file}  binary
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /leaves/{id}/export/pdf [get]
func (h *LeaveHandler) ExportLeaveFormPDF(c *gin.Context) {
	id := c.Param("id")

	lang := c.DefaultQuery("lang", "en")
	if lang != "bn" {
		lang = "en"
	}

	var leave models.Leave
	if err := database.DB.
		Preload("Company").
		Preload("Employee.Department").
		Preload("Employee.SectionRef").
		Preload("Employee.DesignationRef").
		Preload("Employee.Shift").
		Preload("Employee.Manager").
		Preload("LeaveType").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&leave).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "leave not found"})
		return
	}

	labels := leaveEnLabels
	if lang == "bn" {
		labels = leaveBnLabels
	}

	year := time.Now().Year()
	if len(leave.FromDate) >= 4 {
		if y, err := strconv.Atoi(leave.FromDate[:4]); err == nil {
			year = y
		}
	}
	allocs, _, _ := h.leaveRepo.ListAllocations(leave.EmployeeID, year, 1, 100)

	data := buildLeaveFormData(&leave, lang, labels, allocs)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	font := leaveFormFont(pdf, lang)
	renderLeaveFormPDFPage(pdf, font, lang, data, labels)
	if pdf.Error() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF error: " + pdf.Error().Error()})
		return
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: " + err.Error()})
		return
	}

	langSuffix := lang
	if langSuffix == "" {
		langSuffix = "en"
	}
	filename := fmt.Sprintf("leave_application_%s_%s.pdf", leave.EmployeeID, langSuffix)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// ExportLeavesExcel godoc
//
//	@Summary      Export leaves to Excel
//	@Description  Export filtered leave applications to Excel with company header
//	@Tags         Leaves
//	@Security     BearerAuth
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param        company_id    query string false "Filter by company"
//	@Param        department_id query string false "Filter by department"
//	@Param        employee_id   query string false "Filter by employee"
//	@Param        status        query string false "Filter by status"
//	@Param        from_date     query string false "Start date (YYYY-MM-DD)"
//	@Param        to_date       query string false "End date (YYYY-MM-DD)"
//	@Success      200  {file}  binary
//	@Failure      500  {object}  map[string]string
//	@Router       /leaves/export/excel [get]
func (h *LeaveHandler) ExportExcel(c *gin.Context) {
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	employeeID := c.Query("employee_id")
	status := c.Query("status")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	leaves, err := h.leaveRepo.ListLeavesExport(companyID, departmentID, employeeID, status, fromDate, toDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if companyID != "" {
		_ = database.DB.Where("id = ? AND deleted_at IS NULL", companyID).First(&company)
	}
	if company.ID == "" {
		_ = database.DB.Where("deleted_at IS NULL").First(&company)
	}

	periodStr := "All Time"
	if fromDate != "" && toDate != "" {
		periodStr = fmt.Sprintf("Period: %s to %s", fromDate, toDate)
	} else if fromDate != "" {
		periodStr = fmt.Sprintf("From: %s", fromDate)
	} else if toDate != "" {
		periodStr = fmt.Sprintf("Up to: %s", toDate)
	}

	renderLeaveExcel(c, company, periodStr, leaves)
}

// ExportLeavesPDF godoc
//
//	@Summary      Export leaves to PDF
//	@Description  Export filtered leave applications to PDF report
//	@Tags         Leaves
//	@Security     BearerAuth
//	@Produce      application/pdf
//	@Param        company_id    query string false "Filter by company"
//	@Param        department_id query string false "Filter by department"
//	@Param        employee_id   query string false "Filter by employee"
//	@Param        status        query string false "Filter by status"
//	@Param        from_date     query string false "Start date (YYYY-MM-DD)"
//	@Param        to_date       query string false "End date (YYYY-MM-DD)"
//	@Param        lang          query string false "Language: en or bn"
//	@Success      200  {file}  binary
//	@Failure      500  {object}  map[string]string
//	@Router       /leaves/export/pdf [get]
func (h *LeaveHandler) ExportPDF(c *gin.Context) {
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")
	employeeID := c.Query("employee_id")
	status := c.Query("status")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")
	lang := c.DefaultQuery("lang", "en")
	isBn := strings.ToLower(lang) == "bn" || strings.ToLower(lang) == "bangla"

	leaves, err := h.leaveRepo.ListLeavesExport(companyID, departmentID, employeeID, status, fromDate, toDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if companyID != "" {
		_ = database.DB.Where("id = ? AND deleted_at IS NULL", companyID).First(&company)
	}
	if company.ID == "" {
		_ = database.DB.Where("deleted_at IS NULL").First(&company)
	}

	periodStr := "All Time"
	if fromDate != "" && toDate != "" {
		periodStr = fmt.Sprintf("Period: %s to %s", fromDate, toDate)
	} else if fromDate != "" {
		periodStr = fmt.Sprintf("From: %s", fromDate)
	} else if toDate != "" {
		periodStr = fmt.Sprintf("Up to: %s", toDate)
	}

	renderLeavePDF(c, company, periodStr, leaves, isBn)
}

func renderLeaveExcel(c *gin.Context, company models.Company, periodStr string, leaves []models.Leave) {
	companyName := company.CompanyNameEn
	if companyName == "" {
		companyName = "EKUSHE FASHIONS LTD."
	}
	companyAddress := company.AddressEn
	if companyAddress == "" {
		companyAddress = "Masterbari, Gazipur City, Gazipur."
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Leave Report"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"SL", "Employee ID", "Employee Name", "Leave Type", "From Date", "To Date", "Days", "Reason", "Status"}
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	lastCol := "I"

	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 25)
	f.SetColWidth(sheet, "D", "D", 20)
	f.SetColWidth(sheet, "E", "E", 14)
	f.SetColWidth(sheet, "F", "F", 14)
	f.SetColWidth(sheet, "G", "G", 8)
	f.SetColWidth(sheet, "H", "H", 28)
	f.SetColWidth(sheet, "I", "I", 14)

	borderColor := "262626"
	thinBorder := []excelize.Border{
		{Type: "left", Color: borderColor, Style: 1},
		{Type: "top", Color: borderColor, Style: 1},
		{Type: "bottom", Color: borderColor, Style: 1},
		{Type: "right", Color: borderColor, Style: 1},
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Family: "Calibri", Color: "000000"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D9E2F3"}, Pattern: 1},
		Border: thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	cellNormal, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border: thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	cellLeft, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "Calibri", Color: "000000"},
		Border: thinBorder,
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	headerRuns := []excelize.RichTextRun{
		{Text: companyName, Font: &excelize.Font{Bold: true, Size: 20}},
		{Text: "\n" + companyAddress, Font: &excelize.Font{Size: 11}},
		{Text: "\nLEAVE REPORT", Font: &excelize.Font{Bold: true, Size: 11, Color: "DC2626"}},
		{Text: "\n" + periodStr, Font: &excelize.Font{Size: 11}},
	}
	_ = f.SetCellRichText(sheet, "A1", headerRuns)
	_ = f.MergeCell(sheet, "A1", lastCol+"1")
	headerStyleA1, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	_ = f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyleA1)
	_ = f.SetRowHeight(sheet, 1, 75)

	headerRow := 2
	f.SetRowHeight(sheet, headerRow, 30)
	for i, h := range headers {
		cell := cols[i] + strconv.Itoa(headerRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	currentRow := headerRow + 1
	for idx, l := range leaves {
		rStr := strconv.Itoa(currentRow)
		f.SetRowHeight(sheet, currentRow, 22)
		empName := ""
		if l.Employee.ID != "" {
			empName = l.Employee.NameEn
		}
		leaveType := ""
		if l.LeaveType.ID != "" {
			leaveType = l.LeaveType.Name
		}
		f.SetCellValue(sheet, "A"+rStr, idx+1)
		f.SetCellValue(sheet, "B"+rStr, l.EmployeeID)
		f.SetCellValue(sheet, "C"+rStr, empName)
		f.SetCellValue(sheet, "D"+rStr, leaveType)
		f.SetCellValue(sheet, "E"+rStr, utils.FormatBillDate(l.FromDate))
		f.SetCellValue(sheet, "F"+rStr, utils.FormatBillDate(l.ToDate))
		f.SetCellValue(sheet, "G"+rStr, l.TotalDays)
		f.SetCellValue(sheet, "H"+rStr, l.Reason)
		f.SetCellValue(sheet, "I"+rStr, l.Status)
		f.SetCellStyle(sheet, "A"+rStr, "A"+rStr, cellNormal)
		f.SetCellStyle(sheet, "B"+rStr, "B"+rStr, cellNormal)
		f.SetCellStyle(sheet, "C"+rStr, "C"+rStr, cellLeft)
		f.SetCellStyle(sheet, "D"+rStr, "D"+rStr, cellLeft)
		f.SetCellStyle(sheet, "E"+rStr, "E"+rStr, cellNormal)
		f.SetCellStyle(sheet, "F"+rStr, "F"+rStr, cellNormal)
		f.SetCellStyle(sheet, "G"+rStr, "G"+rStr, cellNormal)
		f.SetCellStyle(sheet, "H"+rStr, "H"+rStr, cellLeft)
		f.SetCellStyle(sheet, "I"+rStr, "I"+rStr, cellNormal)
		currentRow++
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=leave_report_%s.xlsx", time.Now().Format("20060102_150405")))
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func renderLeavePDF(c *gin.Context, company models.Company, periodStr string, leaves []models.Leave, isBn bool) {
	companyName := companyDisplayName(company, isBn)
	companyAddress := company.AddressEn
	if isBn && company.AddressBn != "" {
		companyAddress = utils.UnicodeToBijoy(company.AddressBn)
	}
	if companyAddress == "" {
		companyAddress = "Masterbari, Gazipur City, Gazipur."
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(true, 10)

	font := "Arial"
	if isBn {
		font = loadBanglaFont(pdf)
	}
	pdf.AddPage()
	pdf.SetFont(font, "B", 14)
	pdf.SetTextColor(18, 58, 99)
	pdf.CellFormat(281, 7, companyName, "", 1, "C", false, 0, "")
	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(281, 5, companyAddress, "", 1, "C", false, 0, "")
	pdf.SetFont(font, "B", 11)
	pdf.SetTextColor(220, 38, 38)
	title := "LEAVE REPORT"
	if isBn {
		title = utils.UnicodeToBijoy("ছুটির রিপোর্ট")
	}
	pdf.CellFormat(281, 6, title, "", 1, "C", false, 0, "")
	pdf.SetFont(font, "", 9)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(281, 5, periodStr, "", 1, "C", false, 0, "")
	pdf.Ln(3)

	wSL := 10.0
	wEmpID := 22.0
	wEmpName := 42.0
	wType := 30.0
	wFrom := 22.0
	wTo := 22.0
	wDays := 12.0
	wReason := 70.0
	wStatus := 22.0

	pdf.SetFillColor(217, 226, 243)
	pdf.SetDrawColor(38, 38, 38)
	pdf.SetLineWidth(0.3)
	pdf.SetFont(font, "B", 8)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(wSL, 7, "SL", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wEmpID, 7, "Emp ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wEmpName, 7, "Employee", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wType, 7, "Leave Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wFrom, 7, "From", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wTo, 7, "To", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wDays, 7, "Days", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wReason, 7, "Reason", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wStatus, 7, "Status", "1", 1, "C", true, 0, "")

	pdf.SetFont(font, "", 7.5)
	pdf.SetTextColor(20, 20, 20)
	for idx, l := range leaves {
		empName := l.Employee.NameEn
		leaveType := l.LeaveType.Name
		if isBn {
			empName = utils.UnicodeToBijoy(empName)
			leaveType = utils.UnicodeToBijoy(leaveType)
		}
		pdf.CellFormat(wSL, 6, strconv.Itoa(idx+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wEmpID, 6, l.EmployeeID, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wEmpName, 6, truncateString(empName, 24), "1", 0, "L", false, 0, "")
		pdf.CellFormat(wType, 6, truncateString(leaveType, 18), "1", 0, "L", false, 0, "")
		pdf.CellFormat(wFrom, 6, utils.FormatBillDate(l.FromDate), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wTo, 6, utils.FormatBillDate(l.ToDate), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wDays, 6, strconv.Itoa(l.TotalDays), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wReason, 6, truncateString(l.Reason, 40), "1", 0, "L", false, 0, "")
		pdf.CellFormat(wStatus, 6, l.Status, "1", 1, "C", false, 0, "")
	}

	pdf.AliasNbPages("{nb}")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-10)
		pdf.SetFont(font, "", 8)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(281, 5, fmt.Sprintf("Page %d of {nb}   |   Print Date: %s", pdf.PageNo(), time.Now().Format("02-01-2006 15:04")), "", 0, "C", false, 0, "")
	})
	if pdf.Error() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": pdf.Error().Error()})
		return
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=leave_report_%s.pdf", time.Now().Format("20060102_150405")))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
