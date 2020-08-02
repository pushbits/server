package api

import (
	"errors"
	"net/http"

	"github.com/eikendev/pushbits/authentication"

	"github.com/gin-gonic/gin"
)

func successOrAbort(ctx *gin.Context, code int, err error) bool {
	if err != nil {
		ctx.AbortWithError(code, err)
	}

	return err == nil
}

func isCurrentUser(ctx *gin.Context, ID uint) bool {
	user := authentication.GetUser(ctx)

	if user.ID != ID {
		ctx.AbortWithError(http.StatusForbidden, errors.New("only owner can delete application"))
		return false
	}

	return true
}
