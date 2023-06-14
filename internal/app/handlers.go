package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	store "github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
)

func ReceiveURL(storage *store.LinkStorage, w http.ResponseWriter, r *http.Request, baseURL string) {
	fmt.Println("ReceiveUrl")
	// сократить ссылку
	// записать в базу
	j, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	short := util.Shorten(string(j))

	storage.SaveLink(short, string(j))

	fmt.Println("ReceiveUrl storage =", storage.Store)
	fmt.Println("ReceiveUrl baseURL =", baseURL)
	fmt.Println("r.Host =", r.Host)

	fmt.Println("ReceiveUrl baseURL =", baseURL)

	path, err := util.MakeURL(baseURL, short)
	if err != nil {
		fmt.Println("err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(path))
}

func GetURL(storage *store.LinkStorage, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")
	id := chi.URLParam(r, "id")

	// проверить наличие ссылки в базе
	// выдать ссылку

	if val, err := storage.GetByID(id); err == nil {
		setLocation(w, val)
	} else if errors.Is(err, store.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func setLocation(w http.ResponseWriter, addr string) {
	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
