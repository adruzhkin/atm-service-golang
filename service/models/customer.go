package models

import (
	"golang.org/x/crypto/bcrypt"
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

func GeneratePINHash(pin string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePINHash(hash, pin string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)) == nil
}

type CustomerRequestBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	PINNumber string `json:"pin_number"`
}

type CustomerCredentials struct {
	PINNumber string `json:"pin_number"`
	Email     string `json:"email"`
}

type CustomerVerified struct {
	JWT      string    `json:"jwt"`
	Customer *Customer `json:"customer"`
}

func (c *Customer) OmitValues() {
	c.PINHash = ""
	c.AccountID = 0
}
