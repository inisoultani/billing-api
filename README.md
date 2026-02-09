# Billing API – Local Setup & Usage Guide

A high-performance Go service for managing loan payments, featuring type-safe SQL with `sqlc` and efficient cursor-based pagination.

---

## Requirements

- **Go**: `1.24.12`
- **Docker Engine**: For running PostgreSQL 16
- **sqlc**: For generating type-safe database code
- **godotenv**: For environment variable management

---

## Installation & Setup

### Install Tooling

Ensure the `sqlc` compiler is installed:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Verify installation:

```bash
sqlc version
```

---

## 1. Configure Environment

Create a `.env` file in the project root. This file **must be ignored by git** for security reasons.

```bash
# Database connection
DATABASE_URL=postgres://billing:billing@localhost:5432/billing?sslmode=disable
DB_MAX_CONNS=30
DB_MIN_CONNS=5
DB_MAX_IDLE_TIME=200 # in seconds
DB_MAX_LIFE_TIME=900 # in seconds
DB_HEALTH_CHECK_PERIOD=30 # in seconds

# Application settings
SERVER_PORT=8081
PAGING_LIMIT_DEFAULT=10
PAGING_LIMIT_MAX=100
APP_ENV=development
```

---

## 2. Database & Code Generation

Start the PostgreSQL container. This will automatically execute the initialization scripts located in `db/migrations`.

```bash
# Start PostgreSQL using Docker
docker-compose up -d
```

Generate Go models and query code from SQL using `sqlc`:

```bash
sqlc generate
```

---

## 3. Running the Application

To start the API server locally:

```bash
go run ./cmd/billing-api/main.go
```

The server will be available at:

```
http://localhost:8081
```

---

## 4. Project Structure

The project follows a clean separation of concerns and is organized as follows:

```
.
├── cmd/
│   └── billing-api/
│       └── main.go          # Application entry point
├── db/
│   ├── migrations/          # Database schema & migrations
│   ├── queries/             # SQL queries used by sqlc
│   └── sqlc.yaml            # sqlc configuration
├── internal/
│   ├── config/              # Environment & application configuration
│   ├── domain/              # Core business entities (Loan, Payment, Cursor)
│   ├── http/                # HTTP layer (handlers, router, DTOs)
│   │   ├── handler.go       # HTTP handlers
│   │   ├── request.go       # Request parsing & cursor decoding
│   │   ├── response.go      # Response DTOs & cursor encoding
│   │   └── router.go        # HTTP route definitions
│   ├── infra/
│   │   └── db/
│   │       ├── sqlc/         # Generated sqlc code
│   │       └── postgres.go  # PostgreSQL connection setup
│   └── service/             # Business logic layer
│       ├── billing_service.go
│       └── tx.go            # Transaction helper
├── .env                     # Local environment variables (gitignored)
├── docker-compose.yml       # Local PostgreSQL setup
├── go.mod
├── go.sum
└── README.md
```

### Layering Overview

- **cmd/**
  Application bootstrap and wiring.

- **internal/config**
  Configuration loading (env vars, defaults).

- **internal/domain**
  Pure domain models with no infrastructure dependencies.

- **internal/http**
  Transport layer: HTTP handlers, routing, request/response mapping.

- **internal/infra**
  Infrastructure concerns such as database connections and generated sqlc code.

- **internal/service**
  Core business logic, transactions, and orchestration.

---

## 5. Billing API Documentation

This API manages loan lifecycles, including creation, automated payment scheduling, payment processing, and real-time delinquency tracking.

## Base URL

`http://<host>:<port>/loan`

---

## 📋 Endpoints Summary

| Method   | Endpoint                | Description                                   |
| -------- | ----------------------- | --------------------------------------------- |
| **POST** | `/`                     | Create a new loan and generate schedules.     |
| **GET**  | `/{loanID}`             | Retrieve loan details and delinquency status. |
| **GET**  | `/{loanID}/outstanding` | Get the remaining balance to be paid.         |
| **GET**  | `/{loanID}/schedule`    | List repayment schedules (paginated).         |
| **POST** | `/{loanID}/payment`     | Submit a weekly payment.                      |
| **GET**  | `/{loanID}/payment`     | List payment history (paginated).             |

---

## Endpoint Details

### 1. Submit Loan

**POST** `/`

Creates a new loan record and generates all weekly schedules for the duration of the loan.

- **Request Body**:

```json
{
  "principal_amount": 5000000,
  "annual_interest_rate": 0.1,
  "total_weeks": 50,
  "start_date": "2026-02-07"
}
```

- **Success Response (201 Created)**:

```json
{
  "loan_id": 123,
  "weekly_payment_amount": 110000,
  "total_payable": 5500000
}
```

### 2. Get Loan Details

**GET** `/{loanID}`

Returns core loan data, including the on-the-fly calculated delinquency status.

- **Success Response (200 OK)**:

```json
{
  "loan_id": 123,
  "weekly_payment_amount": 110000,
  "total_payable": 5500000,
  "total_weeks": 50,
  "created_at": "2026-02-07T10:00:00Z",
  "is_delinquent": false
}
```

### 3. Get Outstanding Balance

**GET** `/{loanID}/outstanding`

Returns the total remaining amount required to fully pay off the loan.

- **Success Response (200 OK)**:

```json
{
  "loan_id": 123,
  "outstanding": 4400000
}
```

### 4. Make Payment

**POST** `/{loanID}/payment`

Processes a weekly payment for the next sequential unpaid week. The system enforces strict validation and idempotency to ensure the payment matches the expected schedule and prevents duplicate transactions.

#### **Headers**

| Name                  | Type     | Required | Description                                                     |
| --------------------- | -------- | -------- | --------------------------------------------------------------- |
| **X-Idempotency-Key** | `string` | **Yes**  | A unique identifier (e.g., UUID) to prevent duplicate payments. |
| **Content-Type**      | `string` | **Yes**  | Must be `application/json`.                                     |

#### **Request Body**

```json
{
  "amount": 110000
}
```

- **amount**: Must exactly match the `weekly_payment_amount` defined during loan creation.

#### **Success Responses**

**1. New Payment Processed (201 Created)**
Returned when a payment is successfully recorded for the first time.

```json
{
  "payment_id": 987
}
```

**2. Duplicate Request (200 OK)**
Returned when the provided `X-Idempotency-Key` has already been processed for this loan.

```json
{
  "message": "payment already processed",
  "status": "success"
}
```

---

### Business Logic Summary

- **Sequential Processing**: Payments apply strictly to the next unpaid week.
- **Exact Amount**: Only the exact weekly amount is accepted.

### 5. List payment Schedules

**GET** `/{loanID}/schedule?limit=10&cursor=...`

Retrieves the generated weekly schedules using sequence-based pagination.

- **Query Params**: `limit` (int), `cursor` (encoded sequence string).

### 6. List Payments

**GET** `/{loanID}/payment?limit=10&cursor=...`

Retrieves the history of payments made for this loan using cursor-based pagination.

---

## Core Business Logic

### Loan Terms

- **Interest Model**: The system uses a flat interest rate applied once to the full principal.
- **Divisibility Constraint**: The total payable amount must be evenly divisible by the total number of weeks to ensure consistent weekly payments.

### Delinquency Criteria

- **Derived State**: Delinquency is calculated on demand rather than stored.
- **Threshold**: A loan is considered delinquent if there is a gap of **2 or more weeks** between the last paid week and the current expected week (based on the loan start date).

### Payment Validation

- **Sequential Payment**: Payments must apply to the next unpaid week in order.
- **Exact Amount**: Every payment must exactly match the `weekly_payment_amount`.
- **Closure**: Payments are rejected once all weeks in the schedule are paid.

---

## Error Responses

| Code    | Meaning        | Cause                                                                  |
| ------- | -------------- | ---------------------------------------------------------------------- |
| **400** | Bad Request    | Invalid input format, invalid loan terms, or incorrect payment amount. |
| **404** | Not Found      | The specified loan ID does not exist.                                  |
| **409** | Conflict       | Attempting to pay for a loan that is already closed/fully paid.        |
| **500** | Internal Error | Database failure or internal processing error.                         |

---

## 6. Troubleshooting

### Database schema changes not reflected

If changes to files like `001_init.sql` are not applied, the Docker volume may be stale.

Reset the database volume:

```bash
docker-compose down -v
docker-compose up -d
```

Then regenerate sqlc code if needed:

```bash
sqlc generate
```

---

## 7. Unit Test

# Running Unit Tests

Run unit tests only for the internal/service package:

```bash
go test ./internal/service
```

## 8. Resilience & Observability

The system implements sampling **Smart Context Timeouts** to protect the database connection pool and simplify debugging:

- **Dynamic Deadlines:** Timeouts scale with workload (e.g., **2s** for single rows, up to **10s** for batch inserts).
- **Precise Attribution:** Uses `context.Cause` to label specific operations in the logs.
- **Fail-Fast:** Aborts database queries immediately upon timeout or user cancellation.
- **Standardized Responses:** Maps repository failures to **HTTP 504 Gateway Timeout**.

**Log Sample:**

```bash
# pinpointing exactly which query exceeded its allocated limit
[TIMEOUT] POST /payments: repo_timeout: CreatePayment (limit 2s)
[TIMEOUT] POST /loans: repo_timeout: BatchSchedules (limit 7.2s)

```
