package service

import (
	"context"

	"github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/storage/model"
)

type Service struct {
	Storage Storage
}

type Storage interface {
	Get(ctx context.Context, short string, logger logger.Logger) (string, error)
	Save(ctx context.Context, link model.Link, logger logger.Logger) error
}

func New(storage Storage) *Service {
	return &Service{
		Storage: storage,
	}
}
