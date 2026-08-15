package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novacart/graph-detective/internal/database"
	"github.com/novacart/graph-detective/internal/repository"
)

type ServiceHandler struct {
	repo *repository.GraphRepository
	db   *database.DB
}

func NewServiceHandler(repo *repository.GraphRepository, db *database.DB) *ServiceHandler {
	return &ServiceHandler{repo: repo, db: db}
}

func (h *ServiceHandler) checkDB(c *gin.Context) bool {
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

func (h *ServiceHandler) GetBlastRadius(c *gin.Context) {
	if !h.checkDB(c) {
		return
	}

	id := c.Param("id")
	graph, err := h.repo.GetBlastRadius(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, graph)
}

func (h *ServiceHandler) GetDependencies(c *gin.Context) {
	if !h.checkDB(c) {
		return
	}

	id := c.Param("id")
	graph, err := h.repo.GetDependencies(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, graph)
}
