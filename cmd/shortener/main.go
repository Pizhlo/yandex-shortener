package main

import (
	"net/http"

	internal "github.com/Pizhlo/yandex-shortener/internal/app"
	"github.com/go-chi/chi"
)

// func webhook(m internal.Model) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("webhook")

// 		if r.Method == http.MethodGet {
// 			fmt.Println("MethodGet")
// 			internal.GetURL(m, w, r)
// 			return
// 		} else if r.Method == http.MethodPost {
// 			fmt.Println("MethodPost")
// 			internal.ReceiveURL(m, w, r)
// 			return
// 		} else {
// 			fmt.Println("StatusBadRequest")
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 	}

// }

func main() {
	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {
	m := make(internal.Model)

	r := chi.NewRouter()
	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(m, rw, r)
	})
	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(m, rw, r)
	})

	return http.ListenAndServe(":8080", r)
}
