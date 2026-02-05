DROP TABLE IF EXISTS public.payments;
DROP TABLE IF EXISTS public.loans;
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
  paid_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE (loan_id, week_number)
);
CREATE INDEX idx_repayments_loan_id ON payments(loan_id);
CREATE INDEX idx_repayments_loan_week ON payments(loan_id, week_number);
CREATE INDEX idx_repayments_loan_paidat_id ON repayments (loan_id, paid_at, id);