package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler holds information for processing requests about the server's health.
type HealthHandler struct {
	DB Database
}

// Health returns the health status of the server.
func (h *HealthHandler) Health(ctx *gin.Context) {
	if err := h.DB.Health(); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
