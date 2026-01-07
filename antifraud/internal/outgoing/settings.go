package outgoing

import (
	"context"

	"github.com/gocql/gocql"

	"antifraud/internal/app/model"
	"antifraud/internal/constants"
)

//go:generate mockgen -source=settings.go -destination=./mocks/feature_store_mock.go -package=mocks
type (
	FeatureStore interface {
		GetFeatureInteger(ctx context.Context, featureID constants.FeatureType, featureFilter constants.FeatureFilter) ([]int64, error)
		GetFeatureString(ctx context.Context, featureID constants.FeatureType, featureFilter constants.FeatureFilter) ([]string, error)
	}
)

func mapDomainTransactionToCassandra(transaction model.DomainTransaction) constants.FeatureFilter {
	cassandraUUID, _ := gocql.ParseUUID(transaction.SenderID.String())
	return constants.FeatureFilter{
		UserID: cassandraUUID,
		Limit:  1,
	}
}
