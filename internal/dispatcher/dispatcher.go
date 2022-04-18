package dispatcher

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/log"
)

// Dispatcher holds information for sending notifications to clients.
type Dispatcher struct {
	mautrixClient *mautrix.Client
	formatting    configuration.Formatting
}

// Create instanciates a dispatcher connection.
func Create(homeserver, username, password string, formatting configuration.Formatting) (*Dispatcher, error) {
	log.L.Println("Setting up dispatcher.")

	matrixClient, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	_, err = matrixClient.Login(&mautrix.ReqLogin{
		Type:             mautrix.AuthTypePassword,
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: username},
		Password:         password,
		DeviceID:         id.DeviceID("PushBits"),
		StoreCredentials: true,
	})
	if err != nil {
		return nil, err
	}

	return &Dispatcher{formatting: formatting, mautrixClient: matrixClient}, nil
}

// Close closes the dispatcher connection.
func (d *Dispatcher) Close() {
	log.L.Printf("Logging out.")

	_, err := d.mautrixClient.Logout()
	if err != nil {
		log.L.Printf("Error while logging out: %s", err)
	}

	d.mautrixClient.ClearCredentials()

	log.L.Printf("Successfully logged out.")
}
