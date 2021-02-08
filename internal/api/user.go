package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/model"

	"github.com/gin-gonic/gin"
)

// UserHandler holds information for processing requests about users.
type UserHandler struct {
	AH *ApplicationHandler
	CM CredentialsManager
	DB Database
	DP Dispatcher
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

func (h *UserHandler) deleteApplications(ctx *gin.Context, u *model.User) error {
	applications, err := h.DB.GetApplications(u)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	for _, application := range applications {
		if err := h.AH.deleteApplication(ctx, &application, u); err != nil {
			return err
		}
	}

	return nil
}

func (h *UserHandler) updateChannels(ctx *gin.Context, u *model.User, matrixID string) error {
	applications, err := h.DB.GetApplications(u)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	for _, application := range applications {
		err := h.DP.DeregisterApplication(&application, u)
		if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
			return err
		}
	}

	u.MatrixID = matrixID

	for _, application := range applications {
		err := h.AH.registerApplication(ctx, &application, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *UserHandler) updateUser(ctx *gin.Context, u *model.User, updateUser model.UpdateUser) error {
	if updateUser.MatrixID != nil && u.MatrixID != *updateUser.MatrixID {
		if err := h.updateChannels(ctx, u, *updateUser.MatrixID); err != nil {
			return err
		}
	}

	log.Printf("Updating user %s.", u.Name)

	if updateUser.Name != nil {
		u.Name = *updateUser.Name
	}
	if updateUser.Password != nil {
		hash, err := h.CM.CreatePasswordHash(*updateUser.Password)
		if success := successOrAbort(ctx, http.StatusBadRequest, err); !success {
			return err
		}

		u.PasswordHash = hash
	}
	if updateUser.MatrixID != nil {
		u.MatrixID = *updateUser.MatrixID
	}
	if updateUser.IsAdmin != nil {
		u.IsAdmin = *updateUser.IsAdmin
	}

	err := h.DB.UpdateUser(u)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
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

	log.Printf("Creating user %s.", createUser.Name)

	user, err := h.DB.CreateUser(createUser)

	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	ctx.JSON(http.StatusOK, user.IntoExternalUser())
}

// GetUsers returns all users.
// This method assumes that the requesting user has privileges.
func (h *UserHandler) GetUsers(ctx *gin.Context) {
	users, err := h.DB.GetUsers()
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	var externalUsers []*model.ExternalUser

	for _, user := range users {
		externalUsers = append(externalUsers, user.IntoExternalUser())
	}

	ctx.JSON(http.StatusOK, &externalUsers)
}

// GetUser returns the user with the specified ID.
// This method assumes that the requesting user has privileges.
func (h *UserHandler) GetUser(ctx *gin.Context) {
	user, err := getUser(ctx, h.DB)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, user.IntoExternalUser())
}

// DeleteUser deletes a user with a certain ID.
//
// This method assumes that the requesting user has privileges.
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	user, err := getUser(ctx, h.DB)
	if err != nil {
		return
	}

	// Last privileged user must not be deleted.
	if user.IsAdmin {
		if err := h.requireMultipleAdmins(ctx); err != nil {
			return
		}
	}

	log.Printf("Deleting user %s.", user.Name)

	if err := h.deleteApplications(ctx, user); err != nil {
		return
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
	user, err := getUser(ctx, h.DB)
	if err != nil {
		return
	}

	var updateUser model.UpdateUser
	if err := ctx.Bind(&updateUser); err != nil {
		return
	}

	requestingUser := authentication.GetUser(ctx)

	// Last privileged user must not be taken privileges. Assumes that the current user has privileges.
	if user.ID == requestingUser.ID && updateUser.IsAdmin != nil && !(*updateUser.IsAdmin) {
		if err := h.requireMultipleAdmins(ctx); err != nil {
			return
		}
	}

	if err := h.updateUser(ctx, user, updateUser); err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
