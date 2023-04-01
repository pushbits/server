package database

import (
	"errors"

	"github.com/pushbits/server/internal/assert"
	"github.com/pushbits/server/internal/model"

	"gorm.io/gorm"
)

// CreateUser creates a user.
func (d *Database) CreateUser(createUser model.CreateUser) (*model.User, error) {
	user, err := createUser.IntoInternalUser(d.credentialsManager)
	if err != nil {
		return nil, err
	}

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
func (d *Database) GetUserByID(id uint) (*model.User, error) {
	var user model.User

	err := d.gormdb.First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	assert.Assert(user.ID == id)

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

// GetUsers returns all users.
func (d *Database) GetUsers() ([]model.User, error) {
	var users []model.User

	err := d.gormdb.Find(&users).Error

	return users, err
}

// AdminUserCount returns the number of admins or an error.
func (d *Database) AdminUserCount() (int64, error) {
	var users []model.User

	query := d.gormdb.Where("is_admin = ?", true).Find(&users)

	return query.RowsAffected, query.Error
}
