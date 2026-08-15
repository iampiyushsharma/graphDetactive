package main

import (
	"context"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/novacart/graph-detective/internal/config"
	"github.com/novacart/graph-detective/internal/database"
)

func main() {
	log.Println("Starting optimized database seeding...")
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

	// 2. Optimized seed query in a single round-trip transaction
	log.Println("Creating nodes and relationships in a single transactional batch...")
	seedQuery := `
	CREATE 
	  // Teams
	  (team_platform:Team {id: "team-platform", name: "Platform Team", slackChannel: "#team-platform"}),
	  (team_commerce:Team {id: "team-commerce", name: "Core Commerce Team", slackChannel: "#team-commerce"}),
	  (team_integration:Team {id: "team-integration", name: "Integration Team", slackChannel: "#team-integration"}),

	  // Developers
	  (dev_sarah:Developer {id: "dev-sarah", name: "Sarah Jenkins", email: "sarah@novacart.io"}),
	  (dev_david:Developer {id: "dev-david", name: "David Miller", email: "david@novacart.io"}),
	  (dev_alex:Developer {id: "dev-alex", name: "Alex Wong", email: "alex@novacart.io"}),

	  // Services
	  (gateway:Service {id: "gateway", name: "Gateway", type: "service", language: "TypeScript", status: "HEALTHY"}),
	  (auth_service:Service {id: "auth-service", name: "AuthService", type: "service", language: "Go", status: "HEALTHY"}),
	  (catalog_service:Service {id: "catalog-service", name: "ProductCatalogService", type: "service", language: "Go", status: "DEGRADED"}),
	  (cart_service:Service {id: "cart-service", name: "CartService", type: "service", language: "Node.js", status: "DEGRADED"}),
	  (order_service:Service {id: "order-service", name: "OrderService", type: "service", language: "Java", status: "CRITICAL"}),
	  (payment_service:Service {id: "payment-service", name: "PaymentService", type: "service", language: "Go", status: "CRITICAL"}),
	  (inventory_service:Service {id: "inventory-service", name: "InventoryService", type: "service", language: "Java", status: "HEALTHY"}),
	  (notification_service:Service {id: "notification-service", name: "NotificationService", type: "service", language: "Python", status: "HEALTHY"}),

	  // Databases
	  (auth_db:Database {id: "auth-db", name: "Auth-DB", type: "PostgreSQL", status: "HEALTHY"}),
	  (catalog_db:Database {id: "catalog-db", name: "Product-DB", type: "MongoDB", status: "HEALTHY"}),
	  (catalog_cache:Database {id: "catalog-cache", name: "Catalog-Cache", type: "Redis", status: "HEALTHY"}),
	  (cart_redis:Database {id: "cart-redis", name: "Cart-Redis", type: "Redis", status: "DEGRADED"}),
	  (order_db:Database {id: "order-db", name: "Order-DB", type: "PostgreSQL", status: "HEALTHY"}),
	  (inventory_db:Database {id: "inventory-db", name: "Inventory-DB", type: "PostgreSQL", status: "HEALTHY"}),

	  // Service -> Team Ownerships
	  (gateway)-[:OWNED_BY]->(team_platform),
	  (auth_service)-[:OWNED_BY]->(team_platform),
	  (catalog_service)-[:OWNED_BY]->(team_commerce),
	  (cart_service)-[:OWNED_BY]->(team_commerce),
	  (order_service)-[:OWNED_BY]->(team_commerce),
	  (payment_service)-[:OWNED_BY]->(team_integration),
	  (inventory_service)-[:OWNED_BY]->(team_integration),
	  (notification_service)-[:OWNED_BY]->(team_integration),

	  // Service -> Service / Database Dependencies
	  (gateway)-[:DEPENDS_ON]->(auth_service),
	  (gateway)-[:DEPENDS_ON]->(catalog_service),
	  (gateway)-[:DEPENDS_ON]->(cart_service),
	  (gateway)-[:DEPENDS_ON]->(order_service),
	  (auth_service)-[:DEPENDS_ON]->(auth_db),
	  (catalog_service)-[:DEPENDS_ON]->(catalog_db),
	  (catalog_service)-[:DEPENDS_ON]->(catalog_cache),
	  (cart_service)-[:DEPENDS_ON]->(cart_redis),
	  (order_service)-[:DEPENDS_ON]->(order_db),
	  (order_service)-[:DEPENDS_ON]->(payment_service),
	  (order_service)-[:DEPENDS_ON]->(inventory_service),
	  (inventory_service)-[:DEPENDS_ON]->(inventory_db),

	  // Incident 1 Scenario (Checkout timeout failure)
	  (inc1:Incident {
		id: "inc-101",
		title: "Order Checkout Failures (HTTP 500)",
		severity: "CRITICAL",
		status: "ACTIVE",
		description: "Customers are unable to complete checkouts. Order Service is returning HTTP 500 errors on the /checkout route.",
		createdAt: datetime("2026-08-15T04:00:00Z")
	  }),
	  (inc1)-[:AFFECTS]->(order_service),
	  (dep1:Deployment {
		id: "dep-992",
		version: "v2.4.12",
		env: "production",
		deployedAt: datetime("2026-08-15T03:50:00Z"),
		status: "SUCCESS"
	  }),
	  (dep1)-[:TO_SERVICE]->(payment_service),
	  (com1:Commit {
		id: "com-442",
		hash: "8bf9a2c",
		message: "refactor(payment): lower timeout for faster client failover",
		author: "Sarah Jenkins",
		committedAt: datetime("2026-08-15T03:30:00Z")
	  }),
	  (com1)-[:AUTHORED_BY]->(dev_sarah),
	  (dep1)-[:CONTAINS]->(com1),
	  (cfg1:ConfigChange {
		id: "cfg-101",
		key: "payment_timeout_ms",
		oldValue: "5000",
		newValue: "50"
	  }),
	  (com1)-[:MODIFIED_CONFIG]->(cfg1),

	  // Incident 2 Scenario (Redis configuration failure)
	  (inc2:Incident {
		id: "inc-102",
		title: "Unable to Add Items to Cart",
		severity: "HIGH",
		status: "ACTIVE",
		description: "Cart operations are failing with write timeouts. Users cannot add items to their shopping carts.",
		createdAt: datetime("2026-08-15T04:15:00Z")
	  }),
	  (inc2)-[:AFFECTS]->(cart_service),
	  (dep2:Deployment {
		id: "dep-881",
		version: "v1.9.3",
		env: "production",
		deployedAt: datetime("2026-08-15T04:05:00Z"),
		status: "SUCCESS"
	  }),
	  (dep2)-[:TO_DATABASE]->(cart_redis),
	  (com2:Commit {
		id: "com-331",
		hash: "f3c2b1a",
		message: "chore(redis): update eviction policy config for caching persistence",
		author: "David Miller",
		committedAt: datetime("2026-08-15T03:00:00Z")
	  }),
	  (com2)-[:AUTHORED_BY]->(dev_david),
	  (dep2)-[:CONTAINS]->(com2),
	  (cfg2:ConfigChange {
		id: "cfg-102",
		key: "maxmemory-policy",
		oldValue: "allkeys-lru",
		newValue: "noeviction"
	  }),
	  (com2)-[:MODIFIED_CONFIG]->(cfg2),

	  // Incident 3 Scenario (Catalog DB resource limit leak)
	  (inc3:Incident {
		id: "inc-103",
		title: "Product details page timeouts",
		severity: "MEDIUM",
		status: "ACTIVE",
		description: "Users are reporting slow load times or gateway timeouts when viewing product detail pages.",
		createdAt: datetime("2026-08-15T04:30:00Z")
	  }),
	  (inc3)-[:AFFECTS]->(catalog_service),
	  (dep3:Deployment {
		id: "dep-772",
		version: "v3.1.0",
		env: "production",
		deployedAt: datetime("2026-08-15T02:00:00Z"),
		status: "SUCCESS"
	  }),
	  (dep3)-[:TO_SERVICE]->(catalog_service),
	  (com3:Commit {
		id: "com-221",
		hash: "a1b2c3d",
		message: "feat(catalog): introduce concurrency in database queries",
		author: "Alex Wong",
		committedAt: datetime("2026-08-15T01:30:00Z")
	  }),
	  (com3)-[:AUTHORED_BY]->(dev_alex),
	  (dep3)-[:CONTAINS]->(com3),
	  (cfg3:ConfigChange {
		id: "cfg-103",
		key: "db_max_open_conns",
		oldValue: "100",
		newValue: "200"
	  }),
	  (com3)-[:MODIFIED_CONFIG]->(cfg3)
	`

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(ctx, seedQuery, nil)
	})
	if err != nil {
		log.Fatalf("Failed to execute optimized batch seed: %v", err)
	}

	log.Println("Seeding completed successfully in a single batch!")
}
