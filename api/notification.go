package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/eikendev/pushbits/authentication"
	"github.com/eikendev/pushbits/model"

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
	DB         NotificationDatabase
	Dispatcher NotificationDispatcher
}

// CreateNotification is used to create a new notification for a user.
func (h *NotificationHandler) CreateNotification(ctx *gin.Context) {
	notification := model.Notification{}

	if success := successOrAbort(ctx, http.StatusBadRequest, ctx.Bind(&notification)); !success {
		return
	}

	application := authentication.GetApplication(ctx)
	log.Printf("Sending notification for application %s.\n", application.Name)

	notification.ID = 0
	notification.ApplicationID = application.ID
	if strings.TrimSpace(notification.Title) == "" {
		notification.Title = application.Name
	}
	notification.Date = time.Now()

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.Dispatcher.SendNotification(application, &notification)); !success {
		return
	}

	ctx.JSON(http.StatusOK, &notification)
}
