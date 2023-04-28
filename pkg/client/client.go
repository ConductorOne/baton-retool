package client

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Client struct {
	db *pgxpool.Pool
}

func (c *Client) ValidateConnection(ctx context.Context) error {
	err := c.db.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}

func New(ctx context.Context, dsn string) (*Client, error) {
	db, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	c := &Client{
		db: db,
	}

	return c, nil
}
