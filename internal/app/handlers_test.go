package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUrl(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		model      Model
		statusCode int
	}{
		{
			name:    "positive test",
			request: "/" + util.Shorten("asdasda"),
			model: Model{
				util.Shorten("asdasda"): "asdasda",
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "positive test",
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.request, nil)

			w := httptest.NewRecorder()

			GetURL(test.model, w, request)

			res := w.Result()

			assert.Equal(t, test.statusCode, res.StatusCode)

			defer res.Body.Close()

			s := strings.Replace(test.request, "/", "", -1)

			expectedURL, err := util.MakeURL(request.Host, s)
			require.NoError(t, err)

			if test.statusCode != http.StatusNotFound {
				assert.Equal(t, expectedURL, w.Header().Get("Location"))
			} 

		})
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
			body:       []byte("EwHXdJfB"),
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(string(test.body))
			request := httptest.NewRequest(http.MethodPost, "/", r)

			w := httptest.NewRecorder()
			m := make(Model)

			ReceiveURL(m, w, request)

			res := w.Result()

			assert.Equal(t, test.statusCode, res.StatusCode)

			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			expectedResp, err := util.MakeURL(request.Host, util.Shorten(string(test.body)))
			require.NoError(t, err)

			assert.Equal(t, expectedResp, string(resBody))
			assert.Equal(t, m[util.Shorten(string(test.body))], string(test.body))
		})
	}
}
