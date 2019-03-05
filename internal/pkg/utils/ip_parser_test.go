package utils_test

import (
	"github.com/arkadyb/demo_messenger/internal/pkg/utils"
	"net/http"
	"testing"
)

func TestGetRequestIPAddress(t *testing.T) {
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"X-Forwarded-For",
			args{
				&http.Request{
					Header: map[string][]string{
						"X-Forwarded-For": {"123.123.123.123"},
					},
				},
			},
			"123.123.123.123",
		},
		{
			"RemoteAddr",
			args{
				&http.Request{
					RemoteAddr: "localhost:8080",
				},
			},
			"localhost",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.GetRequestIPAddress(tt.args.request); got != tt.want {
				t.Errorf("GetRequestIPAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
