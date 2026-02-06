package domain

import "time"

type PaymentCursor struct {
	PaidAt time.Time
	ID     int64
}

type ScheduleCursor struct {
	Sequence int32 `json:"seq"`
}
