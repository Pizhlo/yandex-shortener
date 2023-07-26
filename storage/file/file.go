package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	storage "github.com/Pizhlo/yandex-shortener/storage/memory"
	"github.com/Pizhlo/yandex-shortener/storage/model"
	"github.com/google/uuid"
)

type FileStorage struct {
	Memory  storage.Memory
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
	Logger  log.Logger
}

func New(filename string, logger log.Logger) (*FileStorage, error) {
	fileStorage := &FileStorage{}
	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		return fileStorage, err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fileStorage, err
	}

	memory, err := storage.New(logger)
	if err != nil {
		return fileStorage, err
	}

	fileStorage.Memory = *memory
	fileStorage.file = file
	fileStorage.decoder = json.NewDecoder(file)
	fileStorage.encoder = json.NewEncoder(file)
	fileStorage.Logger = logger

	links, err := fileStorage.RecoverData()
	if err != nil {
		return fileStorage, err
	}

	fileStorage.Memory.Store = links

	return fileStorage, nil
}

func (f *FileStorage) RecoverData() ([]model.Link, error) {
	f.Logger.Sugar.Debug("RecoverData")

	links := []model.Link{}

	for {
		var link model.Link
		if err := f.decoder.Decode(&link); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	//logger.Sugar.Debug("links = ", links)

	return links, nil
}

func (f *FileStorage) Save(ctx context.Context, link model.Link) error {
	f.Logger.Sugar.Debug("SaveDataToFile")

	f.Logger.Sugar.Debugf("link: %#v\n", link)

	if err := f.Memory.Save(ctx, link); err != nil {
		return err
	}

	return f.encoder.Encode(&link)
}

func (f *FileStorage) Get(ctx context.Context, short string) (string, error) {
	return f.Memory.Get(ctx, short)
}

func (f *FileStorage) GetUserURLS(ctx context.Context, userID uuid.UUID) ([]models.UserLinks, error) {
	return f.Memory.GetUserURLS(ctx, userID)
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
