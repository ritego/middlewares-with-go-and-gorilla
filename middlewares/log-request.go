package middlewares

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func LogRequest(l io.Writer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			request, _ := json.Marshal(struct {
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
				Header: r.Header,
				Body:   body,
				Type:   RequestType,
			})
			l.Write(request)
			next.ServeHTTP(w, r)
		})
	}
}
