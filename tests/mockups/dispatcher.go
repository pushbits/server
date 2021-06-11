package mockups

import (
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/dispatcher"
)

// GetMatrixDispatcher creates and returns a matrix dispatcher
func GetMatrixDispatcher(homeserver, username, password string) (*dispatcher.Dispatcher, error) {
	db, err := GetEmptyDatabase()

	if err != nil {
		return nil, err
	}

	return dispatcher.Create(db, homeserver, username, password, configuration.Formatting{})
}
