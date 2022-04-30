package models

import (
	"fmt"
	"strconv"
)

type Account struct {
	ID           int           `json:"id"`
	Number       string        `json:"number"`
	Balance      string        `json:"balance,omitempty"`
	Transactions []Transaction `json:"transactions,omitempty"`
}

func GenerateAccountNo(previousAccNo string) string {
	var pointer int
	for i, v := range previousAccNo {
		if v == '0' {
			continue
		}
		pointer = i
		break
	}

	previousAccNoTrimmed := previousAccNo[pointer:]
	previousAccNoInt, _ := strconv.Atoi(previousAccNoTrimmed)
	previousAccNoInt++

	return fmt.Sprintf("%012d", previousAccNoInt)
}

type AccountRequestBody struct {
	Number string `json:"number"`
}
