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
	db     Database
	client *gomatrix.Client
}

// Create instanciates a dispatcher connection.
func Create(db Database, homeserver, username, password string) (*Dispatcher, error) {
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

	return &Dispatcher{client: client}, nil
}

// Close closes the dispatcher connection.
func (d *Dispatcher) Close() {
	log.Printf("Logging out.\n")

	d.client.Logout()
	d.client.ClearCredentials()
}
