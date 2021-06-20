package mockups

import (
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/dispatcher"
)

// GetMatrixDispatcher creates and returns a matrix dispatcher
func GetMatrixDispatcher(homeserver, username, password string, confCrypto configuration.CryptoConfig) (*dispatcher.Dispatcher, error) {
	return dispatcher.Create(homeserver, username, password, configuration.Formatting{})
}
