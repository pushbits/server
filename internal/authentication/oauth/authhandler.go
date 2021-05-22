package oauth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	ginserver "github.com/go-oauth2/gin-server"
	mysql "github.com/imrenagi/go-oauth2-mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/model"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
)

// The Database interface for encapsulating database access.
type Database interface {
	GetApplicationByToken(token string) (*model.Application, error)
	GetUserByName(name string) (*model.User, error)
	GetUserByID(id uint) (*model.User, error)
}

// AuthHandler is the oauth provider for authentication
type AuthHandler struct {
	db      Database
	manager *manage.Manager
	config  configuration.Authentication
}

// Initialize prepares the AuthHandler
func (a *AuthHandler) Initialize(db *database.Database, config configuration.Authentication) error {
	a.db = db
	a.config = config

	// The manager handles the tokens
	a.manager = manage.NewDefaultManager()
	a.manager.SetAuthorizeCodeExp(time.Duration(24) * time.Hour)
	// TODO cubicroot add more token configs
	a.manager.SetPasswordTokenCfg(&manage.Config{
		AccessTokenExp:    time.Duration(24) * time.Hour,      // 1 day
		RefreshTokenExp:   time.Duration(24) * time.Hour * 30, // 30 days
		IsGenerateRefresh: true,
	})
	a.manager.MapAccessGenerate(generates.NewJWTAccessGenerate([]byte("dfhgdfhfg"), jwt.SigningMethodHS256)) // TODO cubicroot get RS256 to work

	// Define a storage for the tokens
	if a.config.Oauth.Storage == "mysql" {
		dbOauth, err := sqlx.Connect("mysql", a.config.Oauth.Connection+"?parseTime=true") // TODO cubicroot add more options and move to settings
		if err != nil {
			log.Fatal(err)
		}

		a.manager.MustTokenStorage(mysql.NewTokenStore(dbOauth))

		clientStore, _ := mysql.NewClientStore(dbOauth, mysql.WithClientStoreTableName("oauth_clients"))
		a.manager.MapClientStorage(clientStore)

		// TODO cubicroot better only store the secret as hashed value and autogenerate?
		clientStore.Create(&models.Client{
			ID:     "000000",
			Secret: "999999",
			Domain: "http://localhost",
		})
	} else {
		// TODO cubicroot add more storage options
		return errors.New("Unknown oauth storage")
	}

	// Initialize and configure the token server
	ginserver.InitServer(a.manager)
	ginserver.SetAllowGetAccessRequest(true)
	ginserver.SetClientInfoHandler(server.ClientFormHandler)
	ginserver.SetUserAuthorizationHandler(UserAuthHandler())
	ginserver.SetPasswordAuthorizationHandler(a.passwordAuthorizationHandler())
	ginserver.SetAllowedGrantType(
		//oauth2.AuthorizationCode,
		oauth2.PasswordCredentials,
		//oauth2.ClientCredentials,
		oauth2.Refreshing,
	)
	ginserver.SetClientScopeHandler(ClientScopeHandler())
	ginserver.SetAccessTokenExpHandler(AccessTokenExpHandler())
	return nil
}

// PasswordAuthorizationHandler returns a PasswordAuthorizationHandler that handles username and password based authentication for access tokens
func (a *AuthHandler) passwordAuthorizationHandler() server.PasswordAuthorizationHandler {
	return func(username string, password string) (string, error) {
		log.Println("Received password based authentication request")

		user, err := a.db.GetUserByName(username)

		if err != nil || user == nil {
			return "", nil
		}

		if !credentials.ComparePassword(user.PasswordHash, []byte(password)) {
			return "", nil
		}

		return fmt.Sprintf("%d", user.ID), nil
	}
}

// UserAuthHandler extracts user information from the query
func UserAuthHandler() server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (string, error) {
		// TODO cubicroot check if we need a check here already
		log.Println("UserAuthorizationHandler")

		return "1", nil
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
