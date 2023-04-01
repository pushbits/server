// Package authentication provides definitions and functionality related to user authentication.
package authentication

import (
	"errors"
	"net/http"

	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/model"

	"github.com/gin-gonic/gin"
)

const (
	headerName = "X-Gotify-Key"
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
}

// Authenticator is the provider for authentication middleware.
type Authenticator struct {
	DB Database
}

type hasUserProperty func(user *model.User) bool

func (a *Authenticator) userFromBasicAuth(ctx *gin.Context) (*model.User, error) {
	if name, password, ok := ctx.Request.BasicAuth(); ok {
		if user, err := a.DB.GetUserByName(name); err != nil {
			return nil, err
		} else if user != nil && credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return user, nil
		} else {
			return nil, errors.New("credentials were invalid")
		}
	}

	return nil, errors.New("no credentials were supplied")
}

func (a *Authenticator) requireUserProperty(has hasUserProperty) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := a.userFromBasicAuth(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		if !has(user) {
			ctx.AbortWithError(http.StatusForbidden, errors.New("authentication failed"))
			return
		}

		ctx.Set("user", user)
	}
}

// RequireUser returns a Gin middleware which requires valid user credentials to be supplied with the request.
func (a *Authenticator) RequireUser() gin.HandlerFunc {
	return a.requireUserProperty(func(user *model.User) bool {
		return true
	})
}

// RequireAdmin returns a Gin middleware which requires valid admin credentials to be supplied with the request.
func (a *Authenticator) RequireAdmin() gin.HandlerFunc {
	return a.requireUserProperty(func(user *model.User) bool {
		return user.IsAdmin
	})
}

func (a *Authenticator) tokenFromQueryOrHeader(ctx *gin.Context) string {
	if token := a.tokenFromQuery(ctx); token != "" {
		return token
	} else if token := a.tokenFromHeader(ctx); token != "" {
		return token
	}

	return ""
}

func (a *Authenticator) tokenFromQuery(ctx *gin.Context) string {
	return ctx.Request.URL.Query().Get("token")
}

func (a *Authenticator) tokenFromHeader(ctx *gin.Context) string {
	return ctx.Request.Header.Get(headerName)
}

// RequireApplicationToken returns a Gin middleware which requires an application token to be supplied with the request.
func (a *Authenticator) RequireApplicationToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := a.tokenFromQueryOrHeader(ctx)

		app, err := a.DB.GetApplicationByToken(token)
		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		ctx.Set("app", app)
	}
}
