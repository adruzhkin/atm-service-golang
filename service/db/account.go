package db

import (
	"database/sql"
	"fmt"

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
			return &models.Account{}, fmt.Errorf("failed to query account by id: %w", err)
		}
	}
	return &acc, nil
}
