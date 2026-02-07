package db

import (
	"billing-api/internal/config"
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(config *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Max number of connections in the pool
	log.Printf("set MaxConns : %d\n", config.MaxConns)
	cfg.MaxConns = int32(config.MaxConns)

	// Min number of connections to keep idle/ready
	log.Printf("set MinConns : %d\n", config.MinConns)
	cfg.MinConns = int32(config.MinConns)

	// Time a connection can be idle before being closed
	log.Printf("set MaxConnIdleTime : %d seconds\n", config.MaxConnIdleTime)
	cfg.MaxConnIdleTime = time.Duration(config.MaxConnIdleTime) * time.Second

	// Total lifetime of a connection
	log.Printf("set MaxConnLifetime : %d seconds \n", config.MaxConnLifeTime)
	cfg.MaxConnLifetime = time.Duration(config.MaxConnLifeTime) * time.Second

	// Time to wait to acquire a connection from the pool
	// before returning a "context deadline exceeded" error
	log.Printf("set HealthCheckPeriod : %d seconds\n", config.HealthCheckPeriod)
	cfg.HealthCheckPeriod = time.Duration(config.HealthCheckPeriod) * time.Second

	return pgxpool.NewWithConfig(ctx, cfg)
}
