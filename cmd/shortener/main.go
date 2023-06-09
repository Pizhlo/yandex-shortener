package main

import (
	"net/http"

	"github.com/Pizhlo/yandex-shortener/config"
	internal "github.com/Pizhlo/yandex-shortener/internal/app"
	"github.com/Pizhlo/yandex-shortener/internal/app/compress"
	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

func main() {
	conf := config.ParseConfigAndFlags()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Sugar.Fatal("error while creating sugar: ", zap.Error(err))
	}
	defer logger.Sync()

	log.Sugar = *logger.Sugar()

	log.Sugar.Infow(
		"Starting server",
		"addr", conf.FlagRunAddr,
	)

	storage, err := storage.New(conf.FlagSaveToFile, conf.FlagPathToFile)
	if err != nil {
		log.Sugar.Fatal("error while creating storage: ", zap.Error(err))
	}

	if conf.FlagSaveToFile {
		defer storage.FileStorage.Close()
	}

	if err := http.ListenAndServe(conf.FlagRunAddr, Run(conf, storage)); err != nil {
		log.Sugar.Fatal("error while executing server: ", zap.Error(err))
	}
}

func Run(conf config.Config, store *storage.LinkStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(log.WithLogging)
	r.Use(compress.UnpackData)

	r.Use(middleware.Compress(5, "application/javascript",
		"application/json",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml"))

	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(store, rw, r)
	})
	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(store, rw, r, conf.FlagBaseAddr, conf.FlagSaveToFile)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveURLAPI(store, rw, r, conf.FlagBaseAddr, conf.FlagSaveToFile)
			})
		})
	})

	return r
}
