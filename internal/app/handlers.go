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

func GetURL(m Model, w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")
	s := strings.Replace(r.URL.Path, "/", "", -1)

	// проверить наличие ссылки в базе
	// выдать ссылку

	fmt.Println("m = ", m)
	fmt.Println("s = ", s)

	if rID.MatchString(s) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Println(s)

		if _, ok := m[s]; ok {
			fmt.Println("val = ", s)

			//url, err := util.MakeURL(r.Host, s)
			// if err != nil {
			// 	fmt.Println("err: ", err)
			// 	w.WriteHeader(http.StatusInternalServerError)
			// 	return
			// }
			w.Header().Set("Location", m[s])
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
