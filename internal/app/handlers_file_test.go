package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	store "github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		DB:             nil,
		FlagSaveToFile: true,
		FlagSaveToDB:   false,
		FlagPathToFile: "tmp/short-url-db-test.json",
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range testCases {
		memory, err := store.New(h.FlagSaveToFile, h.FlagPathToFile)
		require.NoError(t, err)

		h.Memory = memory

		ts := httptest.NewServer(runTestServer(h))
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
		store      store.LinkStorage
		statusCode int
	}{
		{
			name:    "positive test #1",
			request: "/MGRkMTk",
			store: store.LinkStorage{
				Store: []store.Link{
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
			store: store.LinkStorage{
				Store: []store.Link{
					{
						ID:          uuid.New(),
						ShortURL:    util.Shorten("ODczZGQ"),
						OriginalURL: "EwHXdJfB",
					},
				},
			},
			statusCode: http.StatusNotFound,
		},
		{
			name:    "not found",
			request: "/" + util.Shorten("asdasda"),
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode: http.StatusNotFound,
		},
	}

	h := Handler{
		DB:             nil,
		FlagSaveToFile: true,
		FlagSaveToDB:   false,
		FlagPathToFile: "tmp/short-url-db-test.json",
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		memory, err := store.New(h.FlagSaveToFile, h.FlagPathToFile)
		require.NoError(t, err)

		h.Memory = memory

		ts := httptest.NewServer(runTestServer(h))
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
		DB:             nil,
		FlagSaveToFile: true,
		FlagSaveToDB:   false,
		FlagPathToFile: "tmp/short-url-db-test.json",
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		memory, err := store.New(h.FlagSaveToFile, h.FlagPathToFile)
		require.NoError(t, err)

		h.Memory = memory

		ts := httptest.NewServer(runTestServer(h))
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
