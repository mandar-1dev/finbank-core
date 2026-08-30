package services

import (
	"context"
	"database/sql"

	"corebank/backend/internal/models"
)

type BankerService struct {
	db      *sql.DB
	audit   *AuditLogService
	account *AccountService
}

func NewBankerService(db *sql.DB, audit *AuditLogService, account *AccountService) *BankerService {
	return &BankerService{db: db, audit: audit, account: account}
}

func (s *BankerService) DashboardStats(ctx context.Context) (models.DashboardStats, error) {
	var stats models.DashboardStats

	row := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM customers`)
	if err := row.Scan(&stats.TotalCustomers); err != nil {
		return stats, err
	}

	row = s.db.QueryRowContext(ctx, `SELECT COUNT(*), 
		SUM(status = 'ACTIVE'), SUM(status = 'FROZEN') FROM accounts`)
	if err := row.Scan(&stats.TotalAccounts, &stats.ActiveAccounts, &stats.FrozenAccounts); err != nil {
		return stats, err
	}

	row = s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(amount),0) FROM transactions WHERE DATE(created_at) = CURDATE()`)
	if err := row.Scan(&stats.TransactionsToday, &stats.TransactionVolumeToday); err != nil {
		return stats, err
	}

	row = s.db.QueryRowContext(ctx, `SELECT SUM(status = 'ACTIVE'), SUM(status = 'PENDING') FROM loans`)
	var activeLoans, pendingLoans sql.NullInt64
	if err := row.Scan(&activeLoans, &pendingLoans); err != nil {
		return stats, err
	}
	stats.ActiveLoans = int(activeLoans.Int64)
	stats.PendingLoans = int(pendingLoans.Int64)

	row = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM transactions WHERE risk_level IN ('MEDIUM','HIGH')`)
	if err := row.Scan(&stats.SuspiciousTransactions); err != nil {
		return stats, err
	}

	return stats, nil
}

// FreezeAccount and UnfreezeAccount are thin wrappers around AccountService
// that additionally write a BANKER-attributed audit log, matching the
// "every banker action must create an audit log" requirement.
func (s *BankerService) FreezeAccount(ctx context.Context, accountID int64) error {
	if err := s.account.SetStatus(ctx, accountID, "FROZEN"); err != nil {
		return err
	}
	return s.audit.Record(ctx, s.db, "ACCOUNT_FROZEN", "BANKER", nil, "ACCOUNT", &accountID,
		"Account frozen by banker simulation")
}

func (s *BankerService) UnfreezeAccount(ctx context.Context, accountID int64) error {
	if err := s.account.SetStatus(ctx, accountID, "ACTIVE"); err != nil {
		return err
	}
	return s.audit.Record(ctx, s.db, "ACCOUNT_UNFROZEN", "BANKER", nil, "ACCOUNT", &accountID,
		"Account unfrozen by banker simulation")
}

func (s *BankerService) AccountTypeDistribution(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.code, COUNT(*) FROM accounts a JOIN account_types t ON a.account_type_id = t.id GROUP BY t.code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			return nil, err
		}
		out[code] = count
	}
	return out, rows.Err()
}

func (s *BankerService) LoanStatusDistribution(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM loans GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		out[status] = count
	}
	return out, rows.Err()
}

func (s *BankerService) DailyTransactionVolume(ctx context.Context, days int) ([]models.DailyVolume, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT DATE(created_at) AS d, COUNT(*), COALESCE(SUM(amount),0)
		FROM transactions
		WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		GROUP BY DATE(created_at) ORDER BY d`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.DailyVolume
	for rows.Next() {
		var v models.DailyVolume
		var d []byte
		if err := rows.Scan(&d, &v.Count, &v.Volume); err != nil {
			return nil, err
		}
		v.Date = string(d)
		out = append(out, v)
	}
	return out, rows.Err()
}
