package mocks

import (
	"billing-api/internal/domain"
	"billing-api/internal/infra/db/sqlc"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockBillingRepository is a testify mock that satisfies the BillingRepository interface.
type MockBillingRepository struct {
	mock.Mock
}

// WithTx is the "magic" part for unit testing SubmitPayment
func (m *MockBillingRepository) WithTx(ctx context.Context, fn func(repo domain.BillingRepository) error) error {
	args := m.Called(ctx, fn)

	// In the test, we will use .Run() to actually execute 'fn'
	// If the mock was told to return an error, we return it here
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

// GetLoanByID mocks the retrieval of a single loan.
func (m *MockBillingRepository) GetLoanByID(ctx context.Context, id int64) (*domain.Loan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Loan), args.Error(1)
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
func (m *MockBillingRepository) InsertLoan(ctx context.Context, arg domain.CreateLoanCommand) (*domain.Loan, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(*domain.Loan), args.Error(1)
}

// InsertPayment mocks the creation of a payment record.
func (m *MockBillingRepository) InsertPayment(ctx context.Context, arg domain.CreatePaymentComand) (*domain.Payment, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(*domain.Payment), args.Error(1)
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

// CreateLoanSchedule mocks the creation of schedules
func (m *MockBillingRepository) CreateLoanSchedules(ctx context.Context, arg []sqlc.CreateLoanSchedulesParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

// ListSchedulesByLoanID mocks the paginated retrieval of schedules
func (m *MockBillingRepository) ListSchedulesByLoanID(ctx context.Context, arg sqlc.ListSchedulesByLoanIDWithCursorParams) ([]sqlc.Schedule, error) {
	args := m.Called(ctx, arg)
	// We use a type assertion for the slice. If nil is returned, we handle it.
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.Schedule), args.Error(1)
}

// UpdateSchedulePayment mocks the schedule based on payment sequence
func (m *MockBillingRepository) UpdateSchedulePayment(ctx context.Context, arg sqlc.UpdateSchedulePaymentParams) (sqlc.Schedule, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Schedule), args.Error(1)
}

// GetScheduleBySequence mocks the retrieval schedule based on sequence
func (m *MockBillingRepository) GetScheduleBySequence(ctx context.Context, arg sqlc.GetScheduleBySequenceParams) (sqlc.Schedule, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Schedule), args.Error(1)
}
