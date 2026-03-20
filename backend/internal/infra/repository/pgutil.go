package repository

import (
	"errors"

	"github.com/lib/pq"
)

func isUniqueViolation(err error) bool {
	var pgErr *pq.Error
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
