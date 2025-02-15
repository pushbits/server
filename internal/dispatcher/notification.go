package dispatcher

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/gomarkdown/markdown"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	mId "maunium.net/go/mautrix/id"

	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/internal/pberrors"
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
	RelatesTo     *RelatesTo    `json:"m.relates_to,omitempty"`
	Format        MessageFormat `json:"format"`
	NewContent    *NewContent   `json:"m.new_content,omitempty"`
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

// SendNotification sends a notification to a given user.
func (d *Dispatcher) SendNotification(a *model.Application, n *model.Notification) (eventID string, err error) {
	log.L.Printf("Sending notification to room %s.", a.MatrixID)

	plainMessage := strings.TrimSpace(n.Message)
	plainTitle := strings.TrimSpace(n.Title)
	message := d.getFormattedMessage(n)
	title := d.getFormattedTitle(n)

	text := fmt.Sprintf("%s\n\n%s", plainTitle, plainMessage)
	formattedText := fmt.Sprintf("%s %s", title, message)

	messageEvent := &MessageEvent{
		Body:          text,
		FormattedBody: formattedText,
		MsgType:       MsgTypeText,
		Format:        MessageFormatHTML,
	}

	evt, err := d.mautrixClient.SendMessageEvent(context.Background(), mId.RoomID(a.MatrixID), event.EventMessage, &messageEvent)
	if err != nil {
		log.L.Errorln(err)
		return "", err
	}

	return evt.EventID.String(), nil
}

// DeleteNotification sends a notification to a given user that another notification is deleted
func (d *Dispatcher) DeleteNotification(a *model.Application, n *model.DeleteNotification) error {
	log.L.Printf("Sending delete notification to room %s", a.MatrixID)
	var oldFormattedBody string
	var oldBody string

	// Get the message we want to delete
	deleteMessage, err := d.getMessage(a, n.ID)
	if err != nil {
		log.L.Println(err)
		return pberrors.ErrMessageNotFound
	}

	oldBody, oldFormattedBody, err = bodiesFromMessage(deleteMessage)
	if err != nil {
		return err
	}

	// Update the message with strikethrough
	newBody := fmt.Sprintf("<del>%s</del>\n- deleted", oldBody)
	newFormattedBody := fmt.Sprintf("<del>%s</del><br>- deleted", oldFormattedBody)

	_, err = d.replaceMessage(a, newBody, newFormattedBody, deleteMessage.ID.String(), oldBody, oldFormattedBody)
	if err != nil {
		return err
	}

	_, err = d.respondToMessage(a, "This message got deleted", "<i>This message got deleted.</i>", deleteMessage)

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
	message := strings.ReplaceAll(html.EscapeString(trimmedMessage), "\n", "<br />") // default to text/plain

	if optionsDisplayRaw, ok := n.Extras["client::display"]; ok {
		optionsDisplay, ok := optionsDisplayRaw.(map[string]interface{})

		if ok {
			if contentTypeRaw, ok := optionsDisplay["contentType"]; ok {
				contentType := fmt.Sprintf("%v", contentTypeRaw)
				log.L.Printf("Message content type: %s", contentType)

				switch contentType {
				case "html", "text/html":
					message = strings.ReplaceAll(trimmedMessage, "\n", "<br />")
				case "markdown", "md", "text/md", "text/markdown":
					// Allow HTML in Markdown
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
func (d *Dispatcher) getMessage(a *model.Application, id string) (*event.Event, error) {
	start := ""
	end := ""
	maxPages := 10 // Maximum pages to request (10 messages per page)

	for i := 0; i < maxPages; i++ {
		messages, err := d.mautrixClient.Messages(context.Background(), mId.RoomID(a.MatrixID), start, end, 'b', nil, 10)
		if err != nil {
			return nil, err
		}

		for _, event := range messages.Chunk {
			if event.ID.String() == id {
				return event, nil
			}
		}
		start = messages.End
	}

	return nil, pberrors.ErrMessageNotFound
}

// Replaces the content of a matrix message
func (d *Dispatcher) replaceMessage(a *model.Application, newBody, newFormattedBody string, messageID string, oldBody, oldFormattedBody string) (*mautrix.RespSendEvent, error) {
	newMessage := NewContent{
		Body:          newBody,
		FormattedBody: newFormattedBody,
		MsgType:       MsgTypeText,
		Format:        MessageFormatHTML,
	}

	replaceRelation := RelatesTo{
		RelType: "m.replace",
		EventID: messageID,
	}

	replaceEvent := MessageEvent{
		Body:          oldBody,
		FormattedBody: oldFormattedBody,
		MsgType:       MsgTypeText,
		NewContent:    &newMessage,
		RelatesTo:     &replaceRelation,
		Format:        MessageFormatHTML,
	}

	sendEvent, err := d.mautrixClient.SendMessageEvent(context.Background(), mId.RoomID(a.MatrixID), event.EventMessage, &replaceEvent)
	if err != nil {
		log.L.Errorln(err)
		return nil, err
	}

	return sendEvent, nil
}

// Sends a notification in response to another matrix message event
func (d *Dispatcher) respondToMessage(a *model.Application, body, formattedBody string, respondMessage *event.Event) (*mautrix.RespSendEvent, error) {
	oldBody, oldFormattedBody, err := bodiesFromMessage(respondMessage)
	if err != nil {
		return nil, err
	}

	// Formatting according to https://matrix.org/docs/spec/client_server/latest#fallbacks-and-event-representation
	newFormattedBody := fmt.Sprintf("<mx-reply><blockquote><a href='https://matrix.to/#/%s/%s'>In reply to</a> <a href='https://matrix.to/#/%s'>%s</a><br />%s</blockquote>\n</mx-reply>%s", respondMessage.RoomID, respondMessage.ID, respondMessage.Sender, respondMessage.Sender, oldFormattedBody, formattedBody)
	newBody := fmt.Sprintf("> <%s>%s\n\n%s", respondMessage.Sender, oldBody, body)

	notificationEvent := MessageEvent{
		FormattedBody: newFormattedBody,
		Body:          newBody,
		MsgType:       MsgTypeText,
		Format:        MessageFormatHTML,
	}

	notificationReply := make(map[string]string)
	notificationReply["event_id"] = respondMessage.ID.String()

	notificationRelation := RelatesTo{
		InReplyTo: notificationReply,
	}
	notificationEvent.RelatesTo = &notificationRelation

	sendEvent, err := d.mautrixClient.SendMessageEvent(context.Background(), mId.RoomID(a.MatrixID), event.EventMessage, &notificationEvent)
	if err != nil {
		log.L.Errorln(err)
		return nil, err
	}

	return sendEvent, nil
}

// Extracts body and formatted body from a matrix message event
func bodiesFromMessage(message *event.Event) (body, formattedBody string, err error) {
	msgContent := message.Content.AsMessage()
	if msgContent == nil {
		return "", "", pberrors.ErrMessageNotFound
	}

	formattedBody = msgContent.Body
	if msgContent.FormattedBody != "" {
		formattedBody = msgContent.FormattedBody
	}

	return msgContent.Body, formattedBody, nil
}
