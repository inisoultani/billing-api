package repository

import (
	"billing-api/internal/domain"
	"billing-api/internal/infra/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapCreateLoanCommand(clc domain.CreateLoanCommand) *sqlc.InsertLoanParams {
	return &sqlc.InsertLoanParams{
		PrincipalAmount:     clc.PrincipalAmount,
		TotalInterestAmount: clc.TotalInterestAmount,
		TotalPayableAmount:  clc.TotalPayableAmount,
		WeeklyPaymentAmount: clc.WeeklyPaymentAmount,
		TotalWeeks:          clc.TotalWeeks,
		StartDate: pgtype.Date{
			Time:  clc.StartDate,
			Valid: true,
		},
	}
}

func MapLoan(l sqlc.Loan) *domain.Loan {
	return &domain.Loan{
		ID:                  l.ID,
		PrincipalAmount:     l.PrincipalAmount,
		TotalPayableAmount:  l.TotalPayableAmount,
		WeeklyPaymentAmount: l.WeeklyPaymentAmount,
		TotalWeeks:          int(l.TotalWeeks),
		CreatedAt:           l.CreatedAt.Time,
	}
}
