package service

import (
	"billing-api/internal/domain"
	"billing-api/internal/infra/db/sqlc"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLoanNotFound            = errors.New("Loan not found")
	ErrInvalidStateOutstanding = errors.New("Invalid loan payment state")
	ErrInvalidLoanTerms        = errors.New("Invalid loan terms")
	ErrInvalidPayment          = errors.New("Invalid payment")
	ErrLoanAlreadyClosed       = errors.New("Loan already fully paid")
)

type SubmitLoanInput struct {
	PrincipalAmount    int64
	AnnualInterestRate float64 // e.g. 0.10
	TotalWeeks         int
	StartDate          time.Time
}

type SubmitPaymentInput struct {
	LoanID int64
	Amount int64
	PaidAt time.Time
}

type BillingService struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func mapLoan(l sqlc.Loan) *domain.Loan {
	return &domain.Loan{
		ID:                  l.ID,
		PrincipalAmount:     l.PrincipalAmount,
		TotalPayableAmount:  l.TotalPayableAmount,
		WeeklyPaymentAmount: l.WeeklyPaymentAmount,
		TotalWeeks:          int(l.TotalWeeks),
		CreatedAt:           l.CreatedAt.Time,
	}
}

// constructor
func NewBillingService(pool *pgxpool.Pool) *BillingService {
	return &BillingService{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

/*
GetLoan get loan detail based on id
*/
func (s *BillingService) GetLoanByID(ctx context.Context, loanID int64) (*domain.Loan, error) {

	// load loan
	loan, err := s.queries.GetLoanByID(ctx, loanID)
	if err != nil {
		return nil, ErrLoanNotFound
	}

	return mapLoan(loan), nil
}

/*
SubmitLoan creates a new loan and save all necessary billing data

The assumptions (as allowed by the problem statement):
- The loan uses flat interest.
- Interest is applied once to the full principal (per annum).
- Weekly payment is calculated as total_payable / total_weeks.
- The total payable amount must be evenly divisible by total_weeks; otherwise, loan creation fails.
*/
func (s *BillingService) SubmitLoan(ctx context.Context, input SubmitLoanInput) (*domain.Loan, error) {

	return withTx(ctx, s.pool, func(tx pgx.Tx) (*domain.Loan, error) {
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

		return mapLoan(loan), nil
	})
}

/*
GetOutstanding get total amount that user still need to pay
*/
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

/*
SubmitPayment will validate the input and determines the next unpaid week, and persists a loan payment.

When a payment is submitted (as interpreted within the problem statement):
- Loan must exist
- Loan must not be fully paid
- Payment amount must equal weekly_payment_amount
- Payment applies to the next unpaid week
- A week cannot be paid twice
- Operation must be atomic (transaction)
*/
func (s *BillingService) SubmitPayment(ctx context.Context, input SubmitPaymentInput) (int64, error) {

	return withTx(ctx, s.pool, func(tx pgx.Tx) (int64, error) {
		// load loan
		loan, err := s.queries.GetLoanByID(ctx, input.LoanID)
		if err != nil {
			return 0, ErrLoanNotFound
		}

		// check for outstanding
		totalPaid, err := s.queries.GetTotalPaidAmount(ctx, input.LoanID)
		if err != nil {
			return 0, err
		}
		if totalPaid > loan.TotalPayableAmount {
			return 0, ErrLoanAlreadyClosed
		}
		if input.Amount != loan.WeeklyPaymentAmount {
			return 0, ErrInvalidPayment
		}

		// determine next unpaid week
		paidWeeks, err := s.queries.GetPaidWeeksCount(ctx, input.LoanID)
		if err != nil {
			return 0, err
		}
		nextWeek := paidWeeks + 1
		if nextWeek > loan.TotalWeeks {
			return 0, ErrLoanAlreadyClosed
		}

		payment, err := s.queries.InsertPayment(ctx, sqlc.InsertPaymentParams{
			LoanID:     input.LoanID,
			WeekNumber: nextWeek,
			Amount:     input.Amount,
			PaidAt: pgtype.Timestamp{
				Time:  input.PaidAt,
				Valid: true,
			},
		})
		if err != nil {
			return 0, err
		}
		return payment.ID, nil
	})
}

/*
IsDelinquent check if the the loan currently in deliquent state or not

Delinquency is currently modeled as a derived state, calculated on demand based on loan creation time and payment history.
Intentionally not persisting delinquency at this stage to avoid consistency issues.
As the system evolves (eg notifications, collections workflows, regulatory reporting),
delinquency can be promoted to an explicit loan lifecycle state, managed via a state machine and updated through well-defined domain events.

A loan is delinquent if:
There is a gap of 2 or more weeks between:
- the latest paid week
- and the current expected week
*/
func (s *BillingService) IsDelinquent(ctx context.Context, loanID int64, now time.Time) (bool, error) {
	// load loan
	loan, err := s.queries.GetLoanByID(ctx, loanID)
	if err != nil {
		return false, ErrLoanNotFound
	}

	// get loan last paid week
	lastPaidWeek, err := s.queries.GetLastPaidWeek(ctx, loanID)
	if err != nil {
		return false, err
	}

	// the loan start date is assumed to be the loan creation timestamp, as the problem statement does not describe a separate approval or disbursement phase.
	// if later such usecases are introduced in the future, we could use start_date field and used it later in the future for delinquency and repayment scheduling logic.
	expectedWeek := weekSince(loan.CreatedAt.Time, now)

	// loan either just started or not yet
	if expectedWeek <= 1 {
		return false, nil
	}

	gap := expectedWeek - int(lastPaidWeek)
	return gap >= 2, nil
}

/*
weekSince internal helper method to calculte week duration between 2 different time
*/
func weekSince(start, now time.Time) int {
	if now.Before(start) {
		return 0
	}

	duration := now.Sub(start)
	weeks := int(duration.Hours() / (24 * 7))

	return weeks + 1
}
