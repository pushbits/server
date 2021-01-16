package model

import (
	"time"
)

// Notification holds information like the message, the title, and the priority of a notification.
type Notification struct {
	ID            uint                   `json:"id"`
	ApplicationID uint                   `json:"appid"`
	Message       string                 `json:"message" form:"message" query:"message" binding:"required"`
	Title         string                 `json:"title" form:"title" query:"title"`
	Priority      int                    `json:"priority" form:"priority" query:"priority"`
	Extras        map[string]interface{} `json:"extras,omitempty" form:"-" query:"-"`
	Date          time.Time              `json:"date"`
}
