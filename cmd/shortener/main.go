package main

import (
	"net/http"

	internal "github.com/Pizhlo/yandex-shortener/internal/app"
	"github.com/go-chi/chi"
)

func main() {
    http.ListenAndServe(":8080", Run())
} 

func Run() chi.Router {
	m := make(internal.Model)

	r := chi.NewRouter()
	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(m, rw, r)
	})
	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(m, rw, r)
	})

	return r
}
