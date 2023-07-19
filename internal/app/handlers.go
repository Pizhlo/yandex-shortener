package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
)

type Handler struct {
	Memory         *storage.LinkStorage
	DB             *storage.Database
	FlagBaseAddr   string
	FlagPathToFile string
	FlagSaveToFile bool
	FlagSaveToDB   bool
}

func ReceiveURL(handler Handler, w http.ResponseWriter, r *http.Request) {
	fmt.Println("ReceiveUrl")

	// сократить ссылку
	// записать в базу

	j, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := http.StatusCreated
	shortURL := util.Shorten(string(j))

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := handler.Memory.SaveLink(ctx, "", shortURL, string(j), handler.FlagSaveToFile, handler.FlagSaveToDB, handler.DB); err != nil {
		if err.Error() == uniqueViolation {
			statusCode = http.StatusConflict

		}
		fmt.Println("ReceiveUrl SaveLink err = ", err)
	}

	fmt.Println("code = ", statusCode)

	path, err := util.MakeURL(handler.FlagBaseAddr, shortURL)
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setHeader(w, "Content-Type", "text/plain", statusCode)
	w.Write([]byte(path))
}

func GetURL(handler Handler, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")

	// проверить наличие ссылки в базе
	// выдать ссылку

	id := chi.URLParam(r, "id")

	fmt.Println("url = ", id)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	val, err := handler.Memory.GetLinkByID(ctx, id, handler.FlagSaveToFile, handler.FlagSaveToDB, handler.DB)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setHeader(w, "Location", val, http.StatusTemporaryRedirect)
}

func Ping(w http.ResponseWriter, r *http.Request, db *storage.Database, flagDB bool) {
	// ping

	if flagDB {
		err := db.Ping()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}

}

func setHeader(w http.ResponseWriter, header string, val string, statusCode int) {
	w.Header().Set(header, val)
	w.WriteHeader(statusCode)
}
