package outgoing

import (
	"context"
	"fmt"

	"antifraud/internal/app/model"
	"antifraud/internal/constants"
)

type largeCashOutDuring1h struct {
	name      string
	isOn      bool
	amountThs int64
	largeThs  int64
	fs        FeatureStore
}

func NewLargeCashOutDuring1H(
	isOn bool,
	amountThs,
	largeThs int64,
	fs FeatureStore,
) *largeCashOutDuring1h {
	return &largeCashOutDuring1h{
		name:      "large_cash_out_during_1h",
		isOn:      isOn,
		amountThs: amountThs,
		largeThs:  largeThs,
		fs:        fs,
	}
}

func (l *largeCashOutDuring1h) ShouldRun(_ context.Context, transaction *model.DomainTransaction) bool {
	if !l.isOn {
		return false
	}

	if transaction.Amount < l.amountThs {
		return false
	}

	return true
}

func (l *largeCashOutDuring1h) Check(ctx context.Context, transaction *model.DomainTransaction) error {
	filter := mapDomainTransactionToCassandra(*transaction)

	feature, err := l.fs.GetFeatureInteger(ctx, constants.FeatureCashOut30M, filter)
	if err != nil {
		return err
	}

	if feature == nil {
		return nil
	}
	sumOfFeature := int64(0)
	for _, feat := range feature {
		sumOfFeature += feat
	}
	if sumOfFeature+transaction.Amount < l.largeThs {
		return nil
	}
	return fmt.Errorf("declined, [%s] sum more than ths [%v > %v]", l.name, sumOfFeature, l.largeThs)
}
