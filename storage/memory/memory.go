package storage

import (
	"context"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/storage/errors"
	"github.com/Pizhlo/yandex-shortener/storage/model"
)

type LinkStorage struct {
	Store []model.Link
}

func New(logger log.Logger) (*LinkStorage, error) {
	linkStorage := &LinkStorage{}
	linkStorage.Store = []model.Link{}

	return linkStorage, nil
}

func (s *LinkStorage) Get(ctx context.Context, short string, logger log.Logger) (string, error) {
	logger.Sugar.Debug("GetLinkByID")

	logger.Sugar.Debug("shortURL = ", short)
	logger.Sugar.Debug("s.Store = ", s.Store)

	for _, val := range s.Store {
		if val.ShortURL == short {
			return val.OriginalURL, nil
		}
	}

	return "", errors.ErrNotFound
}

func (s *LinkStorage) Save(ctx context.Context, link model.Link, logger log.Logger) error {
	logger.Sugar.Debug("SaveLink")

	logger.Sugar.Debug("shortURL = ", link.ShortURL, "original URL = ", link.OriginalURL)

	s.Store = append(s.Store, link)

	// if s.FileStorage.FlagSaveToFile {
	// 	return s.FileStorage.SaveDataToFile(link, logger)
	// } else if s.DB.FlagSaveToDB {
	// 	return db.SaveLinkDB(ctx, link, logger)
	// }

	return nil
}
