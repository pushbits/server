package oauth

import (
	ginserver "github.com/go-oauth2/gin-server"
	mysql "github.com/imrenagi/go-oauth2-mysql"
	"github.com/jmoiron/sqlx"

	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"

	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"

	"errors"
	"log"
)

// InitializeOauth sets up the basics for oauth authentication
func InitializeOauth(db *database.Database, config configuration.Authentication) error {
	// TODO cubicroot move that to the authenticator?
	manager := manage.NewDefaultManager()

	if config.Oauth.Storage == "mysql" {
		dbOauth, err := sqlx.Connect("mysql", config.Oauth.Connection+"?parseTime=true") // TODO cubicroot add more options and move to settings
		if err != nil {
			log.Fatal(err)
		}

		manager.MustTokenStorage(mysql.NewTokenStore(dbOauth))

		clientStore, _ := mysql.NewClientStore(dbOauth, mysql.WithClientStoreTableName("oauth_clients"))
		manager.MapClientStorage(clientStore)

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

	auth := Authenticator{
		DB: db,
	}

	ginserver.InitServer(manager)
	ginserver.SetAllowGetAccessRequest(true)
	ginserver.SetClientInfoHandler(server.ClientFormHandler)
	ginserver.SetUserAuthorizationHandler(UserAuthHandler())
	ginserver.SetPasswordAuthorizationHandler(auth.PasswordAuthorizationHandler())

	return nil
}
