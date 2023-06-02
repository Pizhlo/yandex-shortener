package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUrl(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		statusCode int
		response   string
	}{
		{
			name:       "positive test",
			request:    "/abcedfgr",
			statusCode: http.StatusTemporaryRedirect,
			response:   "Это страница get/id.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			GetUrl(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.statusCode, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, string(resBody), test.response)
		})
	}
}

func TestReceiveUrl(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		param      string
		statusCode int
		response   string
	}{
		{
			name:       "positive test",
			request:    "/",
			statusCode: http.StatusCreated,
			response:   "Это страница created.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			b := new(bytes.Buffer)
			b.Write([]byte("EwHXdJfB"))

			ReceiveUrl(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.statusCode, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, string(resBody), test.response)
		})
	}
}
