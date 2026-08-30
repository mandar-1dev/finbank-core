package services

import (
	"context"
	"database/sql"

	"corebank/backend/internal/models"
	"corebank/backend/internal/util"
)

type CustomerService struct {
	db *sql.DB
}

func NewCustomerService(db *sql.DB) *CustomerService {
	return &CustomerService{db: db}
}

func (s *CustomerService) List(ctx context.Context) ([]models.Customer, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, customer_number, full_name, email, phone, date_of_birth, address, status, created_at
		FROM customers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Customer
	for rows.Next() {
		c, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *CustomerService) Get(ctx context.Context, id int64) (models.Customer, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, customer_number, full_name, email, phone, date_of_birth, address, status, created_at
		FROM customers WHERE id = ?`, id)
	c, err := scanCustomer(row)
	if err == sql.ErrNoRows {
		return models.Customer{}, util.NotFound("Customer")
	}
	return c, err
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanCustomer(row rowScanner) (models.Customer, error) {
	var c models.Customer
	var dob []byte
	err := row.Scan(&c.ID, &c.CustomerNumber, &c.FullName, &c.Email, &c.Phone, &dob, &c.Address, &c.Status, &c.CreatedAt)
	if err != nil {
		return models.Customer{}, err
	}
	c.DateOfBirth = string(dob)
	return c, nil
}
