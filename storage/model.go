package storage

import "errors"

type LinkStorage struct {
	Store map[string]string
}

var NotFoundErr = errors.New("Not Found")

func New() *LinkStorage {
	return &LinkStorage{}
}

func (s *LinkStorage) GetByID(id string) (string, error) {
	if val, ok := s.Store[id]; ok {
		return val, nil
	} else {
		return "", NotFoundErr
	}
}

func (s *LinkStorage) SaveLink(id, original string) {
	s.Store[id] = original
}
