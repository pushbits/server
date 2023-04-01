package api

import (
	"github.com/gin-gonic/gin"
)

// IDInURI is used to retrieve an ID from a context.
type IDInURI struct {
	ID uint `uri:"id" binding:"required"`
}

// messageIDInURI is used to retrieve an message ID from a context.
type messageIDInURI struct {
	MessageID string `uri:"messageid" binding:"required"`
}

// RequireIDInURI returns a Gin middleware which requires an ID to be supplied in the URI of the request.
func RequireIDInURI() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestModel IDInURI

		if err := ctx.BindUri(&requestModel); err != nil {
			return
		}

		ctx.Set("id", requestModel.ID)
	}
}

// RequireMessageIDInURI returns a Gin middleware which requires an messageID to be supplied in the URI of the request.
func RequireMessageIDInURI() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestModel messageIDInURI

		if err := ctx.BindUri(&requestModel); err != nil {
			return
		}

		ctx.Set("messageid", requestModel.MessageID)
	}
}
