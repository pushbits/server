package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eikendev/pushbits/authentication/credentials"
	"github.com/eikendev/pushbits/configuration"
	"github.com/eikendev/pushbits/database"
	"github.com/eikendev/pushbits/dispatcher"
	"github.com/eikendev/pushbits/router"
	"github.com/eikendev/pushbits/runner"
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
		log.Printf("%+v\n", c)
	}

	cm := credentials.CreateManager(c.Crypto)

	db, err := database.Create(cm, c.Database.Dialect, c.Database.Connection)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Populate(c.Admin.Name, c.Admin.Password, c.Admin.MatrixID); err != nil {
		panic(err)
	}

	dp, err := dispatcher.Create(db, c.Matrix.Homeserver, c.Matrix.Username, c.Matrix.Password)
	if err != nil {
		panic(err)
	}
	defer dp.Close()

	setupCleanup(db, dp)

	engine := router.Create(c.Debug, db, dp)

	runner.Run(engine, c.HTTP.ListenAddress, c.HTTP.Port)
}
