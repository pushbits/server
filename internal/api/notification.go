package api

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/model"

	"github.com/gin-gonic/gin"
)

// The NotificationDatabase interface for encapsulating database access.
type NotificationDatabase interface{}

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
// @ID post-message
// @Tags Application
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
	application := authentication.GetApplication(ctx)
	log.L.Printf("Sending notification for application %s.", application.Name)

	var notification model.Notification
	if err := ctx.Bind(&notification); err != nil {
		return
	}

	notification.Sanitize(application)

	messageID, err := h.DP.SendNotification(application, &notification)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	notification.ID = messageID
	notification.UrlEncodedID = url.QueryEscape(messageID)

	ctx.JSON(http.StatusOK, &notification)
}

// DeleteNotification godoc
// @Summary Delete a Notification
// @Description Informs the channel that the notification is deleted
// @ID deöete-message-id
// @Tags Application
// @Accept json,mpfd
// @Produce json
// @Param message_id path string true "ID of the message to delete"
// @Param token query string true "Channels token, can also be provieded in the header"
// @Success 200 ""
// @Failure 500,404,403 ""
// @Router /message/{message_id} [DELETE]
func (h *NotificationHandler) DeleteNotification(ctx *gin.Context) {
	application := authentication.GetApplication(ctx)
	log.L.Printf("Deleting notification for application %s.", application.Name)

	id, err := getMessageID(ctx)
	if success := SuccessOrAbort(ctx, http.StatusUnprocessableEntity, err); !success {
		return
	}

	n := model.DeleteNotification{
		ID:   id,
		Date: time.Now(),
	}

	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, h.DP.DeleteNotification(application, &n)); !success {
		return
	}

	ctx.Status(http.StatusOK)
}
