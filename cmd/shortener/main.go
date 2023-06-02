package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	internal "github.com/Pizhlo/yandex-shortener/internal/app"
)

var rId = regexp.MustCompile(`[a-zA-Z]{8}`)

func webhook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("webhook")

	s := strings.Replace(r.URL.Path, "/", "", -1)

	if rId.MatchString(s) {
		fmt.Println(r.URL.Path)
		internal.GetUrl(w, r)
	} else {
		internal.ReceiveUrl(w, r)
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
