package api

import (
	"log"
	"net/http"
	"net/url"
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
	SendNotification(a *model.Application, n *model.Notification) (id string, err error)
	DeleteNotification(a *model.Application, n *model.DeleteNotification) error
}

// NotificationHandler holds information for processing requests about notifications.
type NotificationHandler struct {
	DB NotificationDatabase
	DP NotificationDispatcher
}

// CreateNotification godoc
// @Summary Create a Notification
// @Description Creates a new notification for the given channel
// @Accept json,mpfd
// @Produce json
// @Param message query string true "The message to send"
// @Param title query string false "The title to send"
// @Param priority query integer false "The notifications priority"
// @Param extras query model.NotificationExtras false "JSON object with additional information"
// @Param token query string true "Channels token, can also be provieded in the header"
// @Success 200 {object} model.Notification
// @Failure 500,404,403 ""
// @Router /message [post]
func (h *NotificationHandler) CreateNotification(ctx *gin.Context) {
	var notification model.Notification

	if err := ctx.Bind(&notification); err != nil {
		return
	}

	application := authentication.GetApplication(ctx)
	log.Printf("Sending notification for application %s.", application.Name)

	notification.ApplicationID = application.ID
	if strings.TrimSpace(notification.Title) == "" {
		notification.Title = application.Name
	}
	notification.Date = time.Now()

	messageID, err := h.DP.SendNotification(application, &notification)

	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	notification.ID = messageID
	notification.UrlEncodedID = url.QueryEscape(messageID)

	ctx.JSON(http.StatusOK, &notification)
}

// DeleteNotification godoc
// @Summary Delete a Notification
// @Description Informs the channel that the notification is deleted
// @Accept json,mpfd
// @Produce  json
// @Param message_id path string true "ID of the message to delete"
// @Param token query string true "Channels token, can also be provieded in the header"
// @Success 200 ""
// @Failure 500,404,403 ""
// @Router /message/{message_id} [DELETE]
func (h *NotificationHandler) DeleteNotification(ctx *gin.Context) {
	application := authentication.GetApplication(ctx)
	id, err := getMessageID(ctx)

	if success := successOrAbort(ctx, http.StatusUnprocessableEntity, err); !success {
		return
	}

	n := model.DeleteNotification{
		ID:   id,
		Date: time.Now(),
	}

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DP.DeleteNotification(application, &n)); !success {
		return
	}

	ctx.Status(http.StatusOK)
}
