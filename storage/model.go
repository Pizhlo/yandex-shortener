package storage

import "errors"

type LinkStorage struct {
	Store map[string]string
}

var ErrNotFound = errors.New("not found")

func New() *LinkStorage {
	return &LinkStorage{
		Store: map[string]string{},
	}
}

func (s *LinkStorage) GetByID(id string) (string, error) {
	if val, ok := s.Store[id]; ok {
		return val, nil
	} else {
		return "", ErrNotFound
	}
}

func (s *LinkStorage) SaveLink(id, original string) {
	s.Store[id] = original
}
