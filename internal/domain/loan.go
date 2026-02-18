package domain

import (
	"time"
)

type Loan struct {
	ID                  int64
	PrincipalAmount     int64
	TotalPayableAmount  int64
	WeeklyPaymentAmount int64
	TotalWeeks          int
	CreatedAt           time.Time
}

type CreateLoanCommand struct {
	PrincipalAmount     int64
	TotalInterestAmount int64
	TotalPayableAmount  int64
	WeeklyPaymentAmount int64
	TotalWeeks          int32
	StartDate           time.Time
}
