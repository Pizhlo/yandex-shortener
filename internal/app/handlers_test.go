package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/service"
	store "github.com/Pizhlo/yandex-shortener/storage/memory"
	"github.com/Pizhlo/yandex-shortener/storage/model"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		store      store.Memory
		statusCode int
	}{
		{
			name:    "positive test #1",
			request: "/YjhkNDY",
			store: store.Memory{
				Store: []model.Link{
					{
						ID:          uuid.New(),
						ShortURL:    "YjhkNDY",
						OriginalURL: "https://practicum.yandex.ru/",
					},
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "positive test #2",
			request: "/" + util.Shorten("Y2NlMzI"),
			store: store.Memory{
				Store: []model.Link{
					{
						ID:          uuid.New(),
						ShortURL:    util.Shorten("Y2NlMzI"),
						OriginalURL: "Y2NlMzI",
					},
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "not found",
			request: "/" + util.Shorten("asdasda"),
			store: store.Memory{
				Store: []model.Link{},
			},
			statusCode: http.StatusNotFound,
		},
	}

	h := Handler{
		Service:      &service.Service{},
		FlagBaseAddr: "http://localhost:8000/",
	}

	for _, v := range tests {
		memory := &v.store
		srv := service.New(memory)
		h.Service = srv

		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		r, err := runTestServer(h)
		require.NoError(t, err)

		ts := httptest.NewServer(r)
		defer ts.Close()

		resp := testRequest(t, ts, "GET", v.request, nil)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		if v.statusCode != http.StatusNotFound {
			assert.Equal(t, v.store.Store[0].OriginalURL, resp.Header.Get("Location"))
		}
	}
}

func TestReceiveURL(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		store        store.Memory
		statusCode   int
		body         []byte
		expectedBody string
	}{
		{
			name:    "positive test #1",
			request: "/",
			store: store.Memory{
				Store: []model.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte("https://practicum.yandex.ru/"),
			expectedBody: "http://localhost:8000/MGRkMTk",
		},
		{
			name:    "positive test #2",
			request: "/",
			store: store.Memory{
				Store: []model.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte("EwHXdJfB"),
			expectedBody: "http://localhost:8000/ODczZGQ",
		},
		{
			name:    "negative test",
			request: "/",
			store: store.Memory{
				Store: []model.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte(""),
			expectedBody: "http://localhost:8000/ZDQxZDh",
		},
	}

	h := Handler{
		Service:      &service.Service{},
		FlagBaseAddr: "http://localhost:8000/",
	}

	for _, v := range tests {
		memory := &v.store
		srv := service.New(memory)
		h.Service = srv

		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		r, err := runTestServer(h)
		require.NoError(t, err)

		ts := httptest.NewServer(r)
		defer ts.Close()

		body := strings.NewReader(string(v.body))
		resp := testRequest(t, ts, "POST", v.request, body)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		resBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, v.expectedBody, string(resBody))

	}
}
