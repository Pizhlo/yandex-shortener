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
				method:       "post",
				request:      "/api/batch",
				expectedCode: http.StatusCreated,
				conf:         config.Config{FlagPathToFile: "tmp/short-url-db-test.json", FlagSaveToFile: true},
				db:           nil,
				body: []models.RequestAPI{
					{
						ID:  "a293b415-7f49-4f3d-ab12-c081ee691924",
						URL: "https://practicum222.yandex.ru",
					},
					{
						ID:  "1bcd1369-9da2-43cd-b68f-1ec8329cc86b",
						URL: "https://practicum.yandex.ru",
					},
					{
						ID:  "11d48cc2-ea84-44a0-9814-7045e9cb551d",
						URL: "https://practicum333.yandex.ru",
					}},

				expectedBody: []models.ResponseAPI{
					{
						ID:       "a293b415-7f49-4f3d-ab12-c081ee691924",
						ShortURL: "N2YzM2Y",
					},
					{
						ID:       "1bcd1369-9da2-43cd-b68f-1ec8329cc86b",
						ShortURL: "NmJkYjV",
					},
					{
						ID:       "11d48cc2-ea84-44a0-9814-7045e9cb551d",
						ShortURL: "NjRiNjg",
					}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(runTestServer(tt.args.memory, tt.args.conf, tt.args.db))
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
