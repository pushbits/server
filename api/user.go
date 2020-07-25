package api

import (
	"errors"
	"net/http"

	"github.com/eikendev/pushbits/model"

	"github.com/gin-gonic/gin"
)

// The UserDatabase interface for encapsulating database access.
type UserDatabase interface {
	CreateUser(user *model.User) error
	GetUserByName(name string) (*model.User, error)
}

// UserHandler holds information for processing requests about users.
type UserHandler struct {
	DB UserDatabase
}

func (h *UserHandler) userExists(name string) bool {
	user, _ := h.DB.GetUserByName(name)
	return user != nil
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	externalUser := model.ExternalUserWithCredentials{}

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
