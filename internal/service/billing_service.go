package service

import (
	"billing-api/internal/domain"
	"billing-api/internal/infra/db/sqlc"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLoanNotFound            = errors.New("Loan not found")
	ErrInvalidStateOutstanding = errors.New("Invalid loan payment state")
	ErrInvalidLoanTerms        = errors.New("Invalid loan terms")
	ErrInvalidPayment          = errors.New("Invalid payment")
	ErrLoanAlreadyClosed       = errors.New("Loan already fully paid")
	ErrDuplicatePayment        = errors.New("Duplicate payment for current week")
	ErrDelinquencyCheck        = errors.New("Failed to compute loan delinquency")
)

type BillingService struct {
	pool *pgxpool.Pool
	repo domain.BillingRepository
}

// constructor
func NewBillingService(pool *pgxpool.Pool, repo domain.BillingRepository) *BillingService {
	return &BillingService{
		pool: pool,
		repo: repo,
	}
}

/*
GetLoan get loan detail based on id
*/
func (s *BillingService) GetLoanByID(ctx context.Context, loanID int64) (*domain.Loan, error) {

	// load loan
	loan, err := s.repo.GetLoanByID(ctx, loanID)
	if err != nil {
		return nil, ErrLoanNotFound
	}

	return domain.MapLoan(loan), nil
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

	var domainLoan *domain.Loan

	err := s.repo.WithTx(ctx, func(repo domain.BillingRepository) error {
		if input.PrincipalAmount <= 0 || input.TotalWeeks <= 0 {
			return ErrInvalidLoanTerms
		}

		// flat interest
		totalInterest := int64(float64(input.PrincipalAmount) * input.AnnualInterestRate)
		totalPayable := input.PrincipalAmount + totalInterest

		if totalPayable%int64(input.TotalWeeks) != 0 {
			return errors.New("weekly payment is not evenly divisible")
		}

		weeklyPayment := totalPayable / int64(input.TotalWeeks)

		loan, err := repo.InsertLoan(ctx, sqlc.InsertLoanParams{
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
			return err
		}

		// generate Schedules with batch insert instead of multiple insert
		schedules := make([]sqlc.CreateLoanSchedulesParams, input.TotalWeeks)

		for i := 1; i <= input.TotalWeeks; i++ {
			dueDate := input.StartDate.AddDate(0, 0, 7*i)

			schedules[i-1] = sqlc.CreateLoanSchedulesParams{
				LoanID:   loan.ID,
				Sequence: int32(i),
				DueDate:  pgtype.Date{Time: dueDate, Valid: true},
				Amount:   weeklyPayment,
				Status:   "PENDING",
			}
		}

		// One single database call
		_, err = repo.CreateLoanSchedules(ctx, schedules)
		if err != nil {
			return err
		}

		domainLoan = domain.MapLoan(loan)
		return nil
	})

	return domainLoan, err
}

/*
GetOutstanding get total amount that user still need to pay
*/
func (s *BillingService) GetOutstanding(ctx context.Context, loanID int64) (int64, error) {
	loan, err := s.repo.GetLoanByID(ctx, loanID)
	if err != nil {
		return 0, ErrLoanNotFound
	}

	totalPaid, err := s.repo.GetTotalPaidAmount(ctx, loanID)
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

	var paymentID int64
	err := s.repo.WithTx(ctx, func(repo domain.BillingRepository) error {
		// load loan
		loan, err := repo.GetLoanByID(ctx, input.LoanID)
		if err != nil {
			return ErrLoanNotFound
		}

		// check for outstanding
		totalPaid, err := repo.GetTotalPaidAmount(ctx, input.LoanID)
		if err != nil {
			return err
		}
		if totalPaid > loan.TotalPayableAmount {
			return ErrLoanAlreadyClosed
		}
		if input.Amount != loan.WeeklyPaymentAmount {
			return ErrInvalidPayment
		}

		// determine next unpaid week
		paidWeeks, err := repo.GetPaidWeeksCount(ctx, input.LoanID)
		if err != nil {
			return err
		}
		nextWeek := paidWeeks + 1
		if nextWeek > loan.TotalWeeks {
			return ErrLoanAlreadyClosed
		}

		payment, err := repo.InsertPayment(ctx, sqlc.InsertPaymentParams{
			LoanID:         input.LoanID,
			WeekNumber:     nextWeek,
			Amount:         input.Amount,
			IdempotencyKey: input.IdempotencyKey,
			PaidAt: pgtype.Timestamp{
				Time:  input.PaidAt,
				Valid: true,
			},
		})
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				// This is a "Duplicate" error.
				return ErrDuplicatePayment
			}
			return err
		}

		// Update the specific schedule record
		// Use GetScheduleBySequence to find the ID instead of a manual loop
		sch, err := repo.GetScheduleBySequence(ctx, sqlc.GetScheduleBySequenceParams{
			LoanID:   input.LoanID,
			Sequence: int32(nextWeek),
		})
		if err != nil {
			return err
		}
		log.Printf("schedule id that going to be updated : %d\n", sch.ID)
		_, err = repo.UpdateSchedulePayment(ctx, sqlc.UpdateSchedulePaymentParams{
			ID:         sch.ID,
			PaidAmount: input.Amount,
		})
		if err != nil {
			return err
		}

		paymentID = payment.ID
		return nil
	})
	return paymentID, err
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
	loan, err := s.repo.GetLoanByID(ctx, loanID)
	if err != nil {
		return false, ErrLoanNotFound
	}

	// get loan last paid week
	lastPaidWeek, err := s.repo.GetLastPaidWeek(ctx, loanID)
	if err != nil {
		return false, fmt.Errorf("%w %v", ErrDelinquencyCheck, err)
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
ListPayments return all payment records based on loan id
*/
func (s *BillingService) ListPayments(ctx context.Context, loanID int64, limit int, cursor *domain.PaymentCursor) ([]*domain.Payment, *domain.PaymentCursor, error) {

	params := sqlc.ListPaymentsByLoanIDParams{
		LoanID:   loanID,
		LimitVal: int32(limit),
	}
	// ensure cursor not null or by default use nil as value for paidAt and id
	if cursor != nil {
		params.CursorPaidAt = pgtype.Timestamptz{
			Time:  cursor.PaidAt,
			Valid: true,
		}
		params.CursorID = pgtype.Int8{
			Int64: cursor.ID,
			Valid: true,
		}
	} else {
		params.CursorPaidAt = pgtype.Timestamptz{Valid: false}
		params.CursorID = pgtype.Int8{Valid: false}
	}

	rows, err := s.repo.ListPaymentsByLoanID(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	payments := make([]*domain.Payment, 0, len(rows))
	for _, r := range rows {
		payments = append(payments, domain.MapPayment(r))
	}

	var nextCursor *domain.PaymentCursor
	if len(rows) == limit {
		last := rows[len(rows)-1]
		nextCursor = &domain.PaymentCursor{
			PaidAt: last.PaidAt.Time,
			ID:     last.ID,
		}
	}

	return payments, nextCursor, nil
}

/*
ListSchedules return all schedule records based on loan id
*/
func (s *BillingService) ListSchedules(ctx context.Context, loanID int64, limit int, cursor *domain.ScheduleCursor) ([]*domain.LoanSchedule, *domain.ScheduleCursor, error) {

	params := sqlc.ListSchedulesByLoanIDWithCursorParams{
		LoanID:   loanID,
		Limit:    int32(limit),
		Sequence: 0, // Default to start from the beginning
	}

	if cursor != nil {
		params.Sequence = cursor.Sequence
	}

	// Fetch using sequence-based pagination
	rows, err := s.repo.ListSchedulesByLoanID(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	schedules := make([]*domain.LoanSchedule, 0, len(rows))
	for _, r := range rows {
		schedules = append(schedules, domain.MapSchedule(r))
	}

	// Generate Next Cursor
	var nextCursor *domain.ScheduleCursor
	if len(schedules) == limit {
		lastItem := schedules[len(schedules)-1]
		nextCursor = &domain.ScheduleCursor{
			Sequence: int32(lastItem.Sequence),
		}
	}

	return schedules, nextCursor, nil
}
