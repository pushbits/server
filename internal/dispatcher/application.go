package dispatcher

import (
	"fmt"
	"log"

	"github.com/pushbits/server/internal/model"

	"github.com/matrix-org/gomatrix"
)

func buildRoomTopic(id uint) string {
	return fmt.Sprintf("Application %d", id)
}

// RegisterApplication creates a channel for an application.
func (d *Dispatcher) RegisterApplication(id uint, name, token, user string) (string, error) {
	log.Printf("Registering application %s, notifications will be relayed to user %s.\n", name, user)

	response, err := d.client.CreateRoom(&gomatrix.ReqCreateRoom{
		Invite:     []string{user},
		IsDirect:   true,
		Name:       name,
		Preset:     "private_chat",
		Topic:      buildRoomTopic(id),
		Visibility: "private",
	})
	if err != nil {
		log.Print(err)
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

	// The user might have left the channel, but we can still try to remove them.
	if _, err := d.client.KickUser(a.MatrixID, kickUser); err != nil {
		log.Print(err)
	}

	if _, err := d.client.LeaveRoom(a.MatrixID); err != nil {
		log.Print(err)
		return err
	}

	if _, err := d.client.ForgetRoom(a.MatrixID); err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (d *Dispatcher) sendRoomEvent(roomID, eventType string, content interface{}) error {
	if _, err := d.client.SendStateEvent(roomID, eventType, "", content); err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// UpdateApplication updates a channel for an application.
func (d *Dispatcher) UpdateApplication(a *model.Application) error {
	log.Printf("Updating application %s (ID %d) with Matrix ID %s.\n", a.Name, a.ID, a.MatrixID)

	content := map[string]interface{}{
		"name": a.Name,
	}

	if err := d.sendRoomEvent(a.MatrixID, "m.room.name", content); err != nil {
		return err
	}

	content = map[string]interface{}{
		"topic": buildRoomTopic(a.ID),
	}

	if err := d.sendRoomEvent(a.MatrixID, "m.room.topic", content); err != nil {
		return err
	}

	return nil
}

// IsOrphan checks if the user is still connected to the channel.
func (d *Dispatcher) IsOrphan(a *model.Application, u *model.User) (bool, error) {
	resp, err := d.client.JoinedMembers(a.MatrixID)
	if err != nil {
		return false, err
	}

	found := false

	for userID := range resp.Joined {
		found = found || (userID == u.MatrixID)
	}

	return !found, nil
}

// RepairApplication re-invites the user to the channel.
func (d *Dispatcher) RepairApplication(a *model.Application, u *model.User) error {
	_, err := d.client.InviteUser(a.MatrixID, &gomatrix.ReqInviteUser{
		UserID: u.MatrixID,
	})
	if err != nil {
		return err
	}

	return nil
}
