package models

type Account struct {
	ID           int           `json:"id"`
	Number       string        `json:"number"`
	Balance      string        `json:"balance,omitempty"`
	Transactions []Transaction `json:"transactions,omitempty"`
}

type AccountRequestBody struct {
	Number string `json:"number"`
}
