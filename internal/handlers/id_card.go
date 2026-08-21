package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type IdCardHandler struct {
	repo *repository.IdCardRepository
}

func NewIdCardHandler(repo *repository.IdCardRepository) *IdCardHandler {
	return &IdCardHandler{repo: repo}
}

type CreateIdCardRequest struct {
	Employee      string `json:"employee" binding:"required"`
	EmployeeID    string `json:"employee_id"`
	DesignationID string `json:"designation_id"`
	DepartmentID  string `json:"department_id" binding:"required"`
	CardNo        string `json:"card_no" binding:"required"`
	Issued        string `json:"issued"`
	Expiry        string `json:"expiry"`
	Status        string `json:"status"`
}

type UpdateIdCardRequest struct {
	Employee      string `json:"employee" binding:"required"`
	EmployeeID    string `json:"employee_id"`
	DesignationID string `json:"designation_id"`
	DepartmentID  string `json:"department_id" binding:"required"`
	CardNo        string `json:"card_no" binding:"required"`
	Issued        string `json:"issued"`
	Expiry        string `json:"expiry"`
	Status        string `json:"status"`
}

// ListIdCards godoc
//
// @Summary      List ID cards
// @Tags         ID Cards
// @Security     BearerAuth
// @Produce      json
// @Param        employee       query string false "Filter by employee name"
// @Param        employee_id    query string false "Filter by employee ID"
// @Param        department_id  query string false "Filter by department"
// @Param        designation_id query string false "Filter by designation"
// @Param        status         query string false "Filter by status"
// @Param        card_no        query string false "Filter by card number"
// @Param        page           query int    false "Page number (default: 1)"
// @Param        limit          query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /id-cards [get]
func (h *IdCardHandler) List(c *gin.Context) {
	employee := c.Query("employee")
	employeeID := c.Query("employee_id")
	departmentID := c.Query("department_id")
	designationID := c.Query("designation_id")
	status := c.Query("status")
	cardNo := c.Query("card_no")

	p := utils.ParsePagination(c)
	if employee != "" || employeeID != "" || departmentID != "" || designationID != "" || status != "" || cardNo != "" {
		items, total, err := h.repo.ListFiltered(employee, employeeID, departmentID, designationID, status, cardNo, p.Page, p.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, utils.NewPaginatedResponse(items, total, p))
		return
	}

	items, total, err := h.repo.List(p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(items, total, p))
}

func (h *IdCardHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "id card not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *IdCardHandler) Create(c *gin.Context) {
	var req CreateIdCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "Active"
	}

	item := &models.IdCard{
		Employee:      req.Employee,
		EmployeeID:    req.EmployeeID,
		DesignationID: req.DesignationID,
		DepartmentID:  req.DepartmentID,
		CardNo:        req.CardNo,
		Issued:        req.Issued,
		Expiry:        req.Expiry,
		Status:        status,
	}

	if err := h.repo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, _ := h.repo.FindByID(item.ID)
	c.JSON(http.StatusCreated, result)
}

func (h *IdCardHandler) Update(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "id card not found"})
		return
	}

	var req UpdateIdCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.Employee = req.Employee
	item.EmployeeID = req.EmployeeID
	item.DesignationID = req.DesignationID
	item.DepartmentID = req.DepartmentID
	item.CardNo = req.CardNo
	item.Issued = req.Issued
	item.Expiry = req.Expiry
	if req.Status != "" {
		item.Status = req.Status
	}

	if err := h.repo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, _ := h.repo.FindByID(item.ID)
	c.JSON(http.StatusOK, result)
}

func (h *IdCardHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	_, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "id card not found"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "id card deleted"})
}
