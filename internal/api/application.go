package api

import (
	"errors"
	"net/http"

	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/log"
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
	log.L.Printf("Registering application %s.", a.Name)

	channelID, err := h.DP.RegisterApplication(a.ID, a.Name, a.Token, u.MatrixID)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	a.MatrixID = channelID

	err = h.DB.UpdateApplication(a)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
}

func (h *ApplicationHandler) createApplication(ctx *gin.Context, u *model.User, name string, compat bool) (*model.Application, error) {
	log.L.Printf("Creating application %s.", name)

	application := model.Application{}
	application.Name = name
	application.Token = h.generateToken(compat)
	application.UserID = u.ID

	err := h.DB.CreateApplication(&application)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return nil, err
	}

	if err := h.registerApplication(ctx, &application, u); err != nil {
		err := h.DB.DeleteApplication(&application)
		if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
			log.L.Printf("Cannot delete application with ID %d.", application.ID)
		}

		return nil, err
	}

	return &application, nil
}

func (h *ApplicationHandler) deleteApplication(ctx *gin.Context, a *model.Application, u *model.User) error {
	log.L.Printf("Deleting application %s (ID %d).", a.Name, a.ID)

	err := h.DP.DeregisterApplication(a, u)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	err = h.DB.DeleteApplication(a)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
}

func (h *ApplicationHandler) updateApplication(ctx *gin.Context, a *model.Application, updateApplication *model.UpdateApplication) error {
	log.L.Printf("Updating application %s (ID %d).", a.Name, a.ID)

	if updateApplication.Name != nil {
		log.L.Printf("Updating application name to '%s'.", *updateApplication.Name)
		a.Name = *updateApplication.Name
	}

	if updateApplication.RefreshToken != nil && (*updateApplication.RefreshToken) {
		log.L.Print("Updating application token.")
		compat := updateApplication.StrictCompatibility != nil && (*updateApplication.StrictCompatibility)
		a.Token = h.generateToken(compat)
	}

	err := h.DB.UpdateApplication(a)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	err = h.DP.UpdateApplication(a)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return err
	}

	return nil
}

// CreateApplication godoc
// @Summary Create Application
// @Description Create a new application
// @ID post-application
// @Tags Application
// @Accept json,mpfd
// @Produce json
// @Param name query string true "Name of the application"
// @Param strict_compatability query boolean false "Use strict compatability mode"
// @Success 200 {object} model.Application
// @Failure 400 ""
// @Security BasicAuth
// @Router /application [post]
func (h *ApplicationHandler) CreateApplication(ctx *gin.Context) {
	var createApplication model.CreateApplication

	if err := ctx.Bind(&createApplication); err != nil {
		log.L.Println(err)
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

// GetApplications godoc
// @Summary Get Applications
// @Description Get all applications from current user
// @ID get-application
// @Tags Application
// @Accept json,mpfd
// @Produce json
// @Success 200 {array} model.Application
// @Failure 500 ""
// @Security BasicAuth
// @Router /application [get]
func (h *ApplicationHandler) GetApplications(ctx *gin.Context) {
	user := authentication.GetUser(ctx)
	if user == nil {
		return
	}

	applications, err := h.DB.GetApplications(user)
	if success := SuccessOrAbort(ctx, http.StatusInternalServerError, err); !success {
		return
	}

	ctx.JSON(http.StatusOK, &applications)
}

// GetApplication godoc
// @Summary Get Application
// @Description Get single application by ID
// @ID get-application-id
// @Tags Application
// @Accept json,mpfd
// @Produce json
// @Param id path int true "ID of the application"
// @Success 200 {object} model.Application
// @Failure 404,403 ""
// @Security BasicAuth
// @Router /application/{id} [get]
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

// DeleteApplication godoc
// @Summary Delete Application
// @Description Delete an application
// @ID delete-application-id
// @Tags Application
// @Accept json,mpfd
// @Produce json
// @Param id path int true "ID of the application"
// @Success 200 ""
// @Failure 500,404,403 ""
// @Security BasicAuth
// @Router /application/{id} [delete]
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

// UpdateApplication godoc
// @Summary Update Application
// @Description Update an application
// @ID put-application-id
// @Tags Application
// @Accept json,mpfd
// @Produce json
// @Param id path int true "ID of the application"
// @Param name query string false "New name for the application"
// @Param refresh_token query bool false "Generate new refresh token for the application"
// @Param strict_compatability query bool false "Whether to use strict compataibility mode"
// @Success 200 ""
// @Failure 500,404,403 ""
// @Security BasicAuth
// @Router /application/{id} [put]
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
