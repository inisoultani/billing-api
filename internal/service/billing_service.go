package service

import (
	"billing-api/internal/infra/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingService struct {
	quaries *sqlc.Queries
}

// constructor
func NewBillingService(pool *pgxpool.Pool) *BillingService {
	return &BillingService{
		quaries: sqlc.New(pool),
	}
}
