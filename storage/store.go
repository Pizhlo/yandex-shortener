package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

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

		return db, db.createTableURLs()
	}

	return db, nil

}

func (db *Database) createTableURLs() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	q := `CREATE TABLE IF NOT EXISTS urls(id uuid NOT NULL,
		short_url text NOT NULL,
		original_url text NOT NULL
	);
	
CREATE UNIQUE INDEX ON "urls" ("original_url");`

	txOptions := pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}

	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	tx.Exec(ctx, q)
	defer tx.Rollback(ctx)

	return tx.Commit(ctx)
}

func (db *Database) Ping() error {
	return db.PgConn().Ping(context.TODO())
}

func (db *Database) SaveLinkDB(ctx context.Context, link Link) error {
	fmt.Println("SaveLinkDB")

	fmt.Printf("INSERT INTO urls (id, short_url, original_url) VALUES(%s, %s, %s)\n", link.ID, link.ShortURL, link.OriginalURL)

	q := `INSERT INTO urls (id, short_url, original_url) VALUES($1, $2, $3)`

	_, err := db.Exec(ctx, q, link.ID, link.ShortURL, link.OriginalURL)
	if err != nil {
		fmt.Println("SaveLinkDB err = ", err)
		return err
	}

	return nil
}

func (db *Database) GetLinkByIDFromDB(ctx context.Context, short string) (string, error) {
	fmt.Println("GetLinkByIDFromDB")
	var originalURL string

	row := db.QueryRow(ctx, `SELECT original_url from urls where short_url = $1`, short)

	err := row.Scan(&originalURL)
	if err != nil {
		fmt.Println("GetLinkByIDFromDB err = ", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return originalURL, ErrNotFound
		}
		return originalURL, err
	}

	return originalURL, nil
}
