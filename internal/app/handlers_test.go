package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
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

func runTestServer(m Model) chi.Router {
	router := chi.NewRouter()
	router.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		GetURL(m, rw, r)
	})
	router.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		ReceiveURL(m, rw, r)
	})

	return router
}

func TestGetURL(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		model      Model
		statusCode int
	}{
		{
			name:    "positive test #1",
			request: "/YjhkNDY",
			model: Model{
				"YjhkNDY": "https://practicum.yandex.ru/",
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "positive test #2",
			request: "/" + util.Shorten("Y2NlMzI"),
			model: Model{
				util.Shorten("Y2NlMzI"): "Y2NlMzI",
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:       "not found",
			request:    "/" + util.Shorten("asdasda"),
			model:      Model{},
			statusCode: http.StatusNotFound,
		},
	}

	for _, v := range tests {
		ts := httptest.NewServer(runTestServer(v.model))
		defer ts.Close()
		
		resp := testRequest(t, ts, "GET", v.request, nil)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		if v.statusCode != http.StatusNotFound {
			s := strings.Replace(v.request, "/", "", -1)

			assert.Equal(t, v.model[s], resp.Header.Get("Location"))
		}
	}
}

func TestReceiveUrl(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		model      Model
		statusCode int
		body       []byte
	}{
		{
			name:       "positive test #1",
			request:    "/",
			model:      Model{},
			statusCode: http.StatusCreated,
			body:       []byte("https://practicum.yandex.ru/"),
		},
		{
			name:       "positive test #2",
			request:    "/",
			model:      Model{},
			statusCode: http.StatusCreated,
			body:       []byte("EwHXdJfB"),
		},
		{
			name:       "negative test",
			request:    "/",
			model:      Model{},
			statusCode: http.StatusCreated,
			body:       []byte(""),
		},
	}

	for _, v := range tests {
		// w := httptest.NewRecorder()
		ts := httptest.NewServer(runTestServer(v.model))
		defer ts.Close()

		body := strings.NewReader(string(v.body))
		resp := testRequest(t, ts, "POST", v.request, body)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		assert.Equal(t, v.model[util.Shorten(string(v.body))], string(v.body))

	}
}
