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

func (h *ApplicationHandler) generateToken(compat bool) string {
	return authentication.GenerateNotExistingToken(authentication.GenerateApplicationToken, compat, h.applicationExists)
}

func (h *ApplicationHandler) registerApplication(ctx *gin.Context, a *model.Application, u *model.User) error {
	log.Printf("Registering application %s.", a.Name)

	channelID, err := h.DP.RegisterApplication(a.ID, a.Name, a.Token, u.MatrixID)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	a.MatrixID = channelID
	h.DB.UpdateApplication(a)

	return nil
}

func (h *ApplicationHandler) createApplication(ctx *gin.Context, u *model.User, name string, compat bool) (*model.Application, error) {
	log.Printf("Creating application %s.", name)

	application := model.Application{}
	application.Name = name
	application.Token = h.generateToken(compat)
	application.UserID = u.ID

	err := h.DB.CreateApplication(&application)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return nil, err
	}

	if err := h.registerApplication(ctx, &application, u); err != nil {
		err := h.DB.DeleteApplication(&application)

		if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
			log.Printf("Cannot delete application with ID %d.", application.ID)
		}

		return nil, err
	}

	return &application, nil
}

func (h *ApplicationHandler) deleteApplication(ctx *gin.Context, a *model.Application, u *model.User) error {
	log.Printf("Deleting application %s (ID %d).", a.Name, a.ID)

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
	log.Printf("Updating application %s (ID %d).", a.Name, a.ID)

	if updateApplication.Name != nil {
		log.Printf("Updating application name to '%s'.", *updateApplication.Name)
		a.Name = *updateApplication.Name
	}

	if updateApplication.RefreshToken != nil && (*updateApplication.RefreshToken) {
		log.Print("Updating application token.")
		compat := updateApplication.StrictCompatibility != nil && (*updateApplication.StrictCompatibility)
		a.Token = h.generateToken(compat)
	}

	err := h.DB.UpdateApplication(a)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	err = h.DP.UpdateApplication(a)
	if success := successOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
}

// CreateApplication creates an application.
func (h *ApplicationHandler) CreateApplication(ctx *gin.Context) {
	var createApplication model.CreateApplication

	if err := ctx.Bind(&createApplication); err != nil {
		log.Println(err)
		return
	}

	user := authentication.GetUser(ctx)
	if user == nil {
		return
	}

	application, err := h.createApplication(ctx, user, createApplication.Name, createApplication.StrictCompatibility)
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
	if err := ctx.Bind(&updateApplication); err != nil {
		return
	}

	if err := h.updateApplication(ctx, application, &updateApplication); err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
