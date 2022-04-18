package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var L *log.Logger

func init() {
	L = log.New()
	L.SetOutput(os.Stderr)
	L.SetLevel(log.InfoLevel)
	L.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
}

func SetDebug() {
	L.SetLevel(log.DebugLevel)
}
