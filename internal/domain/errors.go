package domain

import "errors"

var (
	ErrLoanNotFound            = errors.New("Loan not found")
	ErrInvalidStateOutstanding = errors.New("Invalid loan payment state")
	ErrInvalidLoanTerms        = errors.New("Invalid loan terms")
	ErrInvalidPayment          = errors.New("Invalid payment")
	ErrLoanAlreadyClosed       = errors.New("Loan already fully paid")
	ErrDuplicatePayment        = errors.New("Duplicate payment for current week")
	ErrDelinquencyCheck        = errors.New("Failed to compute loan delinquency")
	ErrScheduleNotFound        = errors.New("Schedule not found")
)
