
[![tests](https://github.com/ildomm/account-balance-manager/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/ildomm/account-balance-manager/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ildomm/account-balance-manager?cache=v1)](https://goreportcard.com/report/github.com/ildomm/account-balance-manager)
![coverage](https://raw.githubusercontent.com/ildomm/account-balance-manager/badges/.badges/main/coverage.svg)

# Account Balance Manager
User Account Balance Management System

## Main Features
- Record user transactions for balance updates.
- Retrieve user account balance.

## Architecture
The application consists of 2 main components:

### 1. API Handler
The API Handler manages HTTP requests for retrieving balances and processing user transactions.

#### API Endpoints
- `POST /user/{userId}/transactions` - Processes a new transaction for a user.
- `GET /user/{userId}/balance` - Retrieves the current balance for a specific user.

### 2. Database
All account balances and transactions are persisted in a PostgreSQL database to ensure consistency and reliability.

#### Database Schema
```mermaid
---
title: Database Schema
---
erDiagram
   users ||--o{ transactions : "One-to-Many"
   users {
      uint64 userId
      float balance
   }
   transactions {
      string transactionId
      uint64 userId
      float amount
      string state
      string source
      datetime createdAt
   }
```

## Build Process
Run the following command to build the application:
```bash
make build
```
Binaries will be available in the `build` directory.

### Entry Points
- API Handler: `cmd/api/main.go`

## Development

### Environment Variables
The following environment variables are required to configure the application:
- `DATABASE_URL` - The Postgres database connection string, e.g., `postgres://user:password@host:port/database`.

## Deployment

### Using Docker Compose
Run the following command to deploy the application locally using Docker Compose:
```bash
docker-compose up -d
```

### Manual Testing
1. Pre-populated user IDs in the database:
    - `1`, `2`, `3`, `4`, `5`, `6`, `7`, `8`, `9`.
2. Run the application and use `curl` to interact with the API. For example:
    - Process a transaction for a user:
      ```bash
      curl -X POST http://localhost:8080/user/1/transactions -H 'Content-Type: application/json' -H "Source-Type: game" -d "{"state": "win", "amount": "50.00", "transactionId": "abc123"}" 
      ```
   - Retrieve a user's balance:
     ```bash
     curl -X GET http://localhost:8080/user/1/balance
     ```
## Testing

### Local Tests
#### Unit Tests
Execute the following command to run unit tests:
```bash
make unit-test
```

#### Test Coverage
To generate and view test coverage reports:
- HTML Report:
  ```bash
  make coverage-report
  ```
- Total Coverage:
  ```bash
  make coverage-total
  ```

## Future Improvements
- WIP

## Missing Features
- WIP

## Considerations
- WIP
