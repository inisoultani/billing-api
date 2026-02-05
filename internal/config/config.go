package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	PagingLimitDefault int
	PagingLimitMax     int
	ServerPort         string
}

func Load() (*Config, error) {
	log.Println("start loading config from .env config file or current env variables")
	// Loads .env into system environment variables
	err := godotenv.Load()
	if err != nil {
		// .env is missing in Production,
		// as vars are set directly in the OS/Docker.
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://localhost:5432/billing"),
		PagingLimitDefault: getEnvInt("PAGING_LIMIT_DEFAULT", 10),
		PagingLimitMax:     getEnvInt("PAGING_LIMIT_MAX", 100),
		ServerPort:         getEnv("SERVER_PORT", "8081"),
	}, nil
}

// getEnv helper to get string env or return default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	s, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
