package mockups

import (
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
)

// GetEmptyDatabase returns an empty sqlite database object
func GetEmptyDatabase() (*database.Database, error) {
	cm := credentials.CreateManager(false, configuration.CryptoConfig{})
	return database.Create(cm, "sqlite3", "pushbits-test.db")
}
