package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type CompanyHandler struct {
	companyRepo *repository.CompanyRepository
}

func NewCompanyHandler(companyRepo *repository.CompanyRepository) *CompanyHandler {
	return &CompanyHandler{companyRepo: companyRepo}
}

type CreateCompanyRequest struct {
	CompanyNameBn string `json:"company_name_bn"`
	CompanyNameEn string `json:"company_name_en" binding:"required"`
	AddressBn     string `json:"address_bn"`
	AddressEn     string `json:"address_en"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Signature     string `json:"signature"`
	Status        string `json:"status"`
}

type UpdateCompanyRequest struct {
	CompanyNameBn string `json:"company_name_bn"`
	CompanyNameEn string `json:"company_name_en" binding:"required"`
	AddressBn     string `json:"address_bn"`
	AddressEn     string `json:"address_en"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Signature     string `json:"signature"`
	Status        string `json:"status"`
}

// ListCompanies godoc
//
// @Summary      List companies
// @Description  Get all companies
// @Tags         Companies
// @Security     BearerAuth
// @Produce      json
// @Param        page   query int    false "Page number (default: 1)"
// @Param        limit  query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /companies [get]
func (h *CompanyHandler) List(c *gin.Context) {
	p := utils.ParsePagination(c)
	companies, total, err := h.companyRepo.List(p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(companies, total, p))
}

// GetCompany godoc
//
// @Summary      Get company by ID
// @Description  Get a company by its ID
// @Tags         Companies
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Company ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /companies/{id} [get]
func (h *CompanyHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	company, err := h.companyRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}
	c.JSON(http.StatusOK, company)
}

// CreateCompany godoc
//
// @Summary      Create company
// @Description  Create a new company
// @Tags         Companies
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateCompanyRequest true "Company details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /companies [post]
func (h *CompanyHandler) Create(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(req.CompanyNameEn, " ", "-"))

	existing, err := h.companyRepo.FindBySlug(slug)
	if err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "company with this name already exists"})
		return
	}

	userID := c.GetString("user_id")
	status := req.Status
	if status == "" {
		status = "active"
	}
	company := &models.Company{
		CompanyNameBn: req.CompanyNameBn,
		CompanyNameEn: req.CompanyNameEn,
		Slug:          slug,
		AddressBn:     req.AddressBn,
		AddressEn:     req.AddressEn,
		Phone:         req.Phone,
		Email:         req.Email,
		Signature:     req.Signature,
		OwnerID:       &userID,
		Status:        status,
		CreatedBy:     &userID,
	}

	if err := h.companyRepo.Create(company); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, company)
}

// UpdateCompany godoc
//
// @Summary      Update company
// @Description  Update an existing company
// @Tags         Companies
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Company ID"
// @Param        request body UpdateCompanyRequest true "Updated company details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /companies/{id} [put]
func (h *CompanyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	company, err := h.companyRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(req.CompanyNameEn, " ", "-"))
	existing, err := h.companyRepo.FindBySlug(slug)
	if err == nil && existing != nil && existing.ID != company.ID {
		c.JSON(http.StatusConflict, gin.H{"error": "company with this name already exists"})
		return
	}

	userID := c.GetString("user_id")
	company.CompanyNameBn = req.CompanyNameBn
	company.CompanyNameEn = req.CompanyNameEn
	company.Slug = slug
	company.AddressBn = req.AddressBn
	company.AddressEn = req.AddressEn
	company.Phone = req.Phone
	company.Email = req.Email
	company.Signature = req.Signature
	if req.Status != "" {
		company.Status = req.Status
	}
	company.UpdatedBy = &userID

	if err := h.companyRepo.Update(company); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, company)
}

// DeleteCompany godoc
//
// @Summary      Delete company
// @Description  Soft delete a company
// @Tags         Companies
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Company ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /companies/{id} [delete]
func (h *CompanyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	company, err := h.companyRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	userID := c.GetString("user_id")
	company.DeletedBy = &userID

	if err := h.companyRepo.Update(company); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.companyRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "company deleted"})
}
