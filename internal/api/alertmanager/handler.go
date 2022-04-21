package alertmanager

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/internal/pberrors"
)

type AlertmanagerHandler struct {
	DP       api.NotificationDispatcher
	Settings AlertmanagerHandlerSettings
}

type AlertmanagerHandlerSettings struct {
	TitleAnnotation   string
	MessageAnnotation string
}

type hookMessage struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotiations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []alert           `json:"alerts"`
}

type alert struct {
	Labels       map[string]string `json:"labels"`
	Annotiations map[string]string `json:"annotiations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	Status       string            `json:"status"`
}

// CreateAlert godoc
// @Summary Create an Alert
// @Description Creates an alert that is send to the channel as a notification. This endpoint is compatible with alertmanager webhooks.
// @ID post-alert
// @Tags Alertmanager
// @Accept json
// @Produce json
// @Param token query string true "Channels token, can also be provieded in the header"
// @Param data body hookMessage true "alertmanager webhook call"
// @Success 200 {object} []model.Notification
// @Failure 500,404,403 ""
// @Router /alert [post]
func (h *AlertmanagerHandler) CreateAlert(ctx *gin.Context) {
	application := authentication.GetApplication(ctx)
	log.Printf("Sending alert notification for application %s.", application.Name)

	var hook hookMessage
	if err := ctx.Bind(&hook); err != nil {
		return
	}

	notifications := make([]model.Notification, len(hook.Alerts))
	for i, alert := range hook.Alerts {
		notification := alert.ToNotification(h.Settings.TitleAnnotation, h.Settings.MessageAnnotation)
		notification.Sanitize(application)
		messageID, err := h.DP.SendNotification(application, &notification)
		if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
			return
		}

		notification.ID = messageID
		notification.UrlEncodedID = url.QueryEscape(messageID)
		notifications[i] = notification
	}
	ctx.JSON(http.StatusOK, &notifications)
}

func (alert *alert) ToNotification(titleAnnotation, messageAnnotation string) model.Notification {
	title := strings.Builder{}
	message := strings.Builder{}

	switch alert.Status {
	case "firing":
		title.WriteString("[FIR] ")
	case "resolved":
		title.WriteString("[RES] ")
	}
	message.WriteString("STATUS: ")
	message.WriteString(alert.Status)
	message.WriteString("\n\n")

	if titleString, ok := alert.Annotiations[titleAnnotation]; ok {
		title.WriteString(titleString)
	} else if titleString, ok := alert.Labels[titleAnnotation]; ok {
		title.WriteString(titleString)
	} else {
		title.WriteString("Unknown Title")
	}

	if messageString, ok := alert.Annotiations[messageAnnotation]; ok {
		message.WriteString(messageString)
	} else if messageString, ok := alert.Labels[messageAnnotation]; ok {
		message.WriteString(messageString)
	} else {
		message.WriteString("Unknown Message")
	}

	message.WriteString("\n")

	for labelName, labelValue := range alert.Labels {
		message.WriteString("\n")
		message.WriteString(labelName)
		message.WriteString(": ")
		message.WriteString(labelValue)
	}

	return model.Notification{
		Message: message.String(),
		Title:   title.String(),
	}
}

func successOrAbort(ctx *gin.Context, code int, err error) bool {
	if err != nil {
		// If we know the error force error code
		switch err {
		case pberrors.ErrorMessageNotFound:
			ctx.AbortWithError(http.StatusNotFound, err)
		default:
			ctx.AbortWithError(code, err)
		}
	}

	return err == nil
}
