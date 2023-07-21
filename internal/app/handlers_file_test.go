package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/Pizhlo/yandex-shortener/internal/app/logger"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	"github.com/Pizhlo/yandex-shortener/internal/app/service"
	store "github.com/Pizhlo/yandex-shortener/storage/file"
	memory "github.com/Pizhlo/yandex-shortener/storage/memory"
	"github.com/Pizhlo/yandex-shortener/storage/model"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestReceiveURLAPIFileStorage(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		body         models.Request
		request      string
		expectedCode int
		expectedBody models.Response
	}{
		{
			name:         "positive test",
			method:       http.MethodPost,
			body:         models.Request{URL: "https://practicum.yandex.ru"},
			request:      "/api/shorten",
			expectedCode: http.StatusCreated,
			expectedBody: models.Response{
				Result: `http://localhost:8000/NmJkYjV`,
			},
		},
	}

	h := Handler{
		FlagPathToFile: "tmp/short-url-db-test.json",
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range testCases {
		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		memory, err := store.New(h.FlagPathToFile)
		require.NoError(t, err)

		h.Service = service.New(memory)

		r, err := runTestServer(h)
		require.NoError(t, err)

		ts := httptest.NewServer(r)
		defer ts.Close()

		bodyJSON, err := json.Marshal(v.body)
		require.NoError(t, err)

		resp := testRequest(t, ts, v.method, v.request, bytes.NewReader(bodyJSON))
		defer resp.Body.Close()

		assert.Equal(t, v.expectedCode, resp.StatusCode)

		var result models.Response
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, v.expectedBody, result)
	}
}

func TestGetURLFileStorage(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		store      memory.LinkStorage
		statusCode int
	}{
		{
			name:    "positive test #1",
			request: "/MGRkMTk",
			store: memory.LinkStorage{
				Store: []model.Link{
					{
						ID:          uuid.New(),
						ShortURL:    "MGRkMTk",
						OriginalURL: "https://practicum.yandex.ru/",
					},
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "not found",
			request: "/" + util.Shorten("ODczZGQ"),
			store: memory.LinkStorage{
				Store: []model.Link{
					{
						ID:          uuid.New(),
						ShortURL:    util.Shorten("ODczZGQ"),
						OriginalURL: "EwHXdJfB",
					},
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "not found",
			request: "/" + util.Shorten("asdasda"),
			store: memory.LinkStorage{
				Store: []model.Link{},
			},
			statusCode: http.StatusNotFound,
		},
	}

	h := Handler{
		FlagPathToFile: "tmp/short-url-db-test.json",
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		//h.Memory = &v.store
		fs, err := store.New(h.FlagPathToFile)
		require.NoError(t, err)

		h.Service = service.New(fs)

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

func TestReceiveURLFileStorage(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		statusCode   int
		body         []byte
		expectedBody string
	}{
		{
			name:         "positive test #1",
			request:      "/",
			statusCode:   http.StatusCreated,
			body:         []byte("https://practicum.yandex.ru/"),
			expectedBody: "http://localhost:8000/MGRkMTk",
		},
		{
			name:         "positive test #2",
			request:      "/",
			statusCode:   http.StatusCreated,
			body:         []byte("EwHXdJfB"),
			expectedBody: "http://localhost:8000/ODczZGQ",
		},
		{
			name:         "negative test",
			request:      "/",
			statusCode:   http.StatusCreated,
			body:         []byte(""),
			expectedBody: "http://localhost:8000/ZDQxZDh",
		},
	}

	h := Handler{
		FlagPathToFile: "tmp/short-url-db-test.json",
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		memory, err := store.New(h.FlagPathToFile)
		require.NoError(t, err)

		h.Service = service.New(memory)

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

func TestReceiveManyURLAPIFileStorage(t *testing.T) {
	type args struct {
		method       string
		request      string
		expectedCode int
		body         []models.RequestAPI
		expectedBody []models.ResponseAPI
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive test without DB",
			args: args{
				method:       http.MethodPost,
				request:      "/api/shorten/batch",
				expectedCode: http.StatusCreated,
				body: []models.RequestAPI{
					{
						ID:  "e169d217-d3c8-493a-930f-7432368139c7",
						URL: "mail2.ru",
					},
					{
						ID:  "c82b937d-c303-40e1-a655-ab085002dfa0",
						URL: "https://practicum.yandex.ru",
					},
					{
						ID:  "cd53c344-fb57-42cf-b576-823476f90918",
						URL: "EwHXdJfB",
					}},

				expectedBody: []models.ResponseAPI{
					{
						ID:       "e169d217-d3c8-493a-930f-7432368139c7",
						ShortURL: "http://localhost:8000/NjYyNjB",
					},
					{
						ID:       "c82b937d-c303-40e1-a655-ab085002dfa0",
						ShortURL: "http://localhost:8000/NmJkYjV",
					},
					{
						ID:       "cd53c344-fb57-42cf-b576-823476f90918",
						ShortURL: "http://localhost:8000/ODczZGQ",
					}},
			},
		},
	}

	h := Handler{
		FlagPathToFile: "tmp/short-url-db-test.json",
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.Logger{}

			zapLogger, err := zap.NewDevelopment()
			require.NoError(t, err)

			defer zapLogger.Sync()

			sugar := *zapLogger.Sugar()

			logger.Sugar = sugar
			h.Logger = logger

			memory, err := store.New(h.FlagPathToFile)
			require.NoError(t, err)

			h.Service = service.New(memory)

			r, err := runTestServer(h)
			require.NoError(t, err)

			ts := httptest.NewServer(r)
			defer ts.Close()

			bodyJSON, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			resp := testRequest(t, ts, tt.args.method, tt.args.request, bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.args.expectedCode, resp.StatusCode)

			var result []models.ResponseAPI

			dec := json.NewDecoder(resp.Body)
			err = dec.Decode(&result)
			require.NoError(t, err)

			assert.Equal(t, tt.args.expectedBody, result)
		})
	}
}
