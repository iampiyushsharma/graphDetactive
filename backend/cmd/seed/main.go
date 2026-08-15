package main

import (
	"context"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/novacart/graph-detective/internal/config"
	"github.com/novacart/graph-detective/internal/database"
)

func main() {
	log.Println("Starting database seeding...")
	cfg := config.LoadConfig()

	if cfg.CognoDBURI == "" || cfg.CognoDBPassword == "" {
		log.Fatal("Error: COGNO_DB_URI and COGNO_DB_PASSWORD must be set in env")
	}

	db, err := database.Connect(cfg.CognoDBURI, cfg.CognoDBPassword)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	session := db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// 1. Clear database
	log.Println("Clearing existing data...")
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
	})
	if err != nil {
		log.Fatalf("Failed to clear database: %v", err)
	}

	// 2. Seed scenarios
	log.Println("Creating nodes and relationships...")
	queries := []string{
		// Teams
		`CREATE (t1:Team {id: "team-platform", name: "Platform Team", slackChannel: "#team-platform"})`,
		`CREATE (t2:Team {id: "team-commerce", name: "Core Commerce Team", slackChannel: "#team-commerce"})`,
		`CREATE (t3:Team {id: "team-integration", name: "Integration Team", slackChannel: "#team-integration"})`,

		// Developers
		`CREATE (d1:Developer {id: "dev-sarah", name: "Sarah Jenkins", email: "sarah@novacart.io"})`,
		`CREATE (d2:Developer {id: "dev-david", name: "David Miller", email: "david@novacart.io"})`,
		`CREATE (d3:Developer {id: "dev-alex", name: "Alex Wong", email: "alex@novacart.io"})`,

		// Services
		`CREATE (s1:Service {id: "gateway", name: "Gateway", type: "service", language: "TypeScript", status: "HEALTHY"})`,
		`CREATE (s2:Service {id: "auth-service", name: "AuthService", type: "service", language: "Go", status: "HEALTHY"})`,
		`CREATE (s3:Service {id: "catalog-service", name: "ProductCatalogService", type: "service", language: "Go", status: "DEGRADED"})`,
		`CREATE (s4:Service {id: "cart-service", name: "CartService", type: "service", language: "Node.js", status: "DEGRADED"})`,
		`CREATE (s5:Service {id: "order-service", name: "OrderService", type: "service", language: "Java", status: "CRITICAL"})`,
		`CREATE (s6:Service {id: "payment-service", name: "PaymentService", type: "service", language: "Go", status: "CRITICAL"})`,
		`CREATE (s7:Service {id: "inventory-service", name: "InventoryService", type: "service", language: "Java", status: "HEALTHY"})`,
		`CREATE (s8:Service {id: "notification-service", name: "NotificationService", type: "service", language: "Python", status: "HEALTHY"})`,

		// Databases
		`CREATE (db1:Database {id: "auth-db", name: "Auth-DB", type: "PostgreSQL", status: "HEALTHY"})`,
		`CREATE (db2:Database {id: "catalog-db", name: "Product-DB", type: "MongoDB", status: "HEALTHY"})`,
		`CREATE (db3:Database {id: "catalog-cache", name: "Catalog-Cache", type: "Redis", status: "HEALTHY"})`,
		`CREATE (db4:Database {id: "cart-redis", name: "Cart-Redis", type: "Redis", status: "DEGRADED"})`,
		`CREATE (db5:Database {id: "order-db", name: "Order-DB", type: "PostgreSQL", status: "HEALTHY"})`,
		`CREATE (db6:Database {id: "inventory-db", name: "Inventory-DB", type: "PostgreSQL", status: "HEALTHY"})`,

		// Team Ownerships
		`MATCH (t:Team {id: "team-platform"}), (s:Service) WHERE s.id IN ["gateway", "auth-service"] CREATE (s)-[:OWNED_BY]->(t)`,
		`MATCH (t:Team {id: "team-commerce"}), (s:Service) WHERE s.id IN ["catalog-service", "cart-service", "order-service"] CREATE (s)-[:OWNED_BY]->(t)`,
		`MATCH (t:Team {id: "team-integration"}), (s:Service) WHERE s.id IN ["payment-service", "inventory-service", "notification-service"] CREATE (s)-[:OWNED_BY]->(t)`,

		// Dependencies
		`MATCH (g:Service {id: "gateway"}), (s:Service) WHERE s.id IN ["auth-service", "catalog-service", "cart-service", "order-service"] CREATE (g)-[:DEPENDS_ON]->(s)`,
		`MATCH (s:Service {id: "auth-service"}), (db:Database {id: "auth-db"}) CREATE (s)-[:DEPENDS_ON]->(db)`,
		`MATCH (s:Service {id: "catalog-service"}), (db:Database) WHERE db.id IN ["catalog-db", "catalog-cache"] CREATE (s)-[:DEPENDS_ON]->(db)`,
		`MATCH (s:Service {id: "cart-service"}), (db:Database {id: "cart-redis"}) CREATE (s)-[:DEPENDS_ON]->(db)`,
		`MATCH (s:Service {id: "order-service"}), (db:Database {id: "order-db"}) CREATE (s)-[:DEPENDS_ON]->(db)`,
		`MATCH (s:Service {id: "order-service"}), (dep:Service) WHERE dep.id IN ["payment-service", "inventory-service"] CREATE (s)-[:DEPENDS_ON]->(dep)`,
		`MATCH (s:Service {id: "inventory-service"}), (db:Database {id: "inventory-db"}) CREATE (s)-[:DEPENDS_ON]->(db)`,

		// Incident 1: Checkout service failure
		`CREATE (i:Incident {
			id: "inc-101",
			title: "Order Checkout Failures (HTTP 500)",
			severity: "CRITICAL",
			status: "ACTIVE",
			description: "Customers are unable to complete checkouts. Order Service is returning HTTP 500 errors on the /checkout route.",
			createdAt: datetime("2026-08-15T04:00:00Z")
		})`,
		`MATCH (i:Incident {id: "inc-101"}), (s:Service {id: "order-service"}) CREATE (i)-[:AFFECTS]->(s)`,

		`CREATE (dep:Deployment {
			id: "dep-992",
			version: "v2.4.12",
			env: "production",
			deployedAt: datetime("2026-08-15T03:50:00Z"),
			status: "SUCCESS"
		})`,
		`MATCH (d:Deployment {id: "dep-992"}), (s:Service {id: "payment-service"}) CREATE (d)-[:TO_SERVICE]->(s)`,

		`CREATE (c:Commit {
			id: "com-442",
			hash: "8bf9a2c",
			message: "refactor(payment): lower timeout for faster client failover",
			author: "Sarah Jenkins",
			committedAt: datetime("2026-08-15T03:30:00Z")
		})`,
		`MATCH (c:Commit {id: "com-442"}), (dev:Developer {id: "dev-sarah"}) CREATE (c)-[:AUTHORED_BY]->(dev)`,
		`MATCH (dep:Deployment {id: "dep-992"}), (c:Commit {id: "com-442"}) CREATE (dep)-[:CONTAINS]->(c)`,

		`CREATE (cfg:ConfigChange {
			id: "cfg-101",
			key: "payment_timeout_ms",
			oldValue: "5000",
			newValue: "50"
		})`,
		`MATCH (c:Commit {id: "com-442"}), (cfg:ConfigChange {id: "cfg-101"}) CREATE (c)-[:MODIFIED_CONFIG]->(cfg)`,

		// Incident 2: Cart cache out of memory
		`CREATE (i:Incident {
			id: "inc-102",
			title: "Unable to Add Items to Cart",
			severity: "HIGH",
			status: "ACTIVE",
			description: "Cart operations are failing with write timeouts. Users cannot add items to their shopping carts.",
			createdAt: datetime("2026-08-15T04:15:00Z")
		})`,
		`MATCH (i:Incident {id: "inc-102"}), (s:Service {id: "cart-service"}) CREATE (i)-[:AFFECTS]->(s)`,

		`CREATE (dep:Deployment {
			id: "dep-881",
			version: "v1.9.3",
			env: "production",
			deployedAt: datetime("2026-08-15T04:05:00Z"),
			status: "SUCCESS"
		})`,
		`MATCH (d:Deployment {id: "dep-881"}), (db:Database {id: "cart-redis"}) CREATE (d)-[:TO_DATABASE]->(db)`,

		`CREATE (c:Commit {
			id: "com-331",
			hash: "f3c2b1a",
			message: "chore(redis): update eviction policy config for caching persistence",
			author: "David Miller",
			committedAt: datetime("2026-08-15T03:00:00Z")
		})`,
		`MATCH (c:Commit {id: "com-331"}), (dev:Developer {id: "dev-david"}) CREATE (c)-[:AUTHORED_BY]->(dev)`,
		`MATCH (dep:Deployment {id: "dep-881"}), (c:Commit {id: "com-331"}) CREATE (dep)-[:CONTAINS]->(c)`,

		`CREATE (cfg:ConfigChange {
			id: "cfg-102",
			key: "maxmemory-policy",
			oldValue: "allkeys-lru",
			newValue: "noeviction"
		})`,
		`MATCH (c:Commit {id: "com-331"}), (cfg:ConfigChange {id: "cfg-102"}) CREATE (c)-[:MODIFIED_CONFIG]->(cfg)`,

		// Incident 3: Catalog Service Memory Leak
		`CREATE (i:Incident {
			id: "inc-103",
			title: "Product details page timeouts",
			severity: "MEDIUM",
			status: "ACTIVE",
			description: "Users are reporting slow load times or gateway timeouts when viewing product detail pages.",
			createdAt: datetime("2026-08-15T04:30:00Z")
		})`,
		`MATCH (i:Incident {id: "inc-103"}), (s:Service {id: "catalog-service"}) CREATE (i)-[:AFFECTS]->(s)`,

		`CREATE (dep:Deployment {
			id: "dep-772",
			version: "v3.1.0",
			env: "production",
			deployedAt: datetime("2026-08-15T02:00:00Z"),
			status: "SUCCESS"
		})`,
		`MATCH (d:Deployment {id: "dep-772"}), (s:Service {id: "catalog-service"}) CREATE (d)-[:TO_SERVICE]->(s)`,

		`CREATE (c:Commit {
			id: "com-221",
			hash: "a1b2c3d",
			message: "feat(catalog): introduce concurrency in database queries",
			author: "Alex Wong",
			committedAt: datetime("2026-08-15T01:30:00Z")
		})`,
		`MATCH (c:Commit {id: "com-221"}), (dev:Developer {id: "dev-alex"}) CREATE (c)-[:AUTHORED_BY]->(dev)`,
		`MATCH (dep:Deployment {id: "dep-772"}), (c:Commit {id: "com-221"}) CREATE (dep)-[:CONTAINS]->(c)`,

		`CREATE (cfg:ConfigChange {
			id: "cfg-103",
			key: "db_max_open_conns",
			oldValue: "100",
			newValue: "200"
		})`,
		`MATCH (c:Commit {id: "com-221"}), (cfg:ConfigChange {id: "cfg-103"}) CREATE (c)-[:MODIFIED_CONFIG]->(cfg)`,
	}

	for i, q := range queries {
		log.Printf("Executing query %d/%d...", i+1, len(queries))
		_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			return tx.Run(ctx, q, nil)
		})
		if err != nil {
			log.Fatalf("Failed to execute query %d: %v", i+1, err)
		}
	}

	log.Println("Seeding completed successfully!")
}
