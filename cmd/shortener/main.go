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

	logger := log.Logger{}

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		zapLogger.Fatal("error while creating sugar: ", zap.Error(err))
	}
	defer zapLogger.Sync()

	sugar := *zapLogger.Sugar()

	logger.Sugar = sugar

	logger.Sugar.Infow(
		"Starting server",
		"addr", conf.FlagRunAddr,
	)

	memory, err := storage.New(logger) // in-memory and file storage
	if err != nil {
		logger.Sugar.Fatal("error while creating storage: ", zap.Error(err))
	}

	if conf.FlagSaveToFile {
		fileStorage, err := storage.NewFileStorage(conf.FlagPathToFile)
		if err != nil {
			logger.Sugar.Fatal("error while creating file storage: ", zap.Error(err))
		}
		memory.FileStorage = *fileStorage

		if err := memory.RecoverData(logger); err != nil {
			logger.Sugar.Fatal("unable to recover file data: ", zap.Error(err))
		}
	}

	db, err := storage.NewStore(conf.FlagDatabaseAddress)
	if err != nil {
		logger.Sugar.Fatal("error while connecting db: ", zap.Error(err))
	}

	if conf.FlagSaveToFile {
		defer memory.FileStorage.Close()
	}

	handler := internal.Handler{
		Memory:         memory,
		DB:             db,
		Logger:         logger,
		FlagBaseAddr:   conf.FlagBaseAddr,
		FlagSaveToFile: conf.FlagSaveToFile,
		FlagSaveToDB:   conf.FlagSaveToDB,
		FlagPathToFile: conf.FlagPathToFile,
	}

	if err := http.ListenAndServe(conf.FlagRunAddr, Run(handler)); err != nil {
		logger.Sugar.Fatal("error while executing server: ", zap.Error(err))
	}
}

func Run(handler internal.Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(handler.Logger.WithLogging)
	r.Use(compress.UnpackData)

	r.Use(middleware.Compress(5, "application/javascript",
		"application/json",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml"))

	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(handler, rw, r)
	})

	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(handler, rw, r)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveURLAPI(handler, rw, r)
			})

			r.Post("/shorten/batch", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveManyURLAPI(handler, rw, r)
			})
		})
	})

	r.Get("/ping", func(rw http.ResponseWriter, r *http.Request) {
		internal.Ping(rw, r, handler.DB, handler.FlagSaveToDB)
	})

	return r
}
