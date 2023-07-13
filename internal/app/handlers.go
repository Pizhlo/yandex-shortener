package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	"github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func ReceiveURLAPI(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, baseURL string, flag bool) {
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

	memory.SaveLink(short, req.URL, flag)

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

	setHeader(w, "Content-Type", "application/json", http.StatusCreated)

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

func ReceiveURL(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, baseURL string, flag bool) {
	fmt.Println("ReceiveUrl")

	// сократить ссылку
	// записать в базу

	j, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	short := util.Shorten(string(j))

	memory.SaveLink(short, string(j), flag)

	path, err := util.MakeURL(baseURL, short)
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setHeader(w, "Content-Type", "text/plain", http.StatusCreated)
	w.Write([]byte(path))
}

func GetURL(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")

	// проверить наличие ссылки в базе
	// выдать ссылку

	id := chi.URLParam(r, "id")
	val, err := memory.GetLinkByID(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setHeader(w, "Location", val, http.StatusTemporaryRedirect)
}

func Ping(w http.ResponseWriter, r *http.Request, db *storage.Database) {
	// ping
	err := db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func setHeader(w http.ResponseWriter, header string, val string, statusCode int) {
	w.Header().Set(header, val)
	w.WriteHeader(statusCode)
}
