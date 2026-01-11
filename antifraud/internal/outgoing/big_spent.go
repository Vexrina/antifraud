package outgoing

import (
	"context"
	"fmt"

	"antifraud/internal/app/model"
	"antifraud/internal/constants"
)

type bigSpent struct {
	name      string
	isOn      bool
	amountThs int64
	largeThs  int64

	fs FeatureStore
}

func NewBigSpent(
	isOn bool,
	amountThs,
	largeThs int64,
	fs FeatureStore,
) *bigSpent {
	return &bigSpent{
		name:      "big_spent",
		isOn:      isOn,
		amountThs: amountThs,
		largeThs:  largeThs,
		fs:        fs,
	}
}

func (l *bigSpent) Name() string {
	return l.name
}

func (l *bigSpent) ShouldRun(_ context.Context, transaction *model.DomainTransaction) bool {
	if !l.isOn {
		return false
	}

	if transaction.Amount < l.amountThs {
		return false
	}

	return true
}

func (l *bigSpent) Check(ctx context.Context, transaction *model.DomainTransaction) error {
	filter := mapDomainTransactionToCassandra(*transaction)

	feature, err := l.fs.GetFeatureInteger(ctx, constants.FeatureSpent3H, filter)
	if err != nil {
		return nil
	}

	if feature == nil && len(feature) < 1 {
		return nil
	}
	if transaction.Amount+feature[0] < l.largeThs {
		return nil
	}
	return fmt.Errorf("declined. [%s] amount+spent 3h more than ths [%v > %v]", l.name, transaction.Amount+feature[0], l.largeThs)
}
