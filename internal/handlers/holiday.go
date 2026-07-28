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
	attendanceProcessor *service.AttendanceProcessor
}

func NewHolidayHandler(holidayRepo *repository.HolidayRepository, attendanceProcessor *service.AttendanceProcessor) *HolidayHandler {
	return &HolidayHandler{
		holidayRepo:         holidayRepo,
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

	go h.reprocessHolidayAttendance(holiday.Date, holiday.FromDate, holiday.ToDate, holiday.CompanyID)

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

	userID := c.GetString("user_id")
	holiday.UpdatedBy = &userID

	if err := h.holidayRepo.Update(holiday); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go h.reprocessHolidayAttendance(holiday.Date, holiday.FromDate, holiday.ToDate, holiday.CompanyID)

	c.JSON(http.StatusOK, holiday)
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

	go h.reprocessHolidayAttendance(holiday.Date, holiday.FromDate, holiday.ToDate, holiday.CompanyID)

	c.JSON(http.StatusOK, gin.H{"message": "holiday deleted"})
}
