package mockups

import (
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/model"
)

// GetAdminUser returns an admin user
func GetAdminUser(c *configuration.Configuration) *model.User {
	credentialsManager := credentials.CreateManager(false, c.Crypto)
	hash, _ := credentialsManager.CreatePasswordHash(c.Admin.Password)

	return &model.User{
		ID:           1,
		Name:         c.Admin.Name,
		PasswordHash: hash,
		IsAdmin:      true,
		MatrixID:     c.Admin.MatrixID,
	}
}
