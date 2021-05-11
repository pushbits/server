package api

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3"

	ginserver "github.com/go-oauth2/gin-server"
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

// RequireIDFromToken returns a Gin middleware which requires an ID to be supplied by the oauth token
func RequireIDFromToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ti, exists := ctx.Get(ginserver.DefaultConfig.TokenKey)
		ti2, ok := ti.(oauth2.TokenInfo)
		log.Println(fmt.Sprintf("USER ID: %s", ti2.GetUserID()))
		if exists && ok {
			ctx.Set("id", ti2.GetUserID)
			return
		}
	}
}
