package tables

import (
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
)

// TBD: move these in to aaronland/go-sqlite ?

// WrapError returns a new error wrapping 'err' and prepending with the value of 't's Name() method.
func WrapError(t sqlite.Table, err error) error {
	return fmt.Errorf("[%s] %w", t.Name(), err)
}

// InitializeTableError returns a new error with a default message for database initialization problems wrapping 'err' and prepending with the value of 't's Name() method.
func InitializeTableError(t sqlite.Table, err error) error {
	return WrapError(t, fmt.Errorf("Failed to initialize database table, %w", err))
}

// MissingPropertyError returns a new error with a default message for problems deriving a given property ('prop') from a record, wrapping 'err' and prepending with the value of 't's Name() method.
func MissingPropertyError(t sqlite.Table, prop string, err error) error {
	return WrapError(t, fmt.Errorf("Failed to determine value for '%s' property, %w", prop, err))
}

// DatabaseConnectionError returns a new error with a default message for database connection problems wrapping 'err' and prepending with the value of 't's Name() method.
func DatabaseConnectionError(t sqlite.Table, err error) error {
	return WrapError(t, fmt.Errorf("Failed to establish database connection, %w", err))
}

// BeginTransactionError returns a new error with a default message for database transaction initialization problems wrapping 'err' and prepending with the value of 't's Name() method.
func BeginTransactionError(t sqlite.Table, err error) error {
	return WrapError(t, fmt.Errorf("Failed to begin database transaction, %w", err))
}

// CommitTransactionError returns a new error with a default message for problems committing database transactions wrapping 'err' and prepending with the value of 't's Name() method.
func CommitTransactionError(t sqlite.Table, err error) error {
	return WrapError(t, fmt.Errorf("Failed to commit database transaction, %w", err))
}

// PrepareStatementError returns a new error with a default message for problems preparing database (SQL) statements wrapping 'err' and prepending with the value of 't's Name() method.
func PrepareStatementError(t sqlite.Table, err error) error {
	return WrapError(t, fmt.Errorf("Failed to prepare SQL statement, %w", err))
}

// ExecuteStatementError returns a new error with a default message for problems executing database (SQL) statements wrapping 'err' and prepending with the value of 't's Name() method.
func ExecuteStatementError(t sqlite.Table, err error) error {
	return WrapError(t, fmt.Errorf("Failed to execute SQL statement, %w", err))
}
