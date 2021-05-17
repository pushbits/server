package basicauth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/model"
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
}

// Handler is the basic auth provider for authentication
type Handler struct {
	DB Database
}

func (h *Handler) AuthenticationValidator() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user *model.User
		err := errors.New("No authentication method")

		user, err = h.userFromBasicAuth(ctx)

		if err != nil {
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		if user == nil {
			ctx.AbortWithError(http.StatusForbidden, errors.New("authentication failed"))
			return
		}

		ctx.Set("user", user)
	}
}

func (h *Handler) userFromBasicAuth(ctx *gin.Context) (*model.User, error) {
	if name, password, ok := ctx.Request.BasicAuth(); ok {
		if user, err := h.DB.GetUserByName(name); err != nil {
			return nil, err
		} else if user != nil && credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return user, nil
		} else {
			return nil, errors.New("credentials were invalid")
		}
	}

	return nil, errors.New("no credentials were supplied")
}
