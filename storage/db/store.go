package db

import (
	"context"
	"time"
	"errors"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/storage/model"
	errs "github.com/Pizhlo/yandex-shortener/storage/errors"
	"github.com/jackc/pgx/v5"
)

type Store interface {
	CreateTableURLs() error
	Ping() error
	Save(ctx context.Context, link model.Link, logger log.Logger) error
	Get(ctx context.Context, short string, logger log.Logger) (string, error)
}

type URLStorage struct {
	*pgx.Conn
}

func New(conn *pgx.Conn) (*URLStorage, error) {
	db := &URLStorage{conn}	

	return db, db.CreateTableURLs()

}

func (db *URLStorage) CreateTableURLs() error {
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

func (db *URLStorage) Ping(ctx context.Context) error {
	return db.PgConn().Ping(ctx)
}

func (db *URLStorage) Save(ctx context.Context, link model.Link, logger log.Logger) error {
	logger.Sugar.Debug("SaveLinkDB")

	logger.Sugar.Debugf("INSERT INTO urls (id, short_url, original_url) VALUES(%s, %s, %s)\n", link.ID, link.ShortURL, link.OriginalURL)

	q := `INSERT INTO urls (id, short_url, original_url) VALUES($1, $2, $3)`

	_, err := db.Exec(ctx, q, link.ID, link.ShortURL, link.OriginalURL)
	if err != nil {
		logger.Sugar.Debug("SaveLinkDB err = ", err)
		return err
	}

	return nil
}

func (db *URLStorage) Get(ctx context.Context, short string, logger log.Logger) (string, error) {
	logger.Sugar.Debug("GetLinkByIDFromDB")

	var originalURL string

	err := db.QueryRow(ctx, `SELECT original_url from urls where short_url = $1`, short).Scan(&originalURL)

	if err != nil {
		logger.Sugar.Debug("GetLinkByIDFromDB err = ", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return originalURL, errs.ErrNotFound
		}
		return originalURL, err
	}

	return originalURL, nil
}
