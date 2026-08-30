package services

import (
	"context"
	"database/sql"

	"corebank/backend/internal/models"
	"corebank/backend/internal/util"
)

type CardService struct {
	db    *sql.DB
	audit *AuditLogService
}

func NewCardService(db *sql.DB, audit *AuditLogService) *CardService {
	return &CardService{db: db, audit: audit}
}

const cardSelect = `
	SELECT c.id, c.account_id, c.card_number, c.card_holder, c.expiry_month, c.expiry_year, c.status, c.spending_limit
	FROM cards c JOIN accounts a ON c.account_id = a.id`

func (s *CardService) ListByCustomer(ctx context.Context, customerID int64) ([]models.Card, error) {
	rows, err := s.db.QueryContext(ctx, cardSelect+" WHERE a.customer_id = ? ORDER BY c.id", customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Card
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *CardService) ListAll(ctx context.Context) ([]models.Card, error) {
	rows, err := s.db.QueryContext(ctx, cardSelect+" ORDER BY c.id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Card
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *CardService) SetStatus(ctx context.Context, cardID int64, status string) (models.Card, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE cards SET status = ? WHERE id = ?`, status, cardID)
	if err != nil {
		return models.Card{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return models.Card{}, errCardNotFound()
	}
	action := "CARD_UNFROZEN"
	if status == "FROZEN" {
		action = "CARD_FROZEN"
	}
	_ = s.audit.Record(ctx, s.db, action, "CUSTOMER", nil, "CARD", &cardID, "Card status changed to "+status)
	row := s.db.QueryRowContext(ctx, cardSelect+" WHERE c.id = ?", cardID)
	return scanCard(row)
}

// Pay simulates a point-of-sale card payment: it debits the linked account
// after checking the card is active and the amount is within both the
// card's spending limit and the account's available balance.
func (s *CardService) Pay(ctx context.Context, cardID int64, merchant string, amount float64) (models.Transaction, error) {
	if amount <= 0 {
		return models.Transaction{}, errInvalidAmount()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Transaction{}, err
	}
	defer tx.Rollback()

	var accountID int64
	var cardStatus string
	var spendLimit float64
	err = tx.QueryRowContext(ctx, `SELECT account_id, status, spending_limit FROM cards WHERE id = ? FOR UPDATE`, cardID).
		Scan(&accountID, &cardStatus, &spendLimit)
	if err == sql.ErrNoRows {
		return models.Transaction{}, errCardNotFound()
	}
	if err != nil {
		return models.Transaction{}, err
	}
	if cardStatus != "ACTIVE" {
		return models.Transaction{}, util.UnprocessableEntity("Card is not active")
	}
	if amount > spendLimit {
		return models.Transaction{}, errSpendingLimitExceeded()
	}

	acc, err := lockAccountForUpdate(ctx, tx, accountID)
	if err != nil {
		return models.Transaction{}, err
	}
	if acc.Status != "ACTIVE" {
		return models.Transaction{}, errAccountFrozen()
	}
	if acc.Balance < amount {
		return models.Transaction{}, errInsufficientBalance()
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - ?, available_balance = available_balance - ? WHERE id = ?`,
		amount, amount, accountID); err != nil {
		return models.Transaction{}, err
	}

	txnID, _, err := insertTransaction(ctx, tx, "CARD_PAYMENT", "SUCCESS", &accountID, nil, amount, "Card payment: "+merchant, "LOW", nil)
	if err != nil {
		return models.Transaction{}, err
	}
	if err := insertLedgerEntry(ctx, tx, txnID, accountID, "DEBIT", amount); err != nil {
		return models.Transaction{}, err
	}
	if err := s.audit.Record(ctx, tx, "CARD_PAYMENT", "CUSTOMER", nil, "TRANSACTION", &txnID,
		"Simulated card payment to "+merchant); err != nil {
		return models.Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Transaction{}, err
	}

	row := s.db.QueryRowContext(ctx, transactionSelect+" WHERE t.id = ?", txnID)
	return scanTransaction(row)
}

func scanCard(row rowScanner) (models.Card, error) {
	var c models.Card
	var number string
	err := row.Scan(&c.ID, &c.AccountID, &number, &c.CardHolder, &c.ExpiryMonth, &c.ExpiryYear, &c.Status, &c.SpendingLimit)
	if err != nil {
		return models.Card{}, err
	}
	c.MaskedNumber = util.MaskCardNumber(number)
	return c, nil
}
