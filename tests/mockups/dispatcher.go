package mockups

import (
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/dispatcher"
)

// GetMatrixDispatcher creates and returns a matrix dispatcher
func GetMatrixDispatcher(homeserver, username, password string, confCrypto configuration.CryptoConfig) (*dispatcher.Dispatcher, error) {
	db, err := GetEmptyDatabase(confCrypto)

	if err != nil {
		return nil, err
	}

	return dispatcher.Create(db, homeserver, username, password, configuration.Formatting{})
}
