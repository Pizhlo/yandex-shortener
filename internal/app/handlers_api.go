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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	path, err := util.MakeURL(conf.FlagBaseAddr, short)
	if err != nil {
		fmt.Println("cannot make path", zap.Error(err))
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
		fmt.Println("cannot Marshal resp", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respJSON)
	if err != nil {
		fmt.Println("Write err = ", err)
		fmt.Println("cannot Write resp", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("respJSON = ", string(respJSON))

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

	for _, val := range requestArr {
		short := util.Shorten(val.URL)
		err := memory.SaveLink(ctx, val.ID, short, val.URL, conf.FlagSaveToFile, conf.FlagSaveToDB, db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := models.ResponseAPI{
			ID:       val.ID,
			ShortURL: short,
		}

		responseArr = append(responseArr, resp)

	}

	setHeader(w, "Content-Type", "application/json", http.StatusCreated)

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

	fmt.Println("respJSON = ", string(respJSON))

}
