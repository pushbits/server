package model

import (
	"time"
)

// Notification holds information like the message, the title, and the priority of a notification.
type Notification struct {
	ID            string                 `json:"id"`
	UrlEncodedID  string                 `json:"id_url_encoded"`
	ApplicationID uint                   `json:"appid"`
	Message       string                 `json:"message" form:"message" query:"message" binding:"required"`
	Title         string                 `json:"title" form:"title" query:"title"`
	Priority      int                    `json:"priority" form:"priority" query:"priority"`
	Extras        map[string]interface{} `json:"extras,omitempty" form:"-" query:"-"`
	Date          time.Time              `json:"date"`
}

// DeleteNotification holds information like the message ID of a deletion notification.
type DeleteNotification struct {
	ID   string    `json:"id" form:"id"`
	Date time.Time `json:"date"`
}

// NotificationExtras is need to document Notification.Extras in a format that the tool can read.
type NotificationExtras map[string]interface{}
