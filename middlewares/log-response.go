package middlewares

import (
	"encoding/json"
	"io"
	"net/http"
)

type logResponseWriter struct {
	http.ResponseWriter
	Status int
	Body   []byte
}

func (lrw *logResponseWriter) Write(b []byte) (int, error) {
	lrw.Body = b
	return lrw.ResponseWriter.Write(b)
}

func (lrw *logResponseWriter) WriteHeader(status int) {
	lrw.Status = status
	lrw.ResponseWriter.WriteHeader(status)
}

func LogResponse(l io.Writer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lrw := &logResponseWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(w, r)

			response, _ := json.Marshal(struct {
				Host   string
				URL    string
				Method string
				Header http.Header
				Status int
				Body   []byte
				Type   string
			}{
				Host:   r.Host,
				URL:    r.URL.String(),
				Method: r.Method,
				Header: lrw.Header(),
				Status: lrw.Status,
				Body:   lrw.Body,
				Type:   ResponseType,
			})

			l.Write(response)
		})
	}
}
