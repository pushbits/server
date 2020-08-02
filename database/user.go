package database

import (
	"errors"

	"github.com/eikendev/pushbits/assert"
	"github.com/eikendev/pushbits/model"

	"gorm.io/gorm"
)

// CreateUser creates a user.
func (d *Database) CreateUser(externalUser model.ExternalUserWithCredentials) (*model.User, error) {
	user := externalUser.IntoInternalUser(d.credentialsManager)

	return user, d.gormdb.Create(user).Error
}

// DeleteUser deletes a user.
func (d *Database) DeleteUser(user *model.User) error {
	if err := d.gormdb.Where("user_id = ?", user.ID).Delete(model.Application{}).Error; err != nil {
		return err
	}

	return d.gormdb.Delete(user).Error
}

// UpdateUser updates a user.
func (d *Database) UpdateUser(user *model.User) error {
	return d.gormdb.Save(user).Error
}

// GetUserByID returns the user with the given ID or nil.
func (d *Database) GetUserByID(ID uint) (*model.User, error) {
	var user model.User

	err := d.gormdb.First(&user, ID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	assert.Assert(user.ID == ID)

	return &user, err
}

// GetUserByName returns the user with the given name or nil.
func (d *Database) GetUserByName(name string) (*model.User, error) {
	var user model.User

	err := d.gormdb.Where("name = ?", name).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	assert.Assert(user.Name == name)

	return &user, err
}

// GetApplications returns the applications associated with a given user.
func (d *Database) GetApplications(user *model.User) ([]model.Application, error) {
	var applications []model.Application

	err := d.gormdb.Model(user).Association("Applications").Find(&applications)

	return applications, err
}

// AdminUserCount returns the number of admins or an error.
func (d *Database) AdminUserCount() (int64, error) {
	var users []model.User

	query := d.gormdb.Where("is_admin = ?", true).Find(&users)

	return query.RowsAffected, query.Error
}
