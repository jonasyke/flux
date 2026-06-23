package main

import(
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonasyke/flux/db"
)

type DBClient struct {
	Pool	*pgxpool.Pool
	Queries	*db.Queries
}

func NewDatabaseConnection(databaseURL string) (*DBClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}
	
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}


	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to the PostgresSQL database!")

	queries := db.New(pool)

	return &DBClient{
		Pool: pool,
		Queries: queries,
	}, nil
}

func (c *DBClient) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}