package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type GroupHandler struct {
	groupRepo *repository.GroupRepository
}

func NewGroupHandler(groupRepo *repository.GroupRepository) *GroupHandler {
	return &GroupHandler{groupRepo: groupRepo}
}

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

// ListGroups godoc
//
// @Summary      List groups
// @Description  Get all groups
// @Tags         Groups
// @Security     BearerAuth
// @Produce      json
// @Param        page   query int    false "Page number (default: 1)"
// @Param        limit  query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /groups [get]
func (h *GroupHandler) List(c *gin.Context) {
	p := utils.ParsePagination(c)
	groups, total, err := h.groupRepo.List(p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(groups, total, p))
}

// GetGroup godoc
//
// @Summary      Get group by ID
// @Description  Get a group by its ID
// @Tags         Groups
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Group ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /groups/{id} [get]
func (h *GroupHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	group, err := h.groupRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

// CreateGroup godoc
//
// @Summary      Create group
// @Description  Create a new group
// @Tags         Groups
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateGroupRequest true "Group details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /groups [post]
func (h *GroupHandler) Create(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := &models.Group{
		Name: req.Name,
	}

	if err := h.groupRepo.Create(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// UpdateGroup godoc
//
// @Summary      Update group
// @Description  Update an existing group
// @Tags         Groups
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Group ID"
// @Param        request body UpdateGroupRequest true "Updated group details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /groups/{id} [put]
func (h *GroupHandler) Update(c *gin.Context) {
	id := c.Param("id")
	group, err := h.groupRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group.Name = req.Name

	if err := h.groupRepo.Update(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup godoc
//
// @Summary      Delete group
// @Description  Soft delete a group
// @Tags         Groups
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Group ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /groups/{id} [delete]
func (h *GroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	_, err := h.groupRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	if err := h.groupRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}
