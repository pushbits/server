package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getID(ctx *gin.Context) (uint, error) {
	id, ok := ctx.MustGet("user").(uint)
	if !ok {
		err := errors.New("an error occured while retrieving ID from context")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return 0, err
	}

	return id, nil
}
