package api

import (
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
	UpdateApplication(application *model.Application) error
	GetApplicationByID(ID uint) (*model.Application, error)
	GetApplicationByToken(token string) (*model.Application, error)
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

// CreateApplication creates an application.
func (h *ApplicationHandler) CreateApplication(ctx *gin.Context) {
	var createApplication model.CreateApplication

	if err := ctx.Bind(&createApplication); err != nil {
		return
	}

	user := authentication.GetUser(ctx)

	application := model.Application{}
	application.Token = authentication.GenerateNotExistingToken(authentication.GenerateApplicationToken, h.applicationExists)
	application.UserID = user.ID

	log.Printf("User %s will receive notifications for application %s.\n", user.Name, application.Name)

	matrixid, err := h.DP.RegisterApplication(application.Name, user.MatrixID)

	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	application.MatrixID = matrixid

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.CreateApplication(&application)); !success {
		return
	}

	ctx.JSON(http.StatusOK, &application)
}

// DeleteApplication deletes an application with a certain ID.
func (h *ApplicationHandler) DeleteApplication(ctx *gin.Context) {
	id, err := getID(ctx)
	if err != nil {
		return
	}

	application, err := h.DB.GetApplicationByID(id)
	if success := successOrAbort(ctx, http.StatusNotFound, err); !success {
		return
	}

	if !isCurrentUser(ctx, application.UserID) {
		return
	}

	log.Printf("Deleting application %s.\n", application.Name)

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DP.DeregisterApplication(application)); !success {
		return
	}

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.DeleteApplication(application)); !success {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// UpdateApplication updates an application with a certain ID.
func (h *ApplicationHandler) UpdateApplication(ctx *gin.Context) {
	id, err := getID(ctx)
	if err != nil {
		return
	}

	application, err := h.DB.GetApplicationByID(id)
	if success := successOrAbort(ctx, http.StatusNotFound, err); !success {
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

	// TODO: Handle unbound members.
	application.Name = updateApplication.Name

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.UpdateApplication(application)); !success {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
