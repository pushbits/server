package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/model"

	"github.com/gin-gonic/gin"
)

// The NotificationDatabase interface for encapsulating database access.
type NotificationDatabase interface {
}

// The NotificationDispatcher interface for relaying notifications.
type NotificationDispatcher interface {
	SendNotification(a *model.Application, n *model.Notification) error
}

// NotificationHandler holds information for processing requests about notifications.
type NotificationHandler struct {
	DB NotificationDatabase
	DP NotificationDispatcher
}

// CreateNotification is used to create a new notification for a user.
func (h *NotificationHandler) CreateNotification(ctx *gin.Context) {
	var notification model.Notification
	notification.Priority = 8 // set a default value

	if err := ctx.Bind(&notification); err != nil {
		return
	}

	application := authentication.GetApplication(ctx)
	log.Printf("Sending notification for application %s.", application.Name)

	notification.ID = 0
	notification.ApplicationID = application.ID
	if strings.TrimSpace(notification.Title) == "" {
		notification.Title = application.Name
	}
	notification.Date = time.Now()

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DP.SendNotification(application, &notification)); !success {
		return
	}

	ctx.JSON(http.StatusOK, &notification)
}
