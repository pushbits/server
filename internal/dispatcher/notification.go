package dispatcher

import (
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/pushbits/server/internal/model"
)

// SendNotification sends a notification to the specified user.
func (d *Dispatcher) SendNotification(a *model.Application, n *model.Notification) error {
	log.Printf("Sending notification to room %s.\n", a.MatrixID)

	plainTitle := strings.TrimSpace(n.Title)
	plainMessage := strings.TrimSpace(n.Message)
	escapedTitle := html.EscapeString(plainTitle)
	escapedMessage := html.EscapeString(plainMessage)

	text := fmt.Sprintf("%s\n\n%s", plainTitle, plainMessage)
	formattedText := fmt.Sprintf("<b>%s</b><br /><br />%s", escapedTitle, escapedMessage)

	_, err := d.client.SendFormattedText(a.MatrixID, text, formattedText)

	return err
}
