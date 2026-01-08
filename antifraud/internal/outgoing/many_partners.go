package outgoing

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"antifraud/internal/app/model"
	"antifraud/internal/constants"
)

type manyPartners struct {
	name      string
	isOn      bool
	amountThs int64
	partners  int

	mode int // sbp/internal
	fs   FeatureStore
}

func NewManyPartners(
	isOn bool,
	amountThs int64,
	partners,
	mode int,
	fs FeatureStore,
) *manyPartners {
	return &manyPartners{
		name:      "many_partners",
		isOn:      isOn,
		amountThs: amountThs,
		partners:  partners,
		mode:      mode,
		fs:        fs,
	}
}

func (l *manyPartners) Name() string {
	return fmt.Sprintf("%s_%s", l.name, lo.Ternary(l.mode == 0, "internal", "sbp"))
}

func (l *manyPartners) ShouldRun(_ context.Context, transaction *model.DomainTransaction) bool {
	if !l.isOn {
		return false
	}

	if transaction.Amount < l.amountThs {
		return false
	}

	return true
}

func (l *manyPartners) Check(ctx context.Context, transaction *model.DomainTransaction) error {
	filter := mapDomainTransactionToCassandra(*transaction)

	featureName := lo.Ternary(l.mode == 0, constants.FeatureInternalPartners30M, constants.FeatureSbpPartners30M)
	feature, err := l.fs.GetFeatureString(ctx, featureName, filter)
	if err != nil {
		return err
	}
	if feature == nil {
		return nil
	}

	uniqPartners := map[string]struct{}{}
	for _, feat := range feature {
		uniqPartners[feat] = struct{}{}
	}

	if len(uniqPartners) < l.partners {
		return nil
	}
	return fmt.Errorf(
		"declined, [%s] number of partners more than ths [%v > %v]",
		l.Name(),
		len(uniqPartners),
		l.partners,
	)
}
