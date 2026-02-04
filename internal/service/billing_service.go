package service

import (
	"billing-api/internal/infra/db/sqlc"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLoanNotFound            = errors.New("Loan not found")
	ErrInvalidStateOutstanding = errors.New("Invalid loan payment state")
	ErrInvalidLoanTerms        = errors.New("invalid loan terms")
)

type SubmitLoanInput struct {
	PrincipalAmount    int64
	AnnualInterestRate float64 // e.g. 0.10
	TotalWeeks         int
	StartDate          time.Time
}

type BillingService struct {
	queries *sqlc.Queries
}

// constructor
func NewBillingService(pool *pgxpool.Pool) *BillingService {
	return &BillingService{
		queries: sqlc.New(pool),
	}
}

/*
SubmitLoan creates a new loan and save all necessary billing data

The assumptions (as allowed by the problem statement):
- The loan uses flat interest.
- Interest is applied once to the full principal (per annum).
- Weekly payment is calculated as total_payable / total_weeks.
- The total payable amount must be evenly divisible by total_weeks; otherwise, loan creation fails.
*/
func (s *BillingService) SubmitLoan(ctx context.Context, input SubmitLoanInput) (*sqlc.Loan, error) {

	if input.PrincipalAmount <= 0 || input.TotalWeeks <= 0 {
		return nil, ErrInvalidLoanTerms
	}

	// flat interest
	totalInterest := int64(float64(input.PrincipalAmount) * input.AnnualInterestRate)
	totalPayable := input.PrincipalAmount + totalInterest

	if totalPayable%int64(input.TotalWeeks) != 0 {
		return nil, errors.New("weekly payment is not evenly divisible")
	}

	weeklyPayment := totalPayable / int64(input.TotalWeeks)

	loan, err := s.queries.InsertLoan(ctx, sqlc.InsertLoanParams{
		PrincipalAmount:     input.PrincipalAmount,
		TotalInterestAmount: totalInterest,
		TotalPayableAmount:  totalPayable,
		WeeklyPaymentAmount: weeklyPayment,
		TotalWeeks:          int32(input.TotalWeeks),
		StartDate: pgtype.Date{
			Time:  input.StartDate,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

func (s *BillingService) GetOutstanding(ctx context.Context, loanID int64) (int64, error) {
	loan, err := s.queries.GetLoanByID(ctx, loanID)
	if err != nil {
		return 0, ErrLoanNotFound
	}

	totalPaid, err := s.queries.GetTotalPaidAmount(ctx, loanID)
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
