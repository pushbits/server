package model

import "strings"

type AlertmanagerWebhook struct {
	Version           string              `json:"version"`
	GroupKey          string              `json:"groupKey"`
	Receiver          string              `json:"receiver"`
	GroupLabels       map[string]string   `json:"groupLabels"`
	CommonLabels      map[string]string   `json:"commonLabels"`
	CommonAnnotations map[string]string   `json:"commonAnnotiations"`
	ExternalURL       string              `json:"externalURL"`
	Alerts            []AlertmanagerAlert `json:"alerts"`
}

type AlertmanagerAlert struct {
	Labels       map[string]string `json:"labels"`
	Annotiations map[string]string `json:"annotiations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	Status       string            `json:"status"`
}

func (alert *AlertmanagerAlert) ToNotification(titleAnnotation, messageAnnotation string) Notification {
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

	return Notification{
		Message: message.String(),
		Title:   title.String(),
	}
}
