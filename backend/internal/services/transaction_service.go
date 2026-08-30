package services

import (
	"context"
	"database/sql"
	"errors"

	"corebank/backend/internal/models"
	"corebank/backend/internal/util"
)

type TransactionService struct {
	db    *sql.DB
	audit *AuditLogService
}

func NewTransactionService(db *sql.DB, audit *AuditLogService) *TransactionService {
	return &TransactionService{db: db, audit: audit}
}

// lockedAccount is the row shape we read with SELECT ... FOR UPDATE inside
// a transaction, right before mutating a balance.
type lockedAccount struct {
	ID      int64
	Number  string
	Balance float64
	Status  string
}

func lockAccountForUpdate(ctx context.Context, tx *sql.Tx, id int64) (lockedAccount, error) {
	var a lockedAccount
	err := tx.QueryRowContext(ctx,
		`SELECT id, account_number, balance, status FROM accounts WHERE id = ? FOR UPDATE`, id,
	).Scan(&a.ID, &a.Number, &a.Balance, &a.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return lockedAccount{}, errAccountNotFound()
	}
	return a, err
}

// Deposit credits an account. Deposits are allowed even on a FROZEN
// account (a freeze blocks money leaving the account, not entering it) but
// never on a CLOSED one.
func (s *TransactionService) Deposit(ctx context.Context, accountID int64, amount float64, description string) (models.Transaction, error) {
	if amount <= 0 {
		return models.Transaction{}, errInvalidAmount()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Transaction{}, err
	}
	defer tx.Rollback()

	acc, err := lockAccountForUpdate(ctx, tx, accountID)
	if err != nil {
		return models.Transaction{}, err
	}
	if acc.Status == "CLOSED" {
		return models.Transaction{}, errAccountClosed()
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + ?, available_balance = available_balance + ? WHERE id = ?`,
		amount, amount, accountID); err != nil {
		return models.Transaction{}, err
	}

	txnID, ref, err := insertTransaction(ctx, tx, "DEPOSIT", "SUCCESS", nil, &accountID, amount, description, "LOW", nil)
	if err != nil {
		return models.Transaction{}, err
	}
	if err := insertLedgerEntry(ctx, tx, txnID, accountID, "CREDIT", amount); err != nil {
		return models.Transaction{}, err
	}
	if err := s.audit.Record(ctx, tx, "DEPOSIT", "CUSTOMER", nil, "TRANSACTION", &txnID,
		"Deposit simulation of "+util.MaskAccountNumber(acc.Number)); err != nil {
		return models.Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Transaction{}, err
	}
	return s.GetByID(ctx, txnID, ref)
}

// Withdraw debits an account. Only an ACTIVE account may be withdrawn from.
func (s *TransactionService) Withdraw(ctx context.Context, accountID int64, amount float64, description string) (models.Transaction, error) {
	if amount <= 0 {
		return models.Transaction{}, errInvalidAmount()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Transaction{}, err
	}
	defer tx.Rollback()

	acc, err := lockAccountForUpdate(ctx, tx, accountID)
	if err != nil {
		return models.Transaction{}, err
	}
	if acc.Status == "FROZEN" {
		return models.Transaction{}, errAccountFrozen()
	}
	if acc.Status == "CLOSED" {
		return models.Transaction{}, errAccountClosed()
	}
	if acc.Balance < amount {
		return models.Transaction{}, errInsufficientBalance()
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - ?, available_balance = available_balance - ? WHERE id = ?`,
		amount, amount, accountID); err != nil {
		return models.Transaction{}, err
	}

	txnID, ref, err := insertTransaction(ctx, tx, "WITHDRAWAL", "SUCCESS", &accountID, nil, amount, description, "LOW", nil)
	if err != nil {
		return models.Transaction{}, err
	}
	if err := insertLedgerEntry(ctx, tx, txnID, accountID, "DEBIT", amount); err != nil {
		return models.Transaction{}, err
	}
	if err := s.audit.Record(ctx, tx, "WITHDRAWAL", "CUSTOMER", nil, "TRANSACTION", &txnID,
		"Withdrawal simulation from "+util.MaskAccountNumber(acc.Number)); err != nil {
		return models.Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Transaction{}, err
	}
	return s.GetByID(ctx, txnID, ref)
}

// insertTransaction writes the transactions row and returns its id + ref.
func insertTransaction(ctx context.Context, tx *sql.Tx, txnType, status string, sourceID, destID *int64, amount float64, description, riskLevel string, idempotencyKey *string) (int64, string, error) {
	ref := util.GenerateReference("TXN")
	res, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (transaction_ref, type, status, source_account_id, destination_account_id, amount, description, idempotency_key, risk_level)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ref, txnType, status, sourceID, destID, amount, description, idempotencyKey, riskLevel)
	if err != nil {
		return 0, "", err
	}
	id, err := res.LastInsertId()
	return id, ref, err
}

func insertLedgerEntry(ctx context.Context, tx *sql.Tx, txnID, accountID int64, entryType string, amount float64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO ledger_entries (transaction_id, account_id, entry_type, amount) VALUES (?, ?, ?, ?)`,
		txnID, accountID, entryType, amount)
	return err
}

const transactionSelect = `
	SELECT t.id, t.transaction_ref, t.type, t.status, t.source_account_id, t.destination_account_id,
	       sa.account_number, da.account_number, t.amount, t.description, t.risk_level, t.created_at
	FROM transactions t
	LEFT JOIN accounts sa ON t.source_account_id = sa.id
	LEFT JOIN accounts da ON t.destination_account_id = da.id`

func (s *TransactionService) GetByID(ctx context.Context, id int64, _ string) (models.Transaction, error) {
	row := s.db.QueryRowContext(ctx, transactionSelect+" WHERE t.id = ?", id)
	return scanTransaction(row)
}

// ListByAccount returns every transaction where the account is either the
// source or the destination, most recent first.
func (s *TransactionService) ListByAccount(ctx context.Context, accountID int64, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, transactionSelect+`
		WHERE t.source_account_id = ? OR t.destination_account_id = ?
		ORDER BY t.created_at DESC, t.id DESC LIMIT ?`, accountID, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactions(rows)
}

// ListByCustomer joins through accounts to return every transaction that
// touches any account owned by the customer.
func (s *TransactionService) ListByCustomer(ctx context.Context, customerID int64, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, transactionSelect+`
		WHERE t.source_account_id IN (SELECT id FROM accounts WHERE customer_id = ?)
		   OR t.destination_account_id IN (SELECT id FROM accounts WHERE customer_id = ?)
		ORDER BY t.created_at DESC, t.id DESC LIMIT ?`, customerID, customerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactions(rows)
}

func (s *TransactionService) ListAll(ctx context.Context, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, transactionSelect+" ORDER BY t.created_at DESC, t.id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactions(rows)
}

func (s *TransactionService) ListSuspicious(ctx context.Context, limit int) ([]models.Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, transactionSelect+`
		WHERE t.risk_level IN ('MEDIUM','HIGH') ORDER BY t.created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactions(rows)
}

func scanTransactions(rows *sql.Rows) ([]models.Transaction, error) {
	var out []models.Transaction
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanTransaction(row rowScanner) (models.Transaction, error) {
	var t models.Transaction
	var sourceID, destID sql.NullInt64
	var sourceNum, destNum sql.NullString
	err := row.Scan(&t.ID, &t.TransactionRef, &t.Type, &t.Status, &sourceID, &destID,
		&sourceNum, &destNum, &t.Amount, &t.Description, &t.RiskLevel, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Transaction{}, util.NotFound("Transaction")
	}
	if err != nil {
		return models.Transaction{}, err
	}
	if sourceID.Valid {
		v := sourceID.Int64
		t.SourceAccountID = &v
	}
	if destID.Valid {
		v := destID.Int64
		t.DestinationAccountID = &v
	}
	if sourceNum.Valid {
		m := util.MaskAccountNumber(sourceNum.String)
		t.SourceAccountNumber = &m
	}
	if destNum.Valid {
		m := util.MaskAccountNumber(destNum.String)
		t.DestinationAccountNum = &m
	}
	return t, nil
}
