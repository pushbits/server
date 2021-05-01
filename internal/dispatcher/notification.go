package dispatcher

import (
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/pushbits/server/internal/model"
)

// SendNotification sends a notification to the specified user.
func (d *Dispatcher) SendNotification(a *model.Application, n *model.Notification) error {
	log.Printf("Sending notification to room %s.", a.MatrixID)

	plainTitle := strings.TrimSpace(n.Title)
	plainMessage := strings.TrimSpace(n.Message)
	escapedTitle := html.EscapeString(plainTitle)
	message := html.EscapeString(plainMessage) // default to text/plain

	if optionsDisplayRaw, ok := n.Extras["client::display"]; ok {
		optionsDisplay, ok := optionsDisplayRaw.(map[string]interface{})
		log.Printf("%s", optionsDisplay)

		if ok {
			if contentTypeRaw, ok := optionsDisplay["contentType"]; ok {
				contentType := fmt.Sprintf("%v", contentTypeRaw)
				log.Printf("Message content type: %s", contentType)

				switch contentType {
				case "html", "text/html":
					message = plainMessage
				case "markdown", "md", "text/md", "text/markdown":
					message = string(markdown.ToHTML([]byte(plainMessage), nil, nil))
				}
			}
		}
	}

	// TODO cubicroot: add colors for priority https://spec.matrix.org/unstable/client-server-api/#mroommessage-msgtypes
	// maybe make this optional in the settings or so

	// TODO cubicroot: check if we somehow can handle \n or other methods of line breaks

	// TODO cubicroot: add docu

	text := fmt.Sprintf("%s\n\n%s", plainTitle, plainMessage)
	formattedText := fmt.Sprintf("<b>%s</b><br /><br />%s", escapedTitle, message)

	_, err := d.client.SendFormattedText(a.MatrixID, text, formattedText)

	return err
}
