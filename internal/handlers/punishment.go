package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type PunishmentHandler struct {
	punishmentRepo *repository.PunishmentRepository
}

func NewPunishmentHandler(punishmentRepo *repository.PunishmentRepository) *PunishmentHandler {
	return &PunishmentHandler{punishmentRepo: punishmentRepo}
}

type CreatePunishmentRequest struct {
	CompanyID  string   `json:"company_id" binding:"required"`
	EmployeeID string   `json:"employee_id" binding:"required"`
	PunishmentType string `json:"punishment_type" binding:"required"`
	Reason     string   `json:"reason"`
	Amount     float64  `json:"amount"`
	OvertimeLessHours *float64 `json:"overtime_less_hours"`
	OvertimeRate      *float64 `json:"overtime_rate"`
	AbsentDays        *int     `json:"absent_days"`
	PerDayRate        *float64 `json:"per_day_rate"`
	Date       string   `json:"date" binding:"required"`
	Remarks    string   `json:"remarks"`
}

func calculateAmount(pt string, req CreatePunishmentRequest) float64 {
	switch pt {
	case "ot_less":
		hours := 0.0
		rate := 0.0
		if req.OvertimeLessHours != nil {
			hours = *req.OvertimeLessHours
		}
		if req.OvertimeRate != nil {
			rate = *req.OvertimeRate
		}
		return hours * rate
	case "absent":
		days := 0
		rate := 0.0
		if req.AbsentDays != nil {
			days = *req.AbsentDays
		}
		if req.PerDayRate != nil {
			rate = *req.PerDayRate
		}
		return float64(days) * rate
	default:
		return req.Amount
	}
}

// CalculatePunishment godoc
//
//	@Summary      Calculate punishment amount
//	@Description  Calculate the amount based on punishment type and parameters
//	@Tags         HR
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body CreatePunishmentRequest true "Punishment calculation params"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Router       /punishments/calculate [post]
func (h *PunishmentHandler) Calculate(c *gin.Context) {
	var req CreatePunishmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	amount := calculateAmount(req.PunishmentType, req)
	c.JSON(http.StatusOK, gin.H{"amount": amount})
}

// ListPunishments godoc
//
//	@Summary      List punishments
//	@Description  Get all punishments for a company
//	@Tags         HR
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /punishments [get]
func (h *PunishmentHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}

	items, err := h.punishmentRepo.List(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"punishments": items,
		"total":       len(items),
	})
}

// CreatePunishment godoc
//
//	@Summary      Create punishment
//	@Description  Add a new punishment record for an employee. Amount is auto-calculated based on punishment_type.
//	@Tags         HR
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body CreatePunishmentRequest true "Punishment details"
//	@Success      201  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /punishments [post]
func (h *PunishmentHandler) Create(c *gin.Context) {
	var req CreatePunishmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := calculateAmount(req.PunishmentType, req)

	item := &models.Punishment{
		CompanyID:         req.CompanyID,
		EmployeeID:        req.EmployeeID,
		PunishmentType:    req.PunishmentType,
		Reason:            req.Reason,
		Amount:            amount,
		OvertimeLessHours: req.OvertimeLessHours,
		OvertimeRate:      req.OvertimeRate,
		AbsentDays:        req.AbsentDays,
		PerDayRate:        req.PerDayRate,
		Date:              req.Date,
		Status:            "active",
		Remarks:           req.Remarks,
		CreatedBy:         c.GetString("user_id"),
	}

	if err := h.punishmentRepo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdatePunishment godoc
//
//	@Summary      Update punishment
//	@Description  Update an existing punishment record
//	@Tags         HR
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        id      path string true "Punishment ID"
//	@Param        request body CreatePunishmentRequest false "Updated fields"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /punishments/{id} [put]
func (h *PunishmentHandler) Update(c *gin.Context) {
	id := c.Param("id")

	item, err := h.punishmentRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Punishment not found"})
		return
	}

	var req CreatePunishmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := calculateAmount(req.PunishmentType, req)

	item.PunishmentType = req.PunishmentType
	item.Reason = req.Reason
	item.Amount = amount
	item.OvertimeLessHours = req.OvertimeLessHours
	item.OvertimeRate = req.OvertimeRate
	item.AbsentDays = req.AbsentDays
	item.PerDayRate = req.PerDayRate
	item.Date = req.Date
	item.Remarks = req.Remarks
	userID := c.GetString("user_id")
	item.UpdatedBy = &userID

	if err := h.punishmentRepo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeletePunishment godoc
//
//	@Summary      Delete punishment
//	@Description  Soft delete a punishment record
//	@Tags         HR
//	@Security     BearerAuth
//	@Produce      json
//	@Param        id path string true "Punishment ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /punishments/{id} [delete]
func (h *PunishmentHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	item, err := h.punishmentRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Punishment not found"})
		return
	}

	if err := h.punishmentRepo.Delete(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Punishment deleted successfully"})
}
