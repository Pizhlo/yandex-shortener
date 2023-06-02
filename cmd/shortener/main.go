package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func ReceiveUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ReceiveUrl")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("Это страница created."))
	w.WriteHeader(http.StatusCreated)
}

var rId = regexp.MustCompile(`[a-zA-Z]{8}`)

func webhook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("webhook")

	s := strings.Replace(r.URL.Path, "/", "", -1)

	if rId.MatchString(s) {
		fmt.Println(r.URL.Path)
		GetUrl(w, r)
	} else {
		ReceiveUrl(w, r)
	}

}

func GetUrl(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")
	w.Write([]byte("Это страница get/id."))
	w.WriteHeader(http.StatusTemporaryRedirect)
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
