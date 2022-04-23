package api

import (
	"errors"
	"net/http"

	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/log"
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
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	for _, application := range applications {
		application := application // See https://stackoverflow.com/a/68247837

		if err := h.AH.deleteApplication(ctx, &application, u); err != nil {
			return err
		}
	}

	return nil
}

func (h *UserHandler) updateChannels(ctx *gin.Context, u *model.User, matrixID string) error {
	applications, err := h.DB.GetApplications(u)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	for _, application := range applications {
		application := application // See https://stackoverflow.com/a/68247837

		err := h.DP.DeregisterApplication(&application, u)
		if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
			return err
		}
	}

	u.MatrixID = matrixID

	for _, application := range applications {
		application := application // See https://stackoverflow.com/a/68247837

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

	log.L.Printf("Updating user %s.", u.Name)

	if updateUser.Name != nil {
		u.Name = *updateUser.Name
	}
	if updateUser.Password != nil {
		hash, err := h.CM.CreatePasswordHash(*updateUser.Password)
		if success := SuccessOrAbort(ctx, http.StatusBadRequest, err); !success {
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
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
}

// CreateUser godoc
// This method assumes that the requesting user has privileges.
// @Summary Create a User
// @Description Creates a new user
// @ID post-user
// @Tags User
// @Accept json,mpfd
// @Produce json
// @Param name query string true "Name of the user"
// @Param is_admin query bool false "Whether to set the user as admin or not"
// @Param matrix_id query string true "Matrix ID of the user in the format @user:domain.tld"
// @Param password query string true "The users password"
// @Success 200 {object} model.ExternalUser
// @Failure 500,404,403 ""
// @Security BasicAuth
// @Router /user [post]
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var createUser model.CreateUser

	if err := ctx.Bind(&createUser); err != nil {
		return
	}

	if h.userExists(createUser.Name) {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("username already exists"))
		return
	}

	log.L.Printf("Creating user %s.", createUser.Name)

	user, err := h.DB.CreateUser(createUser)

	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	ctx.JSON(http.StatusOK, user.IntoExternalUser())
}

// GetUsers godoc
// This method assumes that the requesting user has privileges.
// @Summary Get Users
// @Description Gets a list of all users
// @ID get-user
// @Tags User
// @Accept json,mpfd
// @Produce json
// @Success 200 {object} []model.ExternalUser
// @Failure 500 ""
// @Security BasicAuth
// @Router /user [get]
func (h *UserHandler) GetUsers(ctx *gin.Context) {
	users, err := h.DB.GetUsers()
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	externalUsers := make([]*model.ExternalUser, len(users))

	for i, user := range users {
		externalUsers[i] = user.IntoExternalUser()
	}

	ctx.JSON(http.StatusOK, &externalUsers)
}

// GetUser godoc
// This method assumes that the requesting user has privileges.
// @Summary Get User
// @Description Gets single user
// @ID get-user-id
// @Tags User
// @Accept json,mpfd
// @Produce json
// @Param id path integer true "The users id"
// @Success 200 {object} model.ExternalUser
// @Failure 500,404 ""
// @Security BasicAuth
// @Router /user/{id} [get]
func (h *UserHandler) GetUser(ctx *gin.Context) {
	user, err := getUser(ctx, h.DB)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, user.IntoExternalUser())
}

// DeleteUser godoc
// This method assumes that the requesting user has privileges.
// @Summary Delete User
// @Description Delete user
// @ID delete-user-id
// @Tags User
// @Accept json,mpfd
// @Produce json
// @Param id path integer true "The users id"
// @Success 200 ""
// @Failure 500,404 ""
// @Security BasicAuth
// @Router /user/{id} [delete]
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

	log.L.Printf("Deleting user %s.", user.Name)

	if err := h.deleteApplications(ctx, user); err != nil {
		return
	}

	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, h.DB.DeleteUser(user)); !success {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// UpdateUser godoc
// This method assumes that the requesting user has privileges. If users can later update their own user, make sure they
// cannot give themselves privileges.
// @Summary Update User
// @Description Update user information
// @ID put-user-id
// @Tags User
// @Accept json,mpfd
// @Produce json
// @Param id path integer true "The users id"
// @Param name query string true "Name of the user"
// @Param is_admin query bool false "Whether to set the user as admin or not"
// @Param matrix_id query string true "Matrix ID of the user in the format @user:domain.tld"
// @Param password query string true "The users password"
// @Success 200 ""
// @Failure 500,404,400 ""
// @Security BasicAuth
// @Router /user/{id} [put]
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
