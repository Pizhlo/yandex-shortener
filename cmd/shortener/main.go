package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Pizhlo/yandex-shortener/config"
	internal "github.com/Pizhlo/yandex-shortener/internal/app"
	"github.com/Pizhlo/yandex-shortener/storage"
	"github.com/go-chi/chi"
)

func main() {
	conf := config.ParseConfigAndFlags()

	fmt.Println("Running server on", conf.FlagRunAddr)

	if err := http.ListenAndServe(conf.FlagRunAddr, Run(conf)); err != nil {
		log.Panic("error while executing server: %w\n", err)
	}
}

func Run(conf config.Config) chi.Router {
	storage := storage.New()

	r := chi.NewRouter()
	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(storage, rw, r)
	})
	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(storage, rw, r, conf.FlagBaseAddr)
	})

	return r
}
