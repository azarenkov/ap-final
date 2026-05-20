package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"api-gateway/internal/middleware"
	notificationv1 "github.com/azarenkov/ap2-final-gen/notification/v1"
)

type NotificationHandler struct {
	client notificationv1.NotificationServiceClient
}

func NewNotificationHandler(c notificationv1.NotificationServiceClient) *NotificationHandler {
	return &NotificationHandler{client: c}
}

func (h *NotificationHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/notifications", h.list)
	rg.GET("/notifications/:id", h.get)
	rg.POST("/notifications/:id/read", h.markRead)
}

func (h *NotificationHandler) list(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	out, err := h.client.ListUserNotifications(c, &notificationv1.ListUserNotificationsRequest{
		UserId: userID.(string),
	})
	respond(c, out, err, http.StatusOK)
}

func (h *NotificationHandler) get(c *gin.Context) {
	out, err := h.client.GetNotificationById(c, &notificationv1.GetNotificationByIdRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *NotificationHandler) markRead(c *gin.Context) {
	out, err := h.client.MarkNotificationAsRead(c, &notificationv1.MarkNotificationAsReadRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}
