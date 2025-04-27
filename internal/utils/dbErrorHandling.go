package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// Validate Foreign Key Constraint Violation
func ValidateFKey(c *gin.Context, err error, columnName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		Fail(c, ErrForeignKeyViolation(columnName), err)
		return false
	}
	return true
}

// Validate Unique Constraint Violation
func ValidateUniqueness(c *gin.Context, err error, columnName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		Fail(c, ErrUniqueViolation(columnName), err)
		return false
	}
	return true
}
