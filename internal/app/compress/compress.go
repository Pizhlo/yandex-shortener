package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// упаковка
// сжимает данные
func PackData(next http.Handler) http.Handler {
	fmt.Println("PackData")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") { // если клиент не готов к сжатым данным, идем дальше
			fmt.Println("PackData !strings.Contains")
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			fmt.Println("NewWriterLevel error")
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// распаковка
// принимает сжатые данные
func UnpackData(next http.Handler) http.Handler {
	fmt.Println("UnpackData")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") { // если клиент не сжимал данные, идем дальше
			fmt.Println("UnpackData !strings.Contains")
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			fmt.Println("gzip.NewReader err")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer gz.Close()

		body, err := io.ReadAll(gz)
		if err != nil {
			fmt.Println("io.ReadAll err")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Length: %d", len(body))
	})
}