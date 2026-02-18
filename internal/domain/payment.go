package domain

import (
	"time"
)

type Payment struct {
	ID         int64
	WeekNumber int
	Amount     int64
	PaidAt     time.Time
}

type CreatePaymentComand struct {
	LoanID         int64
	WeekNumber     int32
	Amount         int64
	IdempotencyKey string
	PaidAt         time.Time
}
