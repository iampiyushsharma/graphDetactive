package database

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type DB struct {
	Driver neo4j.DriverWithContext
}

func Connect(uri, password string) (*DB, error) {
	// CognoDB credentials: username is always "cognodb"
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth("cognodb", password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	ctx := context.Background()
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		log.Printf("WARNING: database connection verification failed: %v", err)
		return &DB{Driver: driver}, err
	}

	log.Println("Successfully connected to CognoDB!")
	return &DB{Driver: driver}, nil
}

func (db *DB) Close() {
	if db.Driver != nil {
		db.Driver.Close(context.Background())
	}
}
