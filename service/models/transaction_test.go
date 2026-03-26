package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToAmountInCents_ValidDeposit(t *testing.T) {
	tests := []struct {
		amount   string
		expected int
	}{
		{"10.50", 1050},
		{"0.01", 1},
		{"100.00", 10000},
		{"0.00", 0},
		{"999999.99", 99999999},
	}

	for _, tc := range tests {
		t.Run(tc.amount, func(t *testing.T) {
			trb := TransactionRequestBody{Type: Deposit, Amount: tc.amount}
			result, err := trb.ParseToAmountInCents()
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseToAmountInCents_ValidWithdraw(t *testing.T) {
	trb := TransactionRequestBody{Type: Withdraw, Amount: "25.00"}
	result, err := trb.ParseToAmountInCents()
	assert.NoError(t, err)
	assert.Equal(t, -2500, result)
}

func TestParseToAmountInCents_InvalidFormat(t *testing.T) {
	invalidAmounts := []string{
		"5.0.0",
		"5",
		"abc",
		"",
		"-1.00",
		".50",
		"5.",
		"5.1",
		"5.123",
		"10,50",
	}

	for _, amount := range invalidAmounts {
		t.Run(amount, func(t *testing.T) {
			trb := TransactionRequestBody{Type: Deposit, Amount: amount}
			_, err := trb.ParseToAmountInCents()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "amount must be in format")
		})
	}
}

func TestParseToAmountInCents_UndefinedType(t *testing.T) {
	trb := TransactionRequestBody{Type: Undefined, Amount: "10.00"}
	_, err := trb.ParseToAmountInCents()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined transaction type")
}

func TestTransactionRequestBody_Validate_Valid(t *testing.T) {
	trb := TransactionRequestBody{Type: Deposit, Amount: "10.00", AccountID: 1}
	assert.NoError(t, trb.Validate())
}

func TestTransactionRequestBody_Validate_Errors(t *testing.T) {
	tests := []struct {
		name    string
		trb     TransactionRequestBody
		wantErr string
	}{
		{"undefined type", TransactionRequestBody{Type: Undefined, Amount: "10.00", AccountID: 1}, "type must be 'deposit' or 'withdraw'"},
		{"empty amount", TransactionRequestBody{Type: Deposit, Amount: "", AccountID: 1}, "amount is required"},
		{"whitespace amount", TransactionRequestBody{Type: Deposit, Amount: "   ", AccountID: 1}, "amount is required"},
		{"zero account_id", TransactionRequestBody{Type: Deposit, Amount: "10.00", AccountID: 0}, "account_id is required"},
		{"negative account_id", TransactionRequestBody{Type: Deposit, Amount: "10.00", AccountID: -1}, "account_id is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.trb.Validate()
			assert.Error(t, err)
			assert.Equal(t, tc.wantErr, err.Error())
		})
	}
}
