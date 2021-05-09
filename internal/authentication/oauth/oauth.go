package oauth

import (
	"github.com/pushbits/server/internal/model"
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
}

// Authenticator is the provider for authentication
type Authenticator struct {
	DB Database
}
