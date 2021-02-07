package model

import (
	"log"

	"github.com/pushbits/server/internal/authentication/credentials"
)

// User holds information like the name, the secret, and the applications of a user.
type User struct {
	ID           uint   `gorm:"AUTO_INCREMENT;primary_key"`
	Name         string `gorm:"type:string;size:128;unique"`
	PasswordHash []byte
	IsAdmin      bool
	MatrixID     string `gorm:"type:string"`
	Applications []Application
}

// ExternalUser represents a user for external purposes.
type ExternalUser struct {
	ID       uint   `json:"id"`
	Name     string `json:"name" form:"name" query:"name" binding:"required"`
	IsAdmin  bool   `json:"is_admin" form:"is_admin" query:"is_admin"`
	MatrixID string `json:"matrix_id" form:"matrix_id" query:"matrix_id" binding:"required"`
}

// UserCredentials holds information for authenticating a user.
type UserCredentials struct {
	Password string `json:"password,omitempty" form:"password" query:"password" binding:"required"`
}

// CreateUser is used to process queries for creating users.
type CreateUser struct {
	ExternalUser
	UserCredentials
}

// NewUser creates a new user.
func NewUser(cm *credentials.Manager, name, password string, isAdmin bool, matrixID string) (*User, error) {
	log.Printf("Creating user %s.\n", name)

	passwordHash, err := cm.CreatePasswordHash(password)
	if err != nil {
		return nil, err
	}

	return &User{
		Name:         name,
		PasswordHash: passwordHash,
		IsAdmin:      isAdmin,
		MatrixID:     matrixID,
	}, nil
}

// IntoInternalUser converts a CreateUser into a User.
func (u *CreateUser) IntoInternalUser(cm *credentials.Manager) (*User, error) {
	passwordHash, err := cm.CreatePasswordHash(u.Password)
	if err != nil {
		return nil, err
	}

	return &User{
		Name:         u.Name,
		PasswordHash: passwordHash,
		IsAdmin:      u.IsAdmin,
		MatrixID:     u.MatrixID,
	}, nil
}

// IntoExternalUser converts a User into a ExternalUser.
func (u *User) IntoExternalUser() *ExternalUser {
	return &ExternalUser{
		ID:       u.ID,
		Name:     u.Name,
		IsAdmin:  u.IsAdmin,
		MatrixID: u.MatrixID,
	}
}

// UpdateUser is used to process queries for updating users.
type UpdateUser struct {
	Name     *string `form:"name" query:"name" json:"name"`
	Password *string `form:"password" query:"password" json:"password"`
	IsAdmin  *bool   `form:"is_admin" query:"is_admin" json:"is_admin"`
	MatrixID *string `form:"matrix_id" query:"matrix_id" json:"matrix_id"`
}
