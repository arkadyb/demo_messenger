package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

// GetRequestIPAddress is helper function to read users IP from the request
func GetRequestIPAddress(request *http.Request) string {
	address := request.Header.Get("X-Forwarded-For")
	if len(address) == 0 {
		address = request.Header.Get("X-Real-IP")
	}
	if len(address) == 0 {
		var err error
		address, _, err = net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			log.WithField("error", fmt.Sprintf("%+v", err)).
				WithField("addr", request.RemoteAddr).
				Error("failed to split remote address")
			return ""
		}
	}
	return address
}
