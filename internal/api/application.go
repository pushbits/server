package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/model"

	"github.com/gin-gonic/gin"
)

// ApplicationHandler holds information for processing requests about applications.
type ApplicationHandler struct {
	DB Database
	DP Dispatcher
}

func (h *ApplicationHandler) applicationExists(token string) bool {
	application, _ := h.DB.GetApplicationByToken(token)
	return application != nil
}

func (h *ApplicationHandler) registerApplication(ctx *gin.Context, a *model.Application, u *model.User) error {
	log.Printf("Registering application %s.\n", a.Name)

	channelID, err := h.DP.RegisterApplication(a.ID, a.Name, a.Token, u.MatrixID)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	a.MatrixID = channelID
	h.DB.UpdateApplication(a)

	return nil
}

func (h *ApplicationHandler) createApplication(ctx *gin.Context, name string, u *model.User) (*model.Application, error) {
	log.Printf("Creating application %s.\n", name)

	application := model.Application{}
	application.Name = name
	application.Token = authentication.GenerateNotExistingToken(authentication.GenerateApplicationToken, h.applicationExists)
	application.UserID = u.ID

	err := h.DB.CreateApplication(&application)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return nil, err
	}

	if err := h.registerApplication(ctx, &application, u); err != nil {
		err := h.DB.DeleteApplication(&application)

		if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
			log.Printf("Cannot delete application with ID %d.\n", application.ID)
		}

		return nil, err
	}

	return &application, nil
}

func (h *ApplicationHandler) deleteApplication(ctx *gin.Context, a *model.Application, u *model.User) error {
	log.Printf("Deleting application %s (ID %d).\n", a.Name, a.ID)

	err := h.DP.DeregisterApplication(a, u)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	err = h.DB.DeleteApplication(a)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
}

func (h *ApplicationHandler) updateApplication(ctx *gin.Context, a *model.Application, updateApplication *model.UpdateApplication) error {
	log.Printf("Updating application %s.\n", a.Name)

	if updateApplication.Name != nil {
		a.Name = *updateApplication.Name
	}

	err := h.DB.UpdateApplication(a)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
}

// CreateApplication creates an application.
func (h *ApplicationHandler) CreateApplication(ctx *gin.Context) {
	var createApplication model.CreateApplication

	if err := ctx.Bind(&createApplication); err != nil {
		return
	}

	user := authentication.GetUser(ctx)
	if user == nil {
		return
	}

	application, err := h.createApplication(ctx, createApplication.Name, user)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, &application)
}

// GetApplications returns all applications of the current user.
func (h *ApplicationHandler) GetApplications(ctx *gin.Context) {
	user := authentication.GetUser(ctx)
	if user == nil {
		return
	}

	applications, err := h.DB.GetApplications(user)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	ctx.JSON(http.StatusOK, &applications)
}

// GetApplication returns the application with the specified ID.
func (h *ApplicationHandler) GetApplication(ctx *gin.Context) {
	application, err := getApplication(ctx, h.DB)
	if err != nil {
		return
	}

	user := authentication.GetUser(ctx)
	if user == nil {
		return
	}

	if user.ID != application.UserID {
		err := errors.New("application belongs to another user")
		ctx.AbortWithError(http.StatusForbidden, err)
		return
	}

	ctx.JSON(http.StatusOK, &application)
}

// DeleteApplication deletes an application with a certain ID.
func (h *ApplicationHandler) DeleteApplication(ctx *gin.Context) {
	application, err := getApplication(ctx, h.DB)
	if err != nil {
		return
	}

	if !isCurrentUser(ctx, application.UserID) {
		return
	}

	if err := h.deleteApplication(ctx, application, authentication.GetUser(ctx)); err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// UpdateApplication updates an application with a certain ID.
func (h *ApplicationHandler) UpdateApplication(ctx *gin.Context) {
	application, err := getApplication(ctx, h.DB)
	if err != nil {
		return
	}

	if !isCurrentUser(ctx, application.UserID) {
		return
	}

	var updateApplication model.UpdateApplication
	if err := ctx.BindUri(&updateApplication); err != nil {
		return
	}

	if err := h.updateApplication(ctx, application, &updateApplication); err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
