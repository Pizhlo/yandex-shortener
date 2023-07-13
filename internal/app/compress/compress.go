package compress

import (
	"compress/gzip"
	"fmt"
	"net/http"
)

// распаковка
// принимает сжатые данные
func UnpackData(next http.Handler) http.Handler {
	fmt.Println("UnpackData")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = gz
			defer gz.Close()
		}

		next.ServeHTTP(w, r)
	})
}
