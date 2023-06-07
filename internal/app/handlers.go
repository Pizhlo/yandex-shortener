package app

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/Pizhlo/yandex-shortener/util"
)

var rID = regexp.MustCompile(`[a-zA-Z0-9]{7}`)

type Model map[string]string

func ReceiveURL(m Model, w http.ResponseWriter, r *http.Request) {
	fmt.Println("ReceiveUrl")
	// сократить ссылку
	// записать в базу
	j, _ := io.ReadAll(r.Body)
	short := util.Shorten(string(j))

	m[short] = string(j)
	fmt.Println("ReceiveUrl m =", m)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
}

func GetURL(m Model, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")
	s := strings.Replace(r.URL.Path, "/", "", -1)

	// проверить наличие ссылки в базе
	// выдать ссылку

	fmt.Println("m = ", m)
	fmt.Println("r.URL.Path = ", r.URL.Path)

	if val, ok := m[s]; ok {
		setLocation(w, val)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func setLocation(w http.ResponseWriter, addr string) {
	w.Header().Set("Location", addr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
