package oauth

import (
	ginserver "github.com/go-oauth2/gin-server"
	mysql "github.com/imrenagi/go-oauth2-mysql"
	"github.com/jmoiron/sqlx"

	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"

	"log"
)

// InitializeOauth sets up the basics for oauth authentication
func InitializeOauth() error {
	// Initialize the database
	dbOauth, err := sqlx.Connect("mysql", "root:FqqVnitR8jkuZZeq8j94@tcp(db-pushbitsdev:3306)/pushbits?parseTime=true") // TODO cubicroot add more options and move to settings
	if err != nil {
		log.Fatal(err)
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
	ginserver.SetPasswordAuthorizationHandler(PasswordAuthorizationHandler())

	return nil
}
