package repository

import (
	"billing-api/internal/domain"
	"billing-api/internal/infra/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

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

func MapCreateLoanCommand(clc *domain.CreateLoanCommand) *sqlc.InsertLoanParams {
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

func MapPayment(p sqlc.Payment) *domain.Payment {
	return &domain.Payment{
		ID:         p.ID,
		WeekNumber: int(p.WeekNumber),
		Amount:     p.Amount,
		PaidAt:     p.PaidAt.Time,
	}
}

func MapListPaymentsByLoanIDRow(p sqlc.ListPaymentsByLoanIDRow) *domain.Payment {
	return &domain.Payment{
		ID:         p.ID,
		WeekNumber: int(p.WeekNumber),
		Amount:     p.Amount,
		PaidAt:     p.PaidAt.Time,
	}
}

func MapCreatePaymentComand(cpc *domain.CreatePaymentComand) *sqlc.InsertPaymentParams {
	return &sqlc.InsertPaymentParams{
		LoanID:         cpc.LoanID,
		WeekNumber:     cpc.WeekNumber,
		Amount:         cpc.Amount,
		IdempotencyKey: cpc.IdempotencyKey,
		PaidAt: pgtype.Timestamp{
			Time:  cpc.PaidAt,
			Valid: true,
		},
	}
}
