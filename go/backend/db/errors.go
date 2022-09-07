package db

import (
	"errors"

	"github.com/lib/pq"
)

const (
	UniqueViolationError = pq.ErrorCode("23505")
)

func IsErrorCode(err error, code pq.ErrorCode) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}

	return false
}
