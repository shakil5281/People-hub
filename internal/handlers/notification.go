package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type NotificationHandler struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationHandler(notifRepo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{notifRepo: notifRepo}
}

// ListNotifications godoc
//
//	@Summary		List user notifications
//	@Description	Get notifications for the current user with optional read filter
//	@Tags			Notifications
//	@Security		BearerAuth
//	@Produce		json
//	@Param			is_read	query	bool	false	"Filter by read status"
//	@Param			page	query	int		false	"Page number"
//	@Param			limit	query	int		false	"Page size"
//	@Success		200	{object}	utils.PaginatedResponse
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	p := utils.ParsePagination(c)

	var isRead *bool
	if ir := c.Query("is_read"); ir != "" {
		v := ir == "true"
		isRead = &v
	}

	list, total, err := h.notifRepo.List(userID, isRead, p.Page, p.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.NewPaginatedResponse(list, total, p))
}

// GetUnreadCount godoc
//
//	@Summary		Get unread notification count
//	@Description	Returns the number of unread notifications for the current user
//	@Tags			Notifications
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")
	count, err := h.notifRepo.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkAsRead godoc
//
//	@Summary		Mark notification as read
//	@Description	Mark a single notification as read by ID
//	@Tags			Notifications
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path	string	true	"Notification ID"
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	if err := h.notifRepo.MarkAsRead(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead godoc
//
//	@Summary		Mark all notifications as read
//	@Description	Mark all unread notifications as read for the current user
//	@Tags			Notifications
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.notifRepo.MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification godoc
//
//	@Summary		Delete a notification
//	@Description	Soft-delete a notification by ID
//	@Tags			Notifications
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path	string	true	"Notification ID"
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/notifications/{id} [delete]
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	if err := h.notifRepo.Delete(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted"})
}

// Create notification (for system use - creates notification for any user)
type CreateNotificationRequest struct {
	UserID       string  `json:"user_id" binding:"required"`
	Title        string  `json:"title" binding:"required"`
	Message      string  `json:"message" binding:"required"`
	Type         string  `json:"type"`
	Metadata     *string `json:"metadata"`
}

// CreateNotification godoc
//
//	@Summary		Create a notification
//	@Description	Create a notification for a user (system use)
//	@Tags			Notifications
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateNotificationRequest	true	"Notification details"
//	@Success		201	{object}	models.Notification
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Router			/notifications [post]
func (h *NotificationHandler) Create(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	notifType := req.Type
	if notifType == "" {
		notifType = "info"
	}
	createdBy := c.GetString("user_id")
	var metaPtr *string
	if req.Metadata != nil && *req.Metadata != "" {
		m := *req.Metadata
		if !strings.HasPrefix(m, "{") && !strings.HasPrefix(m, "[") && !strings.HasPrefix(m, `"`) {
			m = fmt.Sprintf(`"%s"`, m)
		}
		metaPtr = &m
	}
	n := &models.Notification{
		UserID:    req.UserID,
		Title:     req.Title,
		Message:   req.Message,
		Type:      notifType,
		Metadata:  metaPtr,
		CreatedBy: &createdBy,
	}
	if err := h.notifRepo.Create(n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, n)
}
