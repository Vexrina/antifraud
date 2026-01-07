package usecases

import (
	"context"

	"golang.org/x/sync/errgroup"

	"antifraud/internal/app/model"
)

type CashOutCheck struct {
	*RuleChecker
}

func (i *CashOutCheck) Check(
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

			return currentRule.Check(ctx, transaction)
		})
	}

	return g.Wait()
}
