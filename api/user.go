package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/eikendev/pushbits/model"

	"github.com/gin-gonic/gin"
)

// The UserDatabase interface for encapsulating database access.
type UserDatabase interface {
	CreateUser(user *model.User) error
	DeleteUser(user *model.User) error
	GetUserByID(ID uint) (*model.User, error)
	GetUserByName(name string) (*model.User, error)
	GetApplications(user *model.User) ([]model.Application, error)
	AdminUserCount() (int64, error)
}

// The UserDispatcher interface for relaying notifications.
type UserDispatcher interface {
	DeregisterApplication(a *model.Application) error
}

// UserHandler holds information for processing requests about users.
type UserHandler struct {
	DB         UserDatabase
	Dispatcher ApplicationDispatcher
}

func (h *UserHandler) userExists(name string) bool {
	user, _ := h.DB.GetUserByName(name)
	return user != nil
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var externalUser model.ExternalUserWithCredentials

	if success := successOrAbort(ctx, http.StatusBadRequest, ctx.Bind(&externalUser)); !success {
		return
	}

	user := externalUser.IntoInternalUser()

	if h.userExists(user.Name) {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("username already exists"))
		return
	}

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.CreateUser(user)); !success {
		return
	}

	ctx.JSON(http.StatusOK, user.IntoExternalUser())
}

// DeleteUser deletes a user with a certain ID.
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	var deleteUser model.DeleteUser

	if success := successOrAbort(ctx, http.StatusBadRequest, ctx.BindUri(&deleteUser)); !success {
		return
	}

	user, err := h.DB.GetUserByID(deleteUser.ID)
	if success := successOrAbort(ctx, http.StatusBadRequest, err); !success {
		return
	}

	if user.IsAdmin {
		if count, err := h.DB.AdminUserCount(); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		} else if count == 1 {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("cannot delete last admin user"))
			return
		}
	}

	log.Printf("Deleting user %s.\n", user.Name)

	applications, err := h.DB.GetApplications(user)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	for _, app := range applications {
		if success := successOrAbort(ctx, http.StatusInternalServerError, h.Dispatcher.DeregisterApplication(&app)); !success {
			return
		}
	}

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.DeleteUser(user)); !success {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
