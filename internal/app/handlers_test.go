package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pizhlo/yandex-shortener/internal/app/models"
	store "github.com/Pizhlo/yandex-shortener/storage"
	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	//defer resp.Body.Close()

	return resp
}

func runTestServer(storage *store.LinkStorage) chi.Router {
	router := chi.NewRouter()
	baseURL := "http://localhost:8000/"

	router.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		GetURL(storage, rw, r)
	})
	router.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		ReceiveURL(storage, rw, r, baseURL)
	})
	router.Route("/api", func(r chi.Router) {
		r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
			ReceiveURLAPI(storage, rw, r, baseURL)
		})
	})

	return router
}

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
				Store: map[string]string{},
			},
			request:      "/api/shorten",
			expectedCode: http.StatusCreated,
			expectedBody: models.Response{
				Result: `http://localhost:8000/NmJkYjV`,
			},
		},
	}

	for _, v := range testCases {
		ts := httptest.NewServer(runTestServer(&v.store))
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
				Store: map[string]string{
					"YjhkNDY": "https://practicum.yandex.ru/",
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "positive test #2",
			request: "/" + util.Shorten("Y2NlMzI"),
			store: store.LinkStorage{
				Store: map[string]string{
					util.Shorten("Y2NlMzI"): "Y2NlMzI",
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "not found",
			request: "/" + util.Shorten("asdasda"),
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			statusCode: http.StatusNotFound,
		},
	}

	for _, v := range tests {
		ts := httptest.NewServer(runTestServer(&v.store))
		defer ts.Close()

		resp := testRequest(t, ts, "GET", v.request, nil)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		if v.statusCode != http.StatusNotFound {
			s := strings.Replace(v.request, "/", "", -1)

			assert.Equal(t, v.store.Store[s], resp.Header.Get("Location"))
		}
	}
}

func TestReceiveURL(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		store      store.LinkStorage
		statusCode int
		body       []byte
		expectedBody string
	}{
		{
			name:    "positive test #1",
			request: "/",
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			statusCode: http.StatusCreated,
			body:       []byte("https://practicum.yandex.ru/"),
			expectedBody: "http://localhost:8000/MGRkMTk",
		},
		{
			name:    "positive test #2",
			request: "/",
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			statusCode: http.StatusCreated,
			body:       []byte("EwHXdJfB"),
			expectedBody: "http://localhost:8000/ODczZGQ",
		},
		{
			name:    "negative test",
			request: "/",
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			statusCode: http.StatusCreated,
			body:       []byte(""),
			expectedBody: "http://localhost:8000/ZDQxZDh",
		},
	}

	for _, v := range tests {
		// w := httptest.NewRecorder()
		ts := httptest.NewServer(runTestServer(&v.store))
		defer ts.Close()

		body := strings.NewReader(string(v.body))
		resp := testRequest(t, ts, "POST", v.request, body)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		resBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		fmt.Println("resBody = ", string(resBody))

		assert.Equal(t, v.expectedBody, string(resBody))

	}
}
