package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novacart/graph-detective/internal/database"
	"github.com/novacart/graph-detective/internal/handler"
	"github.com/novacart/graph-detective/internal/repository"
)

// CORSMiddleware handles cross-origin requests safely
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SetupRouter initializes handlers and registers HTTP endpoints
func SetupRouter(db *database.DB, repo *repository.GraphRepository) *gin.Engine {
	statusHandler := handler.NewStatusHandler(db)
	incidentHandler := handler.NewIncidentHandler(repo, db)
	serviceHandler := handler.NewServiceHandler(repo, db)

	r := gin.Default()
	r.Use(CORSMiddleware())

	api := r.Group("/api")
	{
		api.GET("/status", statusHandler.CheckStatus)
		api.GET("/incidents", incidentHandler.ListIncidents)
		api.GET("/incidents/:id", incidentHandler.GetIncident)
		api.GET("/incidents/:id/root-cause", incidentHandler.GetRootCause)
		api.GET("/services/:id/blast-radius", serviceHandler.GetBlastRadius)
		api.GET("/services/:id/dependencies", serviceHandler.GetDependencies)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
	})

	return r
}
