package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/eikendev/pushbits/authentication"
	"github.com/eikendev/pushbits/model"

	"github.com/gin-gonic/gin"
)

// The UserDatabase interface for encapsulating database access.
type UserDatabase interface {
	CreateUser(user model.CreateUser) (*model.User, error)
	DeleteUser(user *model.User) error
	UpdateUser(user *model.User) error
	GetUserByID(ID uint) (*model.User, error)
	GetUserByName(name string) (*model.User, error)
	GetApplications(user *model.User) ([]model.Application, error)
	AdminUserCount() (int64, error)
}

// The UserDispatcher interface for relaying notifications.
type UserDispatcher interface {
	DeregisterApplication(a *model.Application) error
}

// The CredentialsManager interface for updating credentials.
type CredentialsManager interface {
	CreatePasswordHash(password string) []byte
}

// UserHandler holds information for processing requests about users.
type UserHandler struct {
	CM CredentialsManager
	DB UserDatabase
	DP UserDispatcher
}

func (h *UserHandler) userExists(name string) bool {
	user, _ := h.DB.GetUserByName(name)
	return user != nil
}

func (h *UserHandler) requireMultipleAdmins(ctx *gin.Context) error {
	if count, err := h.DB.AdminUserCount(); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return err
	} else if count == 1 {
		err := errors.New("instance needs at least one privileged user")
		ctx.AbortWithError(http.StatusBadRequest, err)
		return err
	}

	return nil
}

func (h *UserHandler) getUser(ctx *gin.Context) (*model.User, error) {
	id, err := getID(ctx)
	if err != nil {
		return nil, err
	}

	application, err := h.DB.GetUserByID(id)
	if success := successOrAbort(ctx, http.StatusNotFound, err); !success {
		return nil, err
	}

	return application, nil
}

// CreateUser creates a new user.
// This method assumes that the requesting user has privileges.
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var createUser model.CreateUser

	if err := ctx.Bind(&createUser); err != nil {
		return
	}

	if h.userExists(createUser.Name) {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("username already exists"))
		return
	}

	user, err := h.DB.CreateUser(createUser)

	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	ctx.JSON(http.StatusOK, user.IntoExternalUser())
}

// DeleteUser deletes a user with a certain ID.
//
// This method assumes that the requesting user has privileges.
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	user, err := h.getUser(ctx)
	if err != nil {
		return
	}

	// Last privileged user must not be deleted.
	if user.IsAdmin {
		if err := h.requireMultipleAdmins(ctx); err != nil {
			return
		}
	}

	log.Printf("Deleting user %s.\n", user.Name)

	applications, err := h.DB.GetApplications(user)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	for _, app := range applications {
		if success := successOrAbort(ctx, http.StatusInternalServerError, h.DP.DeregisterApplication(&app)); !success {
			return
		}
	}

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.DeleteUser(user)); !success {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// UpdateUser updates a user with a certain ID.
//
// This method assumes that the requesting user has privileges. If users can later update their own user, make sure they
// cannot give themselves privileges.
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	user, err := h.getUser(ctx)
	if err != nil {
		return
	}

	var updateUser model.UpdateUser
	if err := ctx.BindUri(&updateUser); err != nil {
		return
	}

	currentUser := authentication.GetUser(ctx)

	// Last privileged user must not be taken privileges. Assumes that the current user has privileges.
	if user.ID == currentUser.ID && !updateUser.IsAdmin {
		if err := h.requireMultipleAdmins(ctx); err != nil {
			return
		}
	}

	log.Printf("Updating user %s.\n", user.Name)

	if user.MatrixID != updateUser.MatrixID {
		// TODO: Update correspondent in rooms of applications.
	}

	// TODO: Handle unbound members.
	user.Name = updateUser.Name
	user.PasswordHash = h.CM.CreatePasswordHash(updateUser.Password)
	user.MatrixID = updateUser.MatrixID
	user.IsAdmin = updateUser.IsAdmin

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.UpdateUser(user)); !success {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
