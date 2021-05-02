package dispatcher

import (
	"log"

	"github.com/matrix-org/gomatrix"
)

var (
	loginType = "m.login.password"
)

// The Database interface for encapsulating database access.
type Database interface {
}

// Dispatcher holds information for sending notifications to clients.
type Dispatcher struct {
	db       Database
	client   *gomatrix.Client
	settings map[string]interface{}
}

// Create instanciates a dispatcher connection.
func Create(db Database, homeserver, username, password string, settings map[string]interface{}) (*Dispatcher, error) {
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

	return &Dispatcher{client: client, settings: settings}, nil
}

// Close closes the dispatcher connection.
func (d *Dispatcher) Close() {
	log.Printf("Logging out.")

	d.client.Logout()
	d.client.ClearCredentials()

	log.Printf("Successfully logged out.")
}
