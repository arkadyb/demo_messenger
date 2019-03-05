package server

import (
	"fmt"
	hrx "github.com/afex/hystrix-go/hystrix"
	"github.com/arkadyb/demo_messenger/internal/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// middleware handler for Hystrix implementing basic circuit breaker pattern
func CircuitBreakerMiddleware(commandName string, config hrx.CommandConfig, next http.Handler) http.Handler {
	hrx.ConfigureCommand(commandName, config)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := hrx.Do(commandName, func() (err error) {
			lrw := utils.NewStatusCodeRecorder(w)
			next.ServeHTTP(lrw, req)
			if lrw.StatusCode >= http.StatusInternalServerError {
				return fmt.Errorf("internal server error with command %s", req.URL.Path)
			}
			return nil
		}, nil); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			log.Error(errors.Wrap(err, "circuit breaker error"))
		}
	})
}
