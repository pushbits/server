package mockups

import (
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/model"
)

// GetEmptyDatabase returns an empty sqlite database object
func GetEmptyDatabase(confCrypto configuration.CryptoConfig) (*database.Database, error) {
	cm := credentials.CreateManager(false, confCrypto)
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

// AddUsersToDb adds the users to the database and sets their username as a password, returns list of added users
func AddUsersToDb(db *database.Database, users []*model.User) ([]*model.User, error) {
	addedUsers := make([]*model.User, 0)

	for _, user := range users {
		extUser := model.ExternalUser{
			ID:       user.ID,
			Name:     user.Name,
			IsAdmin:  user.IsAdmin,
			MatrixID: user.MatrixID,
		}
		credentials := model.UserCredentials{
			Password: user.Name,
		}
		createUser := model.CreateUser{ExternalUser: extUser, UserCredentials: credentials}

		newUser, err := db.CreateUser(createUser)
		addedUsers = append(addedUsers, newUser)
		if err != nil {
			return nil, err
		}
	}

	return addedUsers, nil
}
