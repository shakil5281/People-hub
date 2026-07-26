package handlers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type NightBillHandler struct {
	nightBillRepo *repository.NightBillRepository
	employeeRepo  *repository.EmployeeRepository
}

func NewNightBillHandler(nightBillRepo *repository.NightBillRepository, employeeRepo *repository.EmployeeRepository) *NightBillHandler {
	return &NightBillHandler{nightBillRepo: nightBillRepo, employeeRepo: employeeRepo}
}

type CreateNightBillRequest struct {
	CompanyID     string   `json:"company_id" binding:"required"`
	EmployeeID    string   `json:"employee_id" binding:"required"`
	DepartmentID  *string  `json:"department_id"`
	SectionID     *string  `json:"section_id"`
	DesignationID *string  `json:"designation_id"`
	LineID        *string  `json:"line_id"`
	GroupID       *string  `json:"group_id"`
	FloorID       *string  `json:"floor_id"`
	Date          string   `json:"date" binding:"required"`
	NightHours    *float64 `json:"night_hours"`
	Rate          *float64 `json:"rate"`
	Amount        *float64 `json:"amount"`
	Remarks       string   `json:"remarks"`
}

type BulkCreateNightBillRequest struct {
	CompanyID string                    `json:"company_id" binding:"required"`
	Month     int                       `json:"month" binding:"required"`
	Year      int                       `json:"year" binding:"required"`
	Items     []BulkCreateNightBillItem `json:"items" binding:"required"`
}

type BulkCreateNightBillItem struct {
	EmployeeID string   `json:"employee_id" binding:"required"`
	Date       string   `json:"date" binding:"required"`
	NightHours *float64 `json:"night_hours"`
	Rate       *float64 `json:"rate"`
	Amount     *float64 `json:"amount"`
	Remarks    string   `json:"remarks"`
}

type ProcessNightBillRequest struct {
	CompanyID     string  `json:"company_id" binding:"required"`
	Month         int     `json:"month" binding:"required"`
	Year          int     `json:"year" binding:"required"`
	DateFrom      string  `json:"date_from" binding:"required"`
	DateTo        string  `json:"date_to" binding:"required"`
	DepartmentID  string  `json:"department_id"`
	SectionID     string  `json:"section_id"`
	DesignationID string  `json:"designation_id"`
	LineID        string  `json:"line_id"`
	GroupID       string  `json:"group_id"`
	FloorID       string  `json:"floor_id"`
	NightRate     float64 `json:"night_rate"`
	MinHours      float64 `json:"min_hours"`
}

// ListNightBills godoc
//
//	@Summary      List night bills
//	@Description  Get paginated night bills with advanced filters
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id     query string false "Company ID"
//	@Param        department_id  query string false "Department ID"
//	@Param        section_id     query string false "Section ID"
//	@Param        designation_id query string false "Designation ID"
//	@Param        line_id        query string false "Line ID"
//	@Param        group_id       query string false "Group ID"
//	@Param        floor_id       query string false "Floor ID"
//	@Param        employee_id    query string false "Employee ID"
//	@Param        status         query string false "Status"
//	@Param        date_from      query string false "Date from (YYYY-MM-DD)"
//	@Param        date_to        query string false "Date to (YYYY-MM-DD)"
//	@Param        month          query int    false "Month (1-12)"
//	@Param        year           query int    false "Year"
//	@Param        page           query int    false "Page number"
//	@Param        limit          query int    false "Page size"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills [get]
func (h *NightBillHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}

	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := repository.NightBillFilter{
		CompanyID:     companyID,
		DepartmentID:  c.Query("department_id"),
		SectionID:     c.Query("section_id"),
		DesignationID: c.Query("designation_id"),
		LineID:        c.Query("line_id"),
		GroupID:       c.Query("group_id"),
		FloorID:       c.Query("floor_id"),
		EmployeeID:    c.Query("employee_id"),
		Status:        c.Query("status"),
		DateFrom:      c.Query("date_from"),
		DateTo:        c.Query("date_to"),
		Month:         month,
		Year:          year,
		Page:          page,
		Limit:         limit,
	}

	result, err := h.nightBillRepo.ListPaginated(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        result.Data,
		"total":       result.Total,
		"page":        result.Page,
		"limit":       result.Limit,
		"total_pages": result.TotalPages,
	})
}

// CreateNightBill godoc
//
//	@Summary      Create night bill
//	@Description  Create a new night bill entry for an employee
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body CreateNightBillRequest true "Night bill details"
//	@Success      201  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills [post]
func (h *NightBillHandler) Create(c *gin.Context) {
	var req CreateNightBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	nh := 0.0
	rt := 0.0
	amt := 0.0
	if req.NightHours != nil {
		nh = *req.NightHours
	}
	if req.Rate != nil {
		rt = *req.Rate
	}
	if req.Amount != nil {
		amt = *req.Amount
	}
	if nh > 0 && rt > 0 && amt == 0 {
		amt = nh * rt
	}

	item := &models.NightBill{
		CompanyID:     req.CompanyID,
		EmployeeID:    req.EmployeeID,
		DepartmentID:  req.DepartmentID,
		SectionID:     req.SectionID,
		DesignationID: req.DesignationID,
		LineID:        req.LineID,
		GroupID:       req.GroupID,
		FloorID:       req.FloorID,
		Date:          req.Date,
		NightHours:    nh,
		Rate:          rt,
		Amount:        amt,
		Month:         int(date.Month()),
		Year:          date.Year(),
		Status:        "pending",
		Remarks:       req.Remarks,
		CreatedBy:     c.GetString("user_id"),
	}

	if err := h.nightBillRepo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// BulkCreateNightBills godoc
//
//	@Summary      Bulk create night bills
//	@Description  Create multiple night bill entries at once
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body BulkCreateNightBillRequest true "Bulk night bill items"
//	@Success      201  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/bulk [post]
func (h *NightBillHandler) BulkCreate(c *gin.Context) {
	var req BulkCreateNightBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []models.NightBill
	for _, it := range req.Items {
		nh := 0.0
		rt := 0.0
		amt := 0.0
		if it.NightHours != nil {
			nh = *it.NightHours
		}
		if it.Rate != nil {
			rt = *it.Rate
		}
		if it.Amount != nil {
			amt = *it.Amount
		}
		if nh > 0 && rt > 0 && amt == 0 {
			amt = nh * rt
		}

		items = append(items, models.NightBill{
			CompanyID:  req.CompanyID,
			EmployeeID: it.EmployeeID,
			Date:       it.Date,
			NightHours: nh,
			Rate:       rt,
			Amount:     amt,
			Month:      req.Month,
			Year:       req.Year,
			Status:     "pending",
			Remarks:    it.Remarks,
			CreatedBy:  c.GetString("user_id"),
		})
	}

	if err := h.nightBillRepo.BulkCreate(items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Night bills created successfully",
		"count":   len(items),
	})
}

// ProcessNightBills godoc
//
//	@Summary      Process night bills automatically
//	@Description  Auto-generate night bills based on attendance data and conditions
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body ProcessNightBillRequest true "Process conditions"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/process [post]
func (h *NightBillHandler) Process(c *gin.Context) {
	var req ProcessNightBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nightRate := req.NightRate
	if nightRate == 0 {
		nightRate = 80
	}
	minHours := req.MinHours
	if minHours == 0 {
		minHours = 4
	}

	employees, err := h.nightBillRepo.ListEligibleEmployees(
		req.CompanyID, req.DepartmentID, req.SectionID, req.DesignationID,
		req.DateFrom, req.DateTo,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var createdItems []models.NightBill
	var skipped []string
	totalAmount := 0.0

	for _, emp := range employees {
		nightHoursMap, err := h.nightBillRepo.GetAttendanceNightHours(emp.EmployeeID, req.DateFrom, req.DateTo)
		if err != nil {
			continue
		}

		employeeNightHours := 0.0
		for dateStr, hours := range nightHoursMap {
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				continue
			}
			month := int(date.Month())
			year := date.Year()

			if hours < minHours {
				continue
			}

			existing, isNew, err := h.nightBillRepo.FindOrCreateNightBill(req.CompanyID, emp.EmployeeID, dateStr, month, year)
			if err != nil {
				continue
			}

			if isNew {
				existing.NightHours = math.Round(hours*100) / 100
				existing.Rate = nightRate
				existing.Amount = math.Round(existing.NightHours*nightRate*100) / 100
				existing.DepartmentID = emp.DepartmentID
				existing.SectionID = emp.SectionID
				existing.DesignationID = emp.DesignationID
				existing.LineID = emp.LineID
				existing.GroupID = emp.GroupID
				existing.FloorID = emp.FloorID
				existing.CreatedBy = c.GetString("user_id")

				if err := h.nightBillRepo.Create(existing); err != nil {
					continue
				}
				createdItems = append(createdItems, *existing)
				totalAmount += existing.Amount
			} else {
				skipped = append(skipped, emp.EmployeeID+" on "+dateStr+" (already exists)")
			}
			employeeNightHours += hours
		}

		_ = employeeNightHours
	}

	createdCount := len(createdItems)
	process := &models.NightBillProcess{
		CompanyID:      req.CompanyID,
		Month:          req.Month,
		Year:           req.Year,
		TotalEmployees: createdCount,
		TotalAmount:    math.Round(totalAmount*100) / 100,
		Status:         "completed",
		ProcessedAt:    time.Now(),
		ProcessedBy:    c.GetString("user_id"),
		Remarks:        "Auto-processed from attendance",
	}
	_ = h.nightBillRepo.CreateProcess(process)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Night bills processed successfully",
		"total_created": createdCount,
		"total_skipped": len(skipped),
		"total_amount":  totalAmount,
		"process_id":    process.ID,
		"skipped":       skipped,
	})
}

// UpdateNightBill godoc
//
//	@Summary      Update night bill
//	@Description  Update an existing night bill
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        id      path string true "Night Bill ID"
//	@Param        request body CreateNightBillRequest false "Updated fields"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/{id} [put]
func (h *NightBillHandler) Update(c *gin.Context) {
	id := c.Param("id")

	item, err := h.nightBillRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Night bill not found"})
		return
	}

	var req CreateNightBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.EmployeeID = req.EmployeeID
	item.DepartmentID = req.DepartmentID
	item.SectionID = req.SectionID
	item.DesignationID = req.DesignationID
	item.LineID = req.LineID
	item.GroupID = req.GroupID
	item.FloorID = req.FloorID
	item.Date = req.Date
	item.Remarks = req.Remarks

	if req.NightHours != nil {
		item.NightHours = *req.NightHours
	}
	if req.Rate != nil {
		item.Rate = *req.Rate
	}
	if req.Amount != nil {
		item.Amount = *req.Amount
	} else if item.NightHours > 0 && item.Rate > 0 {
		item.Amount = item.NightHours * item.Rate
	}

	userID := c.GetString("user_id")
	item.UpdatedBy = &userID

	if err := h.nightBillRepo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// ApproveNightBill godoc
//
//	@Summary      Approve night bill
//	@Description  Approve a pending night bill
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        id path string true "Night Bill ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/{id}/approve [put]
func (h *NightBillHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.nightBillRepo.UpdateStatus(id, "approved", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Night bill approved"})
}

// RejectNightBill godoc
//
//	@Summary      Reject night bill
//	@Description  Reject a pending night bill
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        id path string true "Night Bill ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/{id}/reject [put]
func (h *NightBillHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.nightBillRepo.UpdateStatus(id, "rejected", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Night bill rejected"})
}

// DeleteNightBill godoc
//
//	@Summary      Delete night bill
//	@Description  Soft delete a night bill
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        id path string true "Night Bill ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/{id} [delete]
func (h *NightBillHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	item, err := h.nightBillRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Night bill not found"})
		return
	}

	if err := h.nightBillRepo.Delete(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Night bill deleted successfully"})
}

// ListNightBillProcesses godoc
//
//	@Summary      List night bill processes
//	@Description  Get processing history for night bills
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Param        month      query int    false "Month (1-12)"
//	@Param        year       query int    false "Year"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/processes [get]
func (h *NightBillHandler) ListProcesses(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	processes, err := h.nightBillRepo.ListProcesses(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  processes,
		"total": len(processes),
	})
}

// ListEligibleEmployees godoc
//
//	@Summary      List eligible employees for night bill
//	@Description  Get employees eligible for night bill based on filters
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id     query string false "Company ID"
//	@Param        department_id  query string false "Department ID"
//	@Param        section_id     query string false "Section ID"
//	@Param        designation_id query string false "Designation ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/eligible-employees [get]
func (h *NightBillHandler) ListEligibleEmployees(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}

	employees, err := h.nightBillRepo.ListEligibleEmployees(
		companyID,
		c.Query("department_id"),
		c.Query("section_id"),
		c.Query("designation_id"),
		"", "",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  employees,
		"total": len(employees),
	})
}

// ListEligibleWithRates godoc
//
//	@Summary      List eligible employees with pre-calculated night bill rates
//	@Description  Get employees with night hours, rate, and amount calculated from attendance for a given date
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id     query string false "Company ID"
//	@Param        department_id  query string false "Department ID"
//	@Param        section_id     query string false "Section ID"
//	@Param        designation_id query string false "Designation ID"
//	@Param        date           query string true "Date (YYYY-MM-DD)"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/employees-with-rates [get]
func (h *NightBillHandler) ListEligibleWithRates(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	employees, err := h.nightBillRepo.ListEligibleWithRates(
		companyID,
		c.Query("department_id"),
		c.Query("section_id"),
		c.Query("designation_id"),
		date,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  employees,
		"total": len(employees),
	})
}

// GetNightBillSummary godoc
//
//	@Summary      Get night bill summary
//	@Description  Get summary totals for night bills
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Param        month      query int    false "Month (1-12)"
//	@Param        year       query int    false "Year"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/summary [get]
func (h *NightBillHandler) Summary(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))

	totalEmployees, totalAmount, err := h.nightBillRepo.GetSummary(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_employees": totalEmployees,
		"total_amount":    totalAmount,
	})
}

type CalculateRateRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	Date       string `json:"date" binding:"required"`
}

// CalculateNightBillRate godoc
//
//	@Summary      Calculate night bill rate
//	@Description  Calculate night hours, rate, and amount for an employee on a given date from attendance
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body CalculateRateRequest true "Employee ID and Date"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /night-bills/calculate-rate [post]
func (h *NightBillHandler) CalculateRate(c *gin.Context) {
	var req CalculateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hours, _, err := h.nightBillRepo.GetAttendanceNightHoursByDate(req.EmployeeID, req.Date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rate := h.nightBillRepo.GetDefaultNightRate()
	amount := math.Round(hours*rate*100) / 100

	c.JSON(http.StatusOK, gin.H{
		"employee_id": req.EmployeeID,
		"date":        req.Date,
		"night_hours": hours,
		"rate":        rate,
		"amount":      amount,
	})
}
