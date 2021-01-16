package dispatcher

import (
	"fmt"
	"log"

	"github.com/pushbits/server/model"

	"github.com/matrix-org/gomatrix"
)

// RegisterApplication creates a channel for an application.
func (d *Dispatcher) RegisterApplication(id uint, name, token, user string) (string, error) {
	log.Printf("Registering application %s, notifications will be relayed to user %s.\n", name, user)

	topic := fmt.Sprintf("Application %d, Token %s", id, token)

	response, err := d.client.CreateRoom(&gomatrix.ReqCreateRoom{
		Invite:     []string{user},
		IsDirect:   true,
		Name:       name,
		Preset:     "private_chat",
		Topic:      topic,
		Visibility: "private",
	})

	if err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Printf("Application %s is now relayed to room with ID %s.\n", name, response.RoomID)

	return response.RoomID, err
}

// DeregisterApplication deletes a channel for an application.
func (d *Dispatcher) DeregisterApplication(a *model.Application, u *model.User) error {
	log.Printf("Deregistering application %s (ID %d) with Matrix ID %s.\n", a.Name, a.ID, a.MatrixID)

	kickUser := &gomatrix.ReqKickUser{
		Reason: "This application was deleted",
		UserID: u.MatrixID,
	}

	if _, err := d.client.KickUser(a.MatrixID, kickUser); err != nil {
		log.Fatal(err)
		return err
	}

	if _, err := d.client.LeaveRoom(a.MatrixID); err != nil {
		log.Fatal(err)
		return err
	}

	if _, err := d.client.ForgetRoom(a.MatrixID); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
