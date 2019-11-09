package http

import "net/http"

type response struct {
	http.ResponseWriter
	statusCode int
	bodySize   int
}

func (w *response) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *response) Write(p []byte) (int, error) {
	w.bodySize += len(p)
	return w.ResponseWriter.Write(p)
}
