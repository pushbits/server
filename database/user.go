package database

import (
	"errors"

	"github.com/eikendev/pushbits/model"

	"gorm.io/gorm"
)

// CreateUser creates a user.
func (d *Database) CreateUser(user *model.User) error {
	return d.gormdb.Create(user).Error
}

// GetUserByName returns the user by the given name or nil.
func (d *Database) GetUserByName(name string) (*model.User, error) {
	user := new(model.User)
	err := d.gormdb.Where("name = ?", name).First(user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return user, err
}
