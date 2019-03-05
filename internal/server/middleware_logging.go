package server

import (
	"github.com/arkadyb/demo_messenger/internal/pkg/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// middleware handler to log incoming requests basic parameters and latency
func LoggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		for _, skipPath := range []string{"/health", "/metrics"} {
			if request.URL.Path == skipPath {
				handler.ServeHTTP(writer, request)
				return
			}
		}
		startTime := time.Now()
		recorder := utils.NewStatusCodeRecorder(writer)
		handler.ServeHTTP(recorder, request)
		log.WithFields(log.Fields{
			"elapsedTime": time.Since(startTime),
			"requestIP":   utils.GetRequestIPAddress(request),
			"requestPath": request.URL.Path,
			"statusCode":  recorder.StatusCode,
		}).Info("request")
	})
}
