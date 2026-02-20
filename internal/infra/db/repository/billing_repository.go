package repository

import (
	"billing-api/internal/domain"
	"billing-api/internal/infra/db"
	"billing-api/internal/infra/db/sqlc"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func getContextWithTimeout(ctx context.Context, label string, rowCount int) (context.Context, context.CancelFunc) {
	// base timeout
	timeout := 2 * time.Second

	// add 50ms per row for batch
	if rowCount > 1 {
		timeout += time.Duration(rowCount) * 50 * time.Millisecond
	}

	// cap time to 10sec
	if timeout > 10*time.Second {
		timeout = 10 * time.Second
	}

	cause := fmt.Errorf("repo-timeout : %s limit was (%v)", label, timeout)
	return context.WithTimeoutCause(ctx, timeout, cause)
}

func runWithTimeout[T any](ctx context.Context, label string, rowCount int, fn func(context.Context) (T, error)) (T, error) {
	childCtx, cancel := getContextWithTimeout(ctx, label, rowCount)
	defer cancel()

	t, err := fn(childCtx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// return the custom cause as the error itself
			var zero T
			return zero, context.Cause(childCtx)
		}
		return t, err
	}
	return t, nil
}

func (r *PostgresRepo) WithTx(ctx context.Context, fn func(repo domain.BillingRepository) error) error {
	return db.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Create a new repository that uses the transaction instead of the pool
		txRepo := &PostgresRepo{
			queries: r.queries.WithTx(tx),
			pool:    r.pool,
		}
		return fn(txRepo)
	})
}

// LOAN RELATED
// GetLoanByID retrieves a loan by its primary key
func (r *PostgresRepo) GetLoanByID(ctx context.Context, id int64) (*domain.Loan, error) {

	return runWithTimeout(ctx, "GetLoanByID", 1, func(ctx context.Context) (*domain.Loan, error) {
		l, err := r.queries.GetLoanByID(ctx, id)
		if err != nil {
			var zero *domain.Loan
			return zero, err
		}
		return MapLoan(l), nil
	})

}

// InsertLoan creates a new loan record
func (r *PostgresRepo) InsertLoan(ctx context.Context, arg domain.CreateLoanCommand) (*domain.Loan, error) {
	// set proper timeout for this process
	return runWithTimeout(ctx, "InsertLoan", 1, func(ctx context.Context) (*domain.Loan, error) {
		l, err := r.queries.InsertLoan(ctx, *MapCreateLoanCommand(&arg))
		if err != nil {
			var zero *domain.Loan
			return zero, err
		}
		return MapLoan(l), err
	})
}

// PAYMENT RELATED
// GetTotalPaidAmount calculates the sum of all payments for a loan
func (r *PostgresRepo) GetTotalPaidAmount(ctx context.Context, loanID int64) (int64, error) {
	return r.queries.GetTotalPaidAmount(ctx, loanID)
}

// GetPaidWeeksCount counts how many weekly payments have been made
func (r *PostgresRepo) GetPaidWeeksCount(ctx context.Context, loanID int64) (int32, error) {
	return r.queries.GetPaidWeeksCount(ctx, loanID)
}

// GetLastPaidWeek finds the highest week_number recorded in payments
func (r *PostgresRepo) GetLastPaidWeek(ctx context.Context, loanID int64) (int32, error) {
	return r.queries.GetLastPaidWeek(ctx, loanID)
}

// InsertPayment records a new payment for a specific week
func (r *PostgresRepo) InsertPayment(ctx context.Context, arg domain.CreatePaymentComand) (*domain.Payment, error) {
	return runWithTimeout(ctx, "InsertPayment", 1, func(ctx context.Context) (*domain.Payment, error) {
		p, err := r.queries.InsertPayment(ctx, *MapCreatePaymentComand(&arg))
		if err != nil {
			var zero *domain.Payment
			return zero, err
		}
		return MapPayment(p), nil
	})
}

// ListPaymentsByLoanID handles paginated retrieval of payments
func (r *PostgresRepo) ListPaymentsByLoanID(ctx context.Context, arg domain.ListPaymentsQuery) ([]domain.Payment, error) {
	return runWithTimeout(ctx, "List payments based on loanID", 1, func(ctx context.Context) ([]domain.Payment, error) {
		params := sqlc.ListPaymentsByLoanIDParams{
			LoanID:   arg.LoanID,
			LimitVal: int32(arg.LimitVal),
		}
		// ensure cursor not null or by default use nil as value for paidAt and id
		if !arg.CursorPaidAt.IsZero() {
			params.CursorPaidAt = pgtype.Timestamptz{
				Time:  arg.CursorPaidAt,
				Valid: true,
			}
			params.CursorID = pgtype.Int8{
				Int64: *arg.CursorID,
				Valid: true,
			}
		} else {
			params.CursorPaidAt = pgtype.Timestamptz{Valid: false}
			params.CursorID = pgtype.Int8{Valid: false}
		}
		paymentRows, err := r.queries.ListPaymentsByLoanID(ctx, params)
		if err != nil {
			return nil, err
		}
		payments := make([]domain.Payment, 0, len(paymentRows))
		for _, r := range paymentRows {
			payments = append(payments, MapListPaymentsByLoanIDRow(r))
		}
		return payments, nil
	})
}

// SCHEDULE RELATED
// CreateLoanSchedule record schedule during loan creation
func (r *PostgresRepo) CreateLoanSchedules(ctx context.Context, arg []domain.LoanSchedule) (int64, error) {
	// specifically set timeout for this particular process
	// for insert loan in self there will dedicate timout
	return runWithTimeout(ctx, "Batch insert schedule", len(arg), func(ctx context.Context) (int64, error) {
		// uncomment bellow code to simulate context time out and examining the error handler
		// _, err := simulateContextTimeout[int64](ctx)
		// if err != nil {
		// 	return 0, err
		// }
		params := make([]sqlc.CreateLoanSchedulesParams, len(arg))
		for i, s := range arg {
			params[i] = sqlc.CreateLoanSchedulesParams{
				LoanID:   s.LoanID,
				Sequence: int32(i + 1),
				DueDate:  pgtype.Date{Time: s.DueDate, Valid: true},
				Amount:   s.Amount,
				Status:   "PENDING",
			}
		}
		return r.queries.CreateLoanSchedules(ctx, params)
	})
}

// ListSchedulesByLoanID handle paginated retrieval of schedules
func (r *PostgresRepo) ListSchedulesByLoanID(ctx context.Context, arg sqlc.ListSchedulesByLoanIDWithCursorParams) ([]sqlc.Schedule, error) {
	return runWithTimeout(ctx, "List schedule based on loan ID", 10, func(ctx context.Context) ([]sqlc.Schedule, error) {
		return r.queries.ListSchedulesByLoanIDWithCursor(ctx, arg)
	})
}

// UpdateSchedulePayment update related schedule based payment sequence
func (r *PostgresRepo) UpdateSchedulePayment(ctx context.Context, arg sqlc.UpdateSchedulePaymentParams) (sqlc.Schedule, error) {
	return runWithTimeout(ctx, "Update schedule payment", 1, func(ctx context.Context) (sqlc.Schedule, error) {
		return r.queries.UpdateSchedulePayment(ctx, arg)
	})
}

// GetScheduleBySequence retrieve schedule based on sequence
func (r *PostgresRepo) GetScheduleBySequence(ctx context.Context, arg sqlc.GetScheduleBySequenceParams) (sqlc.Schedule, error) {
	return r.queries.GetScheduleBySequence(ctx, arg)
}

// timeout simulator
func simulateContextTimeout[T any](ctx context.Context) (T, error) {
	var zero T
	select {
	case <-time.After(20 * time.Second):
		log.Println("finished sleep timer")
	case <-ctx.Done():
		return zero, ctx.Err()
	}
	return zero, nil
}
