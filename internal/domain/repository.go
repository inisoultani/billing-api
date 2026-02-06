package domain

import (
	"billing-api/internal/infra/db/sqlc"
	"context"
)

type BillingRepository interface {

	// transcaction
	WithTx(ctx context.Context, fn func(repo BillingRepository) error) error

	// Loan-related actions
	GetLoanByID(ctx context.Context, id int64) (sqlc.Loan, error)
	InsertLoan(ctx context.Context, arg sqlc.InsertLoanParams) (sqlc.Loan, error)

	// Payment-related actions
	GetTotalPaidAmount(ctx context.Context, loanID int64) (int64, error)
	GetPaidWeeksCount(ctx context.Context, loanID int64) (int32, error)
	GetLastPaidWeek(ctx context.Context, loanID int64) (int32, error)
	InsertPayment(ctx context.Context, arg sqlc.InsertPaymentParams) (sqlc.Payment, error)
	ListPaymentsByLoanID(ctx context.Context, arg sqlc.ListPaymentsByLoanIDParams) ([]sqlc.ListPaymentsByLoanIDRow, error)

	// Schedule-related actions
	CreateLoanSchedule(ctx context.Context, arg sqlc.CreateLoanScheduleParams) (sqlc.Schedule, error)
	ListSchedulesByLoanID(ctx context.Context, arg sqlc.ListSchedulesByLoanIDWithCursorParams) ([]sqlc.Schedule, error)
	UpdateSchedulePayment(ctx context.Context, arg sqlc.UpdateSchedulePaymentParams) (sqlc.Schedule, error)
	GetScheduleBySequence(ctx context.Context, arg sqlc.GetScheduleBySequenceParams) (sqlc.Schedule, error)
}
