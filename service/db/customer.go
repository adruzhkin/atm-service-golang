package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/adruzhkin/atm-service-golang/service/models"
)

func (p *Postgres) GetCustomerByEmail(email string) (*models.Customer, error) {
	var cus models.Customer
	err := p.db.
		QueryRow("SELECT id, first_name, last_name, email, pin_hash, account_id FROM customers WHERE email=$1", email).
		Scan(&cus.ID, &cus.FirstName, &cus.LastName, &cus.Email, &cus.PINHash, &cus.AccountID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &models.Customer{}, err
		default:
			return &models.Customer{}, fmt.Errorf("failed to query customer by email: %w", err)
		}
	}

	acc, err := p.GetAccountByID(cus.AccountID)
	if err != nil {
		return &models.Customer{}, err
	}
	cus.Account = acc

	return &cus, nil
}

func (p *Postgres) GetCustomerByCredentials(crd *models.CustomerCredentials) (*models.Customer, error) {
	cus, err := p.GetCustomerByEmail(crd.Email)
	if err != nil {
		return &models.Customer{}, errors.New("invalid login credentials")
	}

	if !models.ComparePINHash(cus.PINHash, crd.PINNumber) {
		return &models.Customer{}, errors.New("invalid login credentials")
	}

	return cus, nil
}

func (p *Postgres) CreateCustomer(cus *models.Customer) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin customer transaction: %w", err)
	}

	// Lock last account row and generate next account number.
	var lastAccNo string
	err = tx.QueryRow("SELECT number FROM accounts ORDER BY number DESC LIMIT 1 FOR UPDATE").Scan(&lastAccNo)
	if err != nil && err != sql.ErrNoRows {
		_ = tx.Rollback()
		return fmt.Errorf("failed to lock accounts for number generation: %w", err)
	}

	acc := cus.Account
	acc.Number = models.GenerateAccountNo(lastAccNo)

	err = tx.QueryRow("INSERT INTO accounts(number) VALUES($1) RETURNING id", acc.Number).Scan(&acc.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to insert account: %w", err)
	}

	err = tx.QueryRow("INSERT INTO customers(first_name, last_name, email, pin_hash, account_id) VALUES($1,$2,$3,$4,$5) RETURNING id",
		cus.FirstName, cus.LastName, cus.Email, cus.PINHash, acc.ID).Scan(&cus.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to insert customer: %w", err)
	}
	cus.AccountID = acc.ID

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit customer creation: %w", err)
	}

	return nil
}
