package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Pizhlo/yandex-shortener/config"
	internal "github.com/Pizhlo/yandex-shortener/internal/app"
	"github.com/Pizhlo/yandex-shortener/internal/app/compress"
	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/service"
	"github.com/Pizhlo/yandex-shortener/internal/app/session"
	storage "github.com/Pizhlo/yandex-shortener/storage/db"
	file "github.com/Pizhlo/yandex-shortener/storage/file"
	memory "github.com/Pizhlo/yandex-shortener/storage/memory"
	"github.com/Pizhlo/yandex-shortener/storage/model"
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

	doneCh := make(chan struct{})
	defer close(doneCh)

	// канал с данными
	// inputCh := generator(doneCh, input)

	// // получаем слайс каналов из 10 рабочих add
	// channels := fanOut(doneCh, inputCh)

	// // а теперь объединяем десять каналов в один
	// addResultCh := fanIn(doneCh, channels...)

	var srv *service.Service
	var db *storage.URLStorage

	if conf.FlagSaveToDB {
		conn, err := storage.Connect(conf.FlagDatabaseAddress)
		if err != nil {
			logger.Sugar.Fatal("error while creating db connection: ", zap.Error(err))
		}

		db, err = storage.New(conn, logger)
		if err != nil {
			logger.Sugar.Fatal("error while creating db: ", zap.Error(err))
		}

		srv = service.New(db)
	} else if conf.FlagSaveToFile {
		storage, err := file.New(conf.FlagPathToFile, logger)
		if err != nil {
			logger.Sugar.Fatal("error while creating file storage: ", zap.Error(err))
		}

		srv = service.New(storage)
	} else {
		storage, err := memory.New(logger)
		if err != nil {
			logger.Sugar.Fatal("error while creating memory storage: ", zap.Error(err))
		}
		srv = service.New(storage)
	}

	handler := internal.Handler{
		Service:      srv,
		Logger:       logger,
		FlagBaseAddr: conf.FlagBaseAddr,
	}

	if conf.FlagSaveToDB {
		handler.LinksChan = make(chan model.DeleteLink, 1024)
		go flushLinks(handler.LinksChan, db, handler.Logger)
	}

	if err := http.ListenAndServe(conf.FlagRunAddr, Run(handler, db)); err != nil {
		logger.Sugar.Fatal("error while executing server: ", zap.Error(err))
	}
}

func Run(handler internal.Handler, db *storage.URLStorage) chi.Router {
	r := chi.NewRouter()

	r.Use(session.CookieMiddleware)
	r.Use(compress.UnpackData)
	r.Use(handler.Logger.WithLogging)

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

			r.Get("/user/urls", func(w http.ResponseWriter, r *http.Request) {
				internal.GetUserURLS(handler, w, r)
			})

			r.Delete("/user/urls", func(w http.ResponseWriter, r *http.Request) {
				internal.DeleteURL(handler, w, r)
			})
		})
	})

	r.Get("/ping", func(rw http.ResponseWriter, r *http.Request) {
		internal.Ping(rw, r, db)
	})

	return r
}

// flushLinks постоянно сохраняет несколько сообщений в хранилище с определённым интервалом
func flushLinks(ch <-chan model.DeleteLink, db *storage.URLStorage, logger log.Logger) {
	ticker := time.NewTicker(10 * time.Second)

	var links []model.DeleteLink

	for {
		select {
		case link := <-ch:
			links = append(links, link)
		case <-ticker.C:
			if len(links) == 0 {
				continue
			}
			err := db.DeleteURLS(context.TODO(), links...)
			if err != nil {
				logger.Sugar.Debug("cannot delete urls: ", zap.Error(err))
				continue
			}
			links = nil
		}
	}
}

// // fanIn объединяет несколько каналов resultChs в один.
// func fanIn(doneCh chan struct{}, resultChs ...chan int) chan int {
// 	finalCh := make(chan int)

// 	var wg sync.WaitGroup

// 	for _, ch := range resultChs {
// 		chClosure := ch

// 		wg.Add(1)

// 		go func() {
// 			defer wg.Done()

// 			for data := range chClosure {
// 				select {
// 				case <-doneCh:
// 					return
// 				case finalCh <- data:
// 				}
// 			}
// 		}()
// 	}

// 	go func() {
// 		wg.Wait()
// 		close(finalCh)
// 	}()

// 	return finalCh
// }

// func generator(doneCh chan struct{}, input []int) chan int {
// 	inputCh := make(chan int)

// 	go func() {
// 		defer close(inputCh)

// 		for _, data := range input {
// 			select {
// 			case <-doneCh:
// 				return
// 			case inputCh <- data:
// 			}
// 		}
// 	}()

// 	return inputCh
// }

// fanOut принимает канал данных, порождает 10 горутин
// func fanOut(doneCh chan struct{}, inputCh chan int) []chan int {
// 	// количество горутин add
// 	numWorkers := 10
// 	// каналы, в которые отправляются результаты
// 	channels := make([]chan int, numWorkers)

// 	for i := 0; i < numWorkers; i++ {
// 		// получаем канал из горутины add
// 		addResultCh := add(doneCh, inputCh)
// 		// отправляем его в слайс каналов
// 		channels[i] = addResultCh
// 	}

// 	// возвращаем слайс каналов
// 	return channels
// }
