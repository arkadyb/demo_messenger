package utils

import "net/http"

// NewStatusCodeRecorder returns new instance of StatusCodeRecorder
func NewStatusCodeRecorder(writer http.ResponseWriter) *StatusCodeRecorder {
	return &StatusCodeRecorder{
		ResponseWriter: writer,
	}
}

// StatusCodeRecorder - composed with ResponseWriter is used to record response status code
type StatusCodeRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func (rec *StatusCodeRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}
