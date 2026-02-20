package domain

import (
	"context"
)

type BillingRepository interface {

	// transcaction
	WithTx(ctx context.Context, fn func(repo BillingRepository) error) error

	// Loan-related actions
	GetLoanByID(ctx context.Context, id int64) (*Loan, error)
	InsertLoan(ctx context.Context, arg CreateLoanCommand) (*Loan, error)

	// Payment-related actions
	GetTotalPaidAmount(ctx context.Context, loanID int64) (int64, error)
	GetPaidWeeksCount(ctx context.Context, loanID int64) (int32, error)
	GetLastPaidWeek(ctx context.Context, loanID int64) (int32, error)
	InsertPayment(ctx context.Context, arg CreatePaymentComand) (*Payment, error)
	ListPaymentsByLoanID(ctx context.Context, arg ListPaymentsQuery) ([]Payment, error)

	// Schedule-related actions
	CreateLoanSchedules(ctx context.Context, arg []LoanSchedule) (int64, error)
	ListSchedulesByLoanID(ctx context.Context, arg ListScheduleQuery) ([]LoanSchedule, error)
	UpdateSchedulePayment(ctx context.Context, arg UpdateLoanSchedulePaymentCommand) (int64, error)
	GetScheduleBySequence(ctx context.Context, arg GetLoanScheduleBySequenceQuery) (LoanSchedule, error)
}
