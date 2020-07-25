package dispatcher

import (
	"fmt"
	"log"

	"github.com/eikendev/pushbits/model"
)

// SendNotification sends a notification to the specified user.
func (d *Dispatcher) SendNotification(a *model.Application, n *model.Notification) error {
	log.Printf("Sending notification to room %s.\n", a.MatrixID)

	text := fmt.Sprintf("%s\n\n%s", n.Title, n.Message)

	_, err := d.client.SendText(a.MatrixID, text)

	return err
}
