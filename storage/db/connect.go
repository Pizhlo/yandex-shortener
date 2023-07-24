package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Connect(databaseAddr string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), databaseAddr)
	if err != nil {
		return &pgx.Conn{}, err
	}

	return conn, nil

}
