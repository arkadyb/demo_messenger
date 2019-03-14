package server_test

import (
	"github.com/arkadyb/demo_messenger/internal/server"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockedRateLimiter struct {
	mock.Mock
}

func (m *MockedRateLimiter) Exceeded(commandName string) (res bool, err error) {
	args := m.Called(commandName)
	if args.Get(0) != nil {
		res = args.Get(0).(bool)
	}

	if args.Get(1) != nil {
		err = args.Error(1)
	}

	return
}

func Test_rateLimitingMiddleware(t *testing.T) {
	type args struct {
		rateLimiter func() *MockedRateLimiter
	}
	tests := []struct {
		name           string
		args           args
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{
			"Has capacity",
			args{
				func() *MockedRateLimiter {
					rl := &MockedRateLimiter{}
					rl.On("Exceeded", mock.Anything).Return(false, nil)
					return rl
				},
			},
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			http.StatusOK,
		},
		{
			"No capacity",
			args{
				func() *MockedRateLimiter {
					rl := &MockedRateLimiter{}
					rl.On("Exceeded", mock.Anything).Return(true, nil)
					return rl
				},
			},
			nil,
			http.StatusTooManyRequests,
		},
		{
			"Error",
			args{
				func() *MockedRateLimiter {
					rl := &MockedRateLimiter{}
					rl.On("Exceeded", mock.Anything).Return(false, errors.New("error"))
					return rl
				},
			},
			nil,
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://fake-url", nil)
			w := httptest.NewRecorder()

			got := server.RateLimitingMiddleware(tt.args.rateLimiter())
			got(http.HandlerFunc(tt.handler)).ServeHTTP(w, req)

			time.Sleep(1 * time.Second)
			if !assert.Equal(t, w.Code, tt.expectedStatus) {
				t.FailNow()
			}
		})
	}
}
