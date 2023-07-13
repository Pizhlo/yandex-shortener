package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type LinkStorage struct {
	Store       []Link
	FileStorage FileStorage
}

var ErrNotFound = errors.New("not found")

func New(flag bool, filename string) (*LinkStorage, error) {
	linkStorage := &LinkStorage{}
	linkStorage.Store = []Link{}

	if flag {
		fileStorage, err := NewFileStorage(filename)
		if err != nil {
			return linkStorage, err
		}
		linkStorage.FileStorage = *fileStorage
		if err := linkStorage.RecoverData(); err != nil {
			return linkStorage, err
		}
	}

	return linkStorage, nil
}

func (s *LinkStorage) RecoverData() error {
	links, err := s.FileStorage.RecoverData()
	if err != nil {
		return err
	}
	s.Store = links
	return nil
}

func (s *LinkStorage) GetLinkByID(ctx context.Context, shortURL string, flagSaveToFile bool, flagSaveToDB bool, db *Database) (string, error) {
	fmt.Println("GetLinkByID")

	fmt.Println("shortURL = ", shortURL)
	fmt.Println("s.Store = ", s.Store)

	if flagSaveToDB {
		return db.GetLinkByIDFromDB(ctx, shortURL)
	}

	for _, val := range s.Store {
		if val.ShortURL == shortURL {
			return val.OriginalURL, nil
		}
	}

	return "", ErrNotFound
}

func (s *LinkStorage) SaveLink(ctx context.Context, shortURL, originalURL string, flagSaveToFile bool, flagSaveToDB bool, db *Database) error {
	fmt.Println("SaveLink")

	fmt.Println("shortURL = ", shortURL, "original URL = ", originalURL)

	link := makeLinkModel(shortURL, originalURL)

	s.Store = append(s.Store, link)

	if flagSaveToFile {
		return s.FileStorage.SaveDataToFile(link)
	} else if flagSaveToDB {
		return db.SaveLinkDB(ctx, link)
	}

	return nil
}

func makeLinkModel(shortURL, originalURL string) Link {
	link := Link{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	return link
}
