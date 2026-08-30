# CoreBank — API Reference

Base URL: `http://localhost:8080/api`

No authentication headers are required or accepted anywhere. Every
response is wrapped as:

```json
{ "success": true, "data": { ... } }
```

or, on error:

```json
{ "success": false, "message": "Insufficient account balance", "timestamp": "2026-08-28T10:30:00Z" }
```

## Health

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Liveness check |

## Customers

| Method | Path | Description |
|---|---|---|
| GET | `/customers` | List all customers |
| GET | `/customers/{id}` | Get one customer |
| GET | `/customers/{id}/accounts` | Customer's accounts |
| GET | `/customers/{id}/transactions` | Customer's transactions (any account) |
| GET | `/customers/{id}/beneficiaries` | Customer's beneficiaries |
| GET | `/customers/{id}/cards` | Customer's cards |
| GET | `/customers/{id}/loans` | Customer's loans |

## Accounts

| Method | Path | Description |
|---|---|---|
| GET | `/accounts` | List all accounts |
| GET | `/accounts/{id}` | Get one account |
| GET | `/accounts/{id}/transactions` | Account's transactions |

## Transactions

| Method | Path | Body | Description |
|---|---|---|---|
| POST | `/transactions/deposit` | `{accountId, amount, description}` | Credit an account |
| POST | `/transactions/withdraw` | `{accountId, amount, description}` | Debit an account |
| POST | `/transactions/transfer` | `{fromAccountId, toAccountId, amount, description}` | Move money between two accounts. Optional `Idempotency-Key` header. |
| GET | `/transactions/{id}` | — | Get one transaction |

## Beneficiaries

| Method | Path | Body | Description |
|---|---|---|---|
| GET | `/beneficiaries/customer/{customerId}` | — | List a customer's beneficiaries |
| POST | `/beneficiaries` | `{customerId, name, accountNumber, ifscCode, nickname}` | Add a beneficiary |
| DELETE | `/beneficiaries/{id}?customerId={customerId}` | — | Remove a beneficiary |

## Cards

| Method | Path | Body | Description |
|---|---|---|---|
| GET | `/cards/customer/{customerId}` | — | List a customer's cards |
| POST | `/cards/{id}/freeze` | — | Freeze a card |
| POST | `/cards/{id}/unfreeze` | — | Unfreeze a card |
| POST | `/cards/{id}/pay` | `{merchant, amount}` | Simulate a card payment |

## Loans

| Method | Path | Body | Description |
|---|---|---|---|
| GET | `/loans/customer/{customerId}` | — | List a customer's loans |
| POST | `/loans/estimate` | `{principal, interestRate, tenureMonths}` | EMI calculator (no DB write) |
| POST | `/loans/apply` | `{customerId, accountId, principal, interestRate, tenureMonths}` | Apply for a simulated loan |

## Banker

| Method | Path | Description |
|---|---|---|
| GET | `/banker/dashboard` | Stats + chart data |
| GET | `/banker/customers` | All customers |
| GET | `/banker/customers/{id}` | One customer with accounts, cards, loans, transactions |
| GET | `/banker/accounts` | All accounts |
| GET | `/banker/transactions` | All transactions |
| GET | `/banker/suspicious-transactions` | Transactions flagged MEDIUM/HIGH risk |
| GET | `/banker/cards` | All cards |
| GET | `/banker/loans` | All loans |
| GET | `/banker/audit-logs` | Recent audit log entries |
| POST | `/banker/accounts/{id}/freeze` | Freeze an account (audited) |
| POST | `/banker/accounts/{id}/unfreeze` | Unfreeze an account (audited) |

## Error responses

| Status | Meaning |
|---|---|
| 400 | Bad request (invalid amount, malformed JSON, missing required field) |
| 404 | Customer / account / card / loan / beneficiary / transaction not found |
| 409 | Conflict — e.g. a raced idempotency key |
| 422 | Business rule violation (insufficient balance, frozen account, spending limit exceeded) |
| 500 | Internal error (never exposes raw database errors) |
