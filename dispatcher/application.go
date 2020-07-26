package dispatcher

import (
	"log"

	"github.com/matrix-org/gomatrix"
)

// RegisterApplication creates a new channel for an application.
func (d *Dispatcher) RegisterApplication(name, user string) (string, error) {
	log.Printf("Registering application %s, notifications will be relayed to user %s.\n", name, user)

	response, err := d.client.CreateRoom(&gomatrix.ReqCreateRoom{
		Visibility: "private",
		Name:       name,
		Invite:     []string{user},
		Preset:     "private_chat",
		IsDirect:   true,
	})

	if err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Printf("Application %s is now relayed to room with ID %s.\n", name, response.RoomID)

	return response.RoomID, err
}

// DeregisterApplication deletes a channel for an application.
func (d *Dispatcher) DeregisterApplication(matrixID string) error {
	log.Printf("Deregistering application with ID %s.\n", matrixID)

	_, err := d.client.LeaveRoom(matrixID)

	if err != nil {
		log.Fatal(err)
	}

	return err
}
