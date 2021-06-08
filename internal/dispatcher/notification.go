package dispatcher

import (
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/matrix-org/gomatrix"
	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/model"
)

// MessageFormat is a matrix message format
type MessageFormat string

// MsgType is a matrix msgtype
type MsgType string

// Define matrix constants
const (
	MessageFormatHTML = MessageFormat("org.matrix.custom.html")
	MsgTypeText       = MsgType("m.text")
)

// MessageEvent is the content of a matrix message event
type MessageEvent struct {
	Body          string        `json:"body"`
	FormattedBody string        `json:"formatted_body"`
	MsgType       MsgType       `json:"msgtype"`
	RelatesTo     RelatesTo     `json:"m.relates_to,omitempty"`
	Format        MessageFormat `json:"format"`
	NewContent    NewContent    `json:"m.new_content,omitempty"`
}

// RelatesTo holds information about relations to other message events
type RelatesTo struct {
	InReplyTo map[string]string `json:"m.in_reply_to,omitempty"`
	RelType   string            `json:"rel_type,omitempty"`
	EventID   string            `json:"event_id,omitempty"`
}

// NewContent holds information about an updated message event
type NewContent struct {
	Body          string        `json:"body"`
	FormattedBody string        `json:"formatted_body"`
	MsgType       MsgType       `json:"msgtype"`
	Format        MessageFormat `json:"format"`
}

// SendNotification sends a notification to the specified user.
func (d *Dispatcher) SendNotification(a *model.Application, n *model.Notification) (id string, err error) {
	log.Printf("Sending notification to room %s.", a.MatrixID)

	plainMessage := strings.TrimSpace(n.Message)
	plainTitle := strings.TrimSpace(n.Title)
	message := d.getFormattedMessage(n)
	title := d.getFormattedTitle(n)

	text := fmt.Sprintf("%s\n\n%s", plainTitle, plainMessage)
	formattedText := fmt.Sprintf("%s %s", title, message)

	respSendEvent, err := d.client.SendFormattedText(a.MatrixID, text, formattedText)

	return respSendEvent.EventID, err
}

// DeleteNotification sends a notification to the specified user that another notificaion is deleted
func (d *Dispatcher) DeleteNotification(a *model.Application, n *model.DeleteNotification) error {
	log.Printf("Sending delete notification to room %s", a.MatrixID)
	var oldFormattedBody string
	var oldBody string

	// get the message we want to delete
	deleteMessage, err := d.getMessage(a, n.ID)

	if err != nil {
		log.Println(err)
		return api.ErrorMessageNotFound
	}

	if val, ok := deleteMessage.Content["body"]; ok {
		body, ok := val.(string)
		if ok {
			oldBody = body
			oldFormattedBody = body
		}
	} else {
		log.Println("Message to delete has wrong format")
		return api.ErrorMessageNotFound
	}

	if val, ok := deleteMessage.Content["formatted_body"]; ok {
		body, ok := val.(string)
		if ok {
			oldFormattedBody = body
		}
	}

	// update the message with strikethrough
	newBody := fmt.Sprintf("<del>%s</del>\n- deleted", oldBody)
	newFormattedBody := fmt.Sprintf("<del>%s</del><br>- deleted", oldFormattedBody)

	newMessage := NewContent{
		Body:          newBody,
		FormattedBody: newFormattedBody,
		MsgType:       MsgTypeText,
		Format:        MessageFormatHTML,
	}

	replaceRelation := RelatesTo{
		RelType: "m.replace",
		EventID: deleteMessage.ID,
	}

	replaceEvent := MessageEvent{
		Body:          oldBody,
		FormattedBody: oldFormattedBody,
		MsgType:       MsgTypeText,
		NewContent:    newMessage,
		RelatesTo:     replaceRelation,
		Format:        MessageFormatHTML,
	}

	_, err = d.client.SendMessageEvent(a.MatrixID, "m.room.message", replaceEvent)

	if err != nil {
		log.Println(err)
		return err
	}

	// send a notification about the deletion
	// formatting according to https://matrix.org/docs/spec/client_server/latest#fallbacks-and-event-representation
	notificationFormattedBody := fmt.Sprintf("<mx-reply><blockquote><a href='https://matrix.to/#/%s/%s'>In reply to</a> <a href='https://matrix.to/#/%s'>%s</a><br />%s</blockquote>\n</mx-reply><i>This message got deleted.</i>", deleteMessage.RoomID, deleteMessage.ID, deleteMessage.Sender, deleteMessage.Sender, oldFormattedBody)
	notificationBody := fmt.Sprintf("> <%s>%s\n\nThis message got deleted", deleteMessage.Sender, oldBody)

	notificationEvent := MessageEvent{
		FormattedBody: notificationFormattedBody,
		Body:          notificationBody,
		MsgType:       MsgTypeText,
		Format:        MessageFormatHTML,
	}

	notificationReply := make(map[string]string)
	notificationReply["event_id"] = deleteMessage.ID

	notificationRelation := RelatesTo{
		InReplyTo: notificationReply,
	}
	notificationEvent.RelatesTo = notificationRelation

	_, err = d.client.SendMessageEvent(a.MatrixID, "m.room.message", notificationEvent)

	return err
}

// HTML-formats the title
func (d *Dispatcher) getFormattedTitle(n *model.Notification) string {
	trimmedTitle := strings.TrimSpace(n.Title)
	title := html.EscapeString(trimmedTitle)

	if d.formatting.ColoredTitle {
		title = d.coloredText(d.priorityToColor(n.Priority), title)
	}

	return "<b>" + title + "</b><br /><br />"
}

// Converts different syntaxes to a HTML-formatted message
func (d *Dispatcher) getFormattedMessage(n *model.Notification) string {
	trimmedMessage := strings.TrimSpace(n.Message)
	message := strings.Replace(html.EscapeString(trimmedMessage), "\n", "<br />", -1) // default to text/plain

	if optionsDisplayRaw, ok := n.Extras["client::display"]; ok {
		optionsDisplay, ok := optionsDisplayRaw.(map[string]interface{})

		if ok {
			if contentTypeRaw, ok := optionsDisplay["contentType"]; ok {
				contentType := fmt.Sprintf("%v", contentTypeRaw)
				log.Printf("Message content type: %s", contentType)

				switch contentType {
				case "html", "text/html":
					message = strings.Replace(trimmedMessage, "\n", "<br />", -1)
				case "markdown", "md", "text/md", "text/markdown":
					// allow HTML in Markdown
					message = string(markdown.ToHTML([]byte(trimmedMessage), nil, nil))
				}
			}
		}
	}

	return message
}

// Maps priorities to hex colors
func (d *Dispatcher) priorityToColor(prio int) string {
	switch {
	case prio < 0:
		return "#828282"
	case prio <= 3: // info - default color
		return ""
	case prio <= 10: // low - yellow
		return "#edd711"
	case prio <= 20: // mid - orange
		return "#ed6d11"
	case prio > 20: // high - red
		return "#ed1f11"
	}

	return ""
}

// Maps a priority to a color tag
func (d *Dispatcher) coloredText(color string, text string) string {
	if color == "" {
		return text
	}

	return "<font data-mx-color='" + color + "'>" + text + "</font>"
}

// Searches in the messages list for the given id
func (d *Dispatcher) getMessage(a *model.Application, id string) (gomatrix.Event, error) {
	start := ""
	end := ""
	maxPages := 10 // maximum pages to request (10 messages per page)

	for i := 0; i < maxPages; i++ {
		messages, _ := d.client.Messages(a.MatrixID, start, end, 'b', 10)
		for _, event := range messages.Chunk {
			if event.ID == id {
				return event, nil
			}
		}
		start = messages.End
	}
	return gomatrix.Event{}, api.ErrorMessageNotFound
}
