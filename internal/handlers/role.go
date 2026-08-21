package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type RoleHandler struct {
	repo *repository.RoleRepository
}

func NewRoleHandler(repo *repository.RoleRepository) *RoleHandler {
	return &RoleHandler{repo: repo}
}

func (h *RoleHandler) List(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	roles, err := h.repo.ListByCompany(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	role, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	permissions, _ := h.repo.GetRolePermissions(id)
	c.JSON(http.StatusOK, gin.H{
		"role":        role,
		"permissions": permissions,
	})
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := c.GetString("company_id")
	userID := c.GetString("user_id")

	role := &models.Role{
		CompanyID:   &companyID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   &userID,
	}
	if err := h.repo.Create(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	role, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot modify system role"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	userID := c.GetString("user_id")
	role.UpdatedBy = &userID

	if err := h.repo.Update(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	role, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete system role"})
		return
	}
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}

func (h *RoleHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.repo.ListAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": permissions})
}

type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.repo.FindByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.ReplaceRolePermissions(id, req.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permissions assigned"})
}
