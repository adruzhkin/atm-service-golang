package db

import (
	"database/sql"
	"errors"

	"github.com/adruzhkin/atm-service-golang/service/models"
)

func (p *Postgres) GetAccountByID(id int) (*models.Account, error) {
	var acc models.Account
	err := p.db.QueryRow("SELECT id, number FROM accounts WHERE id=$1", id).Scan(&acc.ID, &acc.Number)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &models.Account{}, err
		default:
			return &models.Account{}, errors.New("failed to query existing account by id")
		}
	}
	return &acc, nil
}

func (p *Postgres) GetAccountByNumber(strNumber string) (*models.Account, error) {
	var acc models.Account
	err := p.db.QueryRow("SELECT id, number FROM accounts WHERE number=$1", strNumber).Scan(&acc.ID, &acc.Number)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &models.Account{}, err
		default:
			return &models.Account{}, errors.New("failed to query existing account by number")
		}
	}
	return &acc, nil
}