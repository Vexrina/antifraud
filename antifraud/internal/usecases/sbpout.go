package usecases

import (
	"context"

	"golang.org/x/sync/errgroup"

	"antifraud/internal/app/model"
)

type SbpOutCheck struct {
	*RuleChecker
}

func (i *SbpOutCheck) Check(
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
				incrementMetric(err, rule.Name(), SbpTransaction)
			}
			return err
		})
	}

	return g.Wait()
}
