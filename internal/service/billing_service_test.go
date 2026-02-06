package service

import (
	"billing-api/internal/domain"
	"billing-api/internal/infra/db/sqlc"
	"billing-api/internal/mocks"
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsDelinquent_Mock(t *testing.T) {
	// intialize mock repository
	mockRepo := new(mocks.MockBillingRepository)

	// provide nil since we are in mock mode
	svc := NewBillingService(nil, mockRepo)
	ctx := context.Background()
	now := time.Now()

	t.Run("should be delinquent whent gap is 2 weeks or more", func(t *testing.T) {
		loanID := int64(1)

		// mock GetLoanByID
		// setup loan 4 weeks ago
		fourWeeksAgoDate := now.AddDate(0, 0, -28)
		mockRepo.On("GetLoanByID", ctx, loanID).Return(sqlc.Loan{
			ID: loanID,
			CreatedAt: pgtype.Timestamp{
				Time:  fourWeeksAgoDate,
				Valid: true,
			},
		}, nil).Once()

		// mock GetLastPaidWeek, paid only 1 week
		// expected week 4 - paid 1 = 3 means delinquent
		mockRepo.On("GetLastPaidWeek", ctx, loanID).Return(int32(1), nil).Once()

		// execute
		isDelinquent, err := svc.IsDelinquent(ctx, loanID, now)

		assert.NoError(t, err)
		assert.True(t, isDelinquent)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should NOT be delinquent when paid up to date", func(t *testing.T) {
		loanID := int64(2)
		twoWeeksAgo := now.AddDate(0, 0, -14)

		mockRepo.On("GetLoanByID", ctx, loanID).Return(sqlc.Loan{
			ID: loanID,
			CreatedAt: pgtype.Timestamp{
				Time:  twoWeeksAgo,
				Valid: true,
			},
		}, nil).Once()

		// expected week 2 - paid 2 = 0 not delinquent
		mockRepo.On("GetLastPaidWeek", ctx, loanID).Return(int32(2), nil).Once()

		isDelinquent, err := svc.IsDelinquent(ctx, loanID, now)

		assert.NoError(t, err)
		assert.False(t, isDelinquent)
		mockRepo.AssertExpectations(t)
	})

}

func TestGetOutstanding_Unit(t *testing.T) {
	// intialize mock repository
	mockRepo := new(mocks.MockBillingRepository)

	// provide nil since we are in mock mode
	svc := NewBillingService(nil, mockRepo)
	ctx := context.Background()

	loanID := int64(1)

	// simulate a loan with 5,000,000 total and 1,000,000 already paid
	mockRepo.On("GetLoanByID", ctx, loanID).Return(sqlc.Loan{
		ID:                 loanID,
		TotalPayableAmount: 5000000,
	}, nil)
	mockRepo.On("GetTotalPaidAmount", ctx, loanID).Return(int64(1000000), nil)

	outstanding, err := svc.GetOutstanding(ctx, loanID)

	assert.NoError(t, err)
	assert.Equal(t, int64(4000000), outstanding)
	mockRepo.AssertExpectations(t)
}

func TestSubmitPayment_Mock(t *testing.T) {
	mockRepo := new(mocks.MockBillingRepository)
	// We still need a pool to satisfy the struct, but we won't call the real DB
	svc := NewBillingService(nil, mockRepo)
	ctx := context.Background()

	t.Run("successful payment", func(t *testing.T) {
		input := SubmitPaymentInput{
			LoanID: 1,
			Amount: 110000,
			PaidAt: time.Now(),
		}

		mockRepo.On("WithTx", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(domain.BillingRepository) error)
				_ = fn(mockRepo) // Execute the inner logic using the mockRepo
				// assert.NoError(t, err)
			}).Return(nil)

		// mock GetLoanByID (to check if loan exists and get weekly amount)
		mockRepo.On("GetLoanByID", mock.Anything, input.LoanID).Return(sqlc.Loan{
			ID:                  1,
			WeeklyPaymentAmount: 110000,
			TotalPayableAmount:  5500000,
			TotalWeeks:          5,
		}, nil).Once()

		mockRepo.On("GetPaidWeeksCount", mock.Anything, input.LoanID).Return(int32(0), nil).Once()

		mockRepo.On("GetTotalPaidAmount", mock.Anything, input.LoanID).Return(int64(0), nil).Once()

		expectedInsert := sqlc.InsertPaymentParams{
			LoanID:     input.LoanID,
			WeekNumber: 1,
			Amount:     input.Amount,
			PaidAt: pgtype.Timestamp{
				Time:  input.PaidAt,
				Valid: true,
			},
		}
		mockRepo.On("InsertPayment", mock.Anything, expectedInsert).Return(sqlc.Payment{
			ID: 999,
		}, nil).Once()

		id, err := svc.SubmitPayment(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, int64(999), id)
		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})

	t.Run("fails when amount is incorrect", func(t *testing.T) {
		input := SubmitPaymentInput{
			LoanID: 1,
			Amount: 6000, // Incorrect: logic expects 110000
		}

		mockRepo.
			On("WithTx", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(domain.BillingRepository) error)
				_ = fn(mockRepo)
			}).Return(ErrInvalidPayment)

		mockRepo.On("GetTotalPaidAmount", mock.Anything, input.LoanID).Return(int64(0), nil).Once()

		mockRepo.On("GetLoanByID", mock.Anything, input.LoanID).Return(sqlc.Loan{
			ID:                  1,
			WeeklyPaymentAmount: 110000, // The required amount
		}, nil).Once()

		// If the code is correct, it should return early and never call it.
		id, errF := svc.SubmitPayment(ctx, input)

		assert.Error(t, errF)
		assert.Contains(t, errF.Error(), "Invalid payment")
		assert.Equal(t, int64(0), id)

		// This confirms InsertPayment was never called
		mockRepo.AssertExpectations(t)
	})
}