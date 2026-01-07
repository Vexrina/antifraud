package repository

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"
)

func ensureTestKeyspace(t *testing.T) {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Consistency = gocql.Quorum
	// Подключаемся к default keyspace
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })

	err = session.Query(`
		CREATE KEYSPACE IF NOT EXISTS antifraud_test
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
	`).Exec()
	require.NoError(t, err)
}

func newTestSession(t *testing.T) *gocql.Session {
	ensureTestKeyspace(t)

	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "antifraud_test"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

func mustUUID() gocql.UUID {
	u, _ := gocql.RandomUUID()
	return u
}
