package service

import (
	"billing-api/internal/infra/db/sqlc"
	"billing-api/internal/mocks"
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
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
