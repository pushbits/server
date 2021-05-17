package oauth

import (
	"errors"
	"fmt"
	"log"
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

	a.manager = manage.NewDefaultManager()
	a.manager.SetAuthorizeCodeExp(time.Duration(24) * time.Hour)
	// TODO cubicroot add more token configs
	a.manager.SetPasswordTokenCfg(&manage.Config{
		AccessTokenExp:    time.Duration(24) * time.Hour * 12,
		RefreshTokenExp:   time.Duration(24) * time.Hour * 24 * 30,
		IsGenerateRefresh: true,
	})
	a.manager.MapAccessGenerate(generates.NewJWTAccessGenerate([]byte("dfhgdfhfg"), jwt.SigningMethodHS256)) // TODO cubicroot get RS256 to work

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
