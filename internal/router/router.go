package router

import (
	"log"

	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/authentication/basicauth"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/authentication/oauth"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/dispatcher"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	ginserver "github.com/go-oauth2/gin-server"
)

// Create a Gin engine and setup all routes.
func Create(debug bool, cm *credentials.Manager, db *database.Database, dp *dispatcher.Dispatcher, authConfig configuration.Authentication) *gin.Engine {
	log.Println("Setting up HTTP routes.")

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize gin
	r := gin.Default()
	r.Use(location.Default())

	// Set up authentication and handler
	auth := authentication.Authenticator{
		DB:     db,
		Config: authConfig,
	}

	switch authConfig.Method {
	case "oauth":
		authHandler := oauth.AuthHandler{}
		authHandler.Initialize(db, authConfig)
		auth.SetAuthenticationValidator(authHandler.AuthenticationValidator)
		auth.SetUserSetter(authHandler.UserSetter)

		// Register oauth endpoints
		oauthGroup := r.Group("/oauth2")
		{
			oauthGroup.GET("/token", ginserver.HandleTokenRequest)
			// GET TOKEN with client: curl "https://domain.tld/oauth2/token?grant_type=client_credentials&client_id=000000&client_secret=999999&scope=read" -X GET
			// GET TOKEN with password: curl "https://domain.tld/oauth2/token?grant_type=password&client_id=000000&client_secret=999999&scope=read&username=admin&password=123" -X GET -i
			// GET TOKEN with refresh token:  curl "https://domain.tld/oauth2/token?grant_type=refresh_token&client_id=000000&client_secret=999999&refresh_token=OKLLQOOLWP2IFVFBLJVIAA" -X GET
			// GET TOKEN with code: curl "https://domain.tld/oauth2/token?grant_type=authorization_code&client_id=000000&client_secret=999999&code=4T1TJXMBPTOS4NNGILBDYW&redirect_uri=localhost" -X GET -i
			oauthGroup.GET("/auth", ginserver.HandleAuthorizeRequest) // Not very convenient for cli tools as it uses redirects
			// Use auth: curl "https://domain.tld/oauth2/authclient_id=000000&username=admin&password=21132&response_type=token" -X GET
			oauthGroup.GET("/tokeninfo", auth.RequireValidAuthentication(), oauth.GetTokenInfo)
			// curl "https://domain.tld/oauth2/revoke" -X POST -i -H "Authorization: Bearer $token" -d '{"access_token": "$revoke_token"}'
			oauthGroup.POST("/revoke", auth.RequireValidAuthentication(), auth.RequireUser(), auth.RequireAdmin(), authHandler.RevokeAccess)
		}
	default:
		authHandler := basicauth.AuthHandler{
			DB: db,
		}
		auth.SetAuthenticationValidator(authHandler.AuthenticationValidator)
		auth.SetUserSetter(authHandler.UserSetter)
	}

	applicationHandler := api.ApplicationHandler{DB: db, DP: dp}
	healthHandler := api.HealthHandler{DB: db}
	notificationHandler := api.NotificationHandler{DB: db, DP: dp}
	userHandler := api.UserHandler{AH: &applicationHandler, CM: cm, DB: db, DP: dp}

	applicationGroup := r.Group("/application")
	applicationGroup.Use(auth.RequireValidAuthentication())
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
	userGroup.Use(auth.RequireValidAuthentication())
	userGroup.Use(auth.RequireUser())
	userGroup.Use(auth.RequireAdmin()) // TODO cubicroot: stack them so they depend on the lower level ones
	{
		userGroup.POST("", userHandler.CreateUser)
		userGroup.GET("", userHandler.GetUsers)

		userGroup.GET("/:id", api.RequireIDInURI(), userHandler.GetUser)
		userGroup.DELETE("/:id", api.RequireIDInURI(), userHandler.DeleteUser)
		userGroup.PUT("/:id", api.RequireIDInURI(), userHandler.UpdateUser)
	}

	return r
}
