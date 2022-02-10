package dispatcher

import (
	"fmt"
	"log"

	"github.com/pushbits/server/internal/model"
	"maunium.net/go/mautrix"

	"maunium.net/go/mautrix/event"
	mId "maunium.net/go/mautrix/id"
)

func buildRoomTopic(id uint) string {
	return fmt.Sprintf("Application %d", id)
}

// RegisterApplication creates a channel for an application.
func (d *Dispatcher) RegisterApplication(id uint, name, token, user string) (string, error) {
	log.Printf("Registering application %s, notifications will be relayed to user %s.\n", name, user)

	resp, err := d.mautrixClient.CreateRoom(&mautrix.ReqCreateRoom{
		Visibility: "private",
		Invite:     []mId.UserID{mId.UserID(user)},
		IsDirect:   true,
		Name:       name,
		Preset:     "private_chat",
		Topic:      buildRoomTopic(id),
	})
	if err != nil {
		log.Print(err)
		return "", err
	}

	log.Printf("Application %s is now relayed to room with ID %s.\n", name, resp.RoomID.String())

	return resp.RoomID.String(), err
}

// DeregisterApplication deletes a channel for an application.
func (d *Dispatcher) DeregisterApplication(a *model.Application, u *model.User) error {
	log.Printf("Deregistering application %s (ID %d) with Matrix ID %s.\n", a.Name, a.ID, a.MatrixID)

	// The user might have left the channel, but we can still try to remove them.

	if _, err := d.mautrixClient.KickUser(mId.RoomID(a.MatrixID), &mautrix.ReqKickUser{
		Reason: "This application was deleted",
		UserID: mId.UserID(a.MatrixID),
	}); err != nil {
		log.Print(err)
		return err
	}

	if _, err := d.mautrixClient.LeaveRoom(mId.RoomID(a.MatrixID)); err != nil {
		log.Print(err)
		return err
	}

	if _, err := d.mautrixClient.ForgetRoom(mId.RoomID(a.MatrixID)); err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (d *Dispatcher) sendRoomEvent(roomID, eventType string, content interface{}) error {
	if _, err := d.mautrixClient.SendStateEvent(mId.RoomID(roomID), event.NewEventType(eventType), "", content); err != nil {
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
	resp, err := d.mautrixClient.JoinedMembers(mId.RoomID(a.MatrixID))
	if err != nil {
		return false, err
	}

	found := false

	for userID := range resp.Joined {
		found = found || (userID.String() == u.MatrixID)
	}

	return !found, nil
}

// RepairApplication re-invites the user to the channel.
func (d *Dispatcher) RepairApplication(a *model.Application, u *model.User) error {
	_, err := d.mautrixClient.InviteUser(mId.RoomID(a.MatrixID), &mautrix.ReqInviteUser{
		UserID: mId.UserID(u.MatrixID),
	})
	if err != nil {
		return err
	}

	return nil
}
