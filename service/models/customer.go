package models

import (
	"crypto/sha256"
	"encoding/hex"
)

type Customer struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	PINHash   string   `json:"pin_hash,omitempty"`
	AccountID int      `json:"account_id,omitempty"`
	Account   *Account `json:"account"`
}

func GeneratePINHash(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}

type CustomerRequestBody struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	PINNumber     string `json:"pin_number"`
	AccountNumber string `json:"account_number"`
}

type CustomerCredentials struct {
	PINNumber     string `json:"pin_number"`
	AccountNumber string `json:"account_number"`
}

type CustomerVerified struct {
	JWT      string    `json:"jwt"`
	Customer *Customer `json:"customer"`
}

func (c *Customer) OmitValues() {
	c.PINHash = ""
	c.AccountID = 0
}
