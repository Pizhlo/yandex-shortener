package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Pizhlo/yandex-shortener/config"
	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	"github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func ReceiveURLAPI(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, conf config.Config, db *storage.Database) {
	fmt.Println("ReceiveURLAPI")
	var req models.Request

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		log.Sugar.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	short := util.Shorten(req.URL)

	memory.SaveLink(ctx, short, req.URL, conf.FlagSaveToFile, conf.FlagSaveToDB, db)

	path, err := util.MakeURL(conf.FlagBaseAddr, short)
	if err != nil {
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

func ReceiveURL(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, conf config.Config, db *storage.Database) {
	fmt.Println("ReceiveUrl")

	// сократить ссылку
	// записать в базу

	j, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	short := util.Shorten(string(j))

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	memory.SaveLink(ctx, short, string(j), conf.FlagSaveToFile, conf.FlagSaveToDB, db)

	path, err := util.MakeURL(conf.FlagBaseAddr, short)
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setHeader(w, "Content-Type", "text/plain", http.StatusCreated)
	w.Write([]byte(path))
}

func GetURL(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, conf config.Config, db *storage.Database) {
	fmt.Println("GetUrl")

	// проверить наличие ссылки в базе
	// выдать ссылку

	id := chi.URLParam(r, "id")

	fmt.Println("url = ", id)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	val, err := memory.GetLinkByID(ctx, id, conf.FlagSaveToFile, conf.FlagSaveToDB, db)
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
