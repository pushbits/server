package database

import (
	"github.com/pushbits/server/internal/model"
)

// The Dispatcher interface for constructing and destructing channels.
type Dispatcher interface {
	DeregisterApplication(a *model.Application, u *model.User) error
	UpdateApplication(a *model.Application) error
	IsOrphan(a *model.Application, u *model.User) (bool, error)
	RepairApplication(a *model.Application, u *model.User) error
}
