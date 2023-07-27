package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/session"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) *http.Response {

	req, err := http.NewRequest(method, ts.URL+path, body)
	req.Close = true
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("User-Agent", "PostmanRuntime/7.32.3")
	require.NoError(t, err)

	ts.Client()

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}

func runTestServer(handler Handler) (chi.Router, error) {
	router := chi.NewRouter()

	router.Use(session.CookieMiddleware)

	logger, err := makeLogger()
	if err != nil {
		return nil, err
	}

	handler.Logger = logger

	router.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		GetURL(handler, rw, r)
	})
	router.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		ReceiveURL(handler, rw, r)
	})

	router.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
				ReceiveURLAPI(handler, rw, r)
			})

			r.Post("/shorten/batch", func(rw http.ResponseWriter, r *http.Request) {
				ReceiveManyURLAPI(handler, rw, r)
			})
		})
	})

	return router, nil
}

func makeLogger() (log.Logger, error) {
	logger := log.Logger{}

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return logger, err
	}

	defer zapLogger.Sync()

	sugar := *zapLogger.Sugar()

	logger.Sugar = sugar

	return logger, nil
}
