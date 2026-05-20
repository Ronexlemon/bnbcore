package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ronexlemon/bnbcore/internal/config"
)


type PostgresConn struct{
	Pool *pgxpool.Pool
}


func NewPostgressConnect(ctx context.Context,cfg *config.Config)(*PostgresConn,error){
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
	
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour


	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &PostgresConn{
		Pool: pool,
	}, nil


}