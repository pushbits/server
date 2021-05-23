package oauth

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ginserver "github.com/go-oauth2/gin-server"
	"gopkg.in/oauth2.v3"
)

// RevokeAccessRequest holds data required in a revoke request
type RevokeAccessRequest struct {
	Access string `json:"access_token"`
}

// GetTokenInfo answers with information about a access token
func GetTokenInfo(c *gin.Context) {
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

// RevokeAccess revokes an access token
func (a *AuthHandler) RevokeAccess(c *gin.Context) {
	var request RevokeAccessRequest

	err := c.BindJSON(&request)
	if err != nil {
		log.Println("Error when reading request.")
		c.AbortWithError(http.StatusUnprocessableEntity, errors.New("Missing access_token"))
		return
	}

	err = a.manager.RemoveAccessToken(request.Access)
	if err != nil {
		log.Println(err)
		log.Println("Error when revoking")
		c.AbortWithError(http.StatusNotFound, errors.New("Unknown access token"))
		return
	}

	c.JSON(200, request)
}
