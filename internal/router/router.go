package router

import (
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"

	"github.com/pushbits/server/internal/api"
	"github.com/pushbits/server/internal/authentication"
	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/dispatcher"
	"github.com/pushbits/server/internal/log"
)

// Create a Gin engine and setup all routes.
func Create(debug bool, cm *credentials.Manager, db *database.Database, dp *dispatcher.Dispatcher) *gin.Engine {
	log.L.Println("Setting up HTTP routes.")

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	auth := authentication.Authenticator{DB: db}

	applicationHandler := api.ApplicationHandler{DB: db, DP: dp}
	healthHandler := api.HealthHandler{DB: db}
	notificationHandler := api.NotificationHandler{DB: db, DP: dp}
	userHandler := api.UserHandler{AH: &applicationHandler, CM: cm, DB: db, DP: dp}

	r := gin.New()
	r.Use(log.GinLogger(log.L), gin.Recovery())

	r.Use(location.Default())

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
	r.DELETE("/message/:messageid", api.RequireMessageIDInURI(), auth.RequireApplicationToken(), notificationHandler.DeleteNotification)

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
