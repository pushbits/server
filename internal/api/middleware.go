package api

import (
	"github.com/gin-gonic/gin"
)

type idInURI struct {
	ID uint `uri:"id" binding:"required"`
}

type messageIdInURI struct {
	MessageID string `uri:"messageid" binding:"required"`
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

// RequireMessageIDInURI returns a Gin middleware which requires an messageID to be supplied in the URI of the request.
func RequireMessageIDInURI() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestModel messageIdInURI

		if err := ctx.BindUri(&requestModel); err != nil {
			return
		}

		ctx.Set("messageid", requestModel.MessageID)
	}
}
