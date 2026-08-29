package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type HolidayHandler struct {
	holidayRepo         *repository.HolidayRepository
	shiftRepo           *repository.ShiftRepository
	attendanceProcessor *service.AttendanceProcessor
}

func NewHolidayHandler(holidayRepo *repository.HolidayRepository, shiftRepo *repository.ShiftRepository, attendanceProcessor *service.AttendanceProcessor) *HolidayHandler {
	return &HolidayHandler{
		holidayRepo:         holidayRepo,
		shiftRepo:           shiftRepo,
		attendanceProcessor: attendanceProcessor,
	}
}

func (h *HolidayHandler) reprocessHolidayAttendance(date string, fromDate, toDate *string, companyID string) {
	startDate := date
	endDate := date
	if fromDate != nil && toDate != nil {
		startDate = *fromDate
		endDate = *toDate
	}
	_, _ = h.attendanceProcessor.ProcessDateRange(startDate, endDate, companyID)
}

// reprocessWeekendChange re-runs the attendance engine over both dates involved
// in a weekend change: the date that becomes General Duty and the date that
// becomes the Weekend. This guarantees the exchange applies immediately without
// a server restart.
func (h *HolidayHandler) reprocessWeekendChange(date, weekendDate, companyID string) {
	startDate := date
	endDate := date
	if weekendDate != "" {
		if weekendDate < date {
			startDate = weekendDate
			endDate = date
		} else {
			startDate = date
			endDate = weekendDate
		}
	}
	_, _ = h.attendanceProcessor.ProcessDateRange(startDate, endDate, companyID)
}

// validateWeekendChange enforces the weekend-change business rules. Returns an
// error message when the exchange is invalid, otherwise an empty string.
func (h *HolidayHandler) validateWeekendChange(date, weekendDate, companyID, excludeID string) string {
	if weekendDate != "" && weekendDate == date {
		return "Weekend date and general duty date must be different"
	}

	// The general duty date is the day that was originally a weekend and is being
	// turned into a working day, so it must be an official weekend for the company
	// (its weekday must appear in at least one active shift's WeekendDays).
	shifts, err := h.shiftRepo.ListActiveByCompany(companyID)
	if err != nil {
		return "Unable to validate weekend date against company shifts"
	}
	officialWeekend := false
	for _, s := range shifts {
		if s.WeekendDays != "" && utils.IsWeekend(date, s.WeekendDays) {
			officialWeekend = true
			break
		}
	}
	if !officialWeekend {
		return "General duty date is not an official weekend"
	}

	// The general duty date must not already be a government holiday.
	if count, err := h.holidayRepo.CountActiveHolidayOnDate(date, companyID); err != nil {
		return "Failed to validate general duty date against holidays"
	} else if count > 0 {
		return "General duty date is already a holiday"
	}

	// Prevent duplicate active exchanges for either date.
	if count, err := h.holidayRepo.CountActiveWeekendChangeCollision(date, companyID, excludeID); err != nil {
		return "Failed to validate duplicate weekend changes"
	} else if count > 0 {
		return "An active weekend change already exists for general duty date"
	}
	if weekendDate != "" {
		if count, err := h.holidayRepo.CountActiveWeekendChangeCollision(weekendDate, companyID, excludeID); err != nil {
			return "Failed to validate duplicate weekend changes"
		} else if count > 0 {
			return "An active weekend change already exists for weekend date"
		}
	}

	return ""
}

type CreateHolidayRequest struct {
	Name        string  `json:"name" binding:"required"`
	CompanyID   string  `json:"company_id" binding:"required"`
	Date        string  `json:"date" binding:"required"`
	FromDate    *string `json:"from_date"`
	ToDate      *string `json:"to_date"`
	WeekendDate *string `json:"weekend_date"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
}

type UpdateHolidayRequest struct {
	Name        string  `json:"name"`
	Date        string  `json:"date"`
	FromDate    *string `json:"from_date"`
	ToDate      *string `json:"to_date"`
	WeekendDate *string `json:"weekend_date"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
}

// ListHolidays godoc
//
// @Summary      List holidays
// @Description  Get all holidays with optional company filter
// @Tags         Holidays
// @Security     BearerAuth
// @Produce      json
// @Param        company_id query string false "Filter by company ID"
// @Param        page       query int    false "Page number"
// @Param        limit      query int    false "Page size"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /holidays [get]
func (h *HolidayHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	p := utils.ParsePagination(c)
	holidays, total, err := h.holidayRepo.List(companyID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(holidays, total, p))
}

// GetHoliday godoc
//
// @Summary      Get holiday by ID
// @Description  Get a single holiday
// @Tags         Holidays
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Holiday ID"
// @Success      200  {object}  models.Holiday
// @Failure      404  {object}  map[string]string
// @Router       /holidays/{id} [get]
func (h *HolidayHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	holiday, err := h.holidayRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "holiday not found"})
		return
	}
	c.JSON(http.StatusOK, holiday)
}

// CreateHoliday godoc
//
// @Summary      Create holiday
// @Description  Create a new holiday
// @Tags         Holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateHolidayRequest true "Holiday details"
// @Success      201  {object}  models.Holiday
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /holidays [post]
func (h *HolidayHandler) Create(c *gin.Context) {
	var req CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "government"
	}

	if req.FromDate != nil && req.ToDate != nil {
		if _, err := time.Parse("2006-01-02", *req.FromDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date format, use YYYY-MM-DD"})
			return
		}
		if _, err := time.Parse("2006-01-02", *req.ToDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_date format, use YYYY-MM-DD"})
			return
		}
	} else if req.FromDate != nil || req.ToDate != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both from_date and to_date must be provided together"})
		return
	}

	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	if req.Type == "weekend_change" || req.WeekendDate != nil {
		weekendDate := ""
		if req.WeekendDate != nil {
			weekendDate = *req.WeekendDate
		}
		if msg := h.validateWeekendChange(req.Date, weekendDate, req.CompanyID, ""); msg != "" {
			c.JSON(http.StatusConflict, gin.H{"error": msg})
			return
		}
	}

	userID := c.GetString("user_id")

	holiday := models.Holiday{
		CompanyID:   req.CompanyID,
		Name:        req.Name,
		Date:        req.Date,
		FromDate:    req.FromDate,
		ToDate:      req.ToDate,
		WeekendDate: req.WeekendDate,
		Type:        req.Type,
		Description: req.Description,
		Status:      "active",
		CreatedBy:   &userID,
	}

	if err := h.holidayRepo.Create(&holiday); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if holiday.Type == "weekend_change" {
		weekendDate := ""
		if holiday.WeekendDate != nil {
			weekendDate = *holiday.WeekendDate
		}
		go h.reprocessWeekendChange(holiday.Date, weekendDate, holiday.CompanyID)
	} else {
		go h.reprocessHolidayAttendance(holiday.Date, holiday.FromDate, holiday.ToDate, holiday.CompanyID)
	}

	c.JSON(http.StatusCreated, holiday)
}

// UpdateHoliday godoc
//
// @Summary      Update holiday
// @Description  Update an existing holiday
// @Tags         Holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Holiday ID"
// @Param        request body UpdateHolidayRequest true "Holiday details"
// @Success      200  {object}  models.Holiday
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /holidays/{id} [put]
func (h *HolidayHandler) Update(c *gin.Context) {
	id := c.Param("id")
	holiday, err := h.holidayRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "holiday not found"})
		return
	}

	var req UpdateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		holiday.Name = req.Name
	}
	if req.FromDate != nil && req.ToDate != nil {
		if _, err := time.Parse("2006-01-02", *req.FromDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date format, use YYYY-MM-DD"})
			return
		}
		if _, err := time.Parse("2006-01-02", *req.ToDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_date format, use YYYY-MM-DD"})
			return
		}
		holiday.FromDate = req.FromDate
		holiday.ToDate = req.ToDate
	} else if req.FromDate != nil || req.ToDate != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both from_date and to_date must be provided together"})
		return
	}
	if req.Date != "" {
		if _, err := time.Parse("2006-01-02", req.Date); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		holiday.Date = req.Date
	}
	if req.Type != "" {
		holiday.Type = req.Type
	}
	if req.Description != "" {
		holiday.Description = req.Description
	}
	if req.WeekendDate != nil {
		holiday.WeekendDate = req.WeekendDate
	}
	if req.Status != "" {
		holiday.Status = req.Status
	}

	if holiday.Type == "weekend_change" || holiday.WeekendDate != nil {
		weekendDate := ""
		if holiday.WeekendDate != nil {
			weekendDate = *holiday.WeekendDate
		}
		if msg := h.validateWeekendChange(holiday.Date, weekendDate, holiday.CompanyID, id); msg != "" {
			c.JSON(http.StatusConflict, gin.H{"error": msg})
			return
		}
	}

	userID := c.GetString("user_id")
	holiday.UpdatedBy = &userID

	if err := h.holidayRepo.Update(holiday); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if holiday.Type == "weekend_change" {
		weekendDate := ""
		if holiday.WeekendDate != nil {
			weekendDate = *holiday.WeekendDate
		}
		go h.reprocessWeekendChange(holiday.Date, weekendDate, holiday.CompanyID)
	} else {
		go h.reprocessHolidayAttendance(holiday.Date, holiday.FromDate, holiday.ToDate, holiday.CompanyID)
	}

	c.JSON(http.StatusOK, holiday)
}

type BulkAdvanceHolidayRequest struct {
	CompanyID   string   `json:"company_id" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	FromDate    string   `json:"from_date" binding:"required"`
	ToDate      string   `json:"to_date"`
	Dates       []string `json:"dates"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	AdvanceOnly *bool    `json:"advance_only"`
}

type BulkDeleteHolidayRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// BulkAdvanceHoliday godoc
//
// @Summary      Bulk advance holidays
// @Description  Create government holidays in advance for future dates. Expands date range per-day, skips existing active holiday dates.
// @Tags         Holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkAdvanceHolidayRequest true "Bulk advance holiday payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /holidays/bulk-advance [post]
func (h *HolidayHandler) BulkAdvanceHoliday(c *gin.Context) {
	var req BulkAdvanceHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Type == "" {
		req.Type = "government"
	}
	if req.Type != "government" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bulk advance holiday only supports type=government"})
		return
	}
	advanceOnly := true
	if req.AdvanceOnly != nil {
		advanceOnly = *req.AdvanceOnly
	}
	today := time.Now().UTC().Format("2006-01-02")
	var dates []string
	if len(req.Dates) > 0 {
		for _, d := range req.Dates {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dates entry format, use YYYY-MM-DD"})
				return
			}
			if advanceOnly && d < today {
				c.JSON(http.StatusBadRequest, gin.H{"error": "advance_only: date must be today or future: " + d})
				return
			}
		}
		dates = req.Dates
	} else {
		toDate := req.ToDate
		if toDate == "" {
			toDate = req.FromDate
		}
		var err error
		dates, err = utils.GenerateDateRange(req.FromDate, toDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date/to_date format, use YYYY-MM-DD"})
			return
		}
		if advanceOnly {
			for _, d := range dates {
				if d < today {
					c.JSON(http.StatusBadRequest, gin.H{"error": "advance_only: date must be today or future: " + d})
					return
				}
			}
		}
	}
	userID := c.GetString("user_id")
	var created []models.Holiday
	var skipped []string
	for _, d := range dates {
		existing, _ := h.holidayRepo.ListActiveByDate(d, req.CompanyID)
		already := false
		for _, ex := range existing {
			if ex.Type != "weekend_change" {
				already = true
				break
			}
		}
		if already {
			skipped = append(skipped, d)
			continue
		}
		holiday := models.Holiday{
			CompanyID:   req.CompanyID,
			Name:        req.Name,
			Date:        d,
			Type:        req.Type,
			Description: req.Description,
			Status:      "active",
			CreatedBy:   &userID,
		}
		if err := h.holidayRepo.Create(&holiday); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = append(created, holiday)
	}
	if len(created) > 0 {
		startDate := dates[0]
		endDate := dates[len(dates)-1]
		go h.reprocessHolidayAttendance(startDate, &startDate, &endDate, req.CompanyID)
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Bulk advance holidays created",
		"count":   len(created),
		"skipped": skipped,
		"records": created,
	})
}

// AdvancePreview godoc
//
// @Summary      Preview advance holiday/general duty
// @Description  Dry-run expansion of dates and collision check without DB write.
// @Tags         Holidays
// @Security     BearerAuth
// @Produce      json
// @Param        company_id query string true "Company ID"
// @Param        type query string false "Type government|weekend_change"
// @Param        from_date query string false "From date"
// @Param        to_date query string false "To date"
// @Param        dates query string false "Comma-separated dates for preview"
// @Param        advance_only query bool false "Future only"
// @Success      200  {object}  map[string]interface{}
// @Router       /holidays/advance-preview [get]
func (h *HolidayHandler) AdvancePreview(c *gin.Context) {
	companyID := c.Query("company_id")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")
	datesParam := c.Query("dates")
	hType := c.Query("type")
	if hType == "" {
		hType = "government"
	}
	advanceOnlyStr := c.Query("advance_only")
	advanceOnly := advanceOnlyStr != "false"
	today := time.Now().UTC().Format("2006-01-02")
	var dates []string
	if datesParam != "" {
		for _, d := range splitCSV(datesParam) {
			d = utils.NormalizeDate(d)
			if _, err := time.Parse("2006-01-02", d); err != nil {
				continue
			}
			dates = append(dates, d)
		}
	} else if fromDate != "" {
		if toDate == "" {
			toDate = fromDate
		}
		var err error
		dates, err = utils.GenerateDateRange(fromDate, toDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date/to_date"})
			return
		}
	}
	var collisions []string
	var valid []string
	for _, d := range dates {
		if advanceOnly && d < today {
			collisions = append(collisions, d+": past date (advance_only)")
			continue
		}
		if hType == "government" {
			list, _ := h.holidayRepo.ListActiveByDate(d, companyID)
			has := false
			for _, ex := range list {
				if ex.Type != "weekend_change" {
					has = true
					break
				}
			}
			if has {
				collisions = append(collisions, d+": already holiday")
			} else {
				valid = append(valid, d)
			}
		} else {
			valid = append(valid, d)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"expanded":   dates,
		"valid":      valid,
		"collisions": collisions,
	})
}

func splitCSV(s string) []string {
	var out []string
	start := 0
	for i, ch := range s {
		if ch == ',' {
			out = append(out, trimSpace(s[start:i]))
			start = i + 1
		}
	}
	out = append(out, trimSpace(s[start:]))
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// BulkDelete godoc
//
// @Summary      Bulk delete holidays
// @Tags         Holidays
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body BulkDeleteHolidayRequest true "IDs to delete"
// @Success      200  {object}  map[string]string
// @Router       /holidays/bulk-delete [post]
func (h *HolidayHandler) BulkDelete(c *gin.Context) {
	var req BulkDeleteHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.holidayRepo.DeleteBulk(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Holidays deleted", "count": len(req.IDs)})
}

// DeleteHoliday godoc
//
// @Summary      Delete holiday
// @Description  Soft delete a holiday
// @Tags         Holidays
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Holiday ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /holidays/{id} [delete]
func (h *HolidayHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	holiday, err := h.holidayRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "holiday not found"})
		return
	}
	if err := h.holidayRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if holiday.Type == "weekend_change" {
		weekendDate := ""
		if holiday.WeekendDate != nil {
			weekendDate = *holiday.WeekendDate
		}
		go h.reprocessWeekendChange(holiday.Date, weekendDate, holiday.CompanyID)
	} else {
		go h.reprocessHolidayAttendance(holiday.Date, holiday.FromDate, holiday.ToDate, holiday.CompanyID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "holiday deleted"})
}
