package utils

import (
	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

// UUIDArray - конвертирует список uuid в список выражений для jet
func UUIDArray(values []uuid.UUID) []postgres.Expression {
	expressions := make([]postgres.Expression, 0, len(values))
	for _, val := range values {
		expressions = append(expressions, postgres.UUID(val))
	}
	return expressions
}

// UUIDArray - конвертирует список uuid в список выражений для jet
func IntegerArray(values []int64) []postgres.Expression {
	expressions := make([]postgres.Expression, 0, len(values))
	for _, val := range values {
		expressions = append(expressions, postgres.Int(val))
	}
	return expressions
}
