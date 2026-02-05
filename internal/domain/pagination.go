package domain

import "time"

type PaymentCursor struct {
	PaidAt time.Time
	ID     int64
}
