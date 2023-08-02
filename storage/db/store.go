package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	errs "github.com/Pizhlo/yandex-shortener/storage/errors"
	"github.com/Pizhlo/yandex-shortener/storage/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Store interface {
	CreateTableURLs() error
	Ping() error
	Save(ctx context.Context, link model.Link) error
	Get(ctx context.Context, short string) (string, error)
	DeleteURLS(ctx context.Context, link ...models.DeleteLink) error
}

type URLStorage struct {
	*pgx.Conn
	Logger log.Logger
}

func New(conn *pgx.Conn, logger log.Logger) (*URLStorage, error) {
	db := &URLStorage{conn, logger}

	return db, db.CreateTableURLs()

}

func (db *URLStorage) CreateTableURLs() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	q := `CREATE TABLE IF NOT EXISTS urls(id uuid NOT NULL,
		short_url text NOT NULL,
		original_url text NOT NULL,
		"user" uuid NOT NULL,
		is_deleted BOOLEAN NOT NULL DEFAULT FALSE
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

func (db *URLStorage) Save(ctx context.Context, link model.Link) error {
	db.Logger.Sugar.Debug("SaveLinkDB")

	db.Logger.Sugar.Debugf("INSERT INTO urls (id, short_url, original_url, user) VALUES(%s, %s, %s, %s)\n", link.ID, link.ShortURL, link.OriginalURL, link.UserID)

	q := `INSERT INTO urls (id, short_url, original_url, "user") VALUES($1, $2, $3, $4)`

	_, err := db.Exec(ctx, q, link.ID, link.ShortURL, link.OriginalURL, link.UserID)
	if err != nil {
		db.Logger.Sugar.Debug("SaveLinkDB err = ", err)
		return err
	}

	return nil
}

func (db *URLStorage) Get(ctx context.Context, short string) (string, bool, error) {
	db.Logger.Sugar.Debug("GetLinkByIDFromDB")

	var originalURL string
	var deleted bool

	db.Logger.Sugar.Debugf("SELECT original_url from urls where short_url = %s\n", short)

	err := db.QueryRow(ctx, `SELECT original_url, is_deleted from urls where short_url = $1`, short).Scan(&originalURL, &deleted)

	if err != nil {
		db.Logger.Sugar.Debug("GetLinkByIDFromDB err = ", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return originalURL, false, errs.ErrNotFound
		}
		return originalURL, false, err
	}

	return originalURL, deleted, nil
}

func (db *URLStorage) GetUserURLS(ctx context.Context, userID uuid.UUID) ([]models.UserLinks, error) {
	db.Logger.Sugar.Debug("GetUserURLS")

	var result []models.UserLinks

	db.Logger.Sugar.Debugf("SELECT short_url, original_url FROM urls WHERE user=%s\n", userID)

	rows, err := db.Query(ctx, `SELECT short_url, original_url FROM urls WHERE "user"=$1`, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, errs.ErrNotFound
		}
		db.Logger.Sugar.Debug("GetUserURLS err = ", err)
		return result, err
	}

	defer rows.Close()

	for rows.Next() {
		var row models.UserLinks
		err := rows.Scan(&row.ShortURL, &row.OriginalURL)
		if err != nil {
			db.Logger.Sugar.Debug("rows.Scan err = ", err)
			return result, err
		}
		result = append(result, row)
	}

	db.Logger.Sugar.Debug("GetUserURLS result = ", result)

	if len(result) == 0 {
		return result, errs.ErrNotFound
	}

	return result, nil
}

func (db *URLStorage) DeleteURLS(ctx context.Context, link ...model.DeleteLink) error {
	var q string

	for _, link := range link {
		q += fmt.Sprintf(`UPDATE "urls" SET "is_deleted" = true 
  WHERE "user"='%s' AND "short_url" = '%s';`, link.UserID, link.ShortURL)
	}

	reader := db.PgConn().Exec(ctx, q)

	defer reader.Close()

	return nil
}
