DROP TABLE IF EXISTS public.payments;
DROP TABLE IF EXISTS public.loans;
DROP TABLE IF EXISTS public.schedules;
CREATE TABLE loans (
  id BIGSERIAL PRIMARY KEY,
  principal_amount BIGINT NOT NULL,
  total_interest_amount BIGINT NOT NULL,
  total_payable_amount BIGINT NOT NULL,
  weekly_payment_amount BIGINT NOT NULL,
  total_weeks INT NOT NULL,
  start_date DATE NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE TABLE payments (
  id BIGSERIAL PRIMARY KEY,
  loan_id BIGINT NOT NULL REFERENCES loans(id),
  week_number INT NOT NULL,
  amount BIGINT NOT NULL,
  idempotency_key VARCHAR(255) NOT NULL,
  paid_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT uk_payments_loan_id_and_week_number UNIQUE (loan_id, week_number),
  CONSTRAINT uk_payments_idempotency_key UNIQUE (idempotency_key)
);
CREATE INDEX idx_repayments_loan_id ON payments(loan_id);
CREATE INDEX idx_repayments_loan_week ON payments(loan_id, week_number);
CREATE INDEX idx_repayments_loan_paidat_id ON payments (loan_id, paid_at, id);
CREATE TABLE schedules (
  id BIGSERIAL PRIMARY KEY,
  loan_id BIGINT NOT NULL REFERENCES loans(id),
  sequence INT NOT NULL,
  -- W1, W2, W3...
  due_date DATE NOT NULL,
  amount BIGINT NOT NULL,
  paid_amount BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  -- PENDING | PARTIAL | PAID
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE (loan_id, sequence)
);
CREATE INDEX idx_schedules_loan_sequence ON schedules (loan_id, sequence);