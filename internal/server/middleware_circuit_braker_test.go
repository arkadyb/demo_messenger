package server_test

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/arkadyb/demo_messenger/internal/server"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_circuitBreakerMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		cfg                hystrix.CommandConfig
		handler            http.HandlerFunc
		repeats            int
		expectedStatusCode int
	}{
		{
			"Success: 200",
			hystrix.CommandConfig{},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			1,
			http.StatusOK,
		},
		{
			"Fail: 500",
			hystrix.CommandConfig{},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			1,
			http.StatusInternalServerError,
		},
		{
			"Circuit Breaker Opened",
			hystrix.CommandConfig{
				ErrorPercentThreshold: 0,
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			2,
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://fake-url", nil)
			w := httptest.NewRecorder()

			cbHandler := server.CircuitBreakerMiddleware("command", tt.cfg, tt.handler)

			for i := 0; i < tt.repeats; i++ {
				cbHandler.ServeHTTP(w, req)
			}

			time.Sleep(1 * time.Second)

			if !assert.Equal(t, tt.expectedStatusCode, w.Code) {
				t.FailNow()
			}
		})
	}
}
