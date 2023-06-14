package main

import (
	"fmt"
	"net/http"

	"github.com/Pizhlo/yandex-shortener/config"
	internal "github.com/Pizhlo/yandex-shortener/internal/app"
	"github.com/Pizhlo/yandex-shortener/storage"
	"github.com/go-chi/chi"
)

func main() {
	conf := config.ParseConfigAndFlags()

	fmt.Println("Running server on", conf.FlagRunAddr)

	http.ListenAndServe(conf.FlagRunAddr, Run(conf))
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
