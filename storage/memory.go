package storage

import (
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
	linkStorage.Store = []Link{}
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

func (s *LinkStorage) GetLinkByID(shortURL string) (string, error) {
	fmt.Println("GetLinkByID")
	for _, val := range s.Store {
		if val.ShortURL == shortURL {
			return val.OriginalURL, nil
		}
	}

	return "", ErrNotFound
}

func (s *LinkStorage) SaveLink(shortURL, originalURL string, flag bool) error {
	fmt.Println("SaveLink")
	link := Link{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	s.Store = append(s.Store, link)

	if flag {
		return s.FileStorage.SaveDataToFile(link)
	}
	return nil
}
