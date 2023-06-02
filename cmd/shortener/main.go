package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	internal "github.com/Pizhlo/yandex-shortener/internal/app"
)

var rID = regexp.MustCompile(`[a-zA-Z]{8}`)

func webhook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("webhook")

	s := strings.Replace(r.URL.Path, "/", "", -1)

	if rID.MatchString(s) {
		fmt.Println(r.URL.Path)
		internal.GetURL(w, r)
	} else {
		internal.ReceiveURL(w, r)
	}

}

func main() {
	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, http.HandlerFunc(webhook))

	return http.ListenAndServe(`:8080`, mux)
}
