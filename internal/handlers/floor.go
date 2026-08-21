package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type FloorHandler struct {
	floorRepo *repository.FloorRepository
}

func NewFloorHandler(floorRepo *repository.FloorRepository) *FloorHandler {
	return &FloorHandler{floorRepo: floorRepo}
}

type CreateFloorRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateFloorRequest struct {
	Name string `json:"name" binding:"required"`
}

// ListFloors godoc
//
// @Summary      List floors
// @Description  Get all floors
// @Tags         Floors
// @Security     BearerAuth
// @Produce      json
// @Param        page   query int    false "Page number (default: 1)"
// @Param        limit  query int    false "Page size (default: 20, max: 100)"
// @Success      200  {object}  utils.PaginatedResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /floors [get]
func (h *FloorHandler) List(c *gin.Context) {
	p := utils.ParsePagination(c)
	floors, total, err := h.floorRepo.List(p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(floors, total, p))
}

// GetFloor godoc
//
// @Summary      Get floor by ID
// @Description  Get a floor by its ID
// @Tags         Floors
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Floor ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /floors/{id} [get]
func (h *FloorHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	floor, err := h.floorRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "floor not found"})
		return
	}
	c.JSON(http.StatusOK, floor)
}

// CreateFloor godoc
//
// @Summary      Create floor
// @Description  Create a new floor
// @Tags         Floors
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateFloorRequest true "Floor details"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /floors [post]
func (h *FloorHandler) Create(c *gin.Context) {
	var req CreateFloorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	floor := &models.Floor{
		Name: req.Name,
	}

	if err := h.floorRepo.Create(floor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, floor)
}

// UpdateFloor godoc
//
// @Summary      Update floor
// @Description  Update an existing floor
// @Tags         Floors
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path     string true "Floor ID"
// @Param        request body UpdateFloorRequest true "Updated floor details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /floors/{id} [put]
func (h *FloorHandler) Update(c *gin.Context) {
	id := c.Param("id")
	floor, err := h.floorRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "floor not found"})
		return
	}

	var req UpdateFloorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	floor.Name = req.Name

	if err := h.floorRepo.Update(floor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, floor)
}

// DeleteFloor godoc
//
// @Summary      Delete floor
// @Description  Soft delete a floor
// @Tags         Floors
// @Security     BearerAuth
// @Produce      json
// @Param        id   path     string true "Floor ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /floors/{id} [delete]
func (h *FloorHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	_, err := h.floorRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "floor not found"})
		return
	}

	if err := h.floorRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "floor deleted"})
}
