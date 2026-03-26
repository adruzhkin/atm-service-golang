package db

import (
	"errors"

	"github.com/lib/pq"
)

// PostgreSQL error code for unique_violation
const uniqueViolation = "23505"

func IsDuplicateKeyError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == uniqueViolation
	}
	return false
}
