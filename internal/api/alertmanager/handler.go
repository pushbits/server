// Package alertmanager provides definitions and functionality related to Alertmanager notifications.
package alertmanager

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/model"
)

// Handler holds information for processing alerts received via Alertmanager.
type Handler struct {
	DP       api.NotificationDispatcher
	Settings HandlerSettings
}

// HandlerSettings represents the settings for processing alerts received via Alertmanager.
type HandlerSettings struct {
	TitleAnnotation   string
	MessageAnnotation string
}

// CreateAlert godoc
// @Summary Create an Alert
// @Description Creates an alert that is send to the channel as a notification. This endpoint is compatible with alertmanager webhooks.
// @ID post-alert
// @Tags Alertmanager
// @Accept json
// @Produce json
// @Param token query string true "Channels token, can also be provieded in the header"
// @Param data body model.AlertmanagerWebhook true "alertmanager webhook call"
// @Success 200 {object} []model.Notification
// @Failure 500,404,403 ""
// @Router /alert [post]
func (h *Handler) CreateAlert(ctx *gin.Context) {
	application := authentication.GetApplication(ctx)
	if application == nil {
		return
	}

	log.L.Printf("Sending alert notification for application %s.", application.Name)

	var hook model.AlertmanagerWebhook
	if err := ctx.Bind(&hook); err != nil {
		return
	}

	notifications := make([]model.Notification, len(hook.Alerts))
	for i, alert := range hook.Alerts {
		notification := alert.ToNotification(h.Settings.TitleAnnotation, h.Settings.MessageAnnotation)
		notification.Sanitize(application)
		messageID, err := h.DP.SendNotification(application, &notification)
		if success := api.SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
			return
		}

		notification.ID = messageID
		notification.URLEncodedID = url.QueryEscape(messageID)
		notifications[i] = notification
	}
	ctx.JSON(http.StatusOK, &notifications)
}
