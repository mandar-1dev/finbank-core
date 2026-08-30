package services

import (
	"context"
	"database/sql"

	"corebank/backend/internal/models"
	"corebank/backend/internal/util"
)

type BeneficiaryService struct {
	db    *sql.DB
	audit *AuditLogService
}

func NewBeneficiaryService(db *sql.DB, audit *AuditLogService) *BeneficiaryService {
	return &BeneficiaryService{db: db, audit: audit}
}

func (s *BeneficiaryService) ListByCustomer(ctx context.Context, customerID int64) ([]models.Beneficiary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, customer_id, name, account_number, ifsc_code, COALESCE(nickname,''), created_at
		FROM beneficiaries WHERE customer_id = ? ORDER BY id`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Beneficiary
	for rows.Next() {
		var b models.Beneficiary
		if err := rows.Scan(&b.ID, &b.CustomerID, &b.Name, &b.AccountNumber, &b.IFSCCode, &b.Nickname, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *BeneficiaryService) Add(ctx context.Context, customerID int64, name, accountNumber, ifsc, nickname string) (models.Beneficiary, error) {
	if name == "" || accountNumber == "" {
		return models.Beneficiary{}, util.BadRequest("name and accountNumber are required")
	}
	if ifsc == "" {
		ifsc = "CORE0000001"
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO beneficiaries (customer_id, name, account_number, ifsc_code, nickname)
		VALUES (?, ?, ?, ?, ?)`, customerID, name, accountNumber, ifsc, nickname)
	if err != nil {
		return models.Beneficiary{}, err
	}
	id, _ := res.LastInsertId()
	_ = s.audit.Record(ctx, s.db, "BENEFICIARY_ADDED", "CUSTOMER", &customerID, "BENEFICIARY", &id,
		"Beneficiary "+name+" added")
	return s.get(ctx, id)
}

func (s *BeneficiaryService) Remove(ctx context.Context, customerID, beneficiaryID int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM beneficiaries WHERE id = ? AND customer_id = ?`, beneficiaryID, customerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errBeneficiaryNotFound()
	}
	_ = s.audit.Record(ctx, s.db, "BENEFICIARY_REMOVED", "CUSTOMER", &customerID, "BENEFICIARY", &beneficiaryID,
		"Beneficiary removed")
	return nil
}

func (s *BeneficiaryService) get(ctx context.Context, id int64) (models.Beneficiary, error) {
	var b models.Beneficiary
	err := s.db.QueryRowContext(ctx, `
		SELECT id, customer_id, name, account_number, ifsc_code, COALESCE(nickname,''), created_at
		FROM beneficiaries WHERE id = ?`, id,
	).Scan(&b.ID, &b.CustomerID, &b.Name, &b.AccountNumber, &b.IFSCCode, &b.Nickname, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return models.Beneficiary{}, errBeneficiaryNotFound()
	}
	return b, err
}
