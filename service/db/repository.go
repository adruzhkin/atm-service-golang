package db

import "github.com/adruzhkin/atm-service-golang/service/models"

type Repo interface {
	Open() error
	Close() error
	Ping() error

	GetCustomerByEmail(email string) (*models.Customer, error)
	GetCustomerByCredentials(crd *models.CustomerCredentials) (*models.Customer, error)
	CreateCustomer(c *models.Customer) error

	GetAccountByID(id int) (*models.Account, error)

	GetTransactionsByAccountID(id int) ([]models.Transaction, error)
	GetTransactionsBalanceByAccountID(id int) (int, error)
	CreateTransaction(tx *models.Transaction) error
}
