package authentication

import (
	"errors"
	"net/http"

	"github.com/pushbits/server/model"

	"github.com/gin-gonic/gin"
)

// GetApplication returns the application which was previously registered by the authentication middleware.
func GetApplication(ctx *gin.Context) *model.Application {
	app, ok := ctx.MustGet("app").(*model.Application)
	if app == nil || !ok {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("an error occured while retrieving application from context"))
		return nil
	}

	return app
}

// GetUser returns the user which was previously registered by the authentication middleware.
func GetUser(ctx *gin.Context) *model.User {
	user, ok := ctx.MustGet("user").(*model.User)
	if user == nil || !ok {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("an error occured while retrieving user from context"))
		return nil
	}

	return user
}
