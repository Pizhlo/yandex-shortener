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

	memory, err := storage.New(conf.FlagSaveToFile, conf.FlagPathToFile) // in-memory and file storage
	if err != nil {
		log.Sugar.Fatal("error while creating storage: ", zap.Error(err))
	}

	db, err := storage.NewStore(conf.FlagDatabaseAddress)
	if err != nil {
		log.Sugar.Fatal("error while connecting db: ", zap.Error(err))
	}

	if conf.FlagSaveToFile {
		defer memory.FileStorage.Close()
	}

	if err := http.ListenAndServe(conf.FlagRunAddr, Run(conf, memory, db)); err != nil {
		log.Sugar.Fatal("error while executing server: ", zap.Error(err))
	}
}

func Run(conf config.Config, store *storage.LinkStorage, db *storage.Database) chi.Router {
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
		internal.GetURL(store, rw, r, conf, db)
	})

	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(store, rw, r, conf, db)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveURLAPI(store, rw, r, conf, db)
			})

			r.Post("/test", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveManyURLAPI(store, rw, r, conf, db)
			})

			r.Post("/batch", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveManyURLAPI(store, rw, r, conf, db)
			})
		})
	})

	r.Get("/ping", func(rw http.ResponseWriter, r *http.Request) {
		internal.Ping(rw, r, db, conf.FlagSaveToDB)
	})

	return r
}
