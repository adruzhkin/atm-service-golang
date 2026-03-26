package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomerRequestBody_Validate_Valid(t *testing.T) {
	crb := CustomerRequestBody{
		FirstName: "Natasha",
		LastName:  "Romanov",
		Email:     "natasha@gmail.com",
		PINNumber: "1234",
	}
	assert.NoError(t, crb.Validate())
}

func TestCustomerRequestBody_Validate_Errors(t *testing.T) {
	valid := CustomerRequestBody{
		FirstName: "Natasha",
		LastName:  "Romanov",
		Email:     "natasha@gmail.com",
		PINNumber: "1234",
	}

	tests := []struct {
		name    string
		modify  func(crb *CustomerRequestBody)
		wantErr string
	}{
		{"empty first_name", func(crb *CustomerRequestBody) { crb.FirstName = "" }, "first_name is required"},
		{"whitespace first_name", func(crb *CustomerRequestBody) { crb.FirstName = "   " }, "first_name is required"},
		{"first_name too long", func(crb *CustomerRequestBody) { crb.FirstName = strings.Repeat("a", 26) }, "first_name must be 25 characters or less"},
		{"empty last_name", func(crb *CustomerRequestBody) { crb.LastName = "" }, "last_name is required"},
		{"whitespace last_name", func(crb *CustomerRequestBody) { crb.LastName = "   " }, "last_name is required"},
		{"last_name too long", func(crb *CustomerRequestBody) { crb.LastName = strings.Repeat("a", 26) }, "last_name must be 25 characters or less"},
		{"empty email", func(crb *CustomerRequestBody) { crb.Email = "" }, "email is required"},
		{"whitespace email", func(crb *CustomerRequestBody) { crb.Email = "   " }, "email is required"},
		{"invalid email no @", func(crb *CustomerRequestBody) { crb.Email = "notanemail" }, "invalid email format"},
		{"invalid email no domain", func(crb *CustomerRequestBody) { crb.Email = "user@" }, "invalid email format"},
		{"invalid email no tld", func(crb *CustomerRequestBody) { crb.Email = "user@host" }, "invalid email format"},
		{"email too long", func(crb *CustomerRequestBody) { crb.Email = strings.Repeat("a", 243) + "@example.com" }, "email must be 254 characters or less"},
		{"PIN too short", func(crb *CustomerRequestBody) { crb.PINNumber = "123" }, "pin_number must be between 4 and 12 digits"},
		{"PIN too long", func(crb *CustomerRequestBody) { crb.PINNumber = "1234567890123" }, "pin_number must be between 4 and 12 digits"},
		{"PIN empty", func(crb *CustomerRequestBody) { crb.PINNumber = "" }, "pin_number must be between 4 and 12 digits"},
		{"PIN non-digit", func(crb *CustomerRequestBody) { crb.PINNumber = "12ab" }, "pin_number must contain only digits"},
		{"PIN with spaces", func(crb *CustomerRequestBody) { crb.PINNumber = "12 34" }, "pin_number must contain only digits"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			crb := valid
			tc.modify(&crb)
			err := crb.Validate()
			assert.Error(t, err)
			assert.Equal(t, tc.wantErr, err.Error())
		})
	}
}

func TestCustomerCredentials_Validate_Valid(t *testing.T) {
	cc := CustomerCredentials{
		Email:     "natasha@gmail.com",
		PINNumber: "1234",
	}
	assert.NoError(t, cc.Validate())
}

func TestCustomerCredentials_Validate_Errors(t *testing.T) {
	tests := []struct {
		name    string
		cc      CustomerCredentials
		wantErr string
	}{
		{"empty email", CustomerCredentials{Email: "", PINNumber: "1234"}, "email is required"},
		{"whitespace email", CustomerCredentials{Email: "   ", PINNumber: "1234"}, "email is required"},
		{"empty PIN", CustomerCredentials{Email: "natasha@gmail.com", PINNumber: ""}, "pin_number is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cc.Validate()
			assert.Error(t, err)
			assert.Equal(t, tc.wantErr, err.Error())
		})
	}
}
