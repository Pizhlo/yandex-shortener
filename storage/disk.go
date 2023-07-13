package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
)

type Link struct {
	ID          uuid.UUID `json:"id"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

// {"uuid":"1","short_url":"4rSPg8ap","original_url":"http://yandex.ru"}

type FileStorage struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewFileStorage(filename string) (*FileStorage, error) {
	fileStorage := &FileStorage{}
	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		return fileStorage, err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fileStorage, err
	}

	fileStorage.file = file
	fileStorage.decoder = json.NewDecoder(file)
	fileStorage.encoder = json.NewEncoder(file)

	return fileStorage, nil
}

func (f *FileStorage) RecoverData() ([]Link, error) {
	fmt.Println("RecoverData")
	links := []Link{}

	for {
		var link Link
		if err := f.decoder.Decode(&link); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		fmt.Println(link)
		links = append(links, link)
	}

	return links, nil
}

func (f *FileStorage) SaveDataToFile(link Link) error {
	fmt.Println("SaveDataToFile")

	fmt.Printf("link: %#v\n", link)

	return f.encoder.Encode(&link)
}

func (f *FileStorage) Close() error {
	fmt.Println("Close")
	return f.file.Close()
}
