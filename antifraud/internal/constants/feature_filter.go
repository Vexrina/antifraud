package constants

import "github.com/gocql/gocql"

type FeatureFilter struct {
	UserID gocql.UUID
	Limit  int
}
