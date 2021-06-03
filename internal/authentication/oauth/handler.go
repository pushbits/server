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
	if !exists {
		err := errors.New("Token not found")
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	ti, ok := data.(oauth2.TokenInfo)
	if !ok || !exists {
		err := errors.New("Token not found")
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	tdi := TokenDisplayInfo{}
	tdi.ReadFromTi(ti)

	c.JSON(200, tdi)
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
		log.Println("Error when revoking: ", err)
		c.AbortWithError(http.StatusNotFound, errors.New("Unknown access token"))
		return
	}

	c.JSON(200, request)
}
