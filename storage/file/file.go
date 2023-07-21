package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/Pizhlo/yandex-shortener/internal/app/logger"
	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/storage/model"
)

type FileStorage struct {
	FlagSaveToFile bool
	file           *os.File
	encoder        *json.Encoder
	decoder        *json.Decoder
}

func New(filename string) (*FileStorage, error) {
	fileStorage := &FileStorage{}
	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		return fileStorage, err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fileStorage, err
	}

	fileStorage.FlagSaveToFile = true
	fileStorage.file = file
	fileStorage.decoder = json.NewDecoder(file)
	fileStorage.encoder = json.NewEncoder(file)

	return fileStorage, nil
}

func (f *FileStorage) RecoverData(logger log.Logger) ([]model.Link, error) {
	logger.Sugar.Debug("RecoverData")

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

	return links, nil
}

func (f *FileStorage) Save(ctx context.Context, link model.Link, logger log.Logger) error {
	logger.Sugar.Debug("SaveDataToFile")

	logger.Sugar.Debugf("link: %#v\n", link)

	return f.encoder.Encode(&link)
}

func (f *FileStorage) Get(ctx context.Context, short string, logger logger.Logger) (string, error) {
	return "", nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
