package storage

import (
	"context"
	"errors"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/google/uuid"
)

type LinkStorage struct {
	Store       []Link
	FileStorage FileStorage
	DB          Database
}

var ErrNotFound = errors.New("not found")

func New(logger log.Logger) (*LinkStorage, error) {
	linkStorage := &LinkStorage{}
	linkStorage.Store = []Link{}

	return linkStorage, nil
}

func (s *LinkStorage) RecoverData(logger log.Logger) error {
	links, err := s.FileStorage.RecoverData(logger)
	if err != nil {
		return err
	}
	s.Store = links
	return nil
}

func (s *LinkStorage) GetLinkByID(ctx context.Context, shortURL string, flagSaveToFile bool, flagSaveToDB bool, db *Database, logger log.Logger) (string, error) {
	logger.Sugar.Debug("GetLinkByID")

	logger.Sugar.Debug("shortURL = ", shortURL)
	logger.Sugar.Debug("s.Store = ", s.Store)

	if flagSaveToDB {
		return db.GetLinkByIDFromDB(ctx, shortURL, logger)
	}

	for _, val := range s.Store {
		if val.ShortURL == shortURL {
			return val.OriginalURL, nil
		}
	}

	return "", ErrNotFound
}

func (s *LinkStorage) SaveLink(ctx context.Context, id, shortURL, originalURL string, flagSaveToFile bool, flagSaveToDB bool, db *Database, logger log.Logger) error {
	logger.Sugar.Debug("SaveLink")

	logger.Sugar.Debug("shortURL = ", shortURL, "original URL = ", originalURL)

	link, err := makeLinkModel(id, shortURL, originalURL)
	if err != nil {
		return err
	}

	s.Store = append(s.Store, link)

	if flagSaveToFile {
		return s.FileStorage.SaveDataToFile(link, logger)
	} else if flagSaveToDB {
		return db.SaveLinkDB(ctx, link, logger)
	}

	return nil
}

func makeLinkModel(id, shortURL, originalURL string) (Link, error) {
	var realID uuid.UUID
	var err error

	if id == "" { // если запрос пришел через /shorten/batch, id уже есть, если нет - надо сгенерировать
		realID = uuid.New()
	} else {
		realID, err = uuid.Parse(id)
		if err != nil {
			return Link{}, err
		}
	}

	link := Link{
		ID:          realID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	return link, nil
}
