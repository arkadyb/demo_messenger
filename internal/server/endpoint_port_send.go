package server

import (
	"encoding/json"
	"github.com/arkadyb/demo_messenger/internal/messenger"
	"github.com/arkadyb/demo_messenger/internal/types"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"net/http"
)

// SendSMSHandler, implements http.Handler for /v1/send/sms route
func SendSMSHandler(app messenger.Application) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		var (
			sms = &types.SMS{}
		)

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(sms); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Error(errors.Wrap(err, "failed to decode sms from request body"))

			return
		}

		if len(sms.Message) > 160 {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("Message too long"))
			return
		}

		if len(sms.Originator) == 0 {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("Originator required"))
			return
		}

		if len(sms.Recipient) == 0 {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("Recipient required"))
			return
		}

		if err := app.EnqueueSMS(req.Context(), sms); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Error(errors.Wrap(err, "failed to enqueue notification"))

			return
		}

		writer.WriteHeader(http.StatusAccepted)
		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write([]byte(`{"status":"accepted"}`)); err != nil {
			log.Error(errors.Wrap(err, "failed to write response json to output"))
		}
	})
}
