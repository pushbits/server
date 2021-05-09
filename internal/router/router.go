package router

import (
	"log"

	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/authentication/oauth"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/dispatcher"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	ginserver "github.com/go-oauth2/gin-server"
)

// Create a Gin engine and setup all routes.
func Create(debug bool, cm *credentials.Manager, db *database.Database, dp *dispatcher.Dispatcher, authMethod string) *gin.Engine {
	log.Println("Setting up HTTP routes.")

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	auth := authentication.Authenticator{DB: db}

	applicationHandler := api.ApplicationHandler{DB: db, DP: dp}
	healthHandler := api.HealthHandler{DB: db}
	notificationHandler := api.NotificationHandler{DB: db, DP: dp}
	userHandler := api.UserHandler{AH: &applicationHandler, CM: cm, DB: db, DP: dp}

	r := gin.Default()

	r.Use(location.Default())
	// Example from the library: https://github.com/go-oauth2/oauth2/blob/master/example/server/server.go
	// Good Tutorial: https://tutorialedge.net/golang/go-oauth2-tutorial/

	if authMethod == "oauth" {
		oauth.InitializeOauth()

		oauthGroup := r.Group("/oauth2")
		{
			oauthGroup.GET("/token", ginserver.HandleTokenRequest)
			// GET TOKEN with client: curl "https://domain.tld/oauth2/token?grant_type=client_credentials&client_id=000000&client_secret=999999&scope=read" -X GET
			// GET TOKEN with password: curl "https://domain.tld/oauth2/token?grant_type=password&client_id=000000&client_secret=999999&scope=read&user_id=2&username=alex&password=123" -X GET -i
			oauthGroup.GET("/auth", ginserver.HandleAuthorizeRequest)
		}

		// TODO cubicroot remove - currently only for testing
		api := r.Group("/oauthtest")
		{
			api.Use(ginserver.HandleTokenVerify())
			api.GET("/info", func(c *gin.Context) {
				ti, exists := c.Get(ginserver.DefaultConfig.TokenKey)
				if exists {
					c.JSON(200, ti)
					return
				}
				c.String(200, "not found")
			})
		}
	} else {
		// TODO cubicroot add other auth methods here
	}

	applicationGroup := r.Group("/application")
	applicationGroup.Use(auth.RequireUser())
	{
		applicationGroup.POST("", applicationHandler.CreateApplication)
		applicationGroup.GET("", applicationHandler.GetApplications)

		applicationGroup.GET("/:id", api.RequireIDInURI(), applicationHandler.GetApplication)
		applicationGroup.DELETE("/:id", api.RequireIDInURI(), applicationHandler.DeleteApplication)
		applicationGroup.PUT("/:id", api.RequireIDInURI(), applicationHandler.UpdateApplication)
	}

	r.GET("/health", healthHandler.Health)

	r.POST("/message", auth.RequireApplicationToken(), notificationHandler.CreateNotification)

	userGroup := r.Group("/user")
	userGroup.Use(auth.RequireAdmin())
	{
		userGroup.POST("", userHandler.CreateUser)
		userGroup.GET("", userHandler.GetUsers)

		userGroup.GET("/:id", api.RequireIDInURI(), userHandler.GetUser)
		userGroup.DELETE("/:id", api.RequireIDInURI(), userHandler.DeleteUser)
		userGroup.PUT("/:id", api.RequireIDInURI(), userHandler.UpdateUser)
	}

	return r
}
