package server_test

import (
	"context"
	"github.com/arkadyb/demo_messenger/internal/server"
	"github.com/arkadyb/demo_messenger/internal/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockedApplication struct {
	mock.Mock
}

func (ma *MockedApplication) EnqueueSMS(ctx context.Context, sms *types.SMS) (err error) {
	args := ma.Called(ctx, sms)

	if args.Get(0) != nil {
		err = args.Error(0)
	}

	return
}

func TestSendSMSHandler(t *testing.T) {

	type args struct {
		app func() *MockedApplication
	}
	tests := []struct {
		name               string
		args               args
		reqBody            string
		expectedStatusCode int
	}{
		{
			"Success",
			args{
				func() *MockedApplication {
					app := &MockedApplication{}
					app.On("EnqueueSMS", mock.Anything, mock.Anything).Return(nil)
					return app
				},
			},
			`{"recipient": "12345", "originator":"originator", "message":"message"}`,
			http.StatusAccepted,
		},
		{
			"Bad Input",
			args{
				func() *MockedApplication {
					return nil
				},
			},
			`{"recipient": "12345"`,
			http.StatusBadRequest,
		},
		{
			"Empty Input",
			args{
				func() *MockedApplication {
					return nil
				},
			},
			``,
			http.StatusBadRequest,
		},
		{
			"Failed to enqueue sms",
			args{
				func() *MockedApplication {
					app := &MockedApplication{}
					app.On("EnqueueSMS", mock.Anything, mock.Anything).Return(errors.New("error"))
					return app
				},
			},
			`{"recipient": "12345", "originator":"originator", "message":"message"}`,
			http.StatusInternalServerError,
		},
		{
			"Too long message",
			args{
				func() *MockedApplication {
					return nil
				},
			},
			`{"recipient": "12345", "originator":"originator", "message":"too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message too long message "}`,
			http.StatusBadRequest,
		},
		{
			"No originator",
			args{
				func() *MockedApplication {
					return nil
				},
			},
			`{"recipient": "12345", "originator":"", "message":"message"}`,
			http.StatusBadRequest,
		},
		{
			"No recipient",
			args{
				func() *MockedApplication {
					return nil
				},
			},
			`{"recipient": "", "originator":"originator", "message":"message"}`,
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://fake-url", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			server.SendSMSHandler(tt.args.app()).ServeHTTP(w, req)
			if !assert.Equal(t, tt.expectedStatusCode, w.Code) {
				t.Fail()
			}
		})
	}
}
