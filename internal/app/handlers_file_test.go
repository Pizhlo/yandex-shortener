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
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestReceiveURLAPIFileStorage(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		body         models.Request
		filename     string
		request      string
		expectedCode int
		expectedBody models.Response
	}{
		{
			name:         "positive test",
			method:       http.MethodPost,
			body:         models.Request{URL: "https://practicum.yandex.ru"},
			filename:     "tmp/short-url-db-test.json",
			request:      "/api/shorten",
			expectedCode: http.StatusCreated,
			expectedBody: models.Response{
				Result: `http://localhost:8000/NmJkYjV`,
			},
		},
	}

	h := Handler{
		FlagBaseAddr: "http://localhost:8000/",
	}

	for _, v := range testCases {
		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		fs, err := store.New(v.filename, h.Logger)
		require.NoError(t, err)

		h.Service = service.New(fs)

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
		name         string
		request      string
		statusCode   int
		filename     string
		expectedBody models.Response
	}{
		{
			name:       "positive test #1",
			request:    "/NmJkYjV",
			filename:   "tmp/short-url-db-test.json",
			statusCode: http.StatusTemporaryRedirect,
			expectedBody: models.Response{
				Result: "https://practicum.yandex.ru",
			},
		},
		{
			name:       "positive test #2",
			request:    "/NjYyNjB",
			filename:   "tmp/short-url-db-test.json",
			statusCode: http.StatusTemporaryRedirect,
			expectedBody: models.Response{
				Result: "mail2.ru",
			},
		},
		{
			name:         "not found",
			request:      "/" + util.Shorten("not found"),
			filename:     "tmp/short-url-db-test.json",
			statusCode:   http.StatusNotFound,
			expectedBody: models.Response{},
		},
	}

	h := Handler{
		FlagBaseAddr: "http://localhost:8000/",
	}

	for _, v := range tests {
		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		fs, err := store.New(v.filename, h.Logger)
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
			assert.Equal(t, v.expectedBody.Result, resp.Header.Get("Location"))
		}
	}
}

func TestReceiveURLFileStorage(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		statusCode   int
		filename     string
		body         []byte
		expectedBody string
	}{
		{
			name:         "positive test #1",
			request:      "/",
			statusCode:   http.StatusCreated,
			filename:     "tmp/short-url-db-test.json",
			body:         []byte("https://practicum.yandex.ru/"),
			expectedBody: "http://localhost:8000/MGRkMTk",
		},
		{
			name:         "positive test #2",
			request:      "/",
			statusCode:   http.StatusCreated,
			filename:     "tmp/short-url-db-test.json",
			body:         []byte("EwHXdJfB"),
			expectedBody: "http://localhost:8000/ODczZGQ",
		},
		{
			name:         "negative test",
			request:      "/",
			statusCode:   http.StatusCreated,
			filename:     "tmp/short-url-db-test.json",
			body:         []byte(""),
			expectedBody: "http://localhost:8000/ZDQxZDh",
		},
	}

	h := Handler{
		FlagBaseAddr: "http://localhost:8000/",
	}

	for _, v := range tests {
		logger := log.Logger{}

		zapLogger, err := zap.NewDevelopment()
		require.NoError(t, err)

		defer zapLogger.Sync()

		sugar := *zapLogger.Sugar()

		logger.Sugar = sugar
		h.Logger = logger

		fs, err := store.New(v.filename, h.Logger)
		require.NoError(t, err)

		h.Service = service.New(fs)

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
		filename     string
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
				filename:     "tmp/short-url-db-test.json",
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
		FlagBaseAddr: "http://localhost:8000/",
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

			fs, err := store.New(tt.args.filename, h.Logger)
			require.NoError(t, err)

			h.Service = service.New(fs)

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
