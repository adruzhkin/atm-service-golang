package models

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var pinRegex = regexp.MustCompile(`^\d+$`)

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

func (crb *CustomerRequestBody) Validate() error {
	if strings.TrimSpace(crb.FirstName) == "" {
		return errors.New("first_name is required")
	}
	if len(crb.FirstName) > 25 {
		return errors.New("first_name must be 25 characters or less")
	}
	if strings.TrimSpace(crb.LastName) == "" {
		return errors.New("last_name is required")
	}
	if len(crb.LastName) > 25 {
		return errors.New("last_name must be 25 characters or less")
	}
	if strings.TrimSpace(crb.Email) == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(crb.Email) {
		return errors.New("invalid email format")
	}
	if len(crb.Email) > 254 {
		return errors.New("email must be 254 characters or less")
	}
	if len(crb.PINNumber) < 4 || len(crb.PINNumber) > 12 {
		return errors.New("pin_number must be between 4 and 12 digits")
	}
	if !pinRegex.MatchString(crb.PINNumber) {
		return errors.New("pin_number must contain only digits")
	}
	return nil
}

func (cc *CustomerCredentials) Validate() error {
	if strings.TrimSpace(cc.Email) == "" {
		return errors.New("email is required")
	}
	if cc.PINNumber == "" {
		return errors.New("pin_number is required")
	}
	return nil
}
