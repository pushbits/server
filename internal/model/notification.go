// Package model contains structs used in the PushBits API and across the application.
package model

import (
	"strings"
	"time"
)

// Notification holds information like the message, the title, and the priority of a notification.
type Notification struct {
	ID            string                 `json:"id"`
	URLEncodedID  string                 `json:"id_url_encoded"`
	ApplicationID uint                   `json:"appid"`
	Message       string                 `json:"message" form:"message" query:"message" binding:"required"`
	Title         string                 `json:"title" form:"title" query:"title"`
	Priority      int                    `json:"priority" form:"priority" query:"priority"`
	Extras        map[string]interface{} `json:"extras,omitempty" form:"-" query:"-"`
	Date          time.Time              `json:"date"`
}

// Sanitize sets explicit defaults for a notification.
func (n *Notification) Sanitize(application *Application) {
	n.ID = ""
	n.URLEncodedID = ""
	n.ApplicationID = application.ID
	if strings.TrimSpace(n.Title) == "" {
		n.Title = application.Name
	}
	n.Date = time.Now()
}

// DeleteNotification holds information like the message ID of a deletion notification.
type DeleteNotification struct {
	ID   string    `json:"id" form:"id"`
	Date time.Time `json:"date"`
}

// NotificationExtras is need to document Notification.Extras in a format that the tool can read.
type NotificationExtras map[string]interface{}
