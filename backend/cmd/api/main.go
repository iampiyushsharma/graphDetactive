package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novacart/graph-detective/internal/config"
	"github.com/novacart/graph-detective/internal/database"
	"github.com/novacart/graph-detective/internal/handler"
	"github.com/novacart/graph-detective/internal/repository"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	log.Println("Starting Graph Detective API server...")
	cfg := config.LoadConfig()

	// Connect to CognoDB
	db, err := database.Connect(cfg.CognoDBURI, cfg.CognoDBPassword)
	if err != nil {
		log.Printf("ERROR: Database connection failed (running in offline-ready mode): %v", err)
	} else {
		defer db.Close()
	}

	// Initialize Repository and Handlers
	repo := repository.NewGraphRepository(db)
	statusHandler := handler.NewStatusHandler(db)
	incidentHandler := handler.NewIncidentHandler(repo, db)
	serviceHandler := handler.NewServiceHandler(repo, db)

	// Setup Gin
	r := gin.Default()
	r.Use(CORSMiddleware())

	// Routes
	api := r.Group("/api")
	{
		api.GET("/status", statusHandler.CheckStatus)
		api.GET("/incidents", incidentHandler.ListIncidents)
		api.GET("/incidents/:id", incidentHandler.GetIncident)
		api.GET("/incidents/:id/root-cause", incidentHandler.GetRootCause)
		api.GET("/services/:id/blast-radius", serviceHandler.GetBlastRadius)
		api.GET("/services/:id/dependencies", serviceHandler.GetDependencies)
	}

	// Root fallback
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
	})

	log.Printf("Listening and serving HTTP on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run HTTP server: %v", err)
	}
}
