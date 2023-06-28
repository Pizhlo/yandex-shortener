package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	defer resp.Body.Close()

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
		body         string
		store        store.LinkStorage
		request      string
		expectedCode int
		expectedBody string
	}{
		{
			name:   "positive test",
			method: http.MethodPost,
			body:   `{"url": "https://practicum.yandex.ru"}`,
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			request:      "/api/shorten",
			expectedCode: http.StatusCreated,
			expectedBody: `{"result": "http://localhost:8080/NmJkYjV"}`,
		},
	}

	for _, v := range testCases {
		ts := httptest.NewServer(runTestServer(&v.store))
		defer ts.Close()

		resp := testRequest(t, ts, v.method, v.request, bytes.NewReader([]byte(v.body)))
		defer resp.Body.Close()

		assert.Equal(t, v.expectedCode, resp.StatusCode)
		assert.Equal(t, v.body, resp.Body)
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

func TestReceiveUrl(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		store      store.LinkStorage
		statusCode int
		body       []byte
	}{
		{
			name:    "positive test #1",
			request: "/",
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			statusCode: http.StatusCreated,
			body:       []byte("https://practicum.yandex.ru/"),
		},
		{
			name:    "positive test #2",
			request: "/",
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			statusCode: http.StatusCreated,
			body:       []byte("EwHXdJfB"),
		},
		{
			name:    "negative test",
			request: "/",
			store: store.LinkStorage{
				Store: map[string]string{},
			},
			statusCode: http.StatusCreated,
			body:       []byte(""),
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

		assert.Equal(t, v.store.Store[util.Shorten(string(v.body))], string(v.body))

	}
}
