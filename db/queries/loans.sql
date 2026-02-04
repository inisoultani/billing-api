-- name: GetLoanByID :one
SELECT *
FROM loans
WHERE id = $1;
-- name: GetTotalPaidAmount :one
SELECT COALESCE(SUM(amount), 0) AS total_paid
FROM payments
WHERE loan_id = $1;
-- name: InsertRepayment :one
INSERT INTO payments (loan_id, week_number, amount, paid_at)
VALUES ($1, $2, $3, $4)
RETURNING *;
-- name: GetLastPaidWeek :one
SELECT COALESCE(MAX(week_number), 0) AS last_paid_week
FROM payments
WHERE loan_id = $1;