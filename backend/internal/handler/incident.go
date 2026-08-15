package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novacart/graph-detective/internal/database"
	"github.com/novacart/graph-detective/internal/repository"
)

type IncidentHandler struct {
	repo *repository.GraphRepository
	db   *database.DB
}

func NewIncidentHandler(repo *repository.GraphRepository, db *database.DB) *IncidentHandler {
	return &IncidentHandler{repo: repo, db: db}
}

// checkDB returns false if the database is offline and replies with a 503 Service Unavailable
func (h *IncidentHandler) checkDB(c *gin.Context) bool {
	if h.db == nil || h.db.Driver == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database service is offline",
			"code":  "DB_UNREACHABLE",
		})
		return false
	}
	err := h.db.Driver.VerifyConnectivity(context.Background())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database service is temporarily unreachable",
			"code":  "DB_UNREACHABLE",
		})
		return false
	}
	return true
}

func (h *IncidentHandler) ListIncidents(c *gin.Context) {
	if !h.checkDB(c) {
		return
	}

	incidents, err := h.repo.GetIncidents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, incidents)
}

func (h *IncidentHandler) GetIncident(c *gin.Context) {
	if !h.checkDB(c) {
		return
	}

	id := c.Param("id")
	incident, err := h.repo.GetIncidentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, incident)
}

func (h *IncidentHandler) GetRootCause(c *gin.Context) {
	if !h.checkDB(c) {
		return
	}

	id := c.Param("id")
	graph, err := h.repo.GetRootCause(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, graph)
}
