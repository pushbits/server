package dispatcher

import (
	"log"

	"github.com/matrix-org/gomatrix"
	"github.com/pushbits/server/internal/configuration"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

var (
	loginType = "m.login.password"
)

// Dispatcher holds information for sending notifications to clients.
type Dispatcher struct {
	client        *gomatrix.Client // TODO get rid of this client as a dependency
	mautrixClient *mautrix.Client
	formatting    configuration.Formatting
}

// Create instanciates a dispatcher connection.
func Create(homeserver, username, password string, formatting configuration.Formatting) (*Dispatcher, error) {
	log.Println("Setting up dispatcher.")

	// TODO remove from here
	client, err := gomatrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	response, err := client.Login(&gomatrix.ReqLogin{
		Type:     loginType,
		User:     username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	client.SetCredentials(response.UserID, response.AccessToken)
	// To here

	matrixClient, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	_, err = matrixClient.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: username},
		Password:         password,
		DeviceID:         id.DeviceID("my-device"), // TODO make device ID configurable
		StoreCredentials: true,
	})
	if err != nil {
		return nil, err
	}

	return &Dispatcher{client: client, formatting: formatting, mautrixClient: matrixClient}, nil
}

// Close closes the dispatcher connection.
func (d *Dispatcher) Close() {
	log.Printf("Logging out.")

	d.client.Logout()
	d.client.ClearCredentials()

	log.Printf("Successfully logged out.")
}
