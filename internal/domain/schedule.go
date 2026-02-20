package domain

import (
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

type ListScheduleQuery struct {
	LoanID         int64
	Limit          int32
	CursorSequence int32
}

type CreateLoanScheduleCommand struct {
	TotalWeek int
	StartDate time.Time
	LoanID    int64
	Amount    int64
}

type UpdateLoanSchedulePaymentCommand struct {
	ID         int64
	PaidAmount int64
}

type GetLoanScheduleBySequenceQuery struct {
	LoanID   int64
	Sequence int32
}
