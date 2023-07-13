package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pizhlo/yandex-shortener/config"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	store "github.com/Pizhlo/yandex-shortener/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceiveURLAPI(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		body         models.Request
		store        store.LinkStorage
		request      string
		expectedCode int
		expectedBody models.Response
	}{
		{
			name:   "positive test",
			method: http.MethodPost,
			body:   models.Request{URL: "https://practicum.yandex.ru"},
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			request:      "/api/shorten",
			expectedCode: http.StatusCreated,
			expectedBody: models.Response{
				Result: `http://localhost:8000/NmJkYjV`,
			},
		},
	}

	conf := config.Config{
		FlagSaveToFile: false,
		FlagSaveToDB:   false,
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range testCases {
		ts := httptest.NewServer(runTestServer(&v.store, conf, nil))
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

func TestReceiveManyURLAPI(t *testing.T) {
	type args struct {
		memory       *store.LinkStorage
		method       string
		request      string
		expectedCode int
		conf         config.Config
		db           *store.Database
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
				memory:       &store.LinkStorage{},
				method:       http.MethodPost,
				request:      "/api/batch",
				expectedCode: http.StatusCreated,
				conf:         config.Config{FlagPathToFile: "tmp/short-url-db-test.json", FlagSaveToFile: true},
				db:           &store.Database{},
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
						ShortURL: "NjYyNjB",
					},
					{
						ID:       "c82b937d-c303-40e1-a655-ab085002dfa0",
						ShortURL: "NmJkYjV",
					},
					{
						ID:       "cd53c344-fb57-42cf-b576-823476f90918",
						ShortURL: "ODczZGQ",
					}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memory, err := store.New(tt.args.conf.FlagSaveToFile, tt.args.conf.FlagPathToFile)
			require.NoError(t, err)

			ts := httptest.NewServer(runTestServer(memory, tt.args.conf, tt.args.db))
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
