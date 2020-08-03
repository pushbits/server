package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/eikendev/pushbits/authentication"
	"github.com/eikendev/pushbits/model"

	"github.com/gin-gonic/gin"
)

// The ApplicationDatabase interface for encapsulating database access.
type ApplicationDatabase interface {
	CreateApplication(application *model.Application) error
	DeleteApplication(application *model.Application) error
	GetApplicationByID(ID uint) (*model.Application, error)
	GetApplicationByToken(token string) (*model.Application, error)
	GetApplications(user *model.User) ([]model.Application, error)
	UpdateApplication(application *model.Application) error

	GetUserByID(ID uint) (*model.User, error)
}

// The ApplicationDispatcher interface for relaying notifications.
type ApplicationDispatcher interface {
	RegisterApplication(name, user string) (string, error)
	DeregisterApplication(a *model.Application) error
}

// ApplicationHandler holds information for processing requests about applications.
type ApplicationHandler struct {
	DB ApplicationDatabase
	DP ApplicationDispatcher
}

func (h *ApplicationHandler) applicationExists(token string) bool {
	application, _ := h.DB.GetApplicationByToken(token)
	return application != nil
}

func (h *ApplicationHandler) getApplication(ctx *gin.Context) (*model.Application, error) {
	id, err := getID(ctx)
	if err != nil {
		return nil, err
	}

	application, err := h.DB.GetApplicationByID(id)
	if success := successOrAbort(ctx, http.StatusNotFound, err); !success {
		return nil, err
	}

	return application, nil
}

func (h *ApplicationHandler) registerApplication(ctx *gin.Context, a *model.Application, u *model.User) error {
	log.Printf("Registering application %s.\n", a.Name)

	channelID, err := h.DP.RegisterApplication(a.Name, u.MatrixID)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	a.MatrixID = channelID

	return nil
}

func (h *ApplicationHandler) createApplication(ctx *gin.Context, name string, u *model.User) (*model.Application, error) {
	log.Printf("Creating application %s.\n", name)

	application := model.Application{}
	application.Name = name
	application.Token = authentication.GenerateNotExistingToken(authentication.GenerateApplicationToken, h.applicationExists)
	application.UserID = u.ID

	if err := h.registerApplication(ctx, &application, u); err != nil {
		return nil, err
	}

	err := h.DB.CreateApplication(&application)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return nil, err
	}

	return &application, nil
}

func (h *ApplicationHandler) deleteApplication(ctx *gin.Context, a *model.Application) error {
	err := h.DP.DeregisterApplication(a)
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

	application, err := h.createApplication(ctx, createApplication.Name, user)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, &application)
}

// GetApplications returns all applications for the current user.
func (h *ApplicationHandler) GetApplications(ctx *gin.Context) {
	user, err := getUser(ctx, h.DB)
	if err != nil {
		return
	}

	applications, err := h.DB.GetApplications(user)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	ctx.JSON(http.StatusOK, &applications)
}

// GetApplication returns all applications for the current user.
func (h *ApplicationHandler) GetApplication(ctx *gin.Context) {
	application, err := h.getApplication(ctx)
	if err != nil {
		return
	}

	user, err := getUser(ctx, h.DB)
	if err != nil {
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
	application, err := h.getApplication(ctx)
	if err != nil {
		return
	}

	if !isCurrentUser(ctx, application.UserID) {
		return
	}

	log.Printf("Deleting application %s.\n", application.Name)

	if err := h.deleteApplication(ctx, application); err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// UpdateApplication updates an application with a certain ID.
func (h *ApplicationHandler) UpdateApplication(ctx *gin.Context) {
	application, err := h.getApplication(ctx)
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

	log.Printf("Updating application %s.\n", application.Name)

	if err := h.updateApplication(ctx, application, &updateApplication); err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
