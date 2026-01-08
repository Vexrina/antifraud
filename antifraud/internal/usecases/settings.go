package usecases

import (
	"context"
	"errors"
	"strings"

	"antifraud/internal/app/model"

	"github.com/prometheus/client_golang/prometheus"
)

//go:generate mockgen -source=settings.go -destination=./mocks/rule_mock.go -package=mocks
type (
	Rule interface {
		Name() string
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

var FraudErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "processing",
		Subsystem: "fraud",
		Name:      "rule_errors_total",
		Help:      "Total fraud errors by rule, transaction type and reason",
	},
	[]string{"rule", "tx_type", "reason"},
)

func categorizeError(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "context_canceled"
	case strings.Contains(err.Error(), "declined"):
		return "declined_by_af"
	default:
		return "other"
	}
}

const (
	cashOutTransaction  = "CashOut"
	SbpTransaction      = "SbpOutgoing"
	InternalTransaction = "Internal"
)

func incrementMetric(err error, ruleName string, transactionType string) {
	reason := categorizeError(err) // уже используем твою функцию
	FraudErrors.WithLabelValues(
		ruleName,
		transactionType,
		reason,
	).Inc()
}
