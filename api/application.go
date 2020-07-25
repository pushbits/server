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
	GetApplicationByToken(token string) (*model.Application, error)
}

// The ApplicationDispatcher interface for relaying notifications.
type ApplicationDispatcher interface {
	RegisterApplication(name, user string) (string, error)
}

// ApplicationHandler holds information for processing requests about applications.
type ApplicationHandler struct {
	DB         ApplicationDatabase
	Dispatcher ApplicationDispatcher
}

func (h *ApplicationHandler) applicationExists(token string) bool {
	application, _ := h.DB.GetApplicationByToken(token)
	return application != nil
}

// CreateApplication is used to create a new user.
func (h *ApplicationHandler) CreateApplication(ctx *gin.Context) {
	application := model.Application{}

	if success := successOrAbort(ctx, http.StatusBadRequest, ctx.Bind(&application)); !success {
		return
	}

	user := authentication.GetUser(ctx)

	application.Token = authentication.GenerateNotExistingToken(authentication.GenerateApplicationToken, h.applicationExists)
	application.UserID = user.ID

	log.Printf("User %s will receive notifications for application %s.\n", user.Name, application.Name)

	matrixid, err := h.Dispatcher.RegisterApplication(application.Name, user.MatrixID)

	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	application.MatrixID = matrixid

	if success := successOrAbort(ctx, http.StatusInternalServerError, h.DB.CreateApplication(&application)); !success {
		return
	}

	ctx.JSON(http.StatusOK, &application)
}
