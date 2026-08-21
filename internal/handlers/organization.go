package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type DepartmentHandler struct {
	repo *repository.DepartmentRepository
}

func NewDepartmentHandler(repo *repository.DepartmentRepository) *DepartmentHandler {
	return &DepartmentHandler{repo: repo}
}

type OrgNameRequest struct {
	Name   string `json:"name" binding:"required"`
	NameBn string `json:"name_bn"`
}

// ListDepartments godoc
//
// @Summary      List departments
// @Tags         Organization
// @Security     BearerAuth
// @Produce      json
// @Param        page   query int false "Page number (default: 1)"
// @Param        limit  query int false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /departments [get]
func (h *DepartmentHandler) List(c *gin.Context) {
	p := utils.ParsePagination(c)
	list, total, err := h.repo.List(p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

func (h *DepartmentHandler) GetByID(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var req OrgNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m := &models.Department{Name: req.Name, NameBn: req.NameBn}
	if err := h.repo.Create(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req OrgNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.Name = req.Name
	m.NameBn = req.NameBn
	if err := h.repo.Update(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *DepartmentHandler) Delete(c *gin.Context) {
	if _, err := h.repo.FindByID(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := h.repo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Section ---

type SectionHandler struct {
	repo *repository.SectionRepository
}

func NewSectionHandler(repo *repository.SectionRepository) *SectionHandler {
	return &SectionHandler{repo: repo}
}

type SectionRequest struct {
	Name         string `json:"name" binding:"required"`
	NameBn       string `json:"name_bn"`
	DepartmentID string `json:"department_id" binding:"required"`
}

// ListSections godoc
//
// @Summary      List sections
// @Tags         Organization
// @Security     BearerAuth
// @Produce      json
// @Param        department_id query string false "Filter by department"
// @Param        page          query int    false "Page number (default: 1)"
// @Param        limit         query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /sections [get]
func (h *SectionHandler) List(c *gin.Context) {
	departmentID := c.Query("department_id")
	p := utils.ParsePagination(c)
	list, total, err := h.repo.List(departmentID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

func (h *SectionHandler) GetByID(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *SectionHandler) Create(c *gin.Context) {
	var req SectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m := &models.Section{Name: req.Name, NameBn: req.NameBn, DepartmentID: req.DepartmentID}
	if err := h.repo.Create(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *SectionHandler) Update(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req SectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.Name = req.Name
	m.NameBn = req.NameBn
	m.DepartmentID = req.DepartmentID
	if err := h.repo.Update(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *SectionHandler) Delete(c *gin.Context) {
	if _, err := h.repo.FindByID(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := h.repo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Designation ---

type DesignationHandler struct {
	repo *repository.DesignationRepository
}

func NewDesignationHandler(repo *repository.DesignationRepository) *DesignationHandler {
	return &DesignationHandler{repo: repo}
}

type DesignationRequest struct {
	Name      string `json:"name" binding:"required"`
	NameBn    string `json:"name_bn"`
	SectionID string `json:"section_id" binding:"required"`
}

// ListDesignations godoc
//
// @Summary      List designations
// @Tags         Organization
// @Security     BearerAuth
// @Produce      json
// @Param        section_id query string false "Filter by section"
// @Param        page       query int    false "Page number (default: 1)"
// @Param        limit      query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /designations [get]
func (h *DesignationHandler) List(c *gin.Context) {
	sectionID := c.Query("section_id")
	p := utils.ParsePagination(c)
	list, total, err := h.repo.List(sectionID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

func (h *DesignationHandler) GetByID(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *DesignationHandler) Create(c *gin.Context) {
	var req DesignationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m := &models.Designation{Name: req.Name, NameBn: req.NameBn, SectionID: req.SectionID}
	if err := h.repo.Create(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *DesignationHandler) Update(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req DesignationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.Name = req.Name
	m.NameBn = req.NameBn
	m.SectionID = req.SectionID
	if err := h.repo.Update(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *DesignationHandler) Delete(c *gin.Context) {
	if _, err := h.repo.FindByID(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := h.repo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- Line ---

type LineHandler struct {
	repo *repository.LineRepository
}

func NewLineHandler(repo *repository.LineRepository) *LineHandler {
	return &LineHandler{repo: repo}
}

type LineRequest struct {
	Name      string `json:"name" binding:"required"`
	NameBn    string `json:"name_bn"`
	SectionID string `json:"section_id" binding:"required"`
}

// ListLines godoc
//
// @Summary      List lines
// @Tags         Organization
// @Security     BearerAuth
// @Produce      json
// @Param        section_id query string false "Filter by section"
// @Param        page       query int    false "Page number (default: 1)"
// @Param        limit      query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /lines [get]
func (h *LineHandler) List(c *gin.Context) {
	sectionID := c.Query("section_id")
	p := utils.ParsePagination(c)
	list, total, err := h.repo.List(sectionID, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

func (h *LineHandler) GetByID(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *LineHandler) Create(c *gin.Context) {
	var req LineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m := &models.Line{Name: req.Name, NameBn: req.NameBn, SectionID: req.SectionID}
	if err := h.repo.Create(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *LineHandler) Update(c *gin.Context) {
	m, err := h.repo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req LineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.Name = req.Name
	m.NameBn = req.NameBn
	m.SectionID = req.SectionID
	if err := h.repo.Update(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *LineHandler) Delete(c *gin.Context) {
	if _, err := h.repo.FindByID(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := h.repo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
