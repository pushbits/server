// Package log provides functionality to configure the logger.
package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// L is the global logger instance for PushBits.
var L *log.Logger

func init() {
	L = log.New()
	L.SetOutput(os.Stderr)
	L.SetLevel(log.InfoLevel)
	L.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
}

// SetDebug sets the logger to output debug information.
func SetDebug() {
	L.SetLevel(log.DebugLevel)
}
