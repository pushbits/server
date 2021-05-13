package oauth

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

// PasswordAuthorizationHandler returns a PasswordAuthorizationHandler that handles username and password based authentication for access tokens
func (a *Oauth) PasswordAuthorizationHandler() server.PasswordAuthorizationHandler {
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

// ClientScopeHandler returns a ClientScopeHandler that allows or disallows scopes for access tokens
func ClientScopeHandler() server.ClientScopeHandler {
	return func(clientID, scope string) (allowed bool, err error) {
		if scope == "all" || scope == "" { // For now only allow generic scopes so there is place for future expansion
			return true, nil
		}

		return false, nil
	}
}

// AccessTokenExpHandler returns an AccessTokenExpHandler that sets the expiration time of access tokens
func AccessTokenExpHandler() server.AccessTokenExpHandler {
	return func(w http.ResponseWriter, r *http.Request) (exp time.Duration, err error) {
		tokenTypeRaw, ok := r.URL.Query()["token_type"]

		if ok && len(tokenTypeRaw[0]) > 0 {
			tokenType := tokenTypeRaw[0]

			switch tokenType {
			case "longterm", "long":
				return time.Duration(24*365*2) * time.Hour, nil
			}
		}

		return time.Duration(24) * time.Hour, nil // TODO cubicroot -> that is not displayed correctly?
	}
}
