package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"

	"antifraud/internal/constants"
)

type CassandraFeatureStore struct {
	session *gocql.Session
}

// NewCassandraFeatureStore создаёт и возвращает новый экземпляр CassandraFeatureStore.
// host — адрес Cassandra (можно несколько через запятую или slice)
// keyspace — keyspace, с которым будет работать хранилище
func NewCassandraFeatureStore(hosts []string, keyspace string) (*CassandraFeatureStore, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	cluster.ReconnectInterval = 5 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create Cassandra session: %w", err)
	}

	return &CassandraFeatureStore{
		session: session,
	}, nil
}

// Close закрывает сессию Cassandra
func (c *CassandraFeatureStore) Close() {
	if c.session != nil {
		c.session.Close()
	}
}

func (c *CassandraFeatureStore) GetFeatureInteger(
	ctx context.Context,
	featureID constants.FeatureType,
	featureFilter *constants.FeatureFilter,
) ([]int64, error) {
	switch featureID {
	case constants.FeatureCashOut30M:
		return c.getCashOut30M(ctx, featureFilter)

	case constants.FeatureSpent3H:
		return c.getSpent3H(ctx, featureFilter)

	default:
		return nil, fmt.Errorf("unsupported integer feature: %v", featureID)
	}
}

func (c *CassandraFeatureStore) GetFeatureString(
	ctx context.Context,
	featureID constants.FeatureType,
	featureFilter *constants.FeatureFilter,
) ([]string, error) {
	switch featureID {
	case constants.FeatureInternalPartners30M:
		return c.getPartnersLastNBuckets(ctx, "internal_partners_30m", featureFilter, 5)

	case constants.FeatureSbpPartners30M:
		return c.getPartnersLastNBuckets(ctx, "sbp_partners_30m", featureFilter, 5)

	default:
		return nil, fmt.Errorf("unsupported string feature: %v", featureID)
	}
}
