package model

import "github.com/google/uuid"

type Link struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	IsDeleted   bool      `json:"is_deleted"`
}

type DeleteLink struct {
	UserID   uuid.UUID `json:"user_id"`
	ShortURL string    `json:"short_url"`
}

func MakeLinkModel(id string, userID uuid.UUID, shortURL, originalURL string) (Link, error) {
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
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	return link, nil
}
