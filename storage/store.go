package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Store interface {
	Ping() error
}

type Database struct {
	*pgx.Conn
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
	return db.PgConn().Ping(context.TODO())
}

func (db *Database) SaveLinkDB(ctx context.Context, link Link) error {
	fmt.Println("SaveLinkDB")
	return db.Ping()
}

func (db *Database) GetLinkByIDFromDB(ctx context.Context, short string) (string, error) {
	fmt.Println("GetLinkByIDFromDB")
	var originalURL string

	row := db.QueryRow(ctx, `SELECT original_url from urls where short_url = $1`, short)

	err := row.Scan(&originalURL)
	if err != nil {
		return originalURL, err
	}

	return originalURL, nil
}
