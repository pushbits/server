package mockups

import (
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/model"
)

// GetEmptyDatabase returns an empty sqlite database object
func GetEmptyDatabase() (*database.Database, error) {
	cm := credentials.CreateManager(false, configuration.CryptoConfig{})
	return database.Create(cm, "sqlite3", "pushbits-test.db")
}

// AddApplicationsToDb inserts the applications apps into the database db
func AddApplicationsToDb(db *database.Database, apps []*model.Application) error {
	for _, app := range apps {
		err := db.CreateApplication(app)
		if err != nil {
			return err
		}
	}

	return nil
}
