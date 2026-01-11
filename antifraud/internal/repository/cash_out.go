package repository

import (
	"context"
	"errors"

	"github.com/gocql/gocql"
	"gopkg.in/inf.v0"

	"antifraud/internal/constants"
)

func (c *CassandraFeatureStore) getCashOut30M(
	ctx context.Context,
	f *constants.FeatureFilter,
) ([]int64, error) {
	var totalDecimal inf.Dec

	q := c.session.Query(`
		SELECT total_cashout
		FROM cash_out_30m
		WHERE user_id = ?
		ORDER BY window_start DESC
		LIMIT 1
	`, f.UserID).WithContext(ctx)

	if err := q.Scan(&totalDecimal); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return []int64{0}, nil
		}
		return nil, err
	}

	// Конвертируем inf.Dec в int64, отбрасывая дробную часть
	totalInt64 := totalDecimal.UnscaledBig().Int64() / pow10(int32(totalDecimal.Scale()))

	return []int64{totalInt64}, nil
}

// pow10 возвращает 10^scale
func pow10(scale int32) int64 {
	result := 1
	for i := int32(0); i < scale; i++ {
		result *= 10
	}
	return int64(result)
}
