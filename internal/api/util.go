package api

import (
	"errors"
	"net/http"

	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/pberrors"

	"github.com/gin-gonic/gin"
)

func successOrAbort(ctx *gin.Context, code int, err error) bool {
	if err != nil {
		// If we know the error force error code
		switch err {
		case pberrors.ErrorMessageNotFound:
			ctx.AbortWithError(http.StatusNotFound, err)
		default:
			ctx.AbortWithError(code, err)
		}
	}

	return err == nil
}

func isCurrentUser(ctx *gin.Context, ID uint) bool {
	user := authentication.GetUser(ctx)
	if user == nil {
		return false
	}

	if user.ID != ID {
		ctx.AbortWithError(http.StatusForbidden, errors.New("only owner can delete application"))
		return false
	}

	return true
}
