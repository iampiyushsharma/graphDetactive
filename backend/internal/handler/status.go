package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novacart/graph-detective/internal/database"
)

type StatusHandler struct {
	db *database.DB
}

func NewStatusHandler(db *database.DB) *StatusHandler {
	return &StatusHandler{db: db}
}

func (h *StatusHandler) CheckStatus(c *gin.Context) {
	dbStatus := "connected"
	if h.db == nil || h.db.Driver == nil {
		dbStatus = "disconnected"
	} else {
		err := h.db.Driver.VerifyConnectivity(context.Background())
		if err != nil {
			dbStatus = "disconnected"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": dbStatus,
	})
}
