// Package runner provides functions to run the web server.
package runner

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/configuration"
)

// Run starts the Gin engine.
func Run(engine *gin.Engine, c *configuration.Configuration) error {
	var err error
	address := fmt.Sprintf("%s:%d", c.HTTP.ListenAddress, c.HTTP.Port)

	if c.HTTP.CertFile != "" && c.HTTP.KeyFile != "" {
		err = engine.RunTLS(address, c.HTTP.CertFile, c.HTTP.KeyFile)
	} else {
		err = engine.Run(address)
	}

	return err
}
