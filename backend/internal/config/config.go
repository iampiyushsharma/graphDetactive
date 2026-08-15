package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	CognoDBURI      string
	CognoDBPassword string
}

func LoadConfig() *Config {
	// Attempt to load .env file, ignore if it doesn't exist
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	uri := os.Getenv("COGNO_DB_URI")
	if uri == "" {
		log.Println("WARNING: COGNO_DB_URI is not set")
	}

	password := os.Getenv("COGNO_DB_PASSWORD")
	if password == "" {
		log.Println("WARNING: COGNO_DB_PASSWORD is not set")
	}

	return &Config{
		Port:            port,
		CognoDBURI:      uri,
		CognoDBPassword: password,
	}
}
