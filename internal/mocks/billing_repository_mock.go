package mocks

import (
	"billing-api/internal/infra/db/sqlc"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockBillingRepository is a testify mock that satisfies the BillingRepository interface.
type MockBillingRepository struct {
	mock.Mock
}

// GetLoanByID mocks the retrieval of a single loan.
func (m *MockBillingRepository) GetLoanByID(ctx context.Context, id int64) (sqlc.Loan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Loan), args.Error(1)
}

// GetTotalPaidAmount mocks the calculation of total amount paid for a loan.
func (m *MockBillingRepository) GetTotalPaidAmount(ctx context.Context, loanID int64) (int64, error) {
	args := m.Called(ctx, loanID)
	return args.Get(0).(int64), args.Error(1)
}

// GetPaidWeeksCount mocks the count of successful payment weeks.
func (m *MockBillingRepository) GetPaidWeeksCount(ctx context.Context, loanID int64) (int32, error) {
	args := m.Called(ctx, loanID)
	return args.Get(0).(int32), args.Error(1)
}

// GetLastPaidWeek mocks the retrieval of the highest paid week number.
func (m *MockBillingRepository) GetLastPaidWeek(ctx context.Context, loanID int64) (int32, error) {
	args := m.Called(ctx, loanID)
	return args.Get(0).(int32), args.Error(1)
}

// InsertLoan mocks the creation of a new loan.
func (m *MockBillingRepository) InsertLoan(ctx context.Context, arg sqlc.InsertLoanParams) (sqlc.Loan, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Loan), args.Error(1)
}

// InsertPayment mocks the creation of a payment record.
func (m *MockBillingRepository) InsertPayment(ctx context.Context, arg sqlc.InsertPaymentParams) (sqlc.Payment, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Payment), args.Error(1)
}

// ListPaymentsByLoanID mocks the paginated retrieval of payments.
func (m *MockBillingRepository) ListPaymentsByLoanID(ctx context.Context, arg sqlc.ListPaymentsByLoanIDParams) ([]sqlc.ListPaymentsByLoanIDRow, error) {
	args := m.Called(ctx, arg)
	// We use a type assertion for the slice. If nil is returned, we handle it.
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.ListPaymentsByLoanIDRow), args.Error(1)
}
