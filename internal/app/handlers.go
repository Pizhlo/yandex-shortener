package app

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/Pizhlo/yandex-shortener/util"
)

var rID = regexp.MustCompile(`[a-zA-Z]{7}`)

type Model map[string]string

func ReceiveURL(m Model, w http.ResponseWriter, body io.ReadCloser) {
	fmt.Println("ReceiveUrl")
	// сократить ссылку
	// записать в базу
	j, _ := io.ReadAll(body)

	if len(string(j)) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		short := util.Shorten(string(j))
		fmt.Println(m)

		path, err := util.PrependBaseURL("http://localhost:8080/", short)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		m[short] = string(j)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		w.Write([]byte(path))
	}
}

func GetURL(m Model, w http.ResponseWriter, path string) {
	fmt.Println("GetUrl")
	s := strings.Replace(path, "/", "", -1)

	// проверить наличие ссылки в базе
	// выдать ссылку

	fmt.Println(m)

	if rID.MatchString(s) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Println(s)

		if val, ok := m[s]; ok {
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte(val))
			w.Header().Set("Location", val)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not found"))
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}
