package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pizhlo/yandex-shortener/config"
	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	store "github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/google/uuid"
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

func TestGetURL(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		store      store.LinkStorage
		statusCode int
	}{
		{
			name:    "positive test #1",
			request: "/YjhkNDY",
			store: store.LinkStorage{
				Store: []store.Link{
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
			store: store.LinkStorage{
				Store: []store.Link{
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
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode: http.StatusNotFound,
		},
	}
	conf := config.Config{
		FlagSaveToFile: false,
		FlagSaveToDB:   false,
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		ts := httptest.NewServer(runTestServer(&v.store, conf, nil))
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
		store        store.LinkStorage
		statusCode   int
		body         []byte
		expectedBody string
	}{
		{
			name:    "positive test #1",
			request: "/",
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte("https://practicum.yandex.ru/"),
			expectedBody: "http://localhost:8000/MGRkMTk",
		},
		{
			name:    "positive test #2",
			request: "/",
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte("EwHXdJfB"),
			expectedBody: "http://localhost:8000/ODczZGQ",
		},
		{
			name:    "negative test",
			request: "/",
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte(""),
			expectedBody: "http://localhost:8000/ZDQxZDh",
		},
	}

	conf := config.Config{
		FlagSaveToFile: false,
		FlagSaveToDB:   false,
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		ts := httptest.NewServer(runTestServer(&v.store, conf, nil))
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
