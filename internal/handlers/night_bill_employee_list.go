package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type NightBillEmployeeListHandler struct {
	repo *repository.NightBillEmployeeListRepository
}

func NewNightBillEmployeeListHandler(repo *repository.NightBillEmployeeListRepository) *NightBillEmployeeListHandler {
	return &NightBillEmployeeListHandler{repo: repo}
}

// ─── Request / Response types ──────────────────────────────────────────────

type CreateNightBillEmployeeListRequest struct {
	CompanyID   string  `json:"company_id" binding:"required"`
	EmployeeID  string  `json:"employee_id" binding:"required"`
	BillType    string  `json:"bill_type" binding:"required,oneof=fixed hourly"`
	FixedAmount float64 `json:"fixed_amount"`
	HourlyRate  float64 `json:"hourly_rate"`
	Remarks     string  `json:"remarks"`
}

type BulkCreateNightBillEmployeeListRequest struct {
	Entries []CreateNightBillEmployeeListRequest `json:"entries" binding:"required,min=1"`
}

type UpdateNightBillEmployeeListRequest struct {
	BillType    string  `json:"bill_type"`
	FixedAmount float64 `json:"fixed_amount"`
	HourlyRate  float64 `json:"hourly_rate"`
	IsActive    *bool   `json:"is_active"`
	Remarks     string  `json:"remarks"`
}

// ─── Handlers ──────────────────────────────────────────────────────────────

// List godoc
//
//	@Summary      List night bill employee list
//	@Tags         Night Bill Employee List
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /attendance/night-bill/employee-list [get]
func (h *NightBillEmployeeListHandler) List(c *gin.Context) {
	p := utils.ParsePagination(c)
	companyID := c.Query("company_id")

	list, total, err := h.repo.List(companyID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

// Create godoc
//
//	@Summary      Add employee to night bill list
//	@Tags         Night Bill Employee List
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill/employee-list [post]
func (h *NightBillEmployeeListHandler) Create(c *gin.Context) {
	var req CreateNightBillEmployeeListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent duplicates
	if h.repo.ExistsForEmployee(req.EmployeeID) {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Employee %s is already in the night bill list", req.EmployeeID)})
		return
	}

	userID := c.GetString("user_id")
	rec := &models.NightBillEmployeeList{
		CompanyID:   req.CompanyID,
		EmployeeID:  req.EmployeeID,
		BillType:    req.BillType,
		FixedAmount: req.FixedAmount,
		HourlyRate:  req.HourlyRate,
		IsActive:    true,
		Remarks:     req.Remarks,
		CreatedBy:   &userID,
		UpdatedBy:   &userID,
	}

	if err := h.repo.Create(rec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload with preloads
	loaded, err := h.repo.FindByID(rec.ID)
	if err != nil {
		c.JSON(http.StatusCreated, rec)
		return
	}
	c.JSON(http.StatusCreated, loaded)
}

// BulkCreate godoc
//
//	@Summary      Bulk add employees to night bill list
//	@Tags         Night Bill Employee List
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill/employee-list/bulk [post]
func (h *NightBillEmployeeListHandler) BulkCreate(c *gin.Context) {
	var req BulkCreateNightBillEmployeeListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	var recs []*models.NightBillEmployeeList
	skipped := []string{}

	for _, e := range req.Entries {
		if h.repo.ExistsForEmployee(e.EmployeeID) {
			skipped = append(skipped, e.EmployeeID)
			continue
		}
		recs = append(recs, &models.NightBillEmployeeList{
			CompanyID:   e.CompanyID,
			EmployeeID:  e.EmployeeID,
			BillType:    e.BillType,
			FixedAmount: e.FixedAmount,
			HourlyRate:  e.HourlyRate,
			IsActive:    true,
			Remarks:     e.Remarks,
			CreatedBy:   &userID,
			UpdatedBy:   &userID,
		})
	}

	if err := h.repo.BulkCreate(recs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"saved":   len(recs),
		"skipped": skipped,
		"message": fmt.Sprintf("%d employee(s) added, %d skipped (already in list)", len(recs), len(skipped)),
	})
}

// Update godoc
//
//	@Summary      Update night bill employee list entry
//	@Tags         Night Bill Employee List
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill/employee-list/{id} [put]
func (h *NightBillEmployeeListHandler) Update(c *gin.Context) {
	id := c.Param("id")
	rec, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	var req UpdateNightBillEmployeeListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if req.BillType != "" {
		rec.BillType = req.BillType
	}
	rec.FixedAmount = req.FixedAmount
	rec.HourlyRate = req.HourlyRate
	if req.IsActive != nil {
		rec.IsActive = *req.IsActive
	}
	if req.Remarks != "" {
		rec.Remarks = req.Remarks
	}
	rec.UpdatedBy = &userID

	if err := h.repo.Update(rec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rec)
}

// Delete godoc
//
//	@Summary      Remove employee from night bill list
//	@Tags         Night Bill Employee List
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /attendance/night-bill/employee-list/{id} [delete]
func (h *NightBillEmployeeListHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Removed from night bill employee list"})
}

// BulkDelete godoc
//
//	@Summary      Bulk remove employees from night bill list
//	@Tags         Night Bill Employee List
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Router       /attendance/night-bill/employee-list/bulk-delete [post]
func (h *NightBillEmployeeListHandler) BulkDelete(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.BulkDelete(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": strconv.Itoa(len(req.IDs)) + " record(s) removed"})
}

// CheckEmployee godoc
//
//	@Summary      Check if employee is in night bill list
//	@Tags         Night Bill Employee List
//	@Security     BearerAuth
//	@Produce      json
//	@Router       /attendance/night-bill/employee-list/check/{employee_id} [get]
func (h *NightBillEmployeeListHandler) CheckEmployee(c *gin.Context) {
	empID := c.Param("employee_id")
	exists := h.repo.ExistsForEmployee(empID)
	var rec *models.NightBillEmployeeList
	if exists {
		rec, _ = h.repo.FindByEmployeeID(empID)
	}
	c.JSON(http.StatusOK, gin.H{
		"exists": exists,
		"record": rec,
	})
}
