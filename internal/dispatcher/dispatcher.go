package dispatcher

import (
	"log"

	"github.com/matrix-org/gomatrix"
	"github.com/pushbits/server/internal/configuration"
)

var (
	loginType = "m.login.password"
)

// Dispatcher holds information for sending notifications to clients.
type Dispatcher struct {
	client     *gomatrix.Client
	formatting configuration.Formatting
}

// Create instanciates a dispatcher connection.
func Create(homeserver, username, password string, formatting configuration.Formatting) (*Dispatcher, error) {
	log.Println("Setting up dispatcher.")

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

	return &Dispatcher{client: client, formatting: formatting}, nil
}

// Close closes the dispatcher connection.
func (d *Dispatcher) Close() {
	log.Printf("Logging out.")

	d.client.Logout()
	d.client.ClearCredentials()

	log.Printf("Successfully logged out.")
}
