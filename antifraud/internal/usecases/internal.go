package usecases

import (
	"context"

	"antifraud/internal/app/model"

	"golang.org/x/sync/errgroup"
)

type InternalCheck struct {
	*RuleChecker
}

func (i *InternalCheck) Check(
	ctx context.Context,
	transaction *model.DomainTransaction,
) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, rule := range i.rules {
		currentRule := rule
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if !currentRule.ShouldRun(ctx, transaction) {
				return nil
			}

			err := currentRule.Check(ctx, transaction)
			if err != nil {
				incrementMetric(err, rule.Name(), InternalTransaction)
			}
			return err
		})
	}

	return g.Wait()
}
