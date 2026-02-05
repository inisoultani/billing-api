package domain

import "time"

type Payment struct {
	ID         int64
	WeekNumber int
	Amount     int64
	PaidAt     time.Time
}
