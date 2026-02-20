package domain

import (
	"billing-api/internal/infra/db/sqlc"
	"time"
)

type LoanSchedule struct {
	ID         int64
	LoanID     int64
	Sequence   int
	DueDate    time.Time
	Amount     int64
	PaidAmount int64
	Status     string
}

type CreateLoanScheduleCommand struct {
	TotalWeek int
	StartDate time.Time
	LoanID    int64
	Amount    int64
}

func MapSchedule(s sqlc.Schedule) *LoanSchedule {
	return &LoanSchedule{
		ID:         s.ID,
		LoanID:     s.LoanID,
		Sequence:   int(s.Sequence),
		DueDate:    s.DueDate.Time,
		Amount:     s.Amount,
		PaidAmount: s.PaidAmount,
		Status:     s.Status,
	}
}
