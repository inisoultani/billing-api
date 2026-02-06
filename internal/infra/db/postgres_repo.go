package db

import (
	"billing-api/internal/infra/db/sqlc"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	queries *sqlc.Queries
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		queries: sqlc.New(pool),
	}
}

// LOAN RELATED
// GetLoanByID retrieves a loan by its primary key
func (r *PostgresRepo) GetLoanByID(ctx context.Context, id int64) (sqlc.Loan, error) {
	return r.queries.GetLoanByID(ctx, id)
}

// InsertLoan creates a new loan record
func (r *PostgresRepo) InsertLoan(ctx context.Context, arg sqlc.InsertLoanParams) (sqlc.Loan, error) {
	return r.queries.InsertLoan(ctx, arg)
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
func (r *PostgresRepo) InsertPayment(ctx context.Context, arg sqlc.InsertPaymentParams) (sqlc.Payment, error) {
	return r.queries.InsertPayment(ctx, arg)
}

// ListPaymentsByLoanID handles paginated retrieval of payments
func (r *PostgresRepo) ListPaymentsByLoanID(ctx context.Context, arg sqlc.ListPaymentsByLoanIDParams) ([]sqlc.ListPaymentsByLoanIDRow, error) {
	return r.queries.ListPaymentsByLoanID(ctx, arg)
}
