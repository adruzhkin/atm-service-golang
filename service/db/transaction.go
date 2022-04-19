package db

import (
	"database/sql"
	"errors"
	"github.com/adruzhkin/atm-service-golang/service/models"
)

func (p *Postgres) GetTransactionsByAccountID(id int) ([]models.Transaction, error) {
	rows, err := p.db.Query(
		"SELECT id, amount_in_cents, created_at, account_id FROM transactions WHERE account_id=$1", id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, err
		default:
			return nil, errors.New("failed to query transactions by account id")
		}
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err = rows.Scan(&tx.ID, &tx.AmountInCents, &tx.CreatedAt, &tx.AccountID); err != nil {
			return nil, errors.New("failed to query existing transaction by account id")
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

func (p *Postgres) GetTransactionsBalanceByAccountID(id int) (int, error) {
	var balance int
	err := p.db.QueryRow("SELECT coalesce(sum(amount_in_cents), 0) FROM transactions WHERE account_id=$1", id).Scan(&balance)
	if err != nil {
		return 0, errors.New("failed to query transactions balance")
	}

	return balance, nil
}

func (p *Postgres) CreateTransaction(tx *models.Transaction) error {
	err := p.db.QueryRow(
		"INSERT INTO transactions(amount_in_cents, account_id) VALUES($1,$2) RETURNING id, created_at",
		tx.AmountInCents, tx.AccountID).Scan(&tx.ID, &tx.CreatedAt)
	if err != nil {
		return errors.New("failed to create new transaction")
	}
	return nil
}
