package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TransactionType int

const (
	Undefined TransactionType = iota
	Deposit
	Withdraw
)

func (tt TransactionType) String() string {
	switch tt {
	case Deposit:
		return "deposit"
	case Withdraw:
		return "withdraw"
	default:
		return "undefined"
	}
}

func FromStringToTransactionType(strType string) TransactionType {
	switch strType {
	case "deposit":
		return Deposit
	case "withdraw":
		return Withdraw
	default:
		return Undefined
	}
}

func (tt TransactionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(tt.String())
}

func (tt *TransactionType) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if s, ok := v.(string); ok {
		*tt = FromStringToTransactionType(s)
	}
	return nil
}

type Transaction struct {
	ID            uuid.UUID `json:"id"`
	AmountInCents int       `json:"amount_in_cents,omitempty"`
	Amount        string    `json:"amount,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	AccountID     int       `json:"account_id,omitempty"`
}

type TransactionRequestBody struct {
	Type      TransactionType `json:"type"`
	Amount    string          `json:"amount"`
	AccountID int             `json:"account_id"`
}

func (t *Transaction) ParseToAmount() string {
	amount := t.AmountInCents
	sign := ""
	if t.AmountInCents < 0 {
		amount *= -1
		sign = "-"
	}

	return fmt.Sprintf("%s%d.%02d", sign, amount/100, amount%100)
}

func (t *Transaction) OmitAmountInCents() {
	t.AmountInCents = 0
}

func (t *Transaction) OmitAccountID() {
	t.AccountID = 0
}

func (trb *TransactionRequestBody) ParseToAmountInCents() (int, error) {
	trb.Amount = strings.ReplaceAll(trb.Amount, ".", "")
	amountInCents, err := strconv.Atoi(trb.Amount)
	if err != nil {
		return 0, err
	}
	if amountInCents < 0 {
		return 0, errors.New("amount cannot be a negative number")
	}

	switch trb.Type {
	case Deposit:
		return amountInCents, nil
	case Withdraw:
		return -amountInCents, nil
	default:
		return 0, errors.New("undefined transaction type")
	}
}

func (trb *TransactionRequestBody) HasSufficientFunds(balance int, amountInCents int) bool {
	if amountInCents >= 0 {
		return true
	}

	return math.Abs(float64(amountInCents)) <= float64(balance)
}
