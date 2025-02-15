// Package main provides the main function as a starting point of this tool.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/dispatcher"
	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/router"
	"github.com/pushbits/server/internal/runner"
)

var version string

func setupCleanup(db *database.Database, dp *dispatcher.Dispatcher) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		dp.Close()
		db.Close()
		os.Exit(1)
	}()
}

func printStarupMessage() {
	if len(version) == 0 {
		log.L.Panic("Version not set")
	} else {
		log.L.Printf("Starting PushBits %s", version)
	}
}

// @title PushBits Server API Documentation
// @version 0.10.5
// @description Documentation for the PushBits server API.

// @contact.name The PushBits Developers
// @contact.url https://www.pushbits.io

// @license.name ISC
// @license.url https://github.com/pushbits/server/blob/master/LICENSE

// @BasePath /
// @query.collection.format multi
// @schemes http https

// @securityDefinitions.basic BasicAuth
func main() {
	printStarupMessage()

	c := configuration.Get()

	if c.Debug {
		log.SetDebug()
		log.L.Printf("%+v", c)
	}

	cm := credentials.CreateManager(c.Security.CheckHIBP, c.Crypto)

	db, err := database.Create(cm, c.Database.Dialect, c.Database.Connection)
	if err != nil {
		log.L.Fatal(err)
		return
	}
	if db == nil {
		log.L.Fatal("db is nil but error was nil")
		return
	}
	defer db.Close()

	if err := db.Populate(c.Admin.Name, c.Admin.Password, c.Admin.MatrixID); err != nil {
		log.L.Fatal(err)
	}

	dp, err := dispatcher.Create(c.Matrix.Homeserver, c.Matrix.Username, c.Matrix.Password, c.Formatting)
	if err != nil {
		log.L.Fatal(err)
		return
	}
	if dp == nil {
		log.L.Fatal("dp is nil but error was nil")
		return
	}
	defer dp.Close()

	setupCleanup(db, dp)

	err = db.RepairChannels(dp, &c.RepairBehavior)
	if err != nil {
		log.L.Fatal(err)
		return
	}

	engine, err := router.Create(c.Debug, c.HTTP.TrustedProxies, cm, db, dp, &c.Alertmanager)
	if err != nil {
		log.L.Fatal(err)
		return
	}

	err = runner.Run(engine, c)
	if err != nil {
		log.L.Fatal(err)
		return
	}
}
