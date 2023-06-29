package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	store "github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func ReceiveURLAPI(storage *store.LinkStorage, w http.ResponseWriter, r *http.Request, baseURL string) {
	fmt.Println("ReceiveURLAPI")
	var req models.Request

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		fmt.Println("Decode err = ", err)
		log.Sugar.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	short := util.Shorten(req.URL)

	storage.SaveLink(short, req.URL)

	path, err := util.MakeURL(baseURL, short)
	if err != nil {
		fmt.Println("path err = ", err)
		log.Sugar.Debug("cannot make path", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := models.Response{
		Result: path,
	}

	setHeader(w, "Content-Type", "application/json", http.StatusOK)

	respJSON, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Marshal err = ", err)
		log.Sugar.Debug("cannot Marshal resp", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respJSON)
	if err != nil {
		fmt.Println("Write err = ", err)
		log.Sugar.Debug("cannot Write resp", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("respJSON = ", string(respJSON))

}

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

	path, err := util.MakeURL(baseURL, short)
	if err != nil {
		fmt.Println("err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setHeader(w, "Content-Type", "text/plain", http.StatusCreated)
	w.Write([]byte(path))
}

func GetURL(storage *store.LinkStorage, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")

	// проверить наличие ссылки в базе
	// выдать ссылку

	id := chi.URLParam(r, "id")
	val, err := storage.GetLinkByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setHeader(w, "Location", val, http.StatusTemporaryRedirect)
}

func setHeader(w http.ResponseWriter, header string, val string, statusCode int) {
	w.Header().Set(header, val)
	w.WriteHeader(statusCode)
}
