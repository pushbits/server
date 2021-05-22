package oauth

import (
	"github.com/gin-gonic/gin"
	ginserver "github.com/go-oauth2/gin-server"
	"gopkg.in/oauth2.v3"
)

// TokenInfoHandler returns a gin middleware that answers with information about a access token
func TokenInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, exists := c.Get(ginserver.DefaultConfig.TokenKey)
		ti, ok := data.(oauth2.TokenInfo)
		resp := JSONResponse{
			Status:  StatusError,
			Message: "Token not found",
		}

		if !ok || !exists {
			c.JSON(404, resp)
			return
		}

		tdi := TokenDisplayInfo{}
		tdi.ReadFromTi(ti)
		resp.Status = StatusSuccess
		resp.Message = ""
		resp.Data = tdi

		c.JSON(200, resp)
	}
}
