package runner

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Run starts the Gin engine.
func Run(engine *gin.Engine, address string, port int) {
	engine.Run(fmt.Sprintf("%s:%d", address, port))
}
