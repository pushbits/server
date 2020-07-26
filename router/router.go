package router

import (
	"log"

	"github.com/eikendev/pushbits/api"
	"github.com/eikendev/pushbits/authentication"
	"github.com/eikendev/pushbits/database"
	"github.com/eikendev/pushbits/dispatcher"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

// Create a Gin engine and setup all routes.
func Create(db *database.Database, dp *dispatcher.Dispatcher) *gin.Engine {
	log.Println("Setting up HTTP routes.")

	auth := authentication.Authenticator{DB: db}

	applicationHandler := api.ApplicationHandler{DB: db, Dispatcher: dp}
	notificationHandler := api.NotificationHandler{DB: db, Dispatcher: dp}
	userHandler := api.UserHandler{DB: db}

	r := gin.Default()
	r.Use(location.Default())

	applicationGroup := r.Group("/application")
	applicationGroup.Use(auth.RequireUser())
	{
		applicationGroup.POST("", applicationHandler.CreateApplication)
		applicationGroup.DELETE("/:id", applicationHandler.DeleteApplication)
	}

	r.POST("/message", auth.RequireApplicationToken(), notificationHandler.CreateNotification)

	userGroup := r.Group("/user")
	userGroup.Use(auth.RequireAdmin())
	{
		userGroup.POST("", userHandler.CreateUser)
		//userGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	return r
}
