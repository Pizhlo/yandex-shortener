package app

import (
	"fmt"
	"io"
	"net/http"
)

func ReceiveURL(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ReceiveUrl")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusCreated)
	b := r.Body
	j, _ := io.ReadAll(b)
	fmt.Println(string(j))
	w.Write([]byte("Это страница created."))
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetUrl")
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte("Это страница get/id."))
}
