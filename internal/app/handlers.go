package app

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Pizhlo/yandex-shortener/util"
	"github.com/go-chi/chi"
)

type Model map[string]string

func ReceiveURL(m Model, w http.ResponseWriter, r *http.Request, baseURL string) {
	fmt.Println("ReceiveUrl")
	// сократить ссылку
	// записать в базу
	j, _ := io.ReadAll(r.Body)
	short := util.Shorten(string(j))

	m[short] = string(j)
	fmt.Println("ReceiveUrl m =", m)
	fmt.Println("ReceiveUrl baseURL =", baseURL)
	fmt.Println("r.Host =", r.Host)

	fmt.Println("ReceiveUrl baseURL =", baseURL)

	path, err := util.MakeURL(baseURL, short)
	if err != nil {
		fmt.Println("err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(path))
}

func GetURL(m Model, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")
	id := chi.URLParam(r, "id")

	// проверить наличие ссылки в базе
	// выдать ссылку

	if val, ok := m[id]; ok {
		setLocation(w, val)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func setLocation(w http.ResponseWriter, addr string) {
	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
