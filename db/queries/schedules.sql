-- name: CreateLoanSchedules :copyfrom
INSERT INTO schedules (
    loan_id,
    sequence,
    due_date,
    amount,
    status
  )
VALUES ($1, $2, $3, $4, $5);
-- name: ListSchedulesByLoanIDWithCursor :many
SELECT *
FROM schedules
WHERE loan_id = $1
  AND sequence > $2
ORDER BY sequence
LIMIT $3;
-- name: UpdateSchedulePayment :one
UPDATE schedules
SET paid_amount = paid_amount + $2,
  status = CASE
    WHEN paid_amount + $2 >= amount THEN 'PAID'
    ELSE 'PARTIAL'
  END,
  updated_at = now()
WHERE id = $1
RETURNING *;
-- name: GetScheduleBySequence :one
SELECT *
FROM schedules
WHERE loan_id = $1
  AND sequence = $2
LIMIT 1;