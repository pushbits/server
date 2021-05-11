package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/dispatcher"
	"github.com/pushbits/server/internal/router"
	"github.com/pushbits/server/internal/runner"
)

func setupCleanup(db *database.Database, dp *dispatcher.Dispatcher) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		dp.Close()
		db.Close()
		os.Exit(1)
	}()
}

func main() {
	log.Println("Starting PushBits.")

	c := configuration.Get()

	if c.Debug {
		log.Printf("%+v", c)
	}

	cm := credentials.CreateManager(c.Security.CheckHIBP, c.Crypto)

	db, err := database.Create(cm, c.Database.Dialect, c.Database.Connection)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Populate(c.Admin.Name, c.Admin.Password, c.Admin.MatrixID); err != nil {
		log.Fatal(err)
	}

	dp, err := dispatcher.Create(db, c.Matrix.Homeserver, c.Matrix.Username, c.Matrix.Password, c.Formatting)
	if err != nil {
		log.Fatal(err)
	}
	defer dp.Close()

	setupCleanup(db, dp)

	err = db.RepairChannels(dp)
	if err != nil {
		log.Fatal(err)
	}

	engine := router.Create(c.Debug, cm, db, dp, c.Authentication)

	runner.Run(engine, c.HTTP.ListenAddress, c.HTTP.Port)
}
