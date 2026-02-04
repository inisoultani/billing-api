package service

import (
	"billing-api/internal/infra/db/sqlc"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrLoanNotFound = errors.New("Loan not found")
var ErrInvalidStateOutstanding = errors.New("Invalid loan payment state")

type BillingService struct {
	quaries *sqlc.Queries
}

// constructor
func NewBillingService(pool *pgxpool.Pool) *BillingService {
	return &BillingService{
		quaries: sqlc.New(pool),
	}
}

func (s *BillingService) GetOutstanding(ctx context.Context, loanID int64) (int64, error) {
	loan, err := s.quaries.GetLoanByID(ctx, loanID)
	if err != nil {
		return 0, ErrLoanNotFound
	}

	totalPaid, err := s.quaries.GetTotalPaidAmount(ctx, loanID)
	if err != nil {
		return 0, err
	}

	outstanding := loan.TotalPayableAmount - totalPaid
	if outstanding < 0 {
		// safety guard, to ack false behaviour
		return 0, ErrInvalidStateOutstanding
	}

	return outstanding, nil
}
