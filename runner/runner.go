package runner

import (
	"github.com/gin-gonic/gin"
)

// Run starts the Gin engine.
func Run(engine *gin.Engine) {
	engine.Run(":8080")
}
