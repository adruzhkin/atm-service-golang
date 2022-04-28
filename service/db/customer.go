package db

import (
	"database/sql"
	"errors"

	"github.com/adruzhkin/atm-service-golang/service/models"
)

func (p *Postgres) GetCustomerByAccountID(id int) (*models.Customer, error) {
	var cus models.Customer
	err := p.db.
		QueryRow("SELECT id, first_name, last_name, email, pin_hash, account_id FROM customers WHERE account_id=$1", id).
		Scan(&cus.ID, &cus.FirstName, &cus.LastName, &cus.Email, &cus.PINHash, &cus.AccountID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &models.Customer{}, err
		default:
			return &models.Customer{}, errors.New("failed to query existing customer by account id")
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
	acc, err := p.GetAccountByNumber(crd.AccountNumber)
	if err != nil {
		return &models.Customer{}, errors.New("invalid login credentials")
	}

	cus, err := p.GetCustomerByAccountID(acc.ID)
	if err != nil {
		return &models.Customer{}, errors.New("invalid login credentials")
	}

	crdPINHash := models.GeneratePINHash(crd.PINNumber)
	if cus.PINHash != crdPINHash {
		return &models.Customer{}, errors.New("invalid login credentials")
	}

	return cus, nil
}

func (p *Postgres) CreateCustomer(cus *models.Customer) error {
	tx, err := p.db.Begin()
	if err != nil {
		return errors.New("failed to create new customer")
	}

	acc := cus.Account
	_, err = tx.Exec("INSERT INTO accounts(number) VALUES($1)", acc.Number)
	if err != nil {
		_ = tx.Rollback()
		return errors.New("failed to create account for new customer")
	}

	err = tx.QueryRow("SELECT id FROM accounts WHERE number=$1", acc.Number).Scan(&acc.ID)
	if err != nil {
		_ = tx.Rollback()
		return errors.New("failed to create account for new customer")
	}

	_, err = tx.Exec("INSERT INTO customers(first_name, last_name, email, pin_hash, account_id) VALUES($1,$2,$3,$4,$5)",
		cus.FirstName, cus.LastName, cus.Email, cus.PINHash, acc.ID)
	if err != nil {
		_ = tx.Rollback()
		return errors.New("failed to create new customer")
	}

	err = tx.QueryRow("SELECT id, account_id FROM customers WHERE account_id=$1", acc.ID).Scan(&cus.ID, &cus.AccountID)
	if err != nil {
		_ = tx.Rollback()
		return errors.New("failed to create new customer")
	}

	err = tx.Commit()
	if err != nil {
		return errors.New("failed to create new customer")
	}

	return nil
}
