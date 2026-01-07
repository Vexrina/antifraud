package usecases

import (
	"context"

	"antifraud/internal/app/model"
)

//go:generate mockgen -source=settings.go -destination=./mocks/rule_mock.go -package=mocks
type (
	Rule interface {
		ShouldRun(ctx context.Context, transaction *model.DomainTransaction) bool
		Check(ctx context.Context, transaction *model.DomainTransaction) error
	}
	RuleChecker struct {
		rules []Rule
	}
)

func NewRuleChecker(rules []Rule) *RuleChecker {
	return &RuleChecker{
		rules: rules,
	}
}
