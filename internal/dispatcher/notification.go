package dispatcher

import (
	"errors"
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/matrix-org/gomatrix"
	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/model"
)

type MessageFormat string

const (
	FormatHTML = MessageFormat("org.matrix.custom.html")
)

type ReplyEvent struct {
	Body          string        `json:"body"`
	FormattedBody string        `json:"formatted_body"`
	Msgtype       string        `json:"msgtype"`
	RelatesTo     RelatesTo     `json:"m.relates_to,omitempty"`
	Format        MessageFormat `json:"format"`
}

type RelatesTo struct {
	InReplyTo map[string]string `json:"m.in_reply_to"`
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
	var deleteMessageFormattedBody string
	var deleteMessageBody string
	log.Printf("Sending delete notification to room %s", a.MatrixID)

	deleteMessage, err := d.getMessage(a, n.ID)

	if err != nil {
		log.Println(err)
		return api.ErrorMessageNotFound
	}

	if val, ok := deleteMessage.Content["body"]; ok {
		body, ok := val.(string)
		if ok {
			deleteMessageBody = body
			deleteMessageFormattedBody = body
		}
	} else {
		log.Println("Message to delete has wrong format")
		return api.ErrorMessageNotFound
	}

	if val, ok := deleteMessage.Content["formatted_body"]; ok {
		body, ok := val.(string)
		if ok {
			deleteMessageFormattedBody = body
		}
	}

	// formatting according to https://matrix.org/docs/spec/client_server/latest#fallbacks-and-event-representation
	formattedBody := fmt.Sprintf("<mx-reply><blockquote><a href='https://matrix.to/#/%s/%s'>In reply to</a> <a href='https://matrix.to/#/%s'>%s</a><br />%s</blockquote>\n</mx-reply><i>This message got deleted.</i>", deleteMessage.RoomID, deleteMessage.ID, deleteMessage.Sender, deleteMessage.Sender, deleteMessageFormattedBody)
	body := fmt.Sprintf("> <%s>%s\n\nThis message got deleted", deleteMessage.Sender, deleteMessageBody)

	event := ReplyEvent{
		FormattedBody: formattedBody,
		Body:          body,
		Msgtype:       "m.text",
		Format:        FormatHTML,
	}

	irt := make(map[string]string)
	rt := RelatesTo{
		InReplyTo: irt,
	}
	event.RelatesTo = rt
	irt["event_id"] = deleteMessage.ID

	_, err = d.client.SendMessageEvent(a.MatrixID, "m.room.message", event)

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

// TODO cubicroot: find a way to only request on specific event
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
	return gomatrix.Event{}, errors.New("Message not found")
}
