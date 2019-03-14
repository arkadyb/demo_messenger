package server

import (
	"fmt"
	"github.com/arkadyb/demo_messenger/internal/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
)

type RateLimiter interface {
	Exceeded(string) (bool, error)
}

// middleware handler for rate limiter
func RateLimitingMiddleware(rl RateLimiter) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			userIP := utils.GetRequestIPAddress(req)
			if len(userIP) == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			exceeded, err := rl.Exceeded(userIP)
			if err != nil {
				logrus.Error(errors.Wrapf(err, "failed to get rate limits for user IP %s", userIP))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if exceeded {
				logrus.Error(fmt.Errorf("requests limit exceeded for ip %s", userIP))
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				next.ServeHTTP(w, req)
			}
		})
	}
}
