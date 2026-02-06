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

}
