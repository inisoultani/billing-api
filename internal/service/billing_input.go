package service

import "time"

type SubmitLoanInput struct {
	PrincipalAmount    int64
	AnnualInterestRate float64 // e.g. 0.10
	TotalWeeks         int
	StartDate          time.Time
}

type SubmitPaymentInput struct {
	LoanID int64
	Amount int64
	PaidAt time.Time
}
