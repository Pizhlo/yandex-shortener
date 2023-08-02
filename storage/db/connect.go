package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseAddr string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), databaseAddr)
	if err != nil {
		return &pgxpool.Pool{}, err
	}

	return conn, nil

}
