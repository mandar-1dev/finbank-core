package services

import (
	"context"
	"database/sql"

	"corebank/backend/internal/models"
	"corebank/backend/internal/util"
)

type AccountService struct {
	db *sql.DB
}

func NewAccountService(db *sql.DB) *AccountService {
	return &AccountService{db: db}
}

const accountSelect = `
	SELECT a.id, a.account_number, a.customer_id, t.code, a.balance, a.available_balance, a.status, a.opened_at
	FROM accounts a JOIN account_types t ON a.account_type_id = t.id`

func (s *AccountService) ListByCustomer(ctx context.Context, customerID int64) ([]models.Account, error) {
	rows, err := s.db.QueryContext(ctx, accountSelect+" WHERE a.customer_id = ? ORDER BY a.id", customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccounts(rows)
}

func (s *AccountService) List(ctx context.Context) ([]models.Account, error) {
	rows, err := s.db.QueryContext(ctx, accountSelect+" ORDER BY a.id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccounts(rows)
}

func (s *AccountService) Get(ctx context.Context, id int64) (models.Account, error) {
	row := s.db.QueryRowContext(ctx, accountSelect+" WHERE a.id = ?", id)
	a, err := scanAccount(row)
	if err == sql.ErrNoRows {
		return models.Account{}, util.NotFound("Account")
	}
	return a, err
}

// SetStatus is used by banker operations to freeze/unfreeze an account.
func (s *AccountService) SetStatus(ctx context.Context, id int64, status string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE accounts SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return util.NotFound("Account")
	}
	return nil
}

func scanAccounts(rows *sql.Rows) ([]models.Account, error) {
	var out []models.Account
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanAccount(row rowScanner) (models.Account, error) {
	var a models.Account
	err := row.Scan(&a.ID, &a.AccountNumber, &a.CustomerID, &a.AccountTypeCode, &a.Balance, &a.AvailableBalance, &a.Status, &a.OpenedAt)
	if err != nil {
		return models.Account{}, err
	}
	a.MaskedNumber = util.MaskAccountNumber(a.AccountNumber)
	return a, nil
}
