# Billing API – Local Setup & Usage Guide

A high-performance Go service for managing loan repayments, featuring type-safe SQL with `sqlc` and efficient cursor-based pagination.

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

## 5. API Usage

### Get Loan Details

Returns detailed information for a specific loan, including derived state.

**Endpoint**

```
GET /loan/:id
```

**Path Parameters**

| Name | Description                   |
| ---- | ----------------------------- |
| id   | Unique identifier of the loan |

---

### Example Request

```bash
curl "http://localhost:8081/loan/1"
```

---

### Success Response

```json
{
  "loan_id": 1,
  "total_payable": 5500000,
  "weekly_payment_amount": 1100000,
  "total_weeks": 5,
  "created_at": "2026-01-05T10:31:33Z",
  "is_delinquent": true
}
```

Notes

- `is_delinquent` is a derived field computed at read time
- The value is not persisted in the database
- See `ADR-002: Derived states are computed dynamically`

---

### List Payments (Cursor-Based Pagination)

Returns a paginated list of payments for a specific loan.

**Endpoint**

```
GET /loan/:id/payment
```

**Query Parameters**

| Name   | Description                                 |
| ------ | ------------------------------------------- |
| limit  | Number of records to return (default: 10)   |
| cursor | Base64URL-encoded cursor from previous page |

---

### Example Request

```bash
curl "http://localhost:8081/loan/3/payment?limit=2"
```

---

### Success Response

```json
{
  "payments": [
    {
      "week_number": 1,
      "amount": 1100000,
      "paid_at": "2026-02-05T17:31:40Z"
    },
    {
      "week_number": 2,
      "amount": 1100000,
      "paid_at": "2026-02-05T17:31:42Z"
    }
  ],
  "next_cursor": "eyJQYWlkQXQiOiIyMDI2LTAyLTA1VDE3OjMxOjQyLjAxNDQzMloiLCJJRCI6Mn0"
}
```

- `payments` are ordered deterministically by payment time.
- `next_cursor` should be passed to the next request to retrieve the next page.
- If `next_cursor` is omitted, there are no more results.

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

## Notes

- Derived states (e.g. delinquency) are computed dynamically and not persisted.
- Pagination is **cursor-based**, not offset-based, for correctness and performance.
