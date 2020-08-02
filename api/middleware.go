package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type idInURI struct {
	ID uint `uri:"id" binding:"required"`
}

// RequireIDInURI returns a Gin middleware which requires an ID to be supplied in the URI of the request.
func RequireIDInURI() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestModel idInURI

		if err := ctx.BindUri(&requestModel); err != nil {
			return
		}

		ctx.Set("id", requestModel.ID)
	}
}

func getID(ctx *gin.Context) (uint, error) {
	id, ok := ctx.MustGet("user").(uint)
	if !ok {
		err := errors.New("an error occured while retrieving ID from context")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return 0, err
	}

	return id, nil
}
