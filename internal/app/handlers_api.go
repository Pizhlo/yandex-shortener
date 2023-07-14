package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Pizhlo/yandex-shortener/config"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	"github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"go.uber.org/zap"
)

const uniqueViolation = `ERROR: duplicate key value violates unique constraint "urls_original_url_idx" (SQLSTATE 23505)`

func ReceiveURLAPI(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, conf config.Config, db *storage.Database) {
	fmt.Println("ReceiveURLAPI")
	var req models.Request

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		fmt.Println("cannot decode request JSON body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	short := util.Shorten(req.URL)

	err := memory.SaveLink(ctx, "", short, req.URL, conf.FlagSaveToFile, conf.FlagSaveToDB, db)
	if err != nil {
		if err.Error() == uniqueViolation {
			sendJSONRespSingleURL(w, conf.FlagBaseAddr, short, http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendJSONRespSingleURL(w, conf.FlagBaseAddr, short, http.StatusCreated)
}

func sendJSONRespSingleURL(w http.ResponseWriter, flagBaseAddr, short string, statusCode int) error {
	resp := models.Response{
		Result: "",
	}

	path, err := util.MakeURL(flagBaseAddr, short)
	if err != nil {
		fmt.Println("cannot make path", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	resp.Result = path

	setHeader(w, "Content-Type", "application/json", statusCode)

	respJSON, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Marshal err = ", err)
		fmt.Println("cannot Marshal resp", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	_, err = w.Write(respJSON)
	if err != nil {
		fmt.Println("Write err = ", err)
		fmt.Println("cannot Write resp", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	fmt.Println("respJSON = ", string(respJSON))

	return nil
}

func ReceiveManyURLAPI(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, conf config.Config, db *storage.Database) {
	fmt.Println("ReceiveManyURLAPI")

	var requestArr []models.RequestAPI
	var responseArr []models.ResponseAPI

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&requestArr); err != nil {
		fmt.Println("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	statusCode := http.StatusCreated
	var path string
	var shortURL string

	for _, val := range requestArr {
		resp := models.ResponseAPI{ID: val.ID}

		err := memory.SaveLink(ctx, val.ID, shortURL, val.URL, conf.FlagSaveToFile, conf.FlagSaveToDB, db)
		if err != nil {
			if err.Error() == uniqueViolation {
				fmt.Println("unique err: ", err)
				statusCode = http.StatusConflict

				shortURL, err = db.GetShortURL(ctx, val.URL)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				resp.ShortURL = shortURL

			} else { // if error is not unique violation
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		} else { // if no errors occured
			shortURL = util.Shorten(val.URL)
		}

		path, err = util.MakeURL(conf.FlagBaseAddr, shortURL)
		if err != nil {
			fmt.Println("cannot make path", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.ShortURL = path

		responseArr = append(responseArr, resp)

	}

	setHeader(w, "Content-Type", "application/json", statusCode)

	respJSON, err := json.Marshal(responseArr)
	if err != nil {
		fmt.Println("Marshal err = ", err)
		fmt.Println("cannot Marshal resp", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respJSON)
	if err != nil {
		fmt.Println("Write err = ", err)
		fmt.Println("cannot Write resp", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("respJSON Many URL= ", string(respJSON))

}
