package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler holds information for processing requests about the server's health.
type HealthHandler struct {
	DB Database
}

// Health godoc
// @Summary Health of the application
// @ID get-health
// @Tags Health
// @Accept json,mpfd
// @Produce json
// @Success 200 ""
// @Failure 500 ""
// @Router /health [get]
func (h *HealthHandler) Health(ctx *gin.Context) {
	if err := h.DB.Health(); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
