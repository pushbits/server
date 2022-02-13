package runner

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Run starts the Gin engine.
func Run(engine *gin.Engine, address string, port int) error {
	err := engine.Run(fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return err
	}

	return nil
}
