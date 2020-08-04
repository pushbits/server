package router

import (
	"log"

	"github.com/eikendev/pushbits/api"
	"github.com/eikendev/pushbits/authentication"
	"github.com/eikendev/pushbits/authentication/credentials"
	"github.com/eikendev/pushbits/database"
	"github.com/eikendev/pushbits/dispatcher"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

// Create a Gin engine and setup all routes.
func Create(debug bool, cm *credentials.Manager, db *database.Database, dp *dispatcher.Dispatcher) *gin.Engine {
	log.Println("Setting up HTTP routes.")

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	auth := authentication.Authenticator{DB: db}

	applicationHandler := api.ApplicationHandler{DB: db, DP: dp}
	notificationHandler := api.NotificationHandler{DB: db, DP: dp}
	userHandler := api.UserHandler{AH: &applicationHandler, CM: cm, DB: db, DP: dp}

	r := gin.Default()

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
