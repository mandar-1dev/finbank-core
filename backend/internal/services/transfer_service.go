package services

import (
	"context"
	"database/sql"
	"errors"

	"corebank/backend/internal/models"
	"corebank/backend/internal/util"
)

type TransferService struct {
	db    *sql.DB
	audit *AuditLogService
}

func NewTransferService(db *sql.DB, audit *AuditLogService) *TransferService {
	return &TransferService{db: db, audit: audit}
}

// suspicious transaction thresholds — a deliberately simple, documented
// rule set, not a real fraud model.
const (
	highRiskAmountThreshold   = 200000.0 // absolute amount considered high risk
	mediumRiskAmountThreshold = 75000.0  // absolute amount considered medium risk
	mediumRiskBalanceFraction = 0.75     // % of sender's pre-transfer balance
)

func classifyRisk(amount, senderBalanceBefore float64) string {
	if amount >= highRiskAmountThreshold {
		return "HIGH"
	}
	if amount >= mediumRiskAmountThreshold {
		return "MEDIUM"
	}
	if senderBalanceBefore > 0 && amount/senderBalanceBefore >= mediumRiskBalanceFraction {
		return "MEDIUM"
	}
	return "LOW"
}

// Transfer atomically moves money from one account to another.
//
// Concurrency: both accounts are locked with SELECT ... FOR UPDATE inside
// a single database transaction, always acquired in ascending account-id
// order. Two transfers racing on the same pair of accounts (in either
// direction) therefore always take the locks in the same order, which
// makes a deadlock between them impossible — the second request simply
// waits for the first to commit or roll back before it can read a balance.
//
// Idempotency: if idempotencyKey is non-empty and a transaction with that
// key already exists, its result is returned as-is and no new transfer is
// performed — retried requests (e.g. a frontend double-click, or a network
// retry) can never move money twice.
func (s *TransferService) Transfer(ctx context.Context, fromAccountID, toAccountID int64, amount float64, description string, idempotencyKey string) (models.Transaction, error) {
	if amount <= 0 {
		return models.Transaction{}, errInvalidAmount()
	}
	if fromAccountID == toAccountID {
		return models.Transaction{}, errSameAccountTransfer()
	}

	if idempotencyKey != "" {
		existing, err := s.findByIdempotencyKey(ctx, idempotencyKey)
		if err == nil {
			return existing, nil
		}
		var apiErr *util.APIError
		if !errors.As(err, &apiErr) || apiErr.Status != 404 {
			return models.Transaction{}, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Transaction{}, err
	}
	defer tx.Rollback()

	// Fixed lock ordering prevents deadlocks between two transfers that
	// touch the same pair of accounts in opposite directions.
	firstID, secondID := fromAccountID, toAccountID
	if secondID < firstID {
		firstID, secondID = secondID, firstID
	}
	first, err := lockAccountForUpdate(ctx, tx, firstID)
	if err != nil {
		return models.Transaction{}, err
	}
	second, err := lockAccountForUpdate(ctx, tx, secondID)
	if err != nil {
		return models.Transaction{}, err
	}

	var from, to lockedAccount
	if first.ID == fromAccountID {
		from, to = first, second
	} else {
		from, to = second, first
	}

	if from.Status == "FROZEN" {
		return models.Transaction{}, errAccountFrozen()
	}
	if from.Status == "CLOSED" || to.Status == "CLOSED" {
		return models.Transaction{}, errAccountClosed()
	}
	if to.Status == "FROZEN" {
		return models.Transaction{}, util.UnprocessableEntity("Destination account is frozen and cannot receive transfers")
	}
	if from.Balance < amount {
		return models.Transaction{}, errInsufficientBalance()
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - ?, available_balance = available_balance - ? WHERE id = ?`,
		amount, amount, from.ID); err != nil {
		return models.Transaction{}, err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + ?, available_balance = available_balance + ? WHERE id = ?`,
		amount, amount, to.ID); err != nil {
		return models.Transaction{}, err
	}

	risk := classifyRisk(amount, from.Balance)

	var keyPtr *string
	if idempotencyKey != "" {
		keyPtr = &idempotencyKey
	}
	txnID, _, err := insertTransaction(ctx, tx, "TRANSFER", "SUCCESS", &from.ID, &to.ID, amount, description, risk, keyPtr)
	if err != nil {
		// A duplicate idempotency key racing with itself hits the unique
		// constraint here; surface it as a clean conflict rather than a
		// raw SQL error.
		return models.Transaction{}, errDuplicateTransaction()
	}
	if err := insertLedgerEntry(ctx, tx, txnID, from.ID, "DEBIT", amount); err != nil {
		return models.Transaction{}, err
	}
	if err := insertLedgerEntry(ctx, tx, txnID, to.ID, "CREDIT", amount); err != nil {
		return models.Transaction{}, err
	}

	desc := "Transfer from " + util.MaskAccountNumber(from.Number) + " to " + util.MaskAccountNumber(to.Number)
	if err := s.audit.Record(ctx, tx, "TRANSFER", "CUSTOMER", nil, "TRANSACTION", &txnID, desc); err != nil {
		return models.Transaction{}, err
	}
	if risk != "LOW" {
		if err := s.audit.Record(ctx, tx, "SUSPICIOUS_TRANSACTION_FLAGGED", "SYSTEM", nil, "TRANSACTION", &txnID,
			"Transfer flagged as "+risk+" risk by simulated rule engine"); err != nil {
			return models.Transaction{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Transaction{}, err
	}

	row := s.db.QueryRowContext(ctx, transactionSelect+" WHERE t.id = ?", txnID)
	return scanTransaction(row)
}

func (s *TransferService) findByIdempotencyKey(ctx context.Context, key string) (models.Transaction, error) {
	row := s.db.QueryRowContext(ctx, transactionSelect+" WHERE t.idempotency_key = ?", key)
	return scanTransaction(row)
}
