package oauth

import (
	"github.com/pushbits/server/internal/model"
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
}

// Oauth is the oauth provider for authentication
type Oauth struct {
	DB Database
}
