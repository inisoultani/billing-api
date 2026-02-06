package service

import (
	"billing-api/internal/infra/db/sqlc"
	"context"
)

type billingQueries interface {
	GetLoanByID(ctx context.Context, loanID int64) (sqlc.Loan, error)
	GetTotalPaidAmount(ctx context.Context, loanID int64) (int64, error)
	GetPaidWeeksCount(ctx context.Context, loanID int64) (int32, error)
	GetLastPaidWeek(ctx context.Context, loanID int64) (int32, error)
	InsertPayment(ctx context.Context, arg sqlc.InsertPaymentParams) (sqlc.Payment, error)
}
