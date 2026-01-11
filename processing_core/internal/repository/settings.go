package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	connPool *pgxpool.Pool
	once     sync.Once
	initErr  error
)

func NewPool(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}


func ConnectToPostgres(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	once.Do(func() {
		config, err := pgxpool.ParseConfig(connString)
		if err != nil {
			initErr = fmt.Errorf("failed to parse connection string: %v", err)
			return
		}

		connPool, initErr = pgxpool.NewWithConfig(ctx, config)
		if initErr != nil {
			initErr = fmt.Errorf("failed to connect to database: %v", initErr)
			return
		}
	})
	return connPool, initErr
}

type pgDb struct {
	conn *pgxpool.Pool
}

func NewDB(ctx context.Context, connStr string) *pgDb {
	connection, err := ConnectToPostgres(ctx, connStr)
	if err != nil {
		panic(err)
	}
	return &pgDb{
		conn: connection,
	}
}
