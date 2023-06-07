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

	path, err := util.MakeURL(r.Host, short)
	if err != nil {
		fmt.Println("err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	m[short] = string(j)
	fmt.Println(m)


	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(path))
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
