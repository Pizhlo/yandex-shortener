package app

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	"github.com/Pizhlo/yandex-shortener/util"
	"go.uber.org/zap"
)

const uniqueViolation = `ERROR: duplicate key value violates unique constraint "urls_original_url_idx" (SQLSTATE 23505)`

func ReceiveURLAPI(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("ReceiveURLAPI")

	var req models.Request

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		handler.Logger.Sugar.Debug("ReceiveURLAPI cannot decode request JSON body; err = ", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	short := util.Shorten(req.URL)

	err := handler.Memory.SaveLink(ctx, "", short, req.URL, handler.FlagSaveToFile, handler.FlagSaveToDB, handler.DB, handler.Logger)
	if err != nil {
		if err.Error() == uniqueViolation {
			sendJSONRespSingleURL(w, handler.FlagBaseAddr, short, http.StatusConflict, handler.Logger)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendJSONRespSingleURL(w, handler.FlagBaseAddr, short, http.StatusCreated, handler.Logger)
}

func sendJSONRespSingleURL(w http.ResponseWriter, flagBaseAddr, short string, statusCode int, logger log.Logger) error {
	resp := models.Response{
		Result: "",
	}

	path, err := util.MakeURL(flagBaseAddr, short)
	if err != nil {
		logger.Sugar.Debug("cannot make path", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	resp.Result = path

	setHeader(w, "Content-Type", "application/json", statusCode)

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logger.Sugar.Debug("cannot Marshal resp: ", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	_, err = w.Write(respJSON)
	if err != nil {
		logger.Sugar.Debug("cannot Write resp: ", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	logger.Sugar.Debug("respJSON: ", string(respJSON))

	return nil
}

func ReceiveManyURLAPI(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("ReceiveManyURLAPI")

	var requestArr []models.RequestAPI
	var responseArr []models.ResponseAPI

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&requestArr); err != nil {
		handler.Logger.Sugar.Debug("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	statusCode := http.StatusCreated
	var path string

	for _, val := range requestArr {
		resp := models.ResponseAPI{ID: val.ID}
		shortURL := util.Shorten(val.URL)

		err := handler.Memory.SaveLink(ctx, val.ID, shortURL, val.URL, handler.FlagSaveToFile, handler.FlagSaveToDB, handler.DB, handler.Logger)
		if err != nil {
			if err.Error() == uniqueViolation {
				statusCode = http.StatusConflict
			} else { // if error is not unique violation
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

		path, err = util.MakeURL(handler.FlagBaseAddr, shortURL)
		if err != nil {
			handler.Logger.Sugar.Debug("cannot make path: ", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.ShortURL = path

		responseArr = append(responseArr, resp)

	}

	setHeader(w, "Content-Type", "application/json", statusCode)

	respJSON, err := json.Marshal(responseArr)
	if err != nil {
		handler.Logger.Sugar.Debug("cannot Marshal resp: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respJSON)
	if err != nil {
		handler.Logger.Sugar.Debug("cannot Write resp: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	handler.Logger.Sugar.Debug("respJSON Many UR: ", string(respJSON))

}
