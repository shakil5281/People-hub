package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type RequirementHandler struct {
	repo *repository.RequirementRepository
}

func NewRequirementHandler(repo *repository.RequirementRepository) *RequirementHandler {
	return &RequirementHandler{repo: repo}
}

type CreateRequirementRequest struct {
	Position      string `json:"position"`
	DepartmentID  string `json:"department_id"`
	SectionID     string `json:"section_id"`
	DesignationID string `json:"designation_id"`
	GroupType     string `json:"group_type"`
	Vacancies     int    `json:"vacancies" binding:"required"`
	Applicants    int    `json:"applicants"`
	Status        string `json:"status"`
	Priority      string `json:"priority"`
	Description   string `json:"description"`
}

type UpdateRequirementRequest struct {
	Position      string `json:"position"`
	DepartmentID  string `json:"department_id"`
	SectionID     string `json:"section_id"`
	DesignationID string `json:"designation_id"`
	GroupType     string `json:"group_type"`
	Vacancies     int    `json:"vacancies" binding:"required"`
	Applicants    int    `json:"applicants"`
	Status        string `json:"status"`
	Priority      string `json:"priority"`
	Description   string `json:"description"`
}

// ListRequirements godoc
//
// @Summary      List requirements
// @Tags         Requirements
// @Security     BearerAuth
// @Produce      json
// @Param        department_id query string false "Filter by department"
// @Param        status        query string false "Filter by status"
// @Param        priority      query string false "Filter by priority"
// @Param        position      query string false "Filter by position"
// @Param        page          query int    false "Page number (default: 1)"
// @Param        limit         query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Router       /requirements [get]
func (h *RequirementHandler) List(c *gin.Context) {
	departmentID := c.Query("department_id")
	status := c.Query("status")
	priority := c.Query("priority")
	position := c.Query("position")

	p := utils.ParsePagination(c)
	if departmentID != "" || status != "" || priority != "" || position != "" {
		items, total, err := h.repo.ListFiltered(departmentID, status, priority, position, p.Page, p.Limit)
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

func (h *RequirementHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "requirement not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *RequirementHandler) Create(c *gin.Context) {
	var req CreateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "Open"
	}
	priority := req.Priority
	if priority == "" {
		priority = "Medium"
	}

	groupType := req.GroupType
	if groupType == "" {
		groupType = "Worker"
	}

	item := &models.Requirement{
		Position:      req.Position,
		DepartmentID:  req.DepartmentID,
		SectionID:     req.SectionID,
		DesignationID: req.DesignationID,
		GroupType:     groupType,
		Vacancies:     req.Vacancies,
		Applicants:    req.Applicants,
		Status:        status,
		Priority:      priority,
		Description:   req.Description,
	}

	if err := h.repo.Create(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, _ := h.repo.FindByID(item.ID)
	c.JSON(http.StatusCreated, result)
}

func (h *RequirementHandler) SectionSummary(c *gin.Context) {
	result, err := h.repo.ListBySection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *RequirementHandler) Update(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "requirement not found"})
		return
	}

	var req UpdateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.Position = req.Position
	item.DepartmentID = req.DepartmentID
	if req.SectionID != "" {
		item.SectionID = req.SectionID
	}
	if req.DesignationID != "" {
		item.DesignationID = req.DesignationID
	}
	if req.GroupType != "" {
		item.GroupType = req.GroupType
	}
	item.Vacancies = req.Vacancies
	item.Applicants = req.Applicants
	item.Description = req.Description
	if req.Status != "" {
		item.Status = req.Status
	}
	if req.Priority != "" {
		item.Priority = req.Priority
	}

	if err := h.repo.Update(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, _ := h.repo.FindByID(item.ID)
	c.JSON(http.StatusOK, result)
}

func (h *RequirementHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	_, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "requirement not found"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "requirement deleted"})
}
