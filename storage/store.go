package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Store interface {
	Ping() error
}

type Database struct {
	postgreSQL *pgx.Conn
}

func NewStore(databaseAddr string) (*Database, error) {
	var db *Database
	if databaseAddr != "" {

		conn, err := pgx.Connect(context.Background(), databaseAddr)
		if err != nil {
			return db, err
		}

		db = &Database{conn}

		return db, nil
	}

	return db, nil

}

func (db *Database) Ping() error {
	return db.postgreSQL.Ping(context.Background())
}
