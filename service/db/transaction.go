package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/adruzhkin/atm-service-golang/service/models"
)

var ErrInsufficientFunds = errors.New("non-sufficient funds")

func (p *Postgres) GetTransactionsByAccountID(id int) ([]models.Transaction, error) {
	rows, err := p.db.Query(
		"SELECT id, amount_in_cents, created_at, account_id FROM transactions WHERE account_id=$1", id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, err
		default:
			return nil, fmt.Errorf("failed to query transactions by account id: %w", err)
		}
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err = rows.Scan(&tx.ID, &tx.AmountInCents, &tx.CreatedAt, &tx.AccountID); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		txs = append(txs, tx)
	}

	return txs, nil
}

func (p *Postgres) GetTransactionsBalanceByAccountID(id int) (int, error) {
	var balance int
	err := p.db.QueryRow("SELECT coalesce(sum(amount_in_cents), 0) FROM transactions WHERE account_id=$1", id).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to query transactions balance: %w", err)
	}

	return balance, nil
}

func (p *Postgres) CreateTransaction(tx *models.Transaction) error {
	err := p.db.QueryRow(
		"INSERT INTO transactions(amount_in_cents, account_id) VALUES($1,$2) RETURNING id, created_at",
		tx.AmountInCents, tx.AccountID).Scan(&tx.ID, &tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (p *Postgres) CreateTransactionWithBalanceCheck(tx *models.Transaction) error {
	dbTx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	var balance int
	err = dbTx.QueryRow(
		"SELECT coalesce(sum(amount_in_cents), 0) FROM transactions WHERE account_id=$1 FOR UPDATE",
		tx.AccountID).Scan(&balance)
	if err != nil {
		_ = dbTx.Rollback()
		return fmt.Errorf("failed to query balance for update: %w", err)
	}

	if tx.AmountInCents < 0 && -tx.AmountInCents > balance {
		_ = dbTx.Rollback()
		return ErrInsufficientFunds
	}

	err = dbTx.QueryRow(
		"INSERT INTO transactions(amount_in_cents, account_id) VALUES($1,$2) RETURNING id, created_at",
		tx.AmountInCents, tx.AccountID).Scan(&tx.ID, &tx.CreatedAt)
	if err != nil {
		_ = dbTx.Rollback()
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	if err = dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
