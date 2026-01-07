package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
	"gopkg.in/inf.v0"

	"antifraud/internal/constants"
)

func (c *CassandraFeatureStore) getSpent3H(
	ctx context.Context,
	f *constants.FeatureFilter,
) ([]int64, error) {
	var totalDecimal inf.Dec

	q := c.session.Query(`
		SELECT total_spent
		FROM spent_3h
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

	totalInt64 := totalDecimal.UnscaledBig().Int64() / pow10(int32(totalDecimal.Scale()))
	return []int64{totalInt64}, nil
}

func (c *CassandraFeatureStore) getPartnersLastNBuckets(
	ctx context.Context,
	table string,
	filter *constants.FeatureFilter,
	n int,
) ([]string, error) {

	iter := c.session.Query(fmt.Sprintf(`
		SELECT partner_id
		FROM %s
		WHERE user_id = ?
		LIMIT ?
	`, table), filter.UserID, n).WithContext(ctx).Iter()

	var partnerID gocql.UUID
	var partners []string
	for iter.Scan(&partnerID) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			partners = append(partners, partnerID.String())
		}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	if len(partners) == 0 {
		return nil, nil
	}

	return partners, nil
}
