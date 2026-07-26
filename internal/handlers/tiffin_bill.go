package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type TiffinBillHandler struct {
	tiffinBillRepo *repository.TiffinBillRepository
}

func NewTiffinBillHandler(tiffinBillRepo *repository.TiffinBillRepository) *TiffinBillHandler {
	return &TiffinBillHandler{tiffinBillRepo: tiffinBillRepo}
}

type CreateTiffinBillRequest struct {
	CompanyID  string  "json:\"company_id\" binding:\"required\""
	EmployeeID string  "json:\"employee_id\" binding:\"required\""
	Date       string  "json:\"date\" binding:\"required\""
	Amount     float64 "json:\"amount\""
	Month      int     "json:\"month\" binding:\"required\""
	Year       int     "json:\"year\" binding:\"required\""
	Remarks    string  "json:\"remarks\""
}

// ListTiffinBills godoc
//
//	@Summary      List tiffin bills
//	@Description  Get all tiffin bills for a company, optionally filtered by month/year
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        company_id query string false "Company ID"
//	@Param        month      query int    false "Month (1-12)"
//	@Param        year       query int    false "Year"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      500  {object}  map[string]string
//	@Router       /tiffin-bills [get]
func (h *TiffinBillHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}

	monthStr := c.Query("month")
	yearStr := c.Query("year")

	var items []models.TiffinBill
	var err error

	if monthStr != "" && yearStr != "" {
		month, _ := strconv.Atoi(monthStr)
		year, _ := strconv.Atoi(yearStr)
		items, err = h.tiffinBillRepo.ListByMonth(companyID, month, year)
	} else {
		items, err = h.tiffinBillRepo.List(companyID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tiffin_bills": items,
		"total":        len(items),
	})
}

// CreateTiffinBill godoc
//
//	@Summary      Create tiffin bill
//	@Description  Create a new tiffin bill entry for an employee
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body CreateTiffinBillRequest true "Tiffin bill details"
//	@Success      201  {object}  map[string]interface{}
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /tiffin-bills [post]
func (h *TiffinBillHandler) Create(c *gin.Context) {
	var req CreateTiffinBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &models.TiffinBill{
		CompanyID:  req.CompanyID,
		EmployeeID: req.EmployeeID,
		Date:       req.Date,
		Amount:     req.Amount,
		Month:      req.Month,
		Year:       req.Year,
		Status:     "pending",
		Remarks:    req.Remarks,
		CreatedBy:  c.GetString("user_id"),
	}

	if err := h.tiffinBillRepo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateTiffinBill godoc
//
//	@Summary      Update tiffin bill
//	@Description  Update an existing tiffin bill
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        id      path string true "Tiffin Bill ID"
//	@Param        request body CreateTiffinBillRequest false "Updated fields"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /tiffin-bills/{id} [put]
func (h *TiffinBillHandler) Update(c *gin.Context) {
	id := c.Param("id")

	item, err := h.tiffinBillRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tiffin bill not found"})
		return
	}

	var req CreateTiffinBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.Date = req.Date
	item.Amount = req.Amount
	item.Month = req.Month
	item.Year = req.Year
	item.Remarks = req.Remarks
	userID := c.GetString("user_id")
	item.UpdatedBy = &userID

	if err := h.tiffinBillRepo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteTiffinBill godoc
//
//	@Summary      Delete tiffin bill
//	@Description  Soft delete a tiffin bill
//	@Tags         Payroll
//	@Security     BearerAuth
//	@Produce      json
//	@Param        id path string true "Tiffin Bill ID"
//	@Success      200  {object}  map[string]interface{}
//	@Failure      404  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /tiffin-bills/{id} [delete]
func (h *TiffinBillHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	item, err := h.tiffinBillRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tiffin bill not found"})
		return
	}

	if err := h.tiffinBillRepo.Delete(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tiffin bill deleted successfully"})
}
