package oauth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pushbits/server/internal/authentication/credentials"
	"gopkg.in/oauth2.v3/server"
)

// UserAuthHandler extracts user information from the query
func UserAuthHandler() server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (string, error) {
		// TODO cubicroot check if we need a check here already
		log.Println("UserAuthorizationHandler")

		return "1", nil
	}
}

// PasswordAuthorizationHandler handles username and password based authentication
func (a *Authenticator) PasswordAuthorizationHandler() server.PasswordAuthorizationHandler {
	return func(username string, password string) (string, error) {
		log.Println("Received password based authentication request")

		user, err := a.DB.GetUserByName(username)

		if err != nil || user == nil {
			return "", nil
		}

		if !credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return "", nil
		}

		return fmt.Sprintf("%d", user.ID), nil
	}
}
