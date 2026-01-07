package constants

type FeatureType int64

const (
	// UnknownFeatureType .
	UnknownFeatureType FeatureType = iota
	// FeatureCashOut30M сколько снял пользак за прошедшие 30 минут
	FeatureCashOut30M
	// FeatureInternalPartners30M скольким перевел за последние 30 минут по внутренним
	FeatureInternalPartners30M
	// FeatureSbpPartners30M скольким перевел за последние 30 минут по сбп
	FeatureSbpPartners30M
	// FeatureSpent3H сколько потратил за последние 3 часа
	FeatureSpent3H
)
