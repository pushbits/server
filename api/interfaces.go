package api

import (
	"github.com/eikendev/pushbits/model"
)

// The Database interface for encapsulating database access.
type Database interface {
	Health() error

	CreateApplication(application *model.Application) error
	DeleteApplication(application *model.Application) error
	GetApplicationByID(ID uint) (*model.Application, error)
	GetApplicationByToken(token string) (*model.Application, error)
	UpdateApplication(application *model.Application) error

	AdminUserCount() (int64, error)
	CreateUser(user model.CreateUser) (*model.User, error)
	DeleteUser(user *model.User) error
	GetApplications(user *model.User) ([]model.Application, error)
	GetUserByID(ID uint) (*model.User, error)
	GetUserByName(name string) (*model.User, error)
	GetUsers() ([]model.User, error)
	UpdateUser(user *model.User) error
}

// The Dispatcher interface for relaying notifications.
type Dispatcher interface {
	RegisterApplication(id uint, name, token, user string) (string, error)
	DeregisterApplication(a *model.Application, u *model.User) error
}

// The CredentialsManager interface for updating credentials.
type CredentialsManager interface {
	CreatePasswordHash(password string) []byte
}
