package dispatcher

import (
	"context"
	"fmt"

	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/model"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	mId "maunium.net/go/mautrix/id"
)

func buildRoomTopic(id uint) string {
	return fmt.Sprintf("Application %d", id)
}

// RegisterApplication creates a channel for an application.
func (d *Dispatcher) RegisterApplication(id uint, name, user string) (string, error) {
	log.L.Printf("Registering application %s, notifications will be relayed to user %s.\n", name, user)

	resp, err := d.mautrixClient.CreateRoom(context.Background(), &mautrix.ReqCreateRoom{
		Visibility: "private",
		Invite:     []mId.UserID{mId.UserID(user)},
		IsDirect:   true,
		Name:       name,
		Preset:     "private_chat",
		Topic:      buildRoomTopic(id),
	})
	if err != nil {
		log.L.Print(err)
		return "", err
	}

	log.L.Printf("Application %s is now relayed to room with ID %s.\n", name, resp.RoomID.String())

	return resp.RoomID.String(), err
}

// DeregisterApplication deletes a channel for an application.
func (d *Dispatcher) DeregisterApplication(a *model.Application, u *model.User) error {
	log.L.Printf("Deregistering application %s (ID %d) with Matrix ID %s.\n", a.Name, a.ID, a.MatrixID)

	// The user might have left the channel, but we can still try to remove them.

	if _, err := d.mautrixClient.KickUser(context.Background(), mId.RoomID(a.MatrixID), &mautrix.ReqKickUser{
		Reason: "This application was deleted",
		UserID: mId.UserID(u.MatrixID),
	}); err != nil {
		log.L.Print(err)
		return err
	}

	if _, err := d.mautrixClient.LeaveRoom(context.Background(), mId.RoomID(a.MatrixID)); err != nil {
		log.L.Print(err)
		return err
	}

	if _, err := d.mautrixClient.ForgetRoom(context.Background(), mId.RoomID(a.MatrixID)); err != nil {
		log.L.Print(err)
		return err
	}

	return nil
}

func (d *Dispatcher) sendRoomEvent(roomID, eventType string, content interface{}) error {
	if _, err := d.mautrixClient.SendStateEvent(context.Background(), mId.RoomID(roomID), event.NewEventType(eventType), "", content); err != nil {
		log.L.Print(err)
		return err
	}

	return nil
}

// UpdateApplication updates a channel for an application.
func (d *Dispatcher) UpdateApplication(a *model.Application, behavior *configuration.RepairBehavior) error {
	log.L.Printf("Updating application %s (ID %d) with Matrix ID %s.\n", a.Name, a.ID, a.MatrixID)

	if behavior.ResetRoomName {
		content := map[string]interface{}{
			"name": a.Name,
		}

		if err := d.sendRoomEvent(a.MatrixID, "m.room.name", content); err != nil {
			return err
		}
	} else {
		log.L.Debugf("Not reseting room name as per configuration.\n")
	}

	if behavior.ResetRoomTopic {
		content := map[string]interface{}{
			"topic": buildRoomTopic(a.ID),
		}

		if err := d.sendRoomEvent(a.MatrixID, "m.room.topic", content); err != nil {
			return err
		}
	} else {
		log.L.Debugf("Not reseting room topic as per configuration.\n")
	}

	return nil
}

// IsOrphan checks if the user is still connected to the channel.
func (d *Dispatcher) IsOrphan(a *model.Application, u *model.User) (bool, error) {
	resp, err := d.mautrixClient.JoinedMembers(context.Background(), mId.RoomID(a.MatrixID))
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
	_, err := d.mautrixClient.InviteUser(context.Background(), mId.RoomID(a.MatrixID), &mautrix.ReqInviteUser{
		UserID: mId.UserID(u.MatrixID),
	})
	if err != nil {
		return err
	}

	return nil
}
