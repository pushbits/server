package oauth

import (
	ginserver "github.com/go-oauth2/gin-server"
	mysql "github.com/imrenagi/go-oauth2-mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pushbits/server/internal/database"

	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"

	"log"
)

// InitializeOauth sets up the basics for oauth authentication
func InitializeOauth(db *database.Database) error {
	// Initialize the database
	dbOauth, err := sqlx.Connect("mysql", "?parseTime=true") // TODO cubicroot add more options and move to settings
	if err != nil {
		log.Fatal(err)
	}

	auth := Authenticator{
		DB: db,
	}

	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(mysql.NewTokenStore(dbOauth))

	clientStore, _ := mysql.NewClientStore(dbOauth, mysql.WithClientStoreTableName("oauth_clients"))
	manager.MapClientStorage(clientStore)

	// TODO cubicroot move to settings
	clientStore.Create(&models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost",
	})

	ginserver.InitServer(manager)
	ginserver.SetAllowGetAccessRequest(true)
	ginserver.SetClientInfoHandler(server.ClientFormHandler)
	ginserver.SetUserAuthorizationHandler(UserAuthHandler())
	ginserver.SetPasswordAuthorizationHandler(auth.PasswordAuthorizationHandler())

	return nil
}
