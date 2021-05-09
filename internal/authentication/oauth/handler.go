package oauth

import (
	"log"
	"net/http"

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
func PasswordAuthorizationHandler() server.PasswordAuthorizationHandler {
	return func(username string, password string) (string, error) {
		// TODO cubicroot get user id
		log.Println("PW Handler")
		return "5", nil
	}
}
